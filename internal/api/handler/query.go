// Package handler 提供 HTTP 处理器。
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/acanx/finvquant/internal/quote"
)

// QuoteQuery 历史行情查询处理器。
type QuoteQuery struct {
	service *quote.Service
}

// NewQuoteQuery 创建查询处理器。
func NewQuoteQuery(service *quote.Service) *QuoteQuery {
	return &QuoteQuery{service: service}
}

// Query 处理 GET /API/V1/Quote/Query：
// 按证券代码 + 交易日期查询分钟级 K 线（周期目前仅支持 Min=1 分钟）。
func (h *QuoteQuery) Query(c *gin.Context) {
	secuCode := c.Query("secu_code")
	dateStr := c.Query("date")
	period := c.Query("period")
	if period == "" {
		period = "Min"
	}

	if secuCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "secu_code 不能为空"})
		return
	}
	date, err := strconv.Atoi(dateStr)
	if err != nil || date <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "date 必须为有效的交易日期（yyyymmdd）"})
		return
	}

	ctx, cancel := contextWithTimeout(c, 30*time.Second)
	defer cancel()

	bars, err := h.service.QueryBars(ctx, secuCode, date, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2006, "message": "查询失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "查询完成",
		"data": gin.H{
			"secu_code": secuCode,
			"date":      date,
			"period":    period,
			"count":     len(bars),
			"bars":      bars,
		},
	})
}
