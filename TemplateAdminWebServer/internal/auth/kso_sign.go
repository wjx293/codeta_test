package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

const ksoContentType = "application/json"

// signKSO1 为 WPS OpenAPI 请求附加 KSO-1 签名头。
// 仅在开发者后台开启「接口签名」时需要，见官方文档。
func signKSO1(req *http.Request, accessKey, secretKey string, body []byte) error {
	if accessKey == "" || secretKey == "" {
		return fmt.Errorf("kso sign: missing access key or secret")
	}

	ksoDate := time.Now().UTC().Format(time.RFC1123)
	sha256Hex := ""
	if len(body) > 0 {
		sum := sha256.Sum256(body)
		sha256Hex = hex.EncodeToString(sum[:])
	}

	mac := hmac.New(sha256.New, []byte(secretKey))
	_, _ = mac.Write([]byte("KSO-1" + req.Method + req.URL.RequestURI() + ksoContentType + ksoDate + sha256Hex))
	signature := hex.EncodeToString(mac.Sum(nil))

	req.Header.Set("Content-Type", ksoContentType)
	req.Header.Set("X-Kso-Date", ksoDate)
	req.Header.Set("X-Kso-Authorization", fmt.Sprintf("KSO-1 %s:%s", accessKey, signature))
	return nil
}
