package service

import (
	"testing"

	"template-mall/PayWebServer/internal/model"
	"template-mall/PayWebServer/internal/repository"
)

// ==================== Mock 实现 ====================

// mockPaymentRepo 模拟支付仓储
type mockPaymentRepo struct {
	payments map[string]*model.PayPayment
}

func newMockPaymentRepo() *mockPaymentRepo {
	return &mockPaymentRepo{payments: make(map[string]*model.PayPayment)}
}

func (m *mockPaymentRepo) CreatePayment(p *model.PayPayment) error {
	if _, ok := m.payments[p.BusinessOrderNo]; ok {
		return repository.ErrDuplicate
	}
	m.payments[p.BusinessOrderNo] = p
	m.payments[p.PaymentNo] = p
	return nil
}

func (m *mockPaymentRepo) FindByBusinessOrderNo(orderNo string) (*model.PayPayment, error) {
	p, ok := m.payments[orderNo]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return p, nil
}

func (m *mockPaymentRepo) FindByPaymentNo(paymentNo string) (*model.PayPayment, error) {
	p, ok := m.payments[paymentNo]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return p, nil
}

func (m *mockPaymentRepo) UpdateStatusConditionally(paymentNo string, fromStatus, toStatus int8) (int64, error) {
	p, ok := m.payments[paymentNo]
	if !ok {
		return 0, repository.ErrNotFound
	}
	if p.Status == fromStatus {
		p.Status = toStatus
		return 1, nil
	}
	return 0, nil
}

// mockCallbackRepo 模拟回调仓储
type mockCallbackRepo struct {
	callbacks map[string]*model.PayCallback
}

func newMockCallbackRepo() *mockCallbackRepo {
	return &mockCallbackRepo{callbacks: make(map[string]*model.PayCallback)}
}

func (m *mockCallbackRepo) InsertCallback(c *model.PayCallback) error {
	if _, ok := m.callbacks[c.CallbackNo]; ok {
		return repository.ErrDuplicate
	}
	m.callbacks[c.CallbackNo] = c
	return nil
}

// mockOutboxRepo 模拟 Outbox 仓储
type mockOutboxRepo struct {
	records map[string]*model.PayOutbox
}

func newMockOutboxRepo() *mockOutboxRepo {
	return &mockOutboxRepo{records: make(map[string]*model.PayOutbox)}
}

func (m *mockOutboxRepo) Insert(o *model.PayOutbox) error {
	if _, ok := m.records[o.EventKey]; ok {
		return repository.ErrDuplicate
	}
	m.records[o.EventKey] = o
	return nil
}

func (m *mockOutboxRepo) FindPending(limit int) ([]model.PayOutbox, error) {
	var pending []model.PayOutbox
	for _, r := range m.records {
		if r.Status == 0 {
			pending = append(pending, *r)
		}
	}
	return pending, nil
}

func (m *mockOutboxRepo) MarkSent(eventKey string) error {
	if r, ok := m.records[eventKey]; ok {
		r.Status = 1
	}
	return nil
}

func newTestPaymentService(pmRepo *mockPaymentRepo, cbRepo *mockCallbackRepo, outboxRepo *mockOutboxRepo) *PaymentService {
	return NewPaymentService(pmRepo, cbRepo, outboxRepo, nil, nil)
}

// ==================== 测试用例 ====================

func TestCreatePayment_New(t *testing.T) {
	pmRepo := newMockPaymentRepo()
	cbRepo := newMockCallbackRepo()
	outboxRepo := newMockOutboxRepo()
	svc := newTestPaymentService(pmRepo, cbRepo, outboxRepo)

	resp, err := svc.CreatePayment("ORD_001", 9900, "PW_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PaymentNo != "PW_001" {
		t.Fatalf("expected PW_001, got %s", resp.PaymentNo)
	}
	if resp.Status != 1 {
		t.Fatalf("expected status 1, got %d", resp.Status)
	}
}

func TestCreatePayment_Idempotent(t *testing.T) {
	pmRepo := newMockPaymentRepo()
	_ = pmRepo.CreatePayment(&model.PayPayment{
		PaymentNo:       "PW_001",
		BusinessOrderNo: "ORD_001",
		Amount:          9900,
		Status:          1,
	})

	cbRepo := newMockCallbackRepo()
	outboxRepo := newMockOutboxRepo()
	svc := newTestPaymentService(pmRepo, cbRepo, outboxRepo)

	resp, err := svc.CreatePayment("ORD_001", 9900, "PW_002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PaymentNo != "PW_001" {
		t.Fatalf("expected PW_001, got %s", resp.PaymentNo)
	}
}

