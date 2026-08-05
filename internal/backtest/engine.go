package backtest

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// ---------------------------------------------------------------------
// 行情 Bar（引擎输入，周期已聚合）
// ---------------------------------------------------------------------

// Bar 回测行情 bar。
type Bar struct {
	TS       int64 // UTC 秒
	Date     int   // yyyymmdd
	Time     int   // hhmmss
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   float64
	Turnover float64
}

// EngineConfig 引擎配置（服务层从策略定义+账户快照+任务参数组装）。
type EngineConfig struct {
	Definition      StrategyDefinition
	Account         AccountSnapshot
	SecuCode        string
	Period          string
	StartTS         int64
	EndTS           int64
	ReportPrecision string
	Options         RunOptions
}

// EngineResult 引擎输出。
type EngineResult struct {
	EquityPoints []EquityPoint
	Trades       []Trade
	Report       *RunReport
	BarCount     int
}

// ProgressFunc 进度回调（processed/total）。
type ProgressFunc func(processed, total int)

// Engine 回测引擎。
type Engine struct {
	cfg    EngineConfig
	onProg ProgressFunc
}

// NewEngine 创建回测引擎。
func NewEngine(cfg EngineConfig, onProg ProgressFunc) *Engine {
	return &Engine{cfg: cfg, onProg: onProg}
}

// 运行时账户状态
type accountState struct {
	cash      float64
	position  float64
	avgCost   float64 // 持仓加权成本
	entryDate int     // 开仓日期
}

// 每日/全局计数
type dayCounters struct {
	trades        int            // 当日总成交笔数
	ruleTrades    map[string]int // 规则当日触发次数
	ruleRunTrades map[string]int // 规则整个回测触发次数
	lastTradeBar  int            // 最近成交 bar 索引
	curDate       int            // 当前日期（用于跨日重置）
}

