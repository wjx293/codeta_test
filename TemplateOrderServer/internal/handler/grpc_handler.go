package handler

import (
	"context"
	"errors"

	pb "template-mall/TemplateOrderServer/api/proto"
	"template-mall/TemplateOrderServer/internal/model"
	"template-mall/TemplateOrderServer/internal/repository"
	"template-mall/TemplateOrderServer/internal/service"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCHandler gRPC handler，实现 TemplateOrderServiceServer 接口。
type GRPCHandler struct {
	pb.UnimplementedTemplateOrderServiceServer
	userSvc     *service.UserService
	templateSvc *service.TemplateService
	orderSvc    *service.OrderService
	memberSvc   *service.MemberService
}

// NewGRPCHandler 创建 gRPC handler。
func NewGRPCHandler(
	userSvc *service.UserService,
	templateSvc *service.TemplateService,
	orderSvc *service.OrderService,
	memberSvc *service.MemberService,
) *GRPCHandler {
	return &GRPCHandler{
		userSvc:     userSvc,
		templateSvc: templateSvc,
		orderSvc:    orderSvc,
		memberSvc:   memberSvc,
	}
}

// ============================================================
// 用户与认证
// ============================================================

func (h *GRPCHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	resp, err := h.userSvc.Register(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUsernameExists) {
			return &pb.AuthResponse{Base: &pb.BaseResponse{Code: 1001, Message: err.Error()}}, nil
		}
		return nil, err
	}
	return &pb.AuthResponse{
		Base:         &pb.BaseResponse{Code: 0, Message: "ok"},
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		UserId:       int64(resp.UserID),
		Username:     resp.Username,
		IsMember:     isMemberFromDB(resp.MemberStatus),
	}, nil
}

func (h *GRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	resp, err := h.userSvc.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return &pb.AuthResponse{Base: &pb.BaseResponse{Code: 1002, Message: err.Error()}}, nil
		}
		return nil, err
	}
	return &pb.AuthResponse{
		Base:         &pb.BaseResponse{Code: 0, Message: "ok"},
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		UserId:       int64(resp.UserID),
		Username:     resp.Username,
		IsMember:     isMemberFromDB(resp.MemberStatus),
	}, nil
}

func (h *GRPCHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.AuthResponse, error) {
	resp, err := h.userSvc.RefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrTokenRevoked) || errors.Is(err, service.ErrTokenExpired) {
			return &pb.AuthResponse{Base: &pb.BaseResponse{Code: 1003, Message: err.Error()}}, nil
		}
		return nil, err
	}
	return &pb.AuthResponse{
		Base:         &pb.BaseResponse{Code: 0, Message: "ok"},
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		UserId:       int64(resp.UserID),
		Username:     resp.Username,
		IsMember:     isMemberFromDB(resp.MemberStatus),
	}, nil
}

func (h *GRPCHandler) RevokeRefreshTokens(ctx context.Context, req *pb.RevokeRefreshTokensRequest) (*pb.BaseResponse, error) {
	if err := h.userSvc.RevokeRefreshTokens(uint64(req.UserId)); err != nil {
		return nil, err
	}
	return &pb.BaseResponse{Code: 0, Message: "ok"}, nil
}

func (h *GRPCHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	page, pageSize := normalizeOptionalPage(req.Page, req.PageSize)
	offset := int((page - 1) * pageSize)

	users, total, err := h.userSvc.ListUsers(req.Username, offset, int(pageSize))
	if err != nil {
		return nil, err
	}

	pbUsers := make([]*pb.User, len(users))
	for i, u := range users {
		pbUsers[i] = &pb.User{
			Id:           int64(u.ID),
			Username:     u.Username,
			IsMember:     isMemberFromDB(u.MemberStatus),
			CreatedAt:    timestamppb.New(u.CreatedAt),
		}
	}

	return &pb.ListUsersResponse{
		Base:  &pb.BaseResponse{Code: 0, Message: "ok"},
		Users: pbUsers,
		Total: total,
	}, nil
}

// ============================================================
// 模板
// ============================================================

func (h *GRPCHandler) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.BaseResponse, error) {
	template := &model.TosTemplate{
		Name:              req.Name,
		Description:       req.Description,
		IsFree:            isFreeToDB(req.IsFree),
		Price:             req.Price,
		Status:            1, // 默认上架
		OSSType:           req.OssType,
		BucketName:        req.BucketName,
		OriginalFilename:  req.OriginalFilename,
		FileType:          req.FileType,
		FileSize:          uint64(req.FileSize),
		FileOssKey:        req.FileOssKey,
		ThumbnailOssKey:   req.ThumbnailOssKey,
		ThumbnailFileSize: uint64(req.ThumbnailFileSize),
	}

	if err := h.templateSvc.CreateTemplate(template); err != nil {
		if errors.Is(err, service.ErrPriceInvalid) {
			return &pb.BaseResponse{Code: 2001, Message: err.Error()}, nil
		}
		return nil, err
	}
	return &pb.BaseResponse{Code: 0, Message: "ok"}, nil
}