func TestHandleCallback_Success(t *testing.T) {
	pmRepo := newMockPaymentRepo()
	_ = pmRepo.CreatePayment(&model.PayPayment{
		PaymentNo:       "PW_001",
		BusinessOrderNo: "ORD_001",
		Amount:          9900,
		Status:          1,
	})

	cbRepo := newMockCallbackRepo()
	outboxRepo := newMockOutboxRepo()
	svc := newTestPaymentService(pmRepo, cbRepo, outboxRepo)

	resp, err := svc.HandleCallback(t.Context(), "CB_001", "PW_001", "success", 9900)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsNew {
		t.Fatal("expected IsNew=true for first callback")
	}

	p, _ := pmRepo.FindByPaymentNo("PW_001")
	if p.Status != 2 {
		t.Fatalf("expected status 2, got %d", p.Status)
	}

	if _, ok := outboxRepo.records["CB_001"]; !ok {
		t.Fatal("expected outbox record for success callback")
	}
}

func TestHandleCallback_Fail(t *testing.T) {
	pmRepo := newMockPaymentRepo()
	_ = pmRepo.CreatePayment(&model.PayPayment{
		PaymentNo:       "PW_001",
		BusinessOrderNo: "ORD_001",
		Amount:          9900,
		Status:          1,
	})

	cbRepo := newMockCallbackRepo()
	outboxRepo := newMockOutboxRepo()
	svc := newTestPaymentService(pmRepo, cbRepo, outboxRepo)

	resp, err := svc.HandleCallback(t.Context(), "CB_FAIL", "PW_001", "fail", 9900)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsNew {
		t.Fatal("expected IsNew=true for first fail callback")
	}

	p, _ := pmRepo.FindByPaymentNo("PW_001")
	if p.Status != 1 {
		t.Fatalf("expected status 1 for fail callback, got %d", p.Status)
	}

	if len(outboxRepo.records) != 0 {
		t.Fatal("expected no outbox record for fail callback")
	}
}

func TestHandleCallback_Duplicate(t *testing.T) {
	pmRepo := newMockPaymentRepo()
	_ = pmRepo.CreatePayment(&model.PayPayment{
		PaymentNo:       "PW_001",
		BusinessOrderNo: "ORD_001",
		Amount:          9900,
		Status:          1,
	})

	cbRepo := newMockCallbackRepo()
	outboxRepo := newMockOutboxRepo()
	svc := newTestPaymentService(pmRepo, cbRepo, outboxRepo)

	resp1, _ := svc.HandleCallback(t.Context(), "CB_001", "PW_001", "success", 9900)
	if !resp1.IsNew {
		t.Fatal("expected IsNew=true for first callback")
	}

	resp2, err := svc.HandleCallback(t.Context(), "CB_001", "PW_001", "success", 9900)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp2.IsNew {
		t.Fatal("expected IsNew=false for duplicate callback")
	}
}

func TestHandleCallback_AmountMismatch(t *testing.T) {
	pmRepo := newMockPaymentRepo()
	_ = pmRepo.CreatePayment(&model.PayPayment{
		PaymentNo:       "PW_001",
		BusinessOrderNo: "ORD_001",
		Amount:          9900,
		Status:          1,
	})

	cbRepo := newMockCallbackRepo()
	outboxRepo := newMockOutboxRepo()
	svc := newTestPaymentService(pmRepo, cbRepo, outboxRepo)

	_, err := svc.HandleCallback(t.Context(), "CB_001", "PW_001", "success", 10000)
	if err != ErrAmountMismatch {
		t.Fatalf("expected ErrAmountMismatch, got %v", err)
	}
}

func TestHandleCallback_PaymentNotFound(t *testing.T) {
	pmRepo := newMockPaymentRepo()
	cbRepo := newMockCallbackRepo()
	outboxRepo := newMockOutboxRepo()
	svc := newTestPaymentService(pmRepo, cbRepo, outboxRepo)

	_, err := svc.HandleCallback(t.Context(), "CB_001", "PW_NONEXIST", "success", 9900)
	if err != ErrPaymentNotFound {
		t.Fatalf("expected ErrPaymentNotFound, got %v", err)
	}
}

func TestHandleCallback_StatusInvalid(t *testing.T) {
	pmRepo := newMockPaymentRepo()
	_ = pmRepo.CreatePayment(&model.PayPayment{
		PaymentNo:       "PW_001",
		BusinessOrderNo: "ORD_001",
		Amount:          9900,
		Status:          2, // 已支付
	})

	cbRepo := newMockCallbackRepo()
	outboxRepo := newMockOutboxRepo()
	svc := newTestPaymentService(pmRepo, cbRepo, outboxRepo)

	_, err := svc.HandleCallback(t.Context(), "CB_001", "PW_001", "success", 9900)
	if err != ErrStatusInvalid {
		t.Fatalf("expected ErrStatusInvalid, got %v", err)
	}
}