// Run 执行回测。ctx 支持取消（任务取消/服务停机）；bars 为已按周期聚合、时间升序的行情。
func (e *Engine) Run(ctx context.Context, bars []Bar) (*EngineResult, error) {
	if len(bars) == 0 {
		return nil, fmt.Errorf("回测区间内无行情数据")
	}
	def := e.cfg.Definition
	acc := e.cfg.Account

	initialCapital := acc.InitialCapital
	if e.cfg.Options.InitialCapital != nil && *e.cfg.Options.InitialCapital > 0 {
		initialCapital = *e.cfg.Options.InitialCapital
	}

	commission := acc.CommissionRate
	slippage := acc.SlippagePct
	if def.Cost != nil {
		if def.Cost.CommissionRate > 0 {
			commission = def.Cost.CommissionRate
		}
		if def.Cost.SlippagePct > 0 {
			slippage = def.Cost.SlippagePct
		}
	}
	if e.cfg.Options.CommissionRate != nil {
		commission = *e.cfg.Options.CommissionRate
	}
	if e.cfg.Options.SlippagePct != nil {
		slippage = *e.cfg.Options.SlippagePct
	}

	maxTradesPerDay := def.Risk.MaxTradesPerDay
	if e.cfg.Options.MaxTradesPerDay != nil {
		maxTradesPerDay = *e.cfg.Options.MaxTradesPerDay
	}
	allowedTimes := map[string]bool{}
	for _, t := range e.cfg.Options.AllowedTimes {
		allowedTimes[strings.TrimSpace(t)] = true
	}
	// 任务级未指定时回退到策略规则的 allowed_times
	if len(allowedTimes) == 0 {
		for _, t := range def.Rules.Buy.AllowedTimes {
			allowedTimes[strings.TrimSpace(t)] = true
		}
		for _, t := range def.Rules.Sell.AllowedTimes {
			allowedTimes[strings.TrimSpace(t)] = true
		}
	}

	// 校验回测开关
	if !e.cfg.Options.EnableBacktest {
		return nil, fmt.Errorf("回测开关未启用（enable_backtest=false），已拒绝启动回测")
	}

	// 编译信号表达式
	buyExpr, err := CompileExpr(def.Signals.Buy)
	if err != nil && strings.TrimSpace(def.Signals.Buy) != "" {
		return nil, err
	}
	sellExpr, err := CompileExpr(def.Signals.Sell)
	if err != nil && strings.TrimSpace(def.Signals.Sell) != "" {
		return nil, err
	}

	// 构建字段序列与指标
	fields := buildFieldSeries(bars)
	indicators := map[string][]float64{}
	for _, indDef := range def.Indicators {
		ind, err := buildIndicator(indDef, fields)
		if err != nil {
			return nil, err
		}
		indicators[ind.ID()] = ind.Values()
	}
	evalCtx := &EvalContext{Fields: fields, Indicators: indicators}

	fillMode := def.Data.FillMode
	if fillMode == "" {
		fillMode = FillNextBarOpen
	}

	state := &accountState{cash: initialCapital}
	dc := &dayCounters{
		ruleTrades:    map[string]int{},
		ruleRunTrades: map[string]int{},
		lastTradeBar:  -1,
	}

	result := &EngineResult{}
	trades := []Trade{}
	equityPoints := []EquityPoint{}

	// 报告桶（按 report_precision 聚合曲线点）
	bucketKeyOf := func(b Bar) string {
		switch e.cfg.ReportPrecision {
		case PeriodMin:
			return fmt.Sprintf("%d", b.TS)
		case PeriodHour:
			return fmt.Sprintf("%d-%02d", b.Date, b.Time/10000)
		default:
			return fmt.Sprintf("%d", b.Date)
		}
	}
	curBucket := ""

	// 投入统计
	var maxInvested, investedSum, investedBars float64
	investedDays := map[int]bool{}

	// 峰值与回撤
	peakEquity := initialCapital
	maxDD := 0.0
	maxDDStart, maxDDEnd := int64(0), int64(0)
	ddStartTS := int64(0)

	// 平仓盈亏统计
	var grossProfit, grossLoss float64
	winCount, lossCount := 0, 0
	signalDetail := map[string]int{}
	totalFee := 0.0
	periodReturns := []float64{}
	prevBucketEquity := initialCapital

	// 挂单（NEXT_BAR_OPEN：信号在当前 bar 收盘确认，下一 bar 开盘价成交）
	var pendingOrder *pendingOrderT

	warmup := def.Data.WarmupBars
	if warmup < 0 {
		warmup = 0
	}

	reportEvery := len(bars) / 100
	if reportEvery < 1 {
		reportEvery = 1
	}

	for i, b := range bars {
		if i < warmup {
			continue
		}
		// 取消检查（每 512 根检查一次，避免热循环频繁访问）
		if i%512 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf("回测任务已取消: %w", err)
			}
		}
		evalCtx.At = i

		// ---------- 1. 执行上一 bar 收盘产生的挂单（当前 bar 开盘价成交） ----------
		if pendingOrder != nil {
			execPrice := b.Open * (1 + pendingOrder.slippageSign*slippage/100)
			executed := e.fill(pendingOrder.action, execPrice, b, i, state, dc, &trades,
				commission, &totalFee, signalDetail, &grossProfit, &grossLoss, &winCount, &lossCount,
				pendingOrder.signal, maxTradesPerDay)
			if executed {
				dc.lastTradeBar = i
			}
			pendingOrder = nil
		}

		// ---------- 2. 风控（止损/止盈，基于当前 bar 高低价内触达） ----------
		if state.position > 0 && (def.Risk.StopLossPct > 0 || def.Risk.TakeProfitPct > 0) {
			stopPrice := state.avgCost * (1 - def.Risk.StopLossPct/100)
			tpPrice := state.avgCost * (1 + def.Risk.TakeProfitPct/100)
			exitPrice := 0.0
			reason := ""
			if def.Risk.StopLossPct > 0 && b.Low <= stopPrice {
				exitPrice = stopPrice
				reason = "止损"
			} else if def.Risk.TakeProfitPct > 0 && b.High >= tpPrice {
				exitPrice = tpPrice
				reason = "止盈"
			}
			if exitPrice > 0 {
				e.fill(ActionSell, exitPrice, b, i, state, dc, &trades, commission, &totalFee,
					signalDetail, &grossProfit, &grossLoss, &winCount, &lossCount, reason, maxTradesPerDay)
				dc.lastTradeBar = i
			}
		}

		// ---------- 3. 信号求值（基于当前 bar 收盘） ----------
		buyHit, sellHit := false, false
		if buyRuleAllowed(def) && strings.TrimSpace(def.Signals.Buy) != "" {
			v, err := buyExpr(evalCtx)
			if err != nil {
				return nil, err
			}
			buyHit = v
		}
		if sellRuleAllowed(def) && strings.TrimSpace(def.Signals.Sell) != "" {
			v, err := sellExpr(evalCtx)
			if err != nil {
				return nil, err
			}
			sellHit = v
		}

		// 限制：交易时间点
		if len(allowedTimes) > 0 {
			barTime := fmt.Sprintf("%02d%02d%02d", b.Time/10000, (b.Time/100)%100, b.Time%100)
			if buyHit && !allowedTimes[barTime] {
				buyHit = false
			}
			if sellHit && !allowedTimes[barTime] {
				sellHit = false
			}
		}

		// ---------- 4. 下单（限制：频率/次数/仓位） ----------
		if buyHit && state.position <= 0 {
			if dc.ruleAllowed(def.Rules.Buy, "buy", maxTradesPerDay, i) {
				dc.ruleTrades["buy"]++
				dc.ruleRunTrades["buy"]++
				pendingOrder = &pendingOrderT{action: ActionBuy, slippageSign: 1, signal: "买入信号"}
			}
		}
		if sellHit && state.position > 0 {
			if dc.ruleAllowed(def.Rules.Sell, "sell", maxTradesPerDay, i) {
				dc.ruleTrades["sell"]++
				dc.ruleRunTrades["sell"]++
				pendingOrder = &pendingOrderT{action: ActionSell, slippageSign: -1, signal: "卖出信号"}
			}
		}

		// ---------- 5. 账户市值快照（持仓换算现金） ----------
		positionValue := state.position * b.Close
		equity := state.cash + positionValue
		if state.position > 0 {
			if positionValue > maxInvested {
				maxInvested = positionValue
			}
			investedSum += positionValue
			investedBars++
			investedDays[b.Date] = true
		}

		// 回撤跟踪
		if equity > peakEquity {
			peakEquity = equity
			ddStartTS = b.TS
		} else {
			dd := (peakEquity - equity) / peakEquity * 100
			if dd > maxDD {
				maxDD = dd
				maxDDStart = ddStartTS
				maxDDEnd = b.TS
			}
		}

		// ---------- 6. 报告桶边界：上一桶收尾 ----------
		key := bucketKeyOf(b)
		if curBucket == "" {
			curBucket = key
		} else if key != curBucket {
			last := bars[i-1]
			curEquity := equityAt(last, state)
			dd := 0.0
			if peakEquity > 0 {
				dd = (peakEquity - curEquity) / peakEquity * 100
			}
			equityPoints = append(equityPoints, buildPoint(len(equityPoints)+1, last, state, initialCapital, curEquity, dd))
			periodReturns = append(periodReturns, pctChange(prevBucketEquity, curEquity))
			prevBucketEquity = curEquity
			curBucket = key
		}

		// ---------- 7. 进度 ----------
		if e.onProg != nil && (i%reportEvery == 0 || i == len(bars)-1) {
			e.onProg(i+1, len(bars))
		}
	}

	// 最后一个桶收尾
	if len(bars) > 0 {
		last := bars[len(bars)-1]
		curEquity := equityAt(last, state)
		dd := 0.0
		if peakEquity > 0 {
			dd = (peakEquity - curEquity) / peakEquity * 100
		}
		equityPoints = append(equityPoints, buildPoint(len(equityPoints)+1, last, state, initialCapital, curEquity, dd))
		periodReturns = append(periodReturns, pctChange(prevBucketEquity, curEquity))
	}

	// 期末总资产（持仓按最后收盘价换算现金）
	finalEquity := equityAt(bars[len(bars)-1], state)

	// 汇总报告
	report := buildReport(e.cfg, bars, initialCapital, finalEquity, trades,
		equityPoints, periodReturns, maxDD, maxDDStart, maxDDEnd, maxInvested,
		investedSum, investedBars, investedDays, grossProfit, grossLoss, winCount, lossCount,
		totalFee, signalDetail)

	result.EquityPoints = equityPoints
	result.Trades = trades
	result.Report = report
	result.BarCount = len(bars)
	return result, nil
}

