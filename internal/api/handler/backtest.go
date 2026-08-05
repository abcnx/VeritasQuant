// Package handler 提供 HTTP 处理器。
package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/acanx/finvquant/internal/backtest"
)

// parseBTPager 解析分页参数为 backtest.Pager。
func parseBTPager(c *gin.Context) backtest.Pager {
	p := parsePager(c)
	return backtest.Pager{Page: p.Page, PageSize: p.PageSize}
}

// btUserID 解析当前用户标识（多用户隔离；接入 JWT/RBAC 后改为从登录态获取）。
func btUserID(c *gin.Context) string {
	uid := strings.TrimSpace(c.Query("user_id"))
	if uid == "" {
		uid = "default"
	}
	return uid
}

// respondError 统一错误响应：按错误语义映射 HTTP 状态码与业务码（评审：原实现全部 500+2006）。
//   - 参数/校验错误 → 400（业务码 4001）
//   - 资源不存在 → 404（业务码 4004）
//   - 状态冲突/禁止操作 → 409（业务码 4009）
//   - 其他 → 500（业务码 2006）
func respondError(c *gin.Context, action string, err error) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "不存在"):
		c.JSON(http.StatusNotFound, gin.H{"code": 4004, "message": action + ": " + msg})
	case strings.Contains(msg, "禁止") || strings.Contains(msg, "已关联") || strings.Contains(msg, "无权") ||
		strings.Contains(msg, "已结束") || strings.Contains(msg, "无法取消") || strings.Contains(msg, "不一致"):
		c.JSON(http.StatusConflict, gin.H{"code": 4009, "message": action + ": " + msg})
	case strings.Contains(msg, "必填") || strings.Contains(msg, "仅支持") || strings.Contains(msg, "不支持") ||
		strings.Contains(msg, "必须") || strings.Contains(msg, "不能") || strings.Contains(msg, "格式错误") ||
		strings.Contains(msg, "非法") || strings.Contains(msg, "错误"):
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": action + ": " + msg})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": action + ": " + msg})
	}
}

// btIDQuery 校验必填 ID 查询参数（缺失返回 400，评审：原实现缺参直接查库）。
func btIDQuery(c *gin.Context, key string) (string, bool) {
	v := strings.TrimSpace(c.Query(key))
	if v == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "缺少必填参数 " + key})
		return "", false
	}
	return v, true
}

// Backtest 通用量化回测处理器。
type Backtest struct {
	service *backtest.Service
}

// NewBacktest 创建回测处理器。
func NewBacktest(service *backtest.Service) *Backtest {
	return &Backtest{service: service}
}

// ---------------------------------------------------------------------
// 策略管理
// ---------------------------------------------------------------------

