package quote

import (
	"strings"
	"testing"
)

func TestBuildUpsertSQLFieldMode(t *testing.T) {
	sql := buildUpsertSQL(UpsertModeField)
	if !strings.Contains(sql, "ON CONFLICT (ts, market_code, secu_code) DO UPDATE SET") {
		t.Fatal("缺少 ON CONFLICT 子句")
	}
	if !strings.Contains(sql, "close = COALESCE(EXCLUDED.close, finv_quote_secu_kline_min.close)") {
		t.Fatal("FIELD 模式应使用 COALESCE 字段级覆盖")
	}
	if !strings.Contains(sql, "RETURNING (xmax = 0) AS is_insert") {
		t.Fatal("缺少 RETURNING 区分新增/覆盖")
	}
	if strings.Contains(sql, "COALESCE") == false {
		t.Fatal("FIELD 模式必须包含 COALESCE")
	}
}

func TestBuildUpsertSQLRowMode(t *testing.T) {
	sql := buildUpsertSQL(UpsertModeRow)
	if strings.Contains(sql, "COALESCE") {
		t.Fatal("ROW 模式不应包含 COALESCE")
	}
	if !strings.Contains(sql, "close = EXCLUDED.close") {
		t.Fatal("ROW 模式应整行覆盖")
	}
}

func TestBuildUpsertSQLCoversAllUpdatableColumns(t *testing.T) {
	sql := buildUpsertSQL(UpsertModeField)
	for _, column := range []string{"date", `"time"`, "prev_close", "open", "high", "low",
		"close", "paocd", "volume", "turnover", "ext_field", "remark"} {
		if !strings.Contains(sql, column) {
			t.Fatalf("upsert 未覆盖列: %s", column)
		}
	}
}