func buyRuleAllowed(def StrategyDefinition) bool  { return def.Rules.Buy.Allow }
func sellRuleAllowed(def StrategyDefinition) bool { return def.Rules.Sell.Allow }

// ruleAllowed 检查规则限制：每日笔数/全回测笔数/全局每日笔数/最小间隔。
func (dc *dayCounters) ruleAllowed(rule RuleDef, key string, maxTradesPerDay, barIdx int) bool {
	if rule.MaxPerDay > 0 && dc.ruleTrades[key] >= rule.MaxPerDay {
		return false
	}
	if rule.MaxPerRun > 0 && dc.ruleRunTrades[key] >= rule.MaxPerRun {
		return false
	}
	return true
}

// equityAt 计算 bar 收盘口径总资产。
func equityAt(b Bar, state *accountState) float64 {
	return state.cash + state.position*b.Close
}

// buildPoint 构造报告点。
func buildPoint(seq int, b Bar, state *accountState, initialCapital float64, equity, dd float64) EquityPoint {
	positionValue := state.position * b.Close
	profit := equity - initialCapital
	roi := 0.0
	if initialCapital > 0 {
		roi = profit / initialCapital * 100
	}
	return EquityPoint{
		Seq:           seq,
		TS:            b.TS,
		Date:          b.Date,
		Time:          b.Time,
		Equity:        round6(equity),
		Cash:          round6(state.cash),
		PositionValue: round6(positionValue),
		PositionQty:   round6(state.position),
		Profit:        round6(profit),
		ROI:           round6(roi),
		Drawdown:      round6(dd),
	}
}

