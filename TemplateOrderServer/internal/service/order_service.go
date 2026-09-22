package service

import (
	"errors"
	"time"

	"template-mall/TemplateOrderServer/internal/idgen"
	"template-mall/TemplateOrderServer/internal/model"
	"template-mall/TemplateOrderServer/internal/oss"
	"template-mall/TemplateOrderServer/internal/payclient"
	"template-mall/TemplateOrderServer/internal/repository"
)

var (
	ErrTemplateNotOnSale = errors.New("template is not on sale")
	ErrOrderNotFound     = errors.New("order not found")
	ErrOrderForbidden    = errors.New("order access denied")
	ErrOrderCannotCancel = errors.New("order cannot be canceled")
	ErrPaymentFailed     = errors.New("payment creation failed")
)

// OrderService 订单服务
type OrderService struct {
	orderRepo    *repository.OrderRepo
	templateRepo *repository.TemplateRepo
	userRepo     *repository.UserRepo
	presigner    *oss.Presigner
	payClient    *payclient.Client
	idGen        *idgen.Generator
}

// NewOrderService 创建订单服务实例。
func NewOrderService(
	orderRepo *repository.OrderRepo,
	templateRepo *repository.TemplateRepo,
	userRepo *repository.UserRepo,
	presigner *oss.Presigner,
	payClient *payclient.Client,
	idGen *idgen.Generator,
) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		templateRepo: templateRepo,
		userRepo:     userRepo,
		presigner:    presigner,
		payClient:    payClient,
		idGen:        idGen,
	}
}

// DownloadResponse 下载/购买响应
type DownloadResponse struct {
	OrderNo       string
	OrderID       uint64
	OrderType     int8 // 1=免费 2=会员 3=零售
	OrderStatus   int8 // 1=未支付 2=已获得下载资格
	PriceSnapshot int64
	DownloadURL   string // 仅 order_status=2 时填充
	PaymentNo     string // 仅零售订单填充
	IsNewOrder    bool   // 是否新建订单
}

// DownloadTemplate 下载资格判断核心逻辑。
// 严格按顺序判断：未上架 → 免费 → 已零售购买 → 会员 → 零售。
func (s *OrderService) DownloadTemplate(userID, templateID uint64) (*DownloadResponse, error) {
	// 1. 查模板：未上架 → 拒绝
	template, err := s.templateRepo.FindByID(templateID)
	if err != nil {
		return nil, ErrTemplateNotFound
	}
	if template.Status != 1 {
		return nil, ErrTemplateNotOnSale
	}

	// 2. 免费模板 → 直接创建免费订单
	if template.IsFree == 1 {
		return s.createFreeOrder(userID, template)
	}

	// 3. 已零售购买 → 返回原订单 + 下载地址
	if paidOrder, err := s.orderRepo.FindRetailPaid(userID, templateID); err == nil {
		url, _ := s.presigner.PresignDownload(template.FileOssKey, 5*time.Minute)
		return &DownloadResponse{
			OrderNo:       paidOrder.OrderNo,
			OrderID:       paidOrder.ID,
			OrderType:     paidOrder.OrderType,
			OrderStatus:   paidOrder.Status,
			PriceSnapshot: paidOrder.PriceSnapshot,
			DownloadURL:   url,
			IsNewOrder:    false,
		}, nil
	}

	// 4. 会员 → 创建会员订单
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if user.MemberStatus == 1 {
		return s.createMemberOrder(userID, template)
	}

	// 5. 非会员 → 创建零售订单（未支付）
	return s.createRetailOrder(userID, template)
}

// createFreeOrder 创建免费订单（order_type=1, status=2, month=yyyyMM, price=0）。
func (s *OrderService) createFreeOrder(userID uint64, template *model.TosTemplate) (*DownloadResponse, error) {
	month := formatMonth(time.Now())

	// 同月已有免费订单 → 不新建，返回原订单 + 新下载地址
	if existing, err := s.orderRepo.FindByUserAndTemplate(userID, template.ID, 1, month); err == nil {
		s.orderRepo.TouchUpdatedAt(existing.ID)
		url, _ := s.presigner.PresignDownload(template.FileOssKey, 5*time.Minute)
		return &DownloadResponse{
			OrderNo:       existing.OrderNo,
			OrderID:       existing.ID,
			OrderType:     1,
			OrderStatus:   2,
			PriceSnapshot: 0,
			DownloadURL:   url,
			IsNewOrder:    false,
		}, nil
	}

	orderNo := s.idGen.GenerateOrderNo()
	now := time.Now()

	order := &model.TosOrder{
		OrderNo:       orderNo,
		UserID:        userID,
		TemplateID:    template.ID,
		OrderType:     1,
		PriceSnapshot: 0,
		Status:        2, // 直接获得下载资格
		Month:         &month,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.orderRepo.CreateFreeOrder(order); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			// 回查已有订单
			existing, findErr := s.orderRepo.FindByUserAndTemplate(userID, template.ID, 1, month)
			if findErr != nil {
				return nil, findErr
			}
			url, _ := s.presigner.PresignDownload(template.FileOssKey, 5*time.Minute)
			return &DownloadResponse{
				OrderNo:       existing.OrderNo,
				OrderID:       existing.ID,
				OrderType:     1,
				OrderStatus:   2,
				PriceSnapshot: 0,
				DownloadURL:   url,
				IsNewOrder:    false,
			}, nil
		}
		return nil, err
	}

	url, _ := s.presigner.PresignDownload(template.FileOssKey, 5*time.Minute)
	return &DownloadResponse{
		OrderNo:       orderNo,
		OrderID:       order.ID,
		OrderType:     1,
		OrderStatus:   2,
		PriceSnapshot: 0,
		DownloadURL:   url,
		IsNewOrder:    true,
	}, nil
}

