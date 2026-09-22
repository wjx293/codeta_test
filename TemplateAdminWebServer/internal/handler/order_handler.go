package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"template-mall/TemplateAdminWebServer/internal/client"

	pb "template-mall/TemplateOrderServer/api/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OrderHandler B 端订单 handler。
type OrderHandler struct {
	grpcClient *client.GRPCClient
}

// NewOrderHandler 创建订单 handler。
func NewOrderHandler(grpcClient *client.GRPCClient) *OrderHandler {
	return &OrderHandler{grpcClient: grpcClient}
}

// List GET /api/admin/orders
// 筛选维度：status、user_id、username、时间范围（CL-12）
func (h *OrderHandler) List(c *gin.Context) {
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	username := c.Query("username")
	userIDStr := c.Query("user_id")
	var userID int64
	if userIDStr != "" {
		if v, err := strconv.ParseInt(userIDStr, 10, 64); err == nil && v > 0 {
			userID = v
		}
	}
	startTimeStr := c.Query("start_time") // RFC3339
	endTimeStr := c.Query("end_time")

	var startTime, endTime *timestamppb.Timestamp
	if startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = timestamppb.New(t)
		}
	}
	if endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = timestamppb.New(t)
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	pageVal := int32(page)
	pageSizeVal := int32(pageSize)
	req := &pb.ListOrdersRequest{
		Page:      &pageVal,
		PageSize:  &pageSizeVal,
		Username:  username,
		StartTime: startTime,
		EndTime:   endTime,
	}
	if userID > 0 {
		req.UserId = &userID
	}
	if status >= 0 {
		req.Status = orderStatusFromQuery(status)
	}
	resp, err := h.grpcClient.Stub.ListOrders(ctx, req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": err.Error()})
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
		"code": 0,
		"data": gin.H{"orders": orders, "total": resp.Total},
	})
}

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