// pendingOrderT 挂单（下一 bar 开盘成交）。
type pendingOrderT struct {
	action       string
	slippageSign float64
	signal       string
}

// fill 执行一笔成交（买入/卖出），更新账户与成交记录。返回是否成交。
func (e *Engine) fill(action string, price float64, b Bar, barIdx int, state *accountState, dc *dayCounters,
	trades *[]Trade, commission float64, totalFee *float64, signalDetail map[string]int,
	grossProfit, grossLoss *float64, winCount, lossCount *int, signal string, maxTradesPerDay int) bool {

	def := e.cfg.Definition
	acc := e.cfg.Account

	// 跨日重置每日计数
	if dc.curDate != b.Date {
		dc.curDate = b.Date
		dc.trades = 0
		dc.ruleTrades = map[string]int{}
	}
	// 全局每日笔数限制
	if maxTradesPerDay > 0 && dc.trades >= maxTradesPerDay {
		return false
	}
	// 相邻交易最小间隔
	if def.Risk.MinIntervalBars > 0 && dc.lastTradeBar >= 0 && barIdx-dc.lastTradeBar < def.Risk.MinIntervalBars {
		return false
	}

	marginRate := acc.MarginRate
	if marginRate <= 0 {
		marginRate = 1
	}
	qty := 0.0
	profit := 0.0
	fee := 0.0
	amount := 0.0

	switch action {
	case ActionBuy:
		qty = e.buyQuantity(b, state, price, def.Rules.Buy, commission, marginRate)
		if qty <= 0 {
			return false
		}
		amount = qty * price
		fee = amount * commission
		if acc.MarginMode == "FUTURES" {
			state.cash -= amount*marginRate + fee
		} else {
			state.cash -= amount + fee
		}
		newPos := state.position + qty
		state.avgCost = (state.position*state.avgCost + qty*price) / newPos
		state.position = newPos
		state.entryDate = b.Date
	case ActionSell:
		qty = state.position
		if qty <= 0 {
			return false
		}
		amount = qty * price
		fee = amount * commission
		state.cash += amount - fee
		profit = (price - state.avgCost) * qty
		if profit > 0 {
			*grossProfit += profit
			*winCount++
		} else if profit < 0 {
			*grossLoss += -profit
			*lossCount++
		}
		if signalDetail != nil {
			signalDetail[signal]++
		}
		state.position = 0
		state.avgCost = 0
	}

	*totalFee += fee

	trade := Trade{
		TS:            b.TS,
		Date:          b.Date,
		Time:          b.Time,
		Action:        action,
		Price:         round6(price),
		Qty:           round6(qty),
		Amount:        round6(amount),
		Fee:           round6(fee),
		Profit:        round6(profit),
		PositionAfter: round6(state.position),
		CashAfter:     round6(state.cash),
		Signal:        signal,
	}
	*trades = append(*trades, trade)
	dc.trades++
	return true
}

