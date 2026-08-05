package backtest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// maxConcurrentRuns 并发回测任务上限（初始版本 4）。
const maxConcurrentRuns = 4

// Service 回测服务：策略/账户/任务 CRUD + 异步回测执行调度。
type Service struct {
	pool  *pgxpool.Pool
	mu    sync.Mutex
	tasks map[string]context.CancelFunc // run_id → 取消函数
	sem   chan struct{}                 // 并发执行信号量
}

// NewService 创建回测服务。
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		pool:  pool,
		tasks: map[string]context.CancelFunc{},
		sem:   make(chan struct{}, maxConcurrentRuns),
	}
}

// ---------------------------------------------------------------------
// 策略 CRUD
// ---------------------------------------------------------------------

// SaveStrategy 新增或更新策略（strategy_id 为空时自动生成）。
func (s *Service) SaveStrategy(ctx context.Context, st *Strategy) (string, error) {
	if err := validateStrategy(st); err != nil {
		return "", err
	}
	if st.StrategyID == "" {
		st.StrategyID = uuid.NewString()
	}
	if st.StrategyType == "" {
		st.StrategyType = StrategyTypeRuleBased
	}
	if st.DataPeriod == "" {
		st.DataPeriod = st.Definition.Data.Period
	}
	if st.DataPeriod == "" {
		st.DataPeriod = PeriodMin
	}
	if st.SecuCode == "" && len(st.Definition.Universe.Securities) > 0 {
		st.SecuCode = st.Definition.Universe.Securities[0]
	}
	if st.AllowBacktest == "" {
		st.AllowBacktest = FlagOn
	}
	if st.Status == "" {
		st.Status = StatusEnabled
	}
	if st.DefinitionVersion == 0 {
		st.DefinitionVersion = 1
	}
	if st.Definition.Version == "" {
		st.Definition.Version = "1"
	}
	defJSON, err := json.Marshal(st.Definition)
	if err != nil {
		return "", fmt.Errorf("策略定义序列化失败: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
INSERT INTO finv_backtest_strategy
    (strategy_id, strategy_code, strategy_name, strategy_type, description,
     definition, definition_version, data_period, secu_code, allow_backtest, status, created_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
ON CONFLICT (strategy_id) DO UPDATE SET
    strategy_code = EXCLUDED.strategy_code,
    strategy_name = EXCLUDED.strategy_name,
    strategy_type = EXCLUDED.strategy_type,
    description = EXCLUDED.description,
    definition = EXCLUDED.definition,
    definition_version = EXCLUDED.definition_version,
    data_period = EXCLUDED.data_period,
    secu_code = EXCLUDED.secu_code,
    allow_backtest = EXCLUDED.allow_backtest,
    status = EXCLUDED.status,
    created_by = EXCLUDED.created_by`,
		st.StrategyID, st.StrategyCode, st.StrategyName, st.StrategyType, st.Description,
		defJSON, st.DefinitionVersion, st.DataPeriod, st.SecuCode, st.AllowBacktest, st.Status, st.CreatedBy)
	if err != nil {
		return "", err
	}
	return st.StrategyID, nil
}

// ListStrategies 分页查询策略。
func (s *Service) ListStrategies(ctx context.Context, pager Pager, keyword, allowBacktest string) ([]Strategy, int, error) {
	pager.Normalize()
	where := []string{"1=1"}
	args := []any{}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		args = append(args, "%"+keyword+"%")
		where = append(where, fmt.Sprintf(`(
			strategy_code ILIKE $%d OR strategy_name ILIKE $%d OR description ILIKE $%d OR secu_code ILIKE $%d
		)`, len(args), len(args), len(args), len(args)))
	}
	if allowBacktest == FlagOn || allowBacktest == FlagOff {
		args = append(args, allowBacktest)
		where = append(where, fmt.Sprintf("allow_backtest = $%d", len(args)))
	}
	cond := strings.Join(where, " AND ")

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM finv_backtest_strategy WHERE `+cond, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx, `
SELECT strategy_id, strategy_code, strategy_name, strategy_type, description,
       definition, definition_version, data_period, secu_code, allow_backtest, status, created_by, gmt_update
FROM finv_backtest_strategy
WHERE `+cond+`
ORDER BY gmt_update DESC
LIMIT $`+fmt.Sprint(len(args)+1)+` OFFSET $`+fmt.Sprint(len(args)+2),
		append(args, pager.PageSize, (pager.Page-1)*pager.PageSize)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []Strategy{}
	for rows.Next() {
		st, err := scanStrategy(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, st)
	}
	return list, total, rows.Err()
}

// GetStrategy 查询单个策略。
func (s *Service) GetStrategy(ctx context.Context, strategyID string) (*Strategy, error) {
	row := s.pool.QueryRow(ctx, `
SELECT strategy_id, strategy_code, strategy_name, strategy_type, description,
       definition, definition_version, data_period, secu_code, allow_backtest, status, created_by, gmt_update
FROM finv_backtest_strategy WHERE strategy_id = $1`, strategyID)
	st, err := scanStrategy(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("策略不存在: %s", strategyID)
		}
		return nil, err
	}
	return &st, nil
}

// ToggleStrategy 切换策略回测开关。
func (s *Service) ToggleStrategy(ctx context.Context, strategyID, allowBacktest string) error {
	if allowBacktest != FlagOn && allowBacktest != FlagOff {
		return fmt.Errorf("allow_backtest 仅支持 0/1")
	}
	tag, err := s.pool.Exec(ctx, `UPDATE finv_backtest_strategy SET allow_backtest=$2 WHERE strategy_id=$1`, strategyID, allowBacktest)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("策略不存在: %s", strategyID)
	}
	return nil
}

// DeleteStrategy 删除策略（存在关联回测任务时拒绝）。
func (s *Service) DeleteStrategy(ctx context.Context, strategyID string) error {
	var runCount int
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM finv_backtest_run WHERE strategy_id = $1`, strategyID).Scan(&runCount); err != nil {
		return err
	}
	if runCount > 0 {
		return fmt.Errorf("策略已关联 %d 个回测任务，禁止删除（可改为禁用）", runCount)
	}
	tag, err := s.pool.Exec(ctx, `DELETE FROM finv_backtest_strategy WHERE strategy_id = $1`, strategyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("策略不存在: %s", strategyID)
	}
	return nil
}

// ---------------------------------------------------------------------
// 账户 CRUD
// ---------------------------------------------------------------------

// SaveAccount 新增或更新回测账户。
func (s *Service) SaveAccount(ctx context.Context, acc *Account) (string, error) {
	if err := validateAccount(acc); err != nil {
		return "", err
	}
	if acc.AccountID == "" {
		acc.AccountID = uuid.NewString()
	}
	if acc.CurrencyType == "" {
		acc.CurrencyType = "USD"
	}
	if acc.MarginMode == "" {
		acc.MarginMode = "FULL"
	}
	if acc.MarginRate <= 0 || acc.MarginRate > 1 {
		acc.MarginRate = 1
	}
	if acc.AllowBacktest == "" {
		acc.AllowBacktest = FlagOn
	}
	if acc.Status == "" {
		acc.Status = StatusEnabled
	}

	_, err := s.pool.Exec(ctx, `
INSERT INTO finv_backtest_account
    (account_id, account_code, account_name, initial_capital, currency_type,
     commission_rate, slippage_pct, margin_mode, margin_rate, allow_backtest, status, remark, created_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
ON CONFLICT (account_id) DO UPDATE SET
    account_code = EXCLUDED.account_code,
    account_name = EXCLUDED.account_name,
    initial_capital = EXCLUDED.initial_capital,
    currency_type = EXCLUDED.currency_type,
    commission_rate = EXCLUDED.commission_rate,
    slippage_pct = EXCLUDED.slippage_pct,
    margin_mode = EXCLUDED.margin_mode,
    margin_rate = EXCLUDED.margin_rate,
    allow_backtest = EXCLUDED.allow_backtest,
    status = EXCLUDED.status,
    remark = EXCLUDED.remark,
    created_by = EXCLUDED.created_by`,
		acc.AccountID, acc.AccountCode, acc.AccountName, acc.InitialCapital, acc.CurrencyType,
		acc.CommissionRate, acc.SlippagePct, acc.MarginMode, acc.MarginRate,
		acc.AllowBacktest, acc.Status, acc.Remark, acc.CreatedBy)
	if err != nil {
		return "", err
	}
	return acc.AccountID, nil
}

// ListAccounts 分页查询账户。
func (s *Service) ListAccounts(ctx context.Context, pager Pager, keyword, allowBacktest string) ([]Account, int, error) {
	pager.Normalize()
	where := []string{"1=1"}
	args := []any{}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		args = append(args, "%"+keyword+"%")
		where = append(where, fmt.Sprintf(`(
			account_code ILIKE $%d OR account_name ILIKE $%d OR remark ILIKE $%d
		)`, len(args), len(args), len(args)))
	}
	if allowBacktest == FlagOn || allowBacktest == FlagOff {
		args = append(args, allowBacktest)
		where = append(where, fmt.Sprintf("allow_backtest = $%d", len(args)))
	}
	cond := strings.Join(where, " AND ")

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM finv_backtest_account WHERE `+cond, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx, `
SELECT account_id, account_code, account_name, initial_capital, currency_type,
       commission_rate, slippage_pct, margin_mode, margin_rate, allow_backtest, status, remark, created_by, gmt_update
FROM finv_backtest_account
WHERE `+cond+`
ORDER BY gmt_update DESC
LIMIT $`+fmt.Sprint(len(args)+1)+` OFFSET $`+fmt.Sprint(len(args)+2),
		append(args, pager.PageSize, (pager.Page-1)*pager.PageSize)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []Account{}
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.AccountID, &a.AccountCode, &a.AccountName, &a.InitialCapital,
			&a.CurrencyType, &a.CommissionRate, &a.SlippagePct, &a.MarginMode, &a.MarginRate,
			&a.AllowBacktest, &a.Status, &a.Remark, &a.CreatedBy, &a.GMTUpdate); err != nil {
			return nil, 0, err
		}
		list = append(list, a)
	}
	return list, total, rows.Err()
}

