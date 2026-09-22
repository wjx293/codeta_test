// Package payclient 提供调 PayWebServer 创建支付单的 HTTP 客户端。
package payclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client PayWebServer HTTP 客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient 创建支付客户端。
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CreatePaymentRequest 创建支付单请求
type CreatePaymentRequest struct {
	BusinessOrderNo string `json:"business_order_no"`
	Amount          int64  `json:"amount"`
}

// CreatePaymentResponse 创建支付单响应
type CreatePaymentResponse struct {
	PaymentNo       string `json:"payment_no"`
	BusinessOrderNo string `json:"business_order_no"`
	Amount          int64  `json:"amount"`
	Status          int8   `json:"status"`
}

type apiResponse struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    *CreatePaymentResponse `json:"data"`
}

// CreatePayment 创建支付单，幂等：相同 business_order_no 返回相同 payment_no。
func (c *Client) CreatePayment(businessOrderNo string, amount int64) (*CreatePaymentResponse, error) {
	reqBody := CreatePaymentRequest{
		BusinessOrderNo: businessOrderNo,
		Amount:          amount,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/payments",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payweb server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var envelope apiResponse
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if envelope.Code != 0 {
		return nil, fmt.Errorf("payweb server error %d: %s", envelope.Code, envelope.Message)
	}
	if envelope.Data == nil || envelope.Data.PaymentNo == "" {
		return nil, fmt.Errorf("payweb server returned empty payment_no")
	}

	return envelope.Data, nil
}