func (h *GRPCHandler) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.BaseResponse, error) {
	template := &model.TosTemplate{
		ID:                uint64(req.Id),
		Name:              req.Name,
		Description:       req.Description,
		IsFree:            isFreeToDB(req.IsFree),
		Price:             req.Price,
		ThumbnailOssKey:   req.ThumbnailOssKey,
		ThumbnailFileSize: uint64(req.ThumbnailFileSize),
	}
	if err := h.templateSvc.UpdateTemplate(template); err != nil {
		return nil, err
	}
	return &pb.BaseResponse{Code: 0, Message: "ok"}, nil
}

func (h *GRPCHandler) ListTemplates(ctx context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
	page, pageSize := normalizeOptionalPage(req.Page, req.PageSize)
	offset := int((page - 1) * pageSize)

	var status *int8
	if req.Status != nil {
		dbStatus, ok := templateStatusToDB(*req.Status)
		if ok {
			status = &dbStatus
		}
	}

	templates, total, err := h.templateSvc.ListTemplates(status, offset, int(pageSize))
	if err != nil {
		return nil, err
	}

	pbTemplates := make([]*pb.Template, len(templates))
	for i, t := range templates {
		pbTemplates[i] = &pb.Template{
			Id:                   int64(t.Template.ID),
			TemplateNo:           t.Template.TemplateNo,
			Name:                 t.Template.Name,
			Description:          t.Template.Description,
			IsFree:               isFreeFromDB(t.Template.IsFree),
			Price:                t.Template.Price,
			Status:               templateStatusFromDB(t.Template.Status),
			FileType:             t.Template.FileType,
			FileSize:             int64(t.Template.FileSize),
			ThumbnailDownloadUrl: t.ThumbnailDownloadURL,
			CreatedAt:            timestamppb.New(t.Template.CreatedAt),
			UpdatedAt:            timestamppb.New(t.Template.UpdatedAt),
		}
	}

	return &pb.ListTemplatesResponse{
		Base:      &pb.BaseResponse{Code: 0, Message: "ok"},
		Templates: pbTemplates,
		Total:     total,
	}, nil
}

func (h *GRPCHandler) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.GetTemplateResponse, error) {
	t, err := h.templateSvc.GetTemplate(uint64(req.Id))
	if err != nil {
		if errors.Is(err, service.ErrTemplateNotFound) {
			return &pb.GetTemplateResponse{Base: &pb.BaseResponse{Code: 3002, Message: err.Error()}}, nil
		}
		return nil, err
	}
	return &pb.GetTemplateResponse{
		Base: &pb.BaseResponse{Code: 0, Message: "ok"},
		Template: &pb.Template{
			Id:                   int64(t.Template.ID),
			TemplateNo:           t.Template.TemplateNo,
			Name:                 t.Template.Name,
			Description:          t.Template.Description,
			IsFree:               isFreeFromDB(t.Template.IsFree),
			Price:                t.Template.Price,
			Status:               templateStatusFromDB(t.Template.Status),
			FileType:             t.Template.FileType,
			FileSize:             int64(t.Template.FileSize),
			ThumbnailDownloadUrl: t.ThumbnailDownloadURL,
			CreatedAt:            timestamppb.New(t.Template.CreatedAt),
			UpdatedAt:            timestamppb.New(t.Template.UpdatedAt),
		},
	}, nil
}

// UpdateTemplateStatus 上架/下架模板（B 端）。
func (h *GRPCHandler) UpdateTemplateStatus(ctx context.Context, req *pb.UpdateTemplateStatusRequest) (*pb.BaseResponse, error) {
	dbStatus, ok := templateStatusToDB(req.Status)
	if !ok {
		return &pb.BaseResponse{Code: 2002, Message: "invalid template status"}, nil
	}
	if err := h.templateSvc.UpdateStatus(uint64(req.Id), dbStatus); err != nil {
		return &pb.BaseResponse{Code: 2003, Message: err.Error()}, nil
	}
	return &pb.BaseResponse{Code: 0, Message: "ok"}, nil
}

// ============================================================
// 下载与订单
// ============================================================