// buyQuantity 计算买入数量。
func (e *Engine) buyQuantity(b Bar, state *accountState, price float64, rule RuleDef, commission, marginRate float64) float64 {
	ruleType := rule.QuantityType
	if ruleType == "" {
		ruleType = "ALL_IN"
	}
	available := state.cash
	if commission > 0 {
		available = state.cash / (1 + commission)
	}
	switch ruleType {
	case "PERCENT":
		pct := rule.Quantity
		if pct <= 0 || pct > 100 {
			pct = 100
		}
		return floorQty(available * pct / 100 / (price * marginRate))
	case "AMOUNT":
		amt := rule.Quantity
		if amt <= 0 {
			amt = available
		}
		if amt > available {
			amt = available
		}
		return floorQty(amt / (price * marginRate))
	case "FIXED":
		qty := rule.Quantity
		if qty <= 0 {
			return 0
		}
		if qty*price*marginRate > state.cash {
			qty = floorQty(available / (price * marginRate))
		}
		return floorQty(qty)
	default: // ALL_IN
		return floorQty(available / (price * marginRate))
	}
}

// floorQty 数量向下取整（保留 6 位小数，兼容股票整数股/期货手数）。
func floorQty(q float64) float64 {
	return math.Floor(q*1e6) / 1e6
}

// round6 保留 6 位小数。
func round6(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return math.Round(v*1e6) / 1e6
}

// pctChange 计算两个净值间的收益率 %。
func pctChange(prev, cur float64) float64 {
	if prev == 0 {
		return 0
	}
	return (cur - prev) / prev * 100
}

// buildFieldSeries 构建字段序列。
func buildFieldSeries(bars []Bar) map[string][]float64 {
	n := len(bars)
	open := make([]float64, n)
	high := make([]float64, n)
	low := make([]float64, n)
	closeV := make([]float64, n)
	volume := make([]float64, n)
	turnover := make([]float64, n)
	for i, b := range bars {
		open[i], high[i], low[i], closeV[i] = b.Open, b.High, b.Low, b.Close
		volume[i], turnover[i] = b.Volume, b.Turnover
	}
	return map[string][]float64{
		"open": open, "high": high, "low": low, "close": closeV,
		"volume": volume, "turnover": turnover,
	}
}

// AggregateBars 将分钟 bar 聚合为指定周期（Min 原样返回）。
// period: Min/Hour/Day。Hour 按 (date, hour) 聚合，Day 按 date 聚合。
func AggregateBars(bars []Bar, period string) ([]Bar, error) {
	if len(bars) == 0 {
		return nil, fmt.Errorf("无行情数据")
	}
	switch period {
	case "", PeriodMin:
		return bars, nil
	case PeriodHour, PeriodDay:
		sorted := make([]Bar, len(bars))
		copy(sorted, bars)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].TS < sorted[j].TS })
		var out []Bar
		for _, b := range sorted {
			key := aggKeyOf(b, period)
			if len(out) > 0 && aggKeyOf(out[len(out)-1], period) == key {
				last := &out[len(out)-1]
				if b.High > last.High {
					last.High = b.High
				}
				if b.Low < last.Low {
					last.Low = b.Low
				}
				last.Close = b.Close
				last.TS = b.TS
				last.Time = b.Time
				last.Volume += b.Volume
				last.Turnover += b.Turnover
			} else {
				out = append(out, b)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("不支持的周期 %q", period)
	}
}

func aggKeyOf(b Bar, period string) int {
	if period == PeriodHour {
		return b.Date*100 + b.Time/10000
	}
	return b.Date
}

// ---------------------------------------------------------------------
// 报告构建
// ---------------------------------------------------------------------

