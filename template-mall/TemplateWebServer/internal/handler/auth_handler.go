package handler

import (
	"net/http"

	"template-mall/TemplateWebServer/internal/client"
	"template-mall/TemplateWebServer/internal/middleware"

	pb "template-mall/TemplateOrderServer/api/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthHandler 认证 handler
type AuthHandler struct {
	grpcClient *client.GRPCClient
}

// NewAuthHandler 创建认证 handler。
func NewAuthHandler(grpcClient *client.GRPCClient) *AuthHandler {
	return &AuthHandler{grpcClient: grpcClient}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// Register 用户注册。
// POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx, cancel := withTimeout(c)
	defer cancel()

	resp, err := h.grpcClient.Stub.Register(ctx, &pb.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
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
			"access_token":  resp.AccessToken,
			"refresh_token": resp.RefreshToken,
			"user_id":       resp.UserId,
			"username":      resp.Username,
			"member_status": memberStatusJSON(resp.IsMember),
		},
	})
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 用户登录。
// POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx, cancel := withTimeout(c)
	defer cancel()

	resp, err := h.grpcClient.Stub.Login(ctx, &pb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		respondGRPCError(c, err)
		return
	}

	if resp.Base != nil && resp.Base.Code != 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": resp.Base.Code, "message": resp.Base.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"access_token":  resp.AccessToken,
			"refresh_token": resp.RefreshToken,
			"user_id":       resp.UserId,
			"username":      resp.Username,
			"member_status": memberStatusJSON(resp.IsMember),
		},
	})
}

// RefreshRequest 刷新 token 请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh 刷新 access token。
// POST /auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx, cancel := withTimeout(c)
	defer cancel()

	resp, err := h.grpcClient.Stub.RefreshToken(ctx, &pb.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		respondGRPCError(c, err)
		return
	}

	if resp.Base != nil && resp.Base.Code != 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": resp.Base.Code, "message": resp.Base.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"access_token":  resp.AccessToken,
			"refresh_token": resp.RefreshToken,
			"user_id":       resp.UserId,
			"username":      resp.Username,
			"member_status": memberStatusJSON(resp.IsMember),
		},
	})
}

// Logout 用户登出，吊销该用户所有 refresh token。
// POST /auth/logout （需 JWT 鉴权）
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "user not found in context"})
		return
	}

	ctx, cancel := withTimeout(c)
	defer cancel()

	resp, err := h.grpcClient.Stub.RevokeRefreshTokens(ctx, &pb.RevokeRefreshTokensRequest{
		UserId: int64(userID),
	})
	if err != nil {
		respondGRPCError(c, err)
		return
	}

	if resp.Code != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"code": resp.Code, "message": resp.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "logout success"})
}

// respondGRPCError 统一处理 gRPC 错误。
func respondGRPCError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	switch st.Code() {
	case codes.Unauthenticated:
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": st.Message()})
	case codes.NotFound:
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": st.Message()})
	case codes.AlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": st.Message()})
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": st.Message()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": st.Message()})
	}
}
