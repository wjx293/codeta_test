package payclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreatePayment_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/payments" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var req CreatePaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.BusinessOrderNo != "TO202608090000000001" || req.Amount != 9900 {
			t.Fatalf("unexpected request body: %+v", req)
		}

		_ = json.NewEncoder(w).Encode(apiResponse{
			Code: 0,
			Data: &CreatePaymentResponse{
				PaymentNo:       "PW202608090000000001",
				BusinessOrderNo: req.BusinessOrderNo,
				Amount:          req.Amount,
				Status:          1,
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	resp, err := client.CreatePayment("TO202608090000000001", 9900)
	if err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}
	if resp.PaymentNo != "PW202608090000000001" {
		t.Fatalf("expected payment_no, got %q", resp.PaymentNo)
	}
}

func TestCreatePayment_BusinessError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(apiResponse{Code: 400, Message: "invalid amount"})
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL).CreatePayment("TO1", 0)
	if err == nil {
		t.Fatal("expected error")
	}
}
