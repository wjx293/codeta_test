package handler

import (
	"errors"
	"net/http"

	"template-mall/PayWebServer/internal/service"

	"github.com/gin-gonic/gin"
)

// CallbackHandler 支付回调 HTTP handler
type CallbackHandler struct {
	svc *service.PaymentService
}

// NewCallbackHandler 创建回调 handler。
func NewCallbackHandler(svc *service.PaymentService) *CallbackHandler {
	return &CallbackHandler{svc: svc}
}

// CallbackRequest 支付回调请求
type CallbackRequest struct {
	PaymentNo       string `json:"payment_no" binding:"required"`  // 支付单编号
	BusinessOrderNo string `json:"business_order_no" binding:"-"`  // 业务订单编号
	Amount          int64  `json:"amount" binding:"required"`      // 金额
	Result          string `json:"result" binding:"required"`      // 结果
	CallbackNo      string `json:"callback_no" binding:"required"` // 回调编号
}

// PayCallback 处理支付回调。
// POST /api/paycallback
func (h *CallbackHandler) PayCallback(c *gin.Context) {
	var req CallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 校验 result 字段
	if req.Result != "success" && req.Result != "fail" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid result, must be success or fail"})
		return
	}

	resp, err := h.svc.HandleCallback(c.Request.Context(), req.CallbackNo, req.PaymentNo, req.Result, req.Amount)
	if err != nil {
		switch {
		case err == service.ErrPaymentNotFound:
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "payment not found"})
		case err == service.ErrAmountMismatch:
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "amount mismatch"})
		case err == service.ErrStatusInvalid:
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "payment status invalid"})
		case errors.Is(err, service.ErrKafkaProduce):
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "kafka produce failed"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		}
		return
	}

	// 金额不一致时不发 Kafka
	_ = resp

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// Health 健康检查。
// GET /health
func (h *CallbackHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "PayWebServer",
		"version": "1.0.0",
	})
}
