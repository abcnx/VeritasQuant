// Package handler 提供 HTTP 处理器。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/acanx/finvquant/internal/backtest"
)

// parseBTPager 解析分页参数为 backtest.Pager。
func parseBTPager(c *gin.Context) backtest.Pager {
	p := parsePager(c)
	return backtest.Pager{Page: p.Page, PageSize: p.PageSize}
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

// ListStrategies GET /API/V1/Backtest/Strategy/List
func (h *Backtest) ListStrategies(c *gin.Context) {
	list, total, err := h.service.ListStrategies(c.Request.Context(),
		parseBTPager(c), c.Query("keyword"), c.Query("allow_backtest"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": "查询策略列表失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}

// GetStrategy GET /API/V1/Backtest/Strategy/Get?strategy_id=xxx
func (h *Backtest) GetStrategy(c *gin.Context) {
	st, err := h.service.GetStrategy(c.Request.Context(), c.Query("strategy_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": st})
}

// SaveStrategy POST /API/V1/Backtest/Strategy/Save
func (h *Backtest) SaveStrategy(c *gin.Context) {
	var st backtest.Strategy
	if err := c.ShouldBindJSON(&st); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	id, err := h.service.SaveStrategy(c.Request.Context(), &st)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": "保存策略失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "保存成功", "data": gin.H{"strategy_id": id}})
}

// ToggleStrategy POST /API/V1/Backtest/Strategy/Toggle
func (h *Backtest) ToggleStrategy(c *gin.Context) {
	var req struct {
		StrategyID    string `json:"strategy_id"`
		AllowBacktest string `json:"allow_backtest"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	if err := h.service.ToggleStrategy(c.Request.Context(), req.StrategyID, req.AllowBacktest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": "切换策略回测开关失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "切换成功"})
}

// DeleteStrategy POST /API/V1/Backtest/Strategy/Delete
func (h *Backtest) DeleteStrategy(c *gin.Context) {
	var req struct {
		StrategyID string `json:"strategy_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	if err := h.service.DeleteStrategy(c.Request.Context(), req.StrategyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": "删除策略失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// ---------------------------------------------------------------------
// 账户管理
// ---------------------------------------------------------------------

// ListAccounts GET /API/V1/Backtest/Account/List
func (h *Backtest) ListAccounts(c *gin.Context) {
	list, total, err := h.service.ListAccounts(c.Request.Context(),
		parseBTPager(c), c.Query("keyword"), c.Query("allow_backtest"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": "查询账户列表失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}

// GetAccount GET /API/V1/Backtest/Account/Get?account_id=xxx
func (h *Backtest) GetAccount(c *gin.Context) {
	acc, err := h.service.GetAccount(c.Request.Context(), c.Query("account_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": acc})
}

// SaveAccount POST /API/V1/Backtest/Account/Save
func (h *Backtest) SaveAccount(c *gin.Context) {
	var acc backtest.Account
	if err := c.ShouldBindJSON(&acc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	id, err := h.service.SaveAccount(c.Request.Context(), &acc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": "保存账户失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "保存成功", "data": gin.H{"account_id": id}})
}

// ToggleAccount POST /API/V1/Backtest/Account/Toggle
func (h *Backtest) ToggleAccount(c *gin.Context) {
	var req struct {
		AccountID     string `json:"account_id"`
		AllowBacktest string `json:"allow_backtest"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	if err := h.service.ToggleAccount(c.Request.Context(), req.AccountID, req.AllowBacktest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": "切换账户回测开关失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "切换成功"})
}

// DeleteAccount POST /API/V1/Backtest/Account/Delete
func (h *Backtest) DeleteAccount(c *gin.Context) {
	var req struct {
		AccountID string `json:"account_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	if err := h.service.DeleteAccount(c.Request.Context(), req.AccountID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": "删除账户失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// ---------------------------------------------------------------------
// 回测任务
// ---------------------------------------------------------------------

// CreateRun POST /API/V1/Backtest/Run/Create
func (h *Backtest) CreateRun(c *gin.Context) {
	var req backtest.CreateRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	run, err := h.service.CreateRun(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": "创建回测任务失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "回测任务已创建并启动", "data": run})
}

// ListRuns GET /API/V1/Backtest/Run/List
func (h *Backtest) ListRuns(c *gin.Context) {
	q := backtest.RunListQuery{
		Pager:      parseBTPager(c),
		Status:     c.Query("status"),
		SecuCode:   c.Query("secu_code"),
		StrategyID: c.Query("strategy_id"),
		Keyword:    c.Query("keyword"),
	}
	list, total, err := h.service.ListRuns(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": "查询回测任务失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}

// GetRun GET /API/V1/Backtest/Run/Get?run_id=xxx
func (h *Backtest) GetRun(c *gin.Context) {
	run, err := h.service.GetRun(c.Request.Context(), c.Query("run_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": run})
}

// CancelRun POST /API/V1/Backtest/Run/Cancel
func (h *Backtest) CancelRun(c *gin.Context) {
	var req struct {
		RunID string `json:"run_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "请求体格式错误: " + err.Error()})
		return
	}
	if err := h.service.CancelRun(c.Request.Context(), req.RunID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": "取消失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "取消请求已受理"})
}

// GetReport GET /API/V1/Backtest/Run/Report?run_id=xxx
func (h *Backtest) GetReport(c *gin.Context) {
	report, err := h.service.GetReport(c.Request.Context(), c.Query("run_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": report})
}

// ListEquity GET /API/V1/Backtest/Run/Equity?run_id=xxx&page=&page_size=
func (h *Backtest) ListEquity(c *gin.Context) {
	list, total, err := h.service.ListEquity(c.Request.Context(),
		c.Query("run_id"), parseBTPager(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": "查询净值曲线失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}

// ListTrades GET /API/V1/Backtest/Run/Trades?run_id=xxx&page=&page_size=
func (h *Backtest) ListTrades(c *gin.Context) {
	list, total, err := h.service.ListTrades(c.Request.Context(),
		c.Query("run_id"), parseBTPager(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": "查询成交记录失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "查询完成", "data": gin.H{"total": total, "list": list}})
}
