package repository

import (
	"fmt"
	"strings"
	"time"

	"template-mall/TemplateOrderServer/internal/model"

	"gorm.io/gorm"
)

// OrderRepo 订单仓储
type OrderRepo struct{ db *gorm.DB }

func NewOrderRepo(db *gorm.DB) *OrderRepo { return &OrderRepo{db: db} }

// CreateFreeOrder 创建免费订单（month 非空，order_type=1）。
// 捕获 uk_free_month 冲突（user_id, template_id, month, order_type）。
func (r *OrderRepo) CreateFreeOrder(order *model.TosOrder) error {
	err := r.db.Create(order).Error
	if err != nil && isDuplicate(err) {
		return ErrDuplicate
	}
	return err
}

// CreateMemberOrder 创建会员订单（month 非空，order_type=2）。
// 捕获 uk_member_month 冲突（user_id, template_id, month, order_type）。
func (r *OrderRepo) CreateMemberOrder(order *model.TosOrder) error {
	err := r.db.Create(order).Error
	if err != nil && isDuplicate(err) {
		return ErrDuplicate
	}
	return err
}

// CreateRetailOrder 创建零售订单（month=NULL，order_type=3）。
// 捕获 uk_active_unpaid 冲突（生成列 unique index）。
// 因 month 为 NULL，MySQL 唯一索引对 NULL 不冲突，故不会触发 uk_free_month/uk_member_month。
func (r *OrderRepo) CreateRetailOrder(order *model.TosOrder) error {
	err := r.db.Create(order).Error
	if err != nil && isDuplicate(err) {
		return ErrDuplicate
	}
	return err
}

// FindActiveUnpaidRetail 查找同一用户×模板的未支付零售订单。
func (r *OrderRepo) FindActiveUnpaidRetail(userID, templateID uint64) (*model.TosOrder, error) {
	var order model.TosOrder
	err := r.db.Where("user_id = ? AND template_id = ? AND order_type = 3 AND status = 1",
		userID, templateID).First(&order).Error
	if err != nil {
		return nil, ErrNotFound
	}
	return &order, nil
}

// FindRetailPaid 查找同一用户×模板的已支付零售订单。
func (r *OrderRepo) FindRetailPaid(userID, templateID uint64) (*model.TosOrder, error) {
	var order model.TosOrder
	err := r.db.Where("user_id = ? AND template_id = ? AND order_type = 3 AND status = 2",
		userID, templateID).First(&order).Error
	if err != nil {
		return nil, ErrNotFound
	}
	return &order, nil
}

// MaxOrderNoToday 返回当日最大订单编号（无记录时返回空串）。
func (r *OrderRepo) MaxOrderNoToday() (string, error) {
	today := time.Now().Format("20060102")
	prefix := "TO" + today + "%"
	var maxNo *string
	err := r.db.Model(&model.TosOrder{}).
		Where("order_no LIKE ?", prefix).
		Select("MAX(order_no)").
		Scan(&maxNo).Error
	if err != nil {
		return "", err
	}
	if maxNo == nil {
		return "", nil
	}
	return *maxNo, nil
}

// FindByOrderNo 按订单编号查找。
func (r *OrderRepo) FindByOrderNo(orderNo string) (*model.TosOrder, error) {
	var order model.TosOrder
	err := r.db.Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, ErrNotFound
	}
	return &order, nil
}

// TouchUpdatedAt 更新订单修改时间（重复下载时刷新）。
func (r *OrderRepo) TouchUpdatedAt(orderID uint64) error {
	return r.db.Model(&model.TosOrder{}).Where("id = ?", orderID).
		Update("updated_at", time.Now()).Error
}

// FindByUserAndTemplate 查找同一用户×模板的免费/会员订单（用于回查）。
func (r *OrderRepo) FindByUserAndTemplate(userID, templateID uint64, orderType int8, month string) (*model.TosOrder, error) {
	var order model.TosOrder
	err := r.db.Where("user_id = ? AND template_id = ? AND order_type = ? AND month = ?",
		userID, templateID, orderType, month).First(&order).Error
	if err != nil {
		return nil, ErrNotFound
	}
	return &order, nil
}

// UpdateStatusConditionally 条件更新订单状态，返回影响行数。
func (r *OrderRepo) UpdateStatusConditionally(orderNo string, fromStatus, toStatus int8) (int64, error) {
	result := r.db.Model(&model.TosOrder{}).
		Where("order_no = ? AND status = ?", orderNo, fromStatus).
		Update("status", toStatus)
	return result.RowsAffected, result.Error
}

// CancelByUser 取消指定用户的未支付订单（status 1→3），返回影响行数。
func (r *OrderRepo) CancelByUser(orderNo string, userID uint64) (int64, error) {
	result := r.db.Model(&model.TosOrder{}).
		Where("order_no = ? AND user_id = ? AND status = ?", orderNo, userID, 1).
		Update("status", 3)
	return result.RowsAffected, result.Error
}

// ListOrders 分页查询订单，JOIN 用户和模板表，支持筛选。
func (r *OrderRepo) ListOrders(params OrderListParams) ([]model.OrderView, int64, error) {
	var orders []model.OrderView
	var total int64

	query := r.db.Table("tos_orders o").
		Select(`o.id, o.order_no, o.user_id, u.username, o.template_id, t.name as template_name,
			o.order_type, o.price_snapshot, o.status, o.month, o.created_at, o.updated_at`).
		Joins("JOIN tos_users u ON o.user_id = u.id").
		Joins("JOIN tos_templates t ON o.template_id = t.id")

	// 筛选条件
	conditions := make([]string, 0)
	args := make([]interface{}, 0)

	if params.UserID != nil {
		conditions = append(conditions, "o.user_id = ?")
		args = append(args, *params.UserID)
	}
	if params.Status != nil {
		conditions = append(conditions, "o.status = ?")
		args = append(args, *params.Status)
	}
	if params.Username != "" {
		conditions = append(conditions, "u.username LIKE ?")
		args = append(args, "%"+params.Username+"%")
	}
	if params.StartTime != nil {
		conditions = append(conditions, "o.created_at >= ?")
		args = append(args, *params.StartTime)
	}
	if params.EndTime != nil {
		conditions = append(conditions, "o.created_at <= ?")
		args = append(args, *params.EndTime)
	}

	if len(conditions) > 0 {
		query = query.Where(strings.Join(conditions, " AND "), args...)
	}

	// 计数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页 + 排序
	offset := 0
	limit := 20
	if params.Offset != nil {
		offset = *params.Offset
	}
	if params.Limit != nil {
		limit = *params.Limit
	}

	if err := query.Offset(offset).Limit(limit).Order("o.created_at DESC").Scan(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// OrderListParams 订单列表查询参数
type OrderListParams struct {
	UserID    *uint64
	Status    *int8
	Username  string
	StartTime *interface{}
	EndTime   *interface{}
	Offset    *int
	Limit     *int
}

// FormatMonth 将 time.Time 格式化为 yyyyMM。
func FormatMonth(t interface{}) string {
	switch v := t.(type) {
	case fmt.Stringer:
		return v.String()
	default:
		return ""
	}
}