// ListStrategies GET /API/V1/Meta/FinvQuant/Backtest/Strategy/List
func (h *Backtest) ListStrategies(c *gin.Context) {
	list, total, err := h.service.ListStrategies(c.Request.Context(),
		parseBTPager(c), c.Query("keyword"), c.Query("allow_backtest"), btUserID(c))
	if err != nil {
		respondError(c, "查询策略列表失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}

// GetStrategy GET /API/V1/Meta/FinvQuant/Backtest/Strategy/Get?strategy_id=xxx
func (h *Backtest) GetStrategy(c *gin.Context) {
	id, ok := btIDQuery(c, "strategy_id")
	if !ok {
		return
	}
	st, err := h.service.GetStrategy(c.Request.Context(), id, btUserID(c))
	if err != nil {
		respondError(c, "查询策略失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": st})
}

// SaveStrategy POST /API/V1/Meta/FinvQuant/Backtest/Strategy/Save
func (h *Backtest) SaveStrategy(c *gin.Context) {
	var st backtest.Strategy
	if err := c.ShouldBindJSON(&st); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	id, err := h.service.SaveStrategy(c.Request.Context(), &st)
	if err != nil {
		respondError(c, "保存策略失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "保存成功", "data": gin.H{"strategy_id": id}})
}

// ToggleStrategy POST /API/V1/Meta/FinvQuant/Backtest/Strategy/Toggle
func (h *Backtest) ToggleStrategy(c *gin.Context) {
	var req struct {
		StrategyID    string `json:"strategy_id"`
		AllowBacktest string `json:"allow_backtest"`
		UserID        string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	uid := req.UserID
	if uid == "" {
		uid = "default"
	}
	if err := h.service.ToggleStrategy(c.Request.Context(), req.StrategyID, req.AllowBacktest, uid); err != nil {
		respondError(c, "切换策略回测开关失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "切换成功"})
}

// DeleteStrategy POST /API/V1/Meta/FinvQuant/Backtest/Strategy/Delete
func (h *Backtest) DeleteStrategy(c *gin.Context) {
	var req struct {
		StrategyID string `json:"strategy_id"`
		UserID     string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	uid := req.UserID
	if uid == "" {
		uid = "default"
	}
	if err := h.service.DeleteStrategy(c.Request.Context(), req.StrategyID, uid); err != nil {
		respondError(c, "删除策略失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// ---------------------------------------------------------------------
// 账户管理
// ---------------------------------------------------------------------

// ListAccounts GET /API/V1/Meta/FinvQuant/Backtest/Account/List
func (h *Backtest) ListAccounts(c *gin.Context) {
	list, total, err := h.service.ListAccounts(c.Request.Context(),
		parseBTPager(c), c.Query("keyword"), c.Query("allow_backtest"), btUserID(c))
	if err != nil {
		respondError(c, "查询账户列表失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}

// GetAccount GET /API/V1/Meta/FinvQuant/Backtest/Account/Get?account_id=xxx
func (h *Backtest) GetAccount(c *gin.Context) {
	id, ok := btIDQuery(c, "account_id")
	if !ok {
		return
	}
	acc, err := h.service.GetAccount(c.Request.Context(), id, btUserID(c))
	if err != nil {
		respondError(c, "查询账户失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": acc})
}

// SaveAccount POST /API/V1/Meta/FinvQuant/Backtest/Account/Save
func (h *Backtest) SaveAccount(c *gin.Context) {
	var acc backtest.Account
	if err := c.ShouldBindJSON(&acc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	id, err := h.service.SaveAccount(c.Request.Context(), &acc)
	if err != nil {
		respondError(c, "保存账户失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "保存成功", "data": gin.H{"account_id": id}})
}

// ToggleAccount POST /API/V1/Meta/FinvQuant/Backtest/Account/Toggle
func (h *Backtest) ToggleAccount(c *gin.Context) {
	var req struct {
		AccountID     string `json:"account_id"`
		AllowBacktest string `json:"allow_backtest"`
		UserID        string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	uid := req.UserID
	if uid == "" {
		uid = "default"
	}
	if err := h.service.ToggleAccount(c.Request.Context(), req.AccountID, req.AllowBacktest, uid); err != nil {
		respondError(c, "切换账户回测开关失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "切换成功"})
}

// DeleteAccount POST /API/V1/Meta/FinvQuant/Backtest/Account/Delete
func (h *Backtest) DeleteAccount(c *gin.Context) {
	var req struct {
		AccountID string `json:"account_id"`
		UserID    string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	uid := req.UserID
	if uid == "" {
		uid = "default"
	}
	if err := h.service.DeleteAccount(c.Request.Context(), req.AccountID, uid); err != nil {
		respondError(c, "删除账户失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// ---------------------------------------------------------------------
// 回测任务
// ---------------------------------------------------------------------

// CreateRun POST /API/V1/Meta/FinvQuant/Backtest/Run/Create
func (h *Backtest) CreateRun(c *gin.Context) {
	var req backtest.CreateRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	run, err := h.service.CreateRun(c.Request.Context(), req)
	if err != nil {
		respondError(c, "创建回测任务失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "回测任务已创建并启动", "data": run})
}

// ListRuns GET /API/V1/Meta/FinvQuant/Backtest/Run/List
func (h *Backtest) ListRuns(c *gin.Context) {
	q := backtest.RunListQuery{
		Pager:      parseBTPager(c),
		Status:     c.Query("status"),
		SecuCode:   c.Query("secu_code"),
		StrategyID: c.Query("strategy_id"),
		Keyword:    c.Query("keyword"),
	}
	list, total, err := h.service.ListRuns(c.Request.Context(), q, btUserID(c))
	if err != nil {
		respondError(c, "查询回测任务失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}

// GetRun GET /API/V1/Meta/FinvQuant/Backtest/Run/Get?run_id=xxx
func (h *Backtest) GetRun(c *gin.Context) {
	id, ok := btIDQuery(c, "run_id")
	if !ok {
		return
	}
	run, err := h.service.GetRun(c.Request.Context(), id, btUserID(c))
	if err != nil {
		respondError(c, "查询回测任务失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": run})
}

// CancelRun POST /API/V1/Meta/FinvQuant/Backtest/Run/Cancel
func (h *Backtest) CancelRun(c *gin.Context) {
	var req struct {
		RunID  string `json:"run_id"`
		UserID string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	uid := req.UserID
	if uid == "" {
		uid = "default"
	}
	if err := h.service.CancelRun(c.Request.Context(), req.RunID, uid); err != nil {
		respondError(c, "取消失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "取消请求已受理"})
}

// GetReport GET /API/V1/Meta/FinvQuant/Backtest/Run/Report?run_id=xxx
func (h *Backtest) GetReport(c *gin.Context) {
	id, ok := btIDQuery(c, "run_id")
	if !ok {
		return
	}
	report, err := h.service.GetReport(c.Request.Context(), id, btUserID(c))
	if err != nil {
		respondError(c, "查询报告失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": report})
}

// ListEquity GET /API/V1/Meta/FinvQuant/Backtest/Run/Equity?run_id=xxx&page=&page_size=
func (h *Backtest) ListEquity(c *gin.Context) {
	id, ok := btIDQuery(c, "run_id")
	if !ok {
		return
	}
	list, total, err := h.service.ListEquity(c.Request.Context(), id, btUserID(c), parseBTPager(c))
	if err != nil {
		respondError(c, "查询净值曲线失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}

// ListTrades GET /API/V1/Meta/FinvQuant/Backtest/Run/Trades?run_id=xxx&page=&page_size=
func (h *Backtest) ListTrades(c *gin.Context) {
	id, ok := btIDQuery(c, "run_id")
	if !ok {
		return
	}
	list, total, err := h.service.ListTrades(c.Request.Context(), id, btUserID(c), parseBTPager(c))
	if err != nil {
		respondError(c, "查询成交记录失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}

// ListCashflows GET /API/V1/Meta/FinvQuant/Backtest/Run/Cashflows?run_id=xxx&page=&page_size=
func (h *Backtest) ListCashflows(c *gin.Context) {
	id, ok := btIDQuery(c, "run_id")
	if !ok {
		return
	}
	list, total, err := h.service.ListCashflows(c.Request.Context(), id, btUserID(c), parseBTPager(c))
	if err != nil {
		respondError(c, "查询资金流水失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}

// ListPositionLogs GET /API/V1/Meta/FinvQuant/Backtest/Run/PositionLogs?run_id=xxx&page=&page_size=
func (h *Backtest) ListPositionLogs(c *gin.Context) {
	id, ok := btIDQuery(c, "run_id")
	if !ok {
		return
	}
	list, total, err := h.service.ListPositionLogs(c.Request.Context(), id, btUserID(c), parseBTPager(c))
	if err != nil {
		respondError(c, "查询持仓变化明细失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}

// ListEventTraces GET /API/V1/Meta/FinvQuant/Backtest/Run/EventTraces?run_id=xxx&page=&page_size=
func (h *Backtest) ListEventTraces(c *gin.Context) {
	id, ok := btIDQuery(c, "run_id")
	if !ok {
		return
	}
	list, total, err := h.service.ListEventTraces(c.Request.Context(), id, btUserID(c), parseBTPager(c))
	if err != nil {
		respondError(c, "查询事件追踪失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}

// ---------------------------------------------------------------------
// 环境管理
// ---------------------------------------------------------------------

// ListEnvironments GET /API/V1/Meta/FinvQuant/Backtest/Environment/List
func (h *Backtest) ListEnvironments(c *gin.Context) {
	list, total, err := h.service.ListEnvironments(c.Request.Context(),
		parseBTPager(c), btUserID(c), c.Query("env_type"), c.Query("keyword"))
	if err != nil {
		respondError(c, "查询环境失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}

// GetEnvironment GET /API/V1/Meta/FinvQuant/Backtest/Environment/Get?env_id=xxx
func (h *Backtest) GetEnvironment(c *gin.Context) {
	id, ok := btIDQuery(c, "env_id")
	if !ok {
		return
	}
	env, err := h.service.GetEnvironment(c.Request.Context(), id, btUserID(c))
	if err != nil {
		respondError(c, "查询环境失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": env})
}

// SaveEnvironment POST /API/V1/Meta/FinvQuant/Backtest/Environment/Save
func (h *Backtest) SaveEnvironment(c *gin.Context) {
	var env backtest.Environment
	if err := c.ShouldBindJSON(&env); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	id, err := h.service.SaveEnvironment(c.Request.Context(), &env)
	if err != nil {
		respondError(c, "保存环境失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "保存成功", "data": gin.H{"env_id": id}})
}

// ToggleEnvironment POST /API/V1/Meta/FinvQuant/Backtest/Environment/Toggle
func (h *Backtest) ToggleEnvironment(c *gin.Context) {
	var req struct {
		EnvID         string `json:"env_id"`
		AllowBacktest string `json:"allow_backtest"`
		UserID        string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	uid := req.UserID
	if uid == "" {
		uid = "default"
	}
	if err := h.service.ToggleEnvironment(c.Request.Context(), req.EnvID, req.AllowBacktest, uid); err != nil {
		respondError(c, "切换环境回测开关失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "切换成功"})
}

// DeleteEnvironment POST /API/V1/Meta/FinvQuant/Backtest/Environment/Delete
func (h *Backtest) DeleteEnvironment(c *gin.Context) {
	var req struct {
		EnvID  string `json:"env_id"`
		UserID string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	uid := req.UserID
	if uid == "" {
		uid = "default"
	}
	if err := h.service.DeleteEnvironment(c.Request.Context(), req.EnvID, uid); err != nil {
		respondError(c, "删除环境失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// ---------------------------------------------------------------------
// 模板管理
// ---------------------------------------------------------------------

// ListTemplates GET /API/V1/Meta/FinvQuant/Backtest/Template/List
func (h *Backtest) ListTemplates(c *gin.Context) {
	list, total, err := h.service.ListTemplates(c.Request.Context(),
		parseBTPager(c), btUserID(c), c.Query("template_type"), c.Query("keyword"))
	if err != nil {
		respondError(c, "查询模板失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}

// GetTemplate GET /API/V1/Meta/FinvQuant/Backtest/Template/Get?template_id=xxx
func (h *Backtest) GetTemplate(c *gin.Context) {
	id, ok := btIDQuery(c, "template_id")
	if !ok {
		return
	}
	tmpl, err := h.service.GetTemplate(c.Request.Context(), id, btUserID(c))
	if err != nil {
		respondError(c, "查询模板失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": tmpl})
}

// SaveTemplate POST /API/V1/Meta/FinvQuant/Backtest/Template/Save
func (h *Backtest) SaveTemplate(c *gin.Context) {
	var tmpl backtest.Template
	if err := c.ShouldBindJSON(&tmpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	id, err := h.service.SaveTemplate(c.Request.Context(), &tmpl)
	if err != nil {
		respondError(c, "保存模板失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "保存成功", "data": gin.H{"template_id": id}})
}

// DeleteTemplate POST /API/V1/Meta/FinvQuant/Backtest/Template/Delete
func (h *Backtest) DeleteTemplate(c *gin.Context) {
	var req struct {
		TemplateID string `json:"template_id"`
		UserID     string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	uid := req.UserID
	if uid == "" {
		uid = "default"
	}
	if err := h.service.DeleteTemplate(c.Request.Context(), req.TemplateID, uid); err != nil {
		respondError(c, "删除模板失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}
