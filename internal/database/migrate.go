// Package database 提供 PostgreSQL 18 连接与 schema 迁移。
package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// 迁移文件名：V<number>__<name>.sql
var migrationName = regexp.MustCompile(`^V(\d+)__[A-Za-z0-9_\-]+\.sql$`)

// Migrate 应用 migrationsDir 下全部未执行的迁移（启动时幂等调用）。
func Migrate(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	if err := ensureVersionTable(ctx, pool); err != nil {
		return err
	}

	applied, err := appliedVersions(ctx, pool)
	if err != nil {
		return err
	}

	paths, err := discoverMigrations(migrationsDir)
	if err != nil {
		return err
	}

	for _, path := range paths {
		version := path.version
		if applied[version] {
			continue
		}
		script, err := os.ReadFile(path.path)
		if err != nil {
			return fmt.Errorf("读取迁移 %s 失败: %w", path.path, err)
		}
		if err := applyOne(ctx, pool, version, string(script)); err != nil {
			return fmt.Errorf("迁移 V%d 失败: %w", version, err)
		}
	}
	return nil
}

type migrationFile struct {
	version int
	path    string
}

func ensureVersionTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_version (
    version      TEXT        NOT NULL,
    description  TEXT        NOT NULL,
    installed_by TEXT        NOT NULL DEFAULT current_user,
    installed_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    success      BOOLEAN     NOT NULL DEFAULT TRUE,
    PRIMARY KEY (version)
)`)
	return err
}

func appliedVersions(ctx context.Context, pool *pgxpool.Pool) (map[int]bool, error) {
	rows, err := pool.Query(ctx,
		"SELECT version FROM schema_version WHERE success = TRUE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	applied := map[int]bool{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		var parsed int
		if _, err := fmt.Sscanf(version, "%d", &parsed); err == nil {
			applied[parsed] = true
		}
	}
	return applied, rows.Err()
}

func discoverMigrations(dir string) ([]migrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("迁移目录不存在: %s", dir)
	}
	var files []migrationFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		match := migrationName.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		var version int
		if _, err := fmt.Sscanf(match[1], "%d", &version); err != nil {
			return nil, fmt.Errorf("迁移版本号非法: %s", entry.Name())
		}
		files = append(files, migrationFile{version: version, path: filepath.Join(dir, entry.Name())})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].version < files[j].version })
	return files, nil
}

func applyOne(ctx context.Context, pool *pgxpool.Pool, version int, script string) error {
	// 单事务执行：失败整体回滚
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, script); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		"INSERT INTO schema_version (version, description) VALUES ($1, $2)",
		fmt.Sprintf("%d", version), descriptionOf(version),
	); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func descriptionOf(version int) string {
	return fmt.Sprintf("V%d migration", version)
}

// WaitForDatabase 等待 PG 就绪（Compose 启动顺序兜底）。
func WaitForDatabase(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		err := pool.Ping(pingCtx)
		cancel()
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return fmt.Errorf("等待 PostgreSQL 就绪超时（%s）", timeout)
}
