package backtest

import (
	"fmt"
	"strings"
)

// validateStrategy 校验策略必填项与定义结构。
func validateStrategy(st *Strategy) error {
	if strings.TrimSpace(st.StrategyCode) == "" {
		return fmt.Errorf("strategy_code（策略编码）必填")
	}
	if strings.TrimSpace(st.StrategyName) == "" {
		return fmt.Errorf("strategy_name（策略名称）必填")
	}
	if st.StrategyType == "" {
		st.StrategyType = StrategyTypeRuleBased
	}
	if st.StrategyType != StrategyTypeRuleBased {
		return fmt.Errorf("当前版本仅支持 RULE_BASED 策略类型（%s 暂未实现）", st.StrategyType)
	}
	def := st.Definition
	if len(def.Universe.Securities) == 0 {
		return fmt.Errorf("definition.universe.securities（标的池）至少需要一个证券代码")
	}
	for _, code := range def.Universe.Securities {
		if strings.TrimSpace(code) == "" {
			return fmt.Errorf("definition.universe.securities 中存在空证券代码")
		}
	}
	if def.Data.Period != "" && def.Data.Period != PeriodMin && def.Data.Period != PeriodHour && def.Data.Period != PeriodDay {
		return fmt.Errorf("definition.data.period 仅支持 Min/Hour/Day")
	}

	// 指标 ID 唯一性 + 指标定义合法性（不加载数据，仅类型校验）
	seen := map[string]bool{}
	for _, ind := range def.Indicators {
		if strings.TrimSpace(ind.ID) == "" {
			return fmt.Errorf("指标定义缺少 id")
		}
		if seen[ind.ID] {
			return fmt.Errorf("指标 id 重复: %s", ind.ID)
		}
		seen[ind.ID] = true
		switch ind.Type {
		case "MA", "EMA", "RSI", "MACD", "BOLL", "ATR", "STDDEV", "HHV", "LLV":
		default:
			return fmt.Errorf("不支持的指标类型 %q（支持: MA/EMA/RSI/MACD/BOLL/ATR/STDDEV/HHV/LLV）", ind.Type)
		}
	}

	// 信号表达式语法校验
	if strings.TrimSpace(def.Signals.Buy) != "" {
		if _, err := CompileExpr(def.Signals.Buy); err != nil {
			return fmt.Errorf("买入信号表达式错误: %w", err)
		}
	}
	if strings.TrimSpace(def.Signals.Sell) != "" {
		if _, err := CompileExpr(def.Signals.Sell); err != nil {
			return fmt.Errorf("卖出信号表达式错误: %w", err)
		}
	}

	// 规则校验
	for name, rule := range map[string]RuleDef{"buy": def.Rules.Buy, "sell": def.Rules.Sell} {
		if rule.QuantityType != "" {
			switch rule.QuantityType {
			case "ALL_IN", "ALL", "FIXED", "PERCENT", "AMOUNT":
			default:
				return fmt.Errorf("rules.%s.quantity_type 不支持 %q", name, rule.QuantityType)
			}
		}
		if (rule.QuantityType == "FIXED" || rule.QuantityType == "AMOUNT" || rule.QuantityType == "PERCENT") && rule.Quantity <= 0 {
			return fmt.Errorf("rules.%s.quantity_type=%s 时 quantity 必须大于 0", name, rule.QuantityType)
		}
	}
	if def.Risk.StopLossPct < 0 || def.Risk.TakeProfitPct < 0 {
		return fmt.Errorf("risk.stop_loss_pct / take_profit_pct 不能为负数")
	}
	if def.Risk.MaxPositionPct < 0 || def.Risk.MaxPositionPct > 100 {
		return fmt.Errorf("risk.max_position_pct 必须在 0~100 之间")
	}
	return nil
}

// validateAccount 校验回测账户必填项。
func validateAccount(acc *Account) error {
	if strings.TrimSpace(acc.AccountCode) == "" {
		return fmt.Errorf("account_code（账户编码）必填")
	}
	if strings.TrimSpace(acc.AccountName) == "" {
		return fmt.Errorf("account_name（账户名称）必填")
	}
	if acc.InitialCapital <= 0 {
		return fmt.Errorf("initial_capital（初始启动资金）必须大于 0")
	}
	if acc.CommissionRate < 0 || acc.SlippagePct < 0 {
		return fmt.Errorf("commission_rate / slippage_pct 不能为负数")
	}
	if acc.MarginMode != "" && acc.MarginMode != "FULL" && acc.MarginMode != "FUTURES" {
		return fmt.Errorf("margin_mode 仅支持 FULL/FUTURES")
	}
	return nil
}

// validateEnvironment 校验环境必填项与配置。
func validateEnvironment(env *Environment) error {
	if strings.TrimSpace(env.EnvCode) == "" {
		return fmt.Errorf("env_code（环境编码）必填")
	}
	if strings.TrimSpace(env.EnvName) == "" {
		return fmt.Errorf("env_name（环境名称）必填")
	}
	switch env.EnvType {
	case "", "BACKTEST", "PAPER", "SIMULATION", "LIVE":
	default:
		return fmt.Errorf("env_type 仅支持 BACKTEST/PAPER/SIMULATION/LIVE")
	}
	for _, s := range env.Config.TradingSessions {
		if len(s.Start) != 6 || len(s.End) != 6 {
			return fmt.Errorf("trading_sessions 起止时间需为 hhmmss（如 093000）")
		}
	}
	if env.Config.TradingRules != nil && env.Config.TradingRules.TickSize < 0 {
		return fmt.Errorf("trading_rules.tick_size 不能为负数")
	}
	return nil
}

// validateTemplate 校验模板必填项。
func validateTemplate(tmpl *Template) error {
	if strings.TrimSpace(tmpl.TemplateCode) == "" {
		return fmt.Errorf("template_code（模板编码）必填")
	}
	if strings.TrimSpace(tmpl.TemplateName) == "" {
		return fmt.Errorf("template_name（模板名称）必填")
	}
	switch tmpl.TemplateType {
	case "STRATEGY", "ACCOUNT", "ENVIRONMENT":
	default:
		return fmt.Errorf("template_type 仅支持 STRATEGY/ACCOUNT/ENVIRONMENT")
	}
	return nil
}