// GetAccount 查询单个账户。
func (s *Service) GetAccount(ctx context.Context, accountID string) (*Account, error) {
	row := s.pool.QueryRow(ctx, `
SELECT account_id, account_code, account_name, initial_capital, currency_type,
       commission_rate, slippage_pct, margin_mode, margin_rate, allow_backtest, status, remark, created_by, gmt_update
FROM finv_backtest_account WHERE account_id = $1`, accountID)
	var a Account
	if err := row.Scan(&a.AccountID, &a.AccountCode, &a.AccountName, &a.InitialCapital,
		&a.CurrencyType, &a.CommissionRate, &a.SlippagePct, &a.MarginMode, &a.MarginRate,
		&a.AllowBacktest, &a.Status, &a.Remark, &a.CreatedBy, &a.GMTUpdate); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("账户不存在: %s", accountID)
		}
		return nil, err
	}
	return &a, nil
}

// ToggleAccount 切换账户回测开关。
func (s *Service) ToggleAccount(ctx context.Context, accountID, allowBacktest string) error {
	if allowBacktest != FlagOn && allowBacktest != FlagOff {
		return fmt.Errorf("allow_backtest 仅支持 0/1")
	}
	tag, err := s.pool.Exec(ctx, `UPDATE finv_backtest_account SET allow_backtest=$2 WHERE account_id=$1`, accountID, allowBacktest)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("账户不存在: %s", accountID)
	}
	return nil
}