func buildReport(cfg EngineConfig, bars []Bar, initialCapital, finalEquity float64,
	trades []Trade, equityPoints []EquityPoint, periodReturns []float64,
	maxDD float64, maxDDStart, maxDDEnd int64, maxInvested, investedSum float64, investedBars float64,
	investedDays map[int]bool, grossProfit, grossLoss float64, winCount, lossCount int,
	totalFee float64, signalDetail map[string]int) *RunReport {

	totalProfit := finalEquity - initialCapital
	totalReturnPct := 0.0
	if initialCapital > 0 {
		totalReturnPct = totalProfit / initialCapital * 100
	}

	// 年化（按自然日）
	annualReturnPct := 0.0
	if len(bars) >= 2 {
		days := (bars[len(bars)-1].TS - bars[0].TS) / 86400
		if days < 1 {
			days = 1
		}
		if initialCapital > 0 && finalEquity > 0 {
			annualReturnPct = (math.Pow(finalEquity/initialCapital, 365.0/float64(days)) - 1) * 100
		}
	}

	// 夏普 / 波动率
	sharpe, volPct := computeSharpeVol(periodReturns, cfg.ReportPrecision)

	// 交易统计
	tradeCount := len(trades)
	buyCount, sellCount := 0, 0
	for _, t := range trades {
		if t.Action == ActionBuy {
			buyCount++
		} else {
			sellCount++
		}
	}
	winRatePct := 0.0
	closed := winCount + lossCount
	if closed > 0 {
		winRatePct = float64(winCount) / float64(closed) * 100
	}
	profitFactor := 0.0
	if grossLoss > 0 {
		profitFactor = grossProfit / grossLoss
	} else if grossProfit > 0 {
		profitFactor = math.Inf(1)
	}

	avgInvested := 0.0
	if investedBars > 0 {
		avgInvested = investedSum / investedBars
	}

	bestDay, worstDay := 0.0, 0.0
	profitDays, lossDays := 0, 0
	for _, r := range periodReturns {
		if r > bestDay {
			bestDay = r
		}
		if r < worstDay {
			worstDay = r
		}
		if r > 0 {
			profitDays++
		} else if r < 0 {
			lossDays++
		}
	}

	return &RunReport{
		SecuCode:           cfg.SecuCode,
		Period:             cfg.Period,
		ReportPrecision:    cfg.ReportPrecision,
		StartDate:          firstDate(bars),
		EndDate:            lastDate(bars),
		BarCount:           len(bars),
		InitialCapital:     round6(initialCapital),
		FinalEquity:        round6(finalEquity),
		TotalProfit:        round6(totalProfit),
		TotalReturnPct:     round6(totalReturnPct),
		AnnualReturnPct:    round6(annualReturnPct),
		MaxDrawdownPct:     round6(maxDD),
		MaxDrawdownStartTS: maxDDStart,
		MaxDrawdownEndTS:   maxDDEnd,
		SharpeRatio:        round6(sharpe),
		VolatilityPct:      round6(volPct),
		TradeCount:         tradeCount,
		BuyCount:           buyCount,
		SellCount:          sellCount,
		WinCount:           winCount,
		LossCount:          lossCount,
		WinRatePct:         round6(winRatePct),
		ProfitFactor:       round6(profitFactor),
		TotalFee:           round6(totalFee),
		MaxInvested:        round6(maxInvested),
		AvgInvested:        round6(avgInvested),
		InvestedDays:       len(investedDays),
		BestDayPct:         round6(bestDay),
		WorstDayPct:        round6(worstDay),
		ProfitDays:         profitDays,
		LossDays:           lossDays,
		TradeSignalDetail:  signalDetail,
		GeneratedAt:        time.Now().UTC().Format(time.RFC3339),
	}
}

func computeSharpeVol(periodReturns []float64, precision string) (float64, float64) {
	if len(periodReturns) < 2 {
		return 0, 0
	}
	var sum float64
	for _, r := range periodReturns {
		sum += r
	}
	mean := sum / float64(len(periodReturns))
	var variance float64
	for _, r := range periodReturns {
		variance += (r - mean) * (r - mean)
	}
	variance /= float64(len(periodReturns) - 1)
	std := math.Sqrt(variance)
	if std == 0 {
		return 0, 0
	}
	periodsPerYear := 252.0
	switch precision {
	case PeriodMin:
		periodsPerYear = 252 * 240
	case PeriodHour:
		periodsPerYear = 252 * 6
	}
	sharpe := mean / std * math.Sqrt(periodsPerYear)
	volPct := std * math.Sqrt(periodsPerYear)
	return sharpe, volPct
}

func firstDate(bars []Bar) int {
	if len(bars) == 0 {
		return 0
	}
	return bars[0].Date
}

func lastDate(bars []Bar) int {
	if len(bars) == 0 {
		return 0
	}
	return bars[len(bars)-1].Date
}
