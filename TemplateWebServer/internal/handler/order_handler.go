package handler

import (
	"net/http"
	"strconv"

	"template-mall/TemplateWebServer/internal/client"
	"template-mall/TemplateWebServer/internal/middleware"

	pb "template-mall/TemplateOrderServer/api/proto"

	"github.com/gin-gonic/gin"
)

// OrderHandler 订单 handler
type OrderHandler struct {
	grpcClient *client.GRPCClient
}

// NewOrderHandler 创建订单 handler。
func NewOrderHandler(grpcClient *client.GRPCClient) *OrderHandler {
	return &OrderHandler{grpcClient: grpcClient}
}

// ListOrders 我的订单列表。
// GET /orders?status=-1&page=1&page_size=20
func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "user not found in context"})
		return
	}

	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	ctx, cancel := withTimeout(c)
	defer cancel()

	userIDVal := int64(userID)
	pageVal := int32(page)
	pageSizeVal := int32(pageSize)
	req := &pb.ListOrdersRequest{
		UserId:   &userIDVal,
		Page:     &pageVal,
		PageSize: &pageSizeVal,
	}
	if status >= 0 {
		req.Status = orderStatusFromQuery(status)
	}
	resp, err := h.grpcClient.Stub.ListOrders(ctx, req)
	if err != nil {
		respondGRPCError(c, err)
		return
	}

	if resp.Base != nil && resp.Base.Code != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.Base.Code, "message": resp.Base.Message})
		return
	}

	orders := make([]gin.H, 0, len(resp.Orders))
	for _, o := range resp.Orders {
		orders = append(orders, orderToJSON(o))
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"data":  gin.H{"orders": orders, "total": resp.Total},
	})
}

// CancelOrder 取消未支付零售订单。
// POST /orders/:order_no/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderNo := c.Param("order_no")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "missing order_no"})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "user not found in context"})
		return
	}

	ctx, cancel := withTimeout(c)
	defer cancel()

	resp, err := h.grpcClient.Stub.CancelOrder(ctx, &pb.CancelOrderRequest{
		UserId:   int64(userID),
		OrderNo:  orderNo,
	})
	if err != nil {
		respondGRPCError(c, err)
		return
	}

	if resp.Code != 0 {
		status := http.StatusBadRequest
		if resp.Code == 3004 {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"code": resp.Code, "message": resp.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "order cancelled"})
}

// orderToJSON 将 proto Order 转换为 JSON 响应。
func orderToJSON(o *pb.Order) gin.H {
	var createdAt, updatedAt string
	if o.CreatedAt != nil {
		createdAt = o.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00")
	}
	if o.UpdatedAt != nil {
		updatedAt = o.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00")
	}

	return gin.H{
		"id":             o.Id,
		"order_no":       o.OrderNo,
		"user_id":        o.UserId,
		"username":       o.Username,
		"template_id":    o.TemplateId,
		"template_name":  o.TemplateName,
		"order_type":     orderTypeJSON(o.OrderType),
		"price_snapshot": o.PriceSnapshot,
		"status":         orderStatusJSON(o.Status),
		"created_at":     createdAt,
		"updated_at":     updatedAt,
	}
}