// DeleteAccount 删除账户（存在关联回测任务时拒绝）。
func (s *Service) DeleteAccount(ctx context.Context, accountID string) error {
	var runCount int
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM finv_backtest_run WHERE account_id = $1`, accountID).Scan(&runCount); err != nil {
		return err
	}
	if runCount > 0 {
		return fmt.Errorf("账户已关联 %d 个回测任务，禁止删除（可改为禁用）", runCount)
	}
	tag, err := s.pool.Exec(ctx, `DELETE FROM finv_backtest_account WHERE account_id = $1`, accountID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("账户不存在: %s", accountID)
	}
	return nil
}

// ---------------------------------------------------------------------
// 回测任务
// ---------------------------------------------------------------------

// CreateRun 创建并启动回测任务。
func (s *Service) CreateRun(ctx context.Context, req CreateRunRequest) (*Run, error) {
	if req.StrategyID == "" || req.AccountID == "" {
		return nil, fmt.Errorf("strategy_id 与 account_id 必填")
	}
	st, err := s.GetStrategy(ctx, req.StrategyID)
	if err != nil {
		return nil, err
	}
	if st.AllowBacktest != FlagOn {
		return nil, fmt.Errorf("策略「%s」回测开关已关闭，请先在策略管理启用", st.StrategyName)
	}
	acc, err := s.GetAccount(ctx, req.AccountID)
	if err != nil {
		return nil, err
	}
	if acc.AllowBacktest != FlagOn {
		return nil, fmt.Errorf("账户「%s」回测开关已关闭，请先在账户管理启用", acc.AccountName)
	}
	if !req.Options.EnableBacktest {
		return nil, fmt.Errorf("回测开关未启用，已拒绝创建回测任务")
	}

	// 标的：请求参数 > 策略 universe 首个证券
	secuCode := strings.TrimSpace(req.SecuCode)
	if secuCode == "" && len(st.Definition.Universe.Securities) > 0 {
		secuCode = st.Definition.Universe.Securities[0]
	}
	if secuCode == "" {
		return nil, fmt.Errorf("未指定回测标的（secu_code 或策略 universe.securities）")
	}

	// 周期与报告精度
	period := req.Period
	if period == "" {
		period = st.Definition.Data.Period
	}
	if period == "" {
		period = PeriodMin
	}
	if period != PeriodMin && period != PeriodHour && period != PeriodDay {
		return nil, fmt.Errorf("不支持的周期 %q", period)
	}
	precision := req.ReportPrecision
	if precision == "" {
		precision = PeriodDay
	}
	if precision != PeriodMin && precision != PeriodHour && precision != PeriodDay {
		return nil, fmt.Errorf("不支持的报告精度 %q", precision)
	}

	// 时间区间（date 维度）
	startDate, endDate, err := s.resolveDateRange(ctx, secuCode, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	startTS := dateToTS(startDate)
	endTS := dateToTS(endDate) + 86400 - 1

	// 快照
	defJSON, _ := json.Marshal(st.Definition)
	accSnapshot := AccountSnapshot{
		AccountID:      acc.AccountID,
		AccountCode:    acc.AccountCode,
		AccountName:    acc.AccountName,
		InitialCapital: acc.InitialCapital,
		CurrencyType:   acc.CurrencyType,
		CommissionRate: acc.CommissionRate,
		SlippagePct:    acc.SlippagePct,
		MarginMode:     acc.MarginMode,
		MarginRate:     acc.MarginRate,
	}
	accJSON, _ := json.Marshal(accSnapshot)
	optsJSON, _ := json.Marshal(req.Options)

	marketCode := 0
	_ = s.pool.QueryRow(ctx,
		`SELECT COALESCE(market_code, 0) FROM finv_security WHERE usc = $1 OR security_code = $1 LIMIT 1`, secuCode).Scan(&marketCode)

	runID := uuid.NewString()
	_, err = s.pool.Exec(ctx, `
INSERT INTO finv_backtest_run
    (run_id, strategy_id, strategy_code, strategy_name, strategy_snapshot,
     account_id, account_code, account_name, account_snapshot,
     secu_code, market_code, period, report_precision,
     start_ts, end_ts, start_date, end_date, options, status, progress, created_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,'PENDING',0,$19)`,
		runID, st.StrategyID, st.StrategyCode, st.StrategyName, defJSON,
		acc.AccountID, acc.AccountCode, acc.AccountName, accJSON,
		secuCode, marketCode, period, precision,
		startTS, endTS, startDate, endDate, optsJSON, req.StrCreatedBy())
	if err != nil {
		return nil, err
	}

	// 异步执行
	runCtx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.tasks[runID] = cancel
	s.mu.Unlock()
	go s.executeRun(runCtx, runID)

	run, err := s.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	return run, nil
}

// CancelRun 取消运行中的回测任务。
func (s *Service) CancelRun(ctx context.Context, runID string) error {
	s.mu.Lock()
	cancel, ok := s.tasks[runID]
	s.mu.Unlock()
	if !ok {
		// 非运行中任务：检查状态，仅 PENDING/RUNNING 可取消
		var status string
		if err := s.pool.QueryRow(ctx, `SELECT status FROM finv_backtest_run WHERE run_id=$1`, runID).Scan(&status); err != nil {
			return fmt.Errorf("任务不存在: %s", runID)
		}
		if status == RunPending || status == RunRunning {
			return fmt.Errorf("任务不在本进程调度器中（可能已重启），无法取消")
		}
		return fmt.Errorf("任务已结束（%s），无需取消", status)
	}
	cancel()
	return nil
}

// ListRuns 分页查询回测任务。
func (s *Service) ListRuns(ctx context.Context, q RunListQuery) ([]Run, int, error) {
	q.Pager.Normalize()
	where := []string{"1=1"}
	args := []any{}
	if q.Status != "" {
		args = append(args, q.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}
	if q.SecuCode != "" {
		args = append(args, q.SecuCode)
		where = append(where, fmt.Sprintf("secu_code = $%d", len(args)))
	}
	if q.StrategyID != "" {
		args = append(args, q.StrategyID)
		where = append(where, fmt.Sprintf("strategy_id = $%d", len(args)))
	}
	if q.Keyword = strings.TrimSpace(q.Keyword); q.Keyword != "" {
		args = append(args, "%"+q.Keyword+"%")
		where = append(where, fmt.Sprintf(`(
			strategy_name ILIKE $%d OR account_name ILIKE $%d OR secu_code ILIKE $%d OR CAST(run_no AS TEXT) ILIKE $%d
		)`, len(args), len(args), len(args), len(args)))
	}
	cond := strings.Join(where, " AND ")

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM finv_backtest_run WHERE `+cond, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx, `
SELECT run_id, run_no, strategy_id, strategy_code, strategy_name, strategy_snapshot,
       account_id, account_code, account_name, account_snapshot,
       secu_code, market_code, period, report_precision,
       start_ts, end_ts, start_date, end_date, options, status, progress, error_message, report,
       started_at, finished_at, created_by, gmt_update
FROM finv_backtest_run
WHERE `+cond+`
ORDER BY run_no DESC
LIMIT $`+fmt.Sprint(len(args)+1)+` OFFSET $`+fmt.Sprint(len(args)+2),
		append(args, q.PageSize, (q.Page-1)*q.PageSize)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []Run{}
	for rows.Next() {
		r, err := scanRun(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, r)
	}
	return list, total, rows.Err()
}

