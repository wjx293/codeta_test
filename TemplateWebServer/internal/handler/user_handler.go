package handler

import (
	"net/http"

	"template-mall/TemplateWebServer/internal/client"
	"template-mall/TemplateWebServer/internal/middleware"

	pb "template-mall/TemplateOrderServer/api/proto"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户 handler
type UserHandler struct {
	grpcClient *client.GRPCClient
}

// NewUserHandler 创建用户 handler。
func NewUserHandler(grpcClient *client.GRPCClient) *UserHandler {
	return &UserHandler{grpcClient: grpcClient}
}

// GetCurrentUser 当前用户信息（从数据库读取，保证会员状态与管理员设置一致）。
// GET /users/me
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "user not found in context"})
		return
	}

	username, _ := middleware.GetUsername(c)

	ctx, cancel := withTimeout(c)
	defer cancel()

	pageVal := int32(1)
	pageSizeVal := int32(20)
	resp, err := h.grpcClient.Stub.ListUsers(ctx, &pb.ListUsersRequest{
		Username: username,
		Page:     &pageVal,
		PageSize: &pageSizeVal,
	})
	if err != nil {
		respondGRPCError(c, err)
		return
	}

	for _, u := range resp.Users {
		if u.Id == int64(userID) {
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"data": gin.H{
					"user_id":       u.Id,
					"username":      u.Username,
					"member_status": memberStatusJSON(u.IsMember),
				},
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "user not found"})
}

// Health 健康检查。
// GET /health
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "TemplateWebServer",
		"version": "1.0.0",
	})
}
