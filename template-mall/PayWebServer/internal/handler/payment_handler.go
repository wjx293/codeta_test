package handler

import (
	"net/http"
	"strings"

	"template-mall/PayWebServer/internal/idgen"
	"template-mall/PayWebServer/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentHandler 支付单 HTTP handler
type PaymentHandler struct {
	svc *service.PaymentService
	idg *idgen.Generator
}

// NewPaymentHandler 创建支付单 handler。
func NewPaymentHandler(svc *service.PaymentService, idg *idgen.Generator) *PaymentHandler {
	return &PaymentHandler{svc: svc, idg: idg}
}

// CreatePaymentRequest 创建支付单请求
type CreatePaymentRequest struct {
	BusinessOrderNo string `json:"business_order_no" binding:"required"`
	Amount          int64  `json:"amount" binding:"required"`
}

// CreatePayment 创建支付单。
// POST /api/payments
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	paymentNo := h.idg.GeneratePaymentNo()
	resp, err := h.svc.CreatePayment(req.BusinessOrderNo, req.Amount, paymentNo)
	if err != nil {
		// 区分业务错误和系统错误
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"payment_no":        resp.PaymentNo,
			"business_order_no": resp.BusinessOrderNo,
			"amount":            resp.Amount,
			"status":            resp.Status,
		},
	})
}