func (h *GRPCHandler) DownloadTemplate(ctx context.Context, req *pb.DownloadTemplateRequest) (*pb.DownloadResponse, error) {
	resp, err := h.orderSvc.DownloadTemplate(uint64(req.UserId), uint64(req.TemplateId))
	if err != nil {
		if errors.Is(err, service.ErrTemplateNotOnSale) {
			return &pb.DownloadResponse{Base: &pb.BaseResponse{Code: 3001, Message: err.Error()}}, nil
		}
		if errors.Is(err, service.ErrTemplateNotFound) {
			return &pb.DownloadResponse{Base: &pb.BaseResponse{Code: 3002, Message: err.Error()}}, nil
		}
		return &pb.DownloadResponse{Base: &pb.BaseResponse{Code: 3003, Message: err.Error()}}, nil
	}

	return &pb.DownloadResponse{
		Base:          &pb.BaseResponse{Code: 0, Message: "ok"},
		OrderNo:       resp.OrderNo,
		OrderType:     orderTypeFromDB(resp.OrderType),
		OrderStatus:   orderStatusFromDB(resp.OrderStatus),
		PriceSnapshot: resp.PriceSnapshot,
		DownloadUrl:   resp.DownloadURL,
		PaymentNo:     resp.PaymentNo,
	}, nil
}

func (h *GRPCHandler) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	page, pageSize := normalizeOptionalPage(req.Page, req.PageSize)
	offset := int((page - 1) * pageSize)

	params := repository.OrderListParams{
		Username: req.Username,
		Offset:   &offset,
		Limit:    func() *int { v := int(pageSize); return &v }(),
	}

	if req.UserId != nil && *req.UserId > 0 {
		uid := uint64(*req.UserId)
		params.UserID = &uid
	}
	if req.Status != nil && *req.Status != pb.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		s := int8(*req.Status)
		params.Status = &s
	}
	if req.StartTime != nil {
		t := req.StartTime.AsTime()
		params.StartTime = func() *interface{} { v := interface{}(t); return &v }()
	}
	if req.EndTime != nil {
		t := req.EndTime.AsTime()
		params.EndTime = func() *interface{} { v := interface{}(t); return &v }()
	}

	orders, total, err := h.orderSvc.ListOrders(params)
	if err != nil {
		return nil, err
	}

	pbOrders := make([]*pb.Order, len(orders))
	for i, o := range orders {
		pbOrders[i] = &pb.Order{
			Id:            int64(o.ID),
			OrderNo:       o.OrderNo,
			UserId:        int64(o.UserID),
			Username:      o.Username,
			TemplateId:    int64(o.TemplateID),
			TemplateName:  o.TemplateName,
			OrderType:     orderTypeFromDB(o.OrderType),
			PriceSnapshot: o.PriceSnapshot,
			Status:        orderStatusFromDB(o.Status),
			CreatedAt:     timestamppb.New(o.CreatedAt),
			UpdatedAt:     timestamppb.New(o.UpdatedAt),
		}
	}

	return &pb.ListOrdersResponse{
		Base:   &pb.BaseResponse{Code: 0, Message: "ok"},
		Orders: pbOrders,
		Total:  total,
	}, nil
}

func (h *GRPCHandler) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.BaseResponse, error) {
	if err := h.orderSvc.CancelOrder(uint64(req.UserId), req.OrderNo); err != nil {
		switch {
		case errors.Is(err, service.ErrOrderNotFound):
			return &pb.BaseResponse{Code: 3002, Message: err.Error()}, nil
		case errors.Is(err, service.ErrOrderForbidden):
			return &pb.BaseResponse{Code: 3004, Message: err.Error()}, nil
		case errors.Is(err, service.ErrOrderCannotCancel):
			return &pb.BaseResponse{Code: 3003, Message: err.Error()}, nil
		}
		return nil, err
	}
	return &pb.BaseResponse{Code: 0, Message: "ok"}, nil
}

// ============================================================
// 会员
// ============================================================

func (h *GRPCHandler) SetMember(ctx context.Context, req *pb.SetMemberRequest) (*pb.BaseResponse, error) {
	if err := h.memberSvc.SetMember(uint64(req.UserId), isMemberToDB(req.IsMember)); err != nil {
		return &pb.BaseResponse{Code: 4001, Message: err.Error()}, nil
	}
	return &pb.BaseResponse{Code: 0, Message: "ok"}, nil
}

// ============================================================
// helpers
// ============================================================

func normalizeOptionalPage(page, pageSize *int32) (int32, int32) {
	p := int32(1)
	if page != nil && *page >= 1 {
		p = *page
	}
	ps := int32(20)
	if pageSize != nil && *pageSize >= 1 && *pageSize <= 100 {
		ps = *pageSize
	}
	return p, ps
}
