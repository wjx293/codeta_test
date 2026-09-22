package handler

import (
	"net/http"
	"strconv"

	"template-mall/TemplateWebServer/internal/client"
	"template-mall/TemplateWebServer/internal/middleware"

	pb "template-mall/TemplateOrderServer/api/proto"

	"github.com/gin-gonic/gin"
)

// TemplateHandler 模板 handler
type TemplateHandler struct {
	grpcClient *client.GRPCClient
}

// NewTemplateHandler 创建模板 handler。
func NewTemplateHandler(grpcClient *client.GRPCClient) *TemplateHandler {
	return &TemplateHandler{grpcClient: grpcClient}
}

// ListTemplates 模板列表（仅上架）。
// GET /templates?page=1&page_size=20
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
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

	pageVal := int32(page)
	pageSizeVal := int32(pageSize)
	online := pb.TemplateStatus_TEMPLATE_STATUS_ONLINE
	resp, err := h.grpcClient.Stub.ListTemplates(ctx, &pb.ListTemplatesRequest{
		Page:     &pageVal,
		PageSize: &pageSizeVal,
		Status:   &online,
	})
	if err != nil {
		respondGRPCError(c, err)
		return
	}

	if resp.Base != nil && resp.Base.Code != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.Base.Code, "message": resp.Base.Message})
		return
	}

	templates := make([]gin.H, 0, len(resp.Templates))
	for _, t := range resp.Templates {
		templates = append(templates, templateToJSON(t))
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"data":  gin.H{"templates": templates, "total": resp.Total},
	})
}

// GetTemplate 模板详情。
// GET /templates/:id
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid template id"})
		return
	}

	ctx, cancel := withTimeout(c)
	defer cancel()

	resp, err := h.grpcClient.Stub.GetTemplate(ctx, &pb.GetTemplateRequest{Id: id})
	if err != nil {
		respondGRPCError(c, err)
		return
	}
	if resp.Base != nil && resp.Base.Code != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.Base.Code, "message": resp.Base.Message})
		return
	}
	if resp.Template == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 3002, "message": "template not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": templateToJSON(resp.Template),
	})
}

// DownloadTemplate 下载模板（触发资格判断）。
// POST /templates/:id/download
func (h *TemplateHandler) DownloadTemplate(c *gin.Context) {
	idStr := c.Param("id")
	templateID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || templateID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid template id"})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "user not found in context"})
		return
	}

	ctx, cancel := withTimeout(c)
	defer cancel()

	resp, err := h.grpcClient.Stub.DownloadTemplate(ctx, &pb.DownloadTemplateRequest{
		UserId:     int64(userID),
		TemplateId: templateID,
	})
	if err != nil {
		respondGRPCError(c, err)
		return
	}

	if resp.Base != nil && resp.Base.Code != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.Base.Code, "message": resp.Base.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"order_no":       resp.OrderNo,
			"order_type":     orderTypeJSON(resp.OrderType),
			"order_status":   orderStatusJSON(resp.OrderStatus),
			"price_snapshot": resp.PriceSnapshot,
			"download_url":   resp.DownloadUrl,
			"payment_no":     resp.PaymentNo,
		},
	})
}

// templateToJSON 将 proto Template 转换为 JSON 响应。
func templateToJSON(t *pb.Template) gin.H {
	var createdAt, updatedAt string
	if t.CreatedAt != nil {
		createdAt = t.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00")
	}
	if t.UpdatedAt != nil {
		updatedAt = t.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00")
	}

	return gin.H{
		"id":                       t.Id,
		"template_no":              t.TemplateNo,
		"name":                     t.Name,
		"description":              t.Description,
		"is_free":                  isFreeJSON(t.IsFree),
		"price":                    t.Price,
		"status":                   templateStatusJSON(t.Status),
		"file_type":                t.FileType,
		"file_size":                t.FileSize,
		"thumbnail_download_url":   t.ThumbnailDownloadUrl,
		"created_at":               createdAt,
		"updated_at":               updatedAt,
	}
}