// createMemberOrder 创建会员订单（order_type=2, status=2, month=yyyyMM, price=0）。
func (s *OrderService) createMemberOrder(userID uint64, template *model.TosTemplate) (*DownloadResponse, error) {
	month := formatMonth(time.Now())

	// 同月已有会员订单 → 不新建，返回原订单 + 新下载地址
	if existing, err := s.orderRepo.FindByUserAndTemplate(userID, template.ID, 2, month); err == nil {
		s.orderRepo.TouchUpdatedAt(existing.ID)
		url, _ := s.presigner.PresignDownload(template.FileOssKey, 5*time.Minute)
		return &DownloadResponse{
			OrderNo:       existing.OrderNo,
			OrderID:       existing.ID,
			OrderType:     2,
			OrderStatus:   2,
			PriceSnapshot: 0,
			DownloadURL:   url,
			IsNewOrder:    false,
		}, nil
	}

	orderNo := s.idGen.GenerateOrderNo()
	now := time.Now()

	order := &model.TosOrder{
		OrderNo:       orderNo,
		UserID:        userID,
		TemplateID:    template.ID,
		OrderType:     2,
		PriceSnapshot: 0,
		Status:        2,
		Month:         &month,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.orderRepo.CreateMemberOrder(order); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			existing, findErr := s.orderRepo.FindByUserAndTemplate(userID, template.ID, 2, month)
			if findErr != nil {
				return nil, findErr
			}
			url, _ := s.presigner.PresignDownload(template.FileOssKey, 5*time.Minute)
			return &DownloadResponse{
				OrderNo:       existing.OrderNo,
				OrderID:       existing.ID,
				OrderType:     2,
				OrderStatus:   2,
				PriceSnapshot: 0,
				DownloadURL:   url,
				IsNewOrder:    false,
			}, nil
		}
		return nil, err
	}

	url, _ := s.presigner.PresignDownload(template.FileOssKey, 5*time.Minute)
	return &DownloadResponse{
		OrderNo:       orderNo,
		OrderID:       order.ID,
		OrderType:     2,
		OrderStatus:   2,
		PriceSnapshot: 0,
		DownloadURL:   url,
		IsNewOrder:    true,
	}, nil
}

// createRetailOrder 创建零售订单（order_type=3, status=1, month=NULL, price_snapshot=template.price）。
func (s *OrderService) createRetailOrder(userID uint64, template *model.TosTemplate) (*DownloadResponse, error) {
	orderNo := s.idGen.GenerateOrderNo()
	now := time.Now()

	order := &model.TosOrder{
		OrderNo:       orderNo,
		UserID:        userID,
		TemplateID:    template.ID,
		OrderType:     3,
		PriceSnapshot: template.Price,
		Status:        1,   // 未支付
		Month:         nil, // 零售订单 month 为 NULL
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.orderRepo.CreateRetailOrder(order); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			// 回查已有未支付订单
			existing, findErr := s.orderRepo.FindActiveUnpaidRetail(userID, template.ID)
			if findErr != nil {
				return nil, findErr
			}
			// 调 payclient 创建/获取支付单
			payResp, payErr := s.payClient.CreatePayment(existing.OrderNo, existing.PriceSnapshot)
			if payErr != nil {
				return nil, payErr
			}
			return &DownloadResponse{
				OrderNo:       existing.OrderNo,
				OrderID:       existing.ID,
				OrderType:     3,
				OrderStatus:   1,
				PriceSnapshot: existing.PriceSnapshot,
				PaymentNo:     payResp.PaymentNo,
				IsNewOrder:    false,
			}, nil
		}
		return nil, err
	}

	// 调 payclient 创建支付单
	payResp, err := s.payClient.CreatePayment(orderNo, template.Price)
	if err != nil {
		return nil, err
	}

	return &DownloadResponse{
		OrderNo:       orderNo,
		OrderID:       order.ID,
		OrderType:     3,
		OrderStatus:   1,
		PriceSnapshot: template.Price,
		PaymentNo:     payResp.PaymentNo,
		IsNewOrder:    true,
	}, nil
}

// ListOrders 分页查询订单，含 username 与 template_name。
func (s *OrderService) ListOrders(params repository.OrderListParams) ([]model.OrderView, int64, error) {
	return s.orderRepo.ListOrders(params)
}

// CancelOrder 取消订单（仅本人、status=1 可取消）。
func (s *OrderService) CancelOrder(userID uint64, orderNo string) error {
	rows, err := s.orderRepo.CancelByUser(orderNo, userID)
	if err != nil {
		return err
	}
	if rows > 0 {
		return nil
	}

	order, findErr := s.orderRepo.FindByOrderNo(orderNo)
	if findErr != nil {
		return ErrOrderNotFound
	}
	if order.UserID != userID {
		return ErrOrderForbidden
	}
	return ErrOrderCannotCancel
}

// formatMonth 返回 Asia/Shanghai 时区的 yyyyMM。
func formatMonth(t time.Time) string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return t.In(loc).Format("200601")
}