// GetRun 查询单个回测任务（含报告）。
func (s *Service) GetRun(ctx context.Context, runID string) (*Run, error) {
	row := s.pool.QueryRow(ctx, `
SELECT run_id, run_no, strategy_id, strategy_code, strategy_name, strategy_snapshot,
       account_id, account_code, account_name, account_snapshot,
       secu_code, market_code, period, report_precision,
       start_ts, end_ts, start_date, end_date, options, status, progress, error_message, report,
       started_at, finished_at, created_by, gmt_update
FROM finv_backtest_run WHERE run_id = $1`, runID)
	r, err := scanRun(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("回测任务不存在: %s", runID)
		}
		return nil, err
	}
	return &r, nil
}

// GetReport 查询回测报告（汇总指标）。
func (s *Service) GetReport(ctx context.Context, runID string) (*RunReport, error) {
	var reportJSON []byte
	var status string
	err := s.pool.QueryRow(ctx,
		`SELECT status, COALESCE(report, '{}') FROM finv_backtest_run WHERE run_id=$1`, runID).Scan(&status, &reportJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("回测任务不存在: %s", runID)
		}
		return nil, err
	}
	if status != RunSucceeded {
		return nil, fmt.Errorf("任务尚未成功完成（当前状态 %s），暂无报告", status)
	}
	var report RunReport
	if err := json.Unmarshal(reportJSON, &report); err != nil {
		return nil, fmt.Errorf("报告数据解析失败: %w", err)
	}
	return &report, nil
}

