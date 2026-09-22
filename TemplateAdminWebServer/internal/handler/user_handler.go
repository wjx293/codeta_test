package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"template-mall/TemplateAdminWebServer/internal/client"

	pb "template-mall/TemplateOrderServer/api/proto"

	"github.com/gin-gonic/gin"
)

// UserHandler B 端用户 handler（检索用户 + 设置会员）。
type UserHandler struct {
	grpcClient *client.GRPCClient
}

// NewUserHandler 创建用户 handler。
func NewUserHandler(grpcClient *client.GRPCClient) *UserHandler {
	return &UserHandler{grpcClient: grpcClient}
}

// List GET /api/admin/users
// 模糊搜索 + 分页（对应 gRPC ListUsers）。
func (h *UserHandler) List(c *gin.Context) {
	username := c.Query("username")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	pageVal := int32(page)
	pageSizeVal := int32(pageSize)
	resp, err := h.grpcClient.Stub.ListUsers(ctx, &pb.ListUsersRequest{
		Page:     &pageVal,
		PageSize: &pageSizeVal,
		Username: username,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": err.Error()})
		return
	}
	if resp.Base != nil && resp.Base.Code != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.Base.Code, "message": resp.Base.Message})
		return
	}

	users := make([]gin.H, 0, len(resp.Users))
	for _, u := range resp.Users {
		var createdAt string
		if u.CreatedAt != nil {
			createdAt = u.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00")
		}
		users = append(users, gin.H{
			"id":            u.Id,
			"username":      u.Username,
			"member_status": memberStatusJSON(u.IsMember),
			"created_at":    createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"users": users, "total": resp.Total},
	})
}

// SetMemberRequest 设置会员请求。
type SetMemberRequest struct {
	MemberStatus int32 `json:"member_status" binding:"oneof=0 1"` // 0=非会员 1=会员
}

// SetMember PUT /api/admin/users/:id/member
func (h *UserHandler) SetMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid user id"})
		return
	}

	var req SetMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if req.MemberStatus != 0 && req.MemberStatus != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "member_status must be 0 or 1"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.grpcClient.Stub.SetMember(ctx, &pb.SetMemberRequest{
		UserId:   id,
		IsMember: req.MemberStatus == 1,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": err.Error()})
		return
	}
	if resp.Code != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.Code, "message": resp.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "member status updated"})
}