// ListEquity 分页查询净值曲线（按报告精度）。
func (s *Service) ListEquity(ctx context.Context, runID string, pager Pager) ([]EquityPoint, int, error) {
	pager = normalizeLargePager(pager)
	var total int
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM finv_backtest_equity WHERE run_id=$1`, runID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.pool.Query(ctx, `
SELECT seq, ts, date, "time", equity, cash, position_value, position_qty, profit, roi, drawdown
FROM finv_backtest_equity WHERE run_id=$1
ORDER BY seq ASC
LIMIT $2 OFFSET $3`, runID, pager.PageSize, (pager.Page-1)*pager.PageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []EquityPoint{}
	for rows.Next() {
		var p EquityPoint
		if err := rows.Scan(&p.Seq, &p.TS, &p.Date, &p.Time, &p.Equity, &p.Cash,
			&p.PositionValue, &p.PositionQty, &p.Profit, &p.ROI, &p.Drawdown); err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	return list, total, rows.Err()
}

// ListTrades 分页查询成交记录。
func (s *Service) ListTrades(ctx context.Context, runID string, pager Pager) ([]Trade, int, error) {
	pager = normalizeLargePager(pager)
	var total int
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM finv_backtest_trade WHERE run_id=$1`, runID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.pool.Query(ctx, `
SELECT trade_id, run_id, ts, date, "time", action, price, qty, amount, fee, profit, position_after, cash_after, signal
FROM finv_backtest_trade WHERE run_id=$1
ORDER BY ts ASC
LIMIT $2 OFFSET $3`, runID, pager.PageSize, (pager.Page-1)*pager.PageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []Trade{}
	for rows.Next() {
		var t Trade
		if err := rows.Scan(&t.TradeID, &t.RunID, &t.TS, &t.Date, &t.Time, &t.Action, &t.Price,
			&t.Qty, &t.Amount, &t.Fee, &t.Profit, &t.PositionAfter, &t.CashAfter, &t.Signal); err != nil {
			return nil, 0, err
		}
		list = append(list, t)
	}
	return list, total, rows.Err()
}

// ---------------------------------------------------------------------
// 任务执行
// ---------------------------------------------------------------------

// executeRun 执行回测任务（goroutine 内运行）。
func (s *Service) executeRun(ctx context.Context, runID string) {
	defer func() {
		s.mu.Lock()
		delete(s.tasks, runID)
		s.mu.Unlock()
	}()

	// 并发信号量（排队等待）
	select {
	case s.sem <- struct{}{}:
		defer func() { <-s.sem }()
	case <-ctx.Done():
		s.markRun(ctx, runID, RunCancelled, 0, "任务已取消", nil)
		return
	}

	run, err := s.GetRun(ctx, runID)
	if err != nil {
		s.markRun(context.Background(), runID, RunFailed, 0, "加载任务失败: "+err.Error(), nil)
		return
	}

	now := time.Now().UTC()
	_, _ = s.pool.Exec(ctx, `UPDATE finv_backtest_run SET status='RUNNING', started_at=$2 WHERE run_id=$1`, runID, now)

	// 加载行情（按 date 范围）
	bars, err := s.loadBars(ctx, run.SecuCode, run.StartDate, run.EndDate)
	if err != nil {
		s.markRun(context.Background(), runID, RunFailed, 0, "加载行情失败: "+err.Error(), nil)
		return
	}
	bars, err = AggregateBars(bars, run.Period)
	if err != nil {
		s.markRun(context.Background(), runID, RunFailed, 0, "行情聚合失败: "+err.Error(), nil)
		return
	}

	// 组装引擎配置
	cfg := EngineConfig{
		Definition:      run.Definition(),
		Account:         run.AccountSnapshot,
		SecuCode:        run.SecuCode,
		Period:          run.Period,
		StartTS:         run.StartTS,
		EndTS:           run.EndTS,
		ReportPrecision: run.ReportPrecision,
		Options:         run.RunOptions(),
	}
	lastProgress := -1
	engine := NewEngine(cfg, func(processed, total int) {
		pct := processed * 100 / total
		if pct != lastProgress {
			lastProgress = pct
			_, _ = s.pool.Exec(ctx, `UPDATE finv_backtest_run SET progress=$2 WHERE run_id=$1`, runID, pct)
		}
	})

	result, err := engine.Run(ctx, bars)
	if err != nil {
		if ctx.Err() != nil {
			s.markRun(context.Background(), runID, RunCancelled, 0, "任务已取消", nil)
			return
		}
		s.markRun(context.Background(), runID, RunFailed, 0, "回测执行失败: "+err.Error(), nil)
		return
	}

	// 持久化曲线/成交/报告
	if err := s.persistResult(ctx, runID, result); err != nil {
		s.markRun(context.Background(), runID, RunFailed, 0, "结果持久化失败: "+err.Error(), nil)
		return
	}
	s.markRun(context.Background(), runID, RunSucceeded, 100, "", result.Report)
}

// loadBars 加载分钟行情（date 范围，时间升序）。
func (s *Service) loadBars(ctx context.Context, secuCode string, startDate, endDate int) ([]Bar, error) {
	rows, err := s.pool.Query(ctx, `
SELECT ts, date, "time",
       COALESCE(open, close), COALESCE(high, close), COALESCE(low, close), COALESCE(close, open),
       COALESCE(volume, 0), COALESCE(turnover, 0)
FROM finv_quote_secu_kline_min
WHERE secu_code = $1 AND date BETWEEN $2 AND $3
ORDER BY ts ASC`, secuCode, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bars := []Bar{}
	for rows.Next() {
		var b Bar
		if err := rows.Scan(&b.TS, &b.Date, &b.Time, &b.Open, &b.High, &b.Low, &b.Close, &b.Volume, &b.Turnover); err != nil {
			return nil, err
		}
		bars = append(bars, b)
	}
	return bars, rows.Err()
}

// persistResult 批量落库曲线/成交/报告。
func (s *Service) persistResult(ctx context.Context, runID string, result *EngineResult) error {
	batch := &pgx.Batch{}
	for _, p := range result.EquityPoints {
		batch.Queue(`
INSERT INTO finv_backtest_equity (run_id, seq, ts, date, "time", equity, cash, position_value, position_qty, profit, roi, drawdown)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
ON CONFLICT (run_id, ts) DO UPDATE SET
  equity=EXCLUDED.equity, cash=EXCLUDED.cash, position_value=EXCLUDED.position_value,
  position_qty=EXCLUDED.position_qty, profit=EXCLUDED.profit, roi=EXCLUDED.roi, drawdown=EXCLUDED.drawdown`,
			runID, p.Seq, p.TS, p.Date, p.Time, p.Equity, p.Cash, p.PositionValue,
			p.PositionQty, p.Profit, p.ROI, p.Drawdown)
	}
	for _, t := range result.Trades {
		batch.Queue(`
INSERT INTO finv_backtest_trade (run_id, ts, date, "time", action, price, qty, amount, fee, profit, position_after, cash_after, signal)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			runID, t.TS, t.Date, t.Time, t.Action, t.Price, t.Qty, t.Amount, t.Fee,
			t.Profit, t.PositionAfter, t.CashAfter, t.Signal)
	}
	reportJSON, _ := json.Marshal(result.Report)
	batch.Queue(`UPDATE finv_backtest_run SET report=$2 WHERE run_id=$1`, runID, reportJSON)

	results := s.pool.SendBatch(ctx, batch)
	defer results.Close()
	for i := 0; i < batch.Len(); i++ {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}
	return nil
}

// normalizeLargePager 大页分页归一化（曲线/成交查询用，单页上限 5000）。
func normalizeLargePager(p Pager) Pager {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 5000
	}
	if p.PageSize > 5000 {
		p.PageSize = 5000
	}
	return p
}

// markRun 更新任务状态/进度/错误/报告。
func (s *Service) markRun(ctx context.Context, runID, status string, progress int, errMsg string, report *RunReport) {
	var reportJSON any
	if report != nil {
		reportJSON = report
	} else {
		reportJSON = nil
	}
	_, _ = s.pool.Exec(ctx, `
UPDATE finv_backtest_run
SET status=$2, progress=$3, error_message=$4, report=COALESCE($5, report), finished_at=now()
WHERE run_id=$1`, runID, status, progress, nullIfEmpty(errMsg), reportJSON)
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// resolveDateRange 解析回测日期区间（缺省取行情最早/最晚日期）。
func (s *Service) resolveDateRange(ctx context.Context, secuCode string, startDate, endDate int) (int, int, error) {
	if startDate == 0 || endDate == 0 {
		var minDate, maxDate int
		err := s.pool.QueryRow(ctx, `
SELECT COALESCE(MIN(date), 0), COALESCE(MAX(date), 0)
FROM finv_quote_secu_kline_min WHERE secu_code = $1`, secuCode).Scan(&minDate, &maxDate)
		if err != nil {
			return 0, 0, err
		}
		if minDate == 0 {
			return 0, 0, fmt.Errorf("标的 %s 无任何行情数据", secuCode)
		}
		if startDate == 0 {
			startDate = minDate
		}
		if endDate == 0 {
			endDate = maxDate
		}
	}
	if startDate > endDate {
		return 0, 0, fmt.Errorf("开始日期（%d）不能晚于结束日期（%d）", startDate, endDate)
	}
	return startDate, endDate, nil
}

// dateToTS 将 yyyymmdd 转为 UTC 秒（当日 00:00）。
func dateToTS(date int) int64 {
	year := date / 10000
	month := (date / 100) % 100
	day := date % 100
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return t.Unix()
}

// ---------------------------------------------------------------------
// 行扫描与解析辅助
// ---------------------------------------------------------------------

type rowScanner interface {
	Scan(dest ...any) error
}

func scanStrategy(row rowScanner) (Strategy, error) {
	var st Strategy
	var defJSON []byte
	err := row.Scan(&st.StrategyID, &st.StrategyCode, &st.StrategyName, &st.StrategyType,
		&st.Description, &defJSON, &st.DefinitionVersion, &st.DataPeriod, &st.SecuCode,
		&st.AllowBacktest, &st.Status, &st.CreatedBy, &st.GMTUpdate)
	if err != nil {
		return st, err
	}
	if err := json.Unmarshal(defJSON, &st.Definition); err != nil {
		return st, fmt.Errorf("策略定义解析失败: %w", err)
	}
	return st, nil
}

func scanRun(row rowScanner) (Run, error) {
	var r Run
	var defJSON, accJSON, optsJSON, reportJSON []byte
	var startedAt, finishedAt *time.Time
	err := row.Scan(&r.RunID, &r.RunNo, &r.StrategyID, &r.StrategyCode, &r.StrategyName, &defJSON,
		&r.AccountID, &r.AccountCode, &r.AccountName, &accJSON,
		&r.SecuCode, &r.MarketCode, &r.Period, &r.ReportPrecision,
		&r.StartTS, &r.EndTS, &r.StartDate, &r.EndDate, &optsJSON,
		&r.Status, &r.Progress, &r.ErrorMessage, &reportJSON,
		&startedAt, &finishedAt, &r.CreatedBy, &r.GMTUpdate)
	if err != nil {
		return r, err
	}
	if len(defJSON) > 0 {
		if err := json.Unmarshal(defJSON, &r.StrategySnapshot); err != nil {
			return r, fmt.Errorf("策略快照解析失败: %w", err)
		}
	}
	if len(accJSON) > 0 {
		if err := json.Unmarshal(accJSON, &r.AccountSnapshot); err != nil {
			return r, fmt.Errorf("账户快照解析失败: %w", err)
		}
	}
	if len(optsJSON) > 0 {
		if err := json.Unmarshal(optsJSON, &r.Options); err != nil {
			return r, fmt.Errorf("任务配置解析失败: %w", err)
		}
	}
	if len(reportJSON) > 0 && string(reportJSON) != "null" && string(reportJSON) != "{}" {
		var report RunReport
		if err := json.Unmarshal(reportJSON, &report); err == nil {
			r.Report = &report
		}
	}
	r.StartedAt = startedAt
	r.FinishedAt = finishedAt
	return r, nil
}

// Definition 解析任务快照为策略定义。
func (r *Run) Definition() StrategyDefinition {
	var def StrategyDefinition
	raw, _ := json.Marshal(r.StrategySnapshot)
	_ = json.Unmarshal(raw, &def)
	return def
}

// RunOptions 解析任务配置。
func (r *Run) RunOptions() RunOptions {
	var opts RunOptions
	raw, _ := json.Marshal(r.Options)
	_ = json.Unmarshal(raw, &opts)
	return opts
}

// StrCreatedBy 返回创建人（请求体可能不含，兼容）。
func (req CreateRunRequest) StrCreatedBy() string { return "console" }
