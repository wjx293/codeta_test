package auth

import (
	"encoding/json"
	"testing"
)

func TestParseWPSUserInfoResponse(t *testing.T) {
	body := []byte(`{
		"code": 0,
		"msg": "ok",
		"data": {
			"id": "wps-user-1",
			"user_name": "张三",
			"avatar": "https://example.com/a.png",
			"company_id": "c1"
		}
	}`)

	var raw struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			ID       string `json:"id"`
			UserName string `json:"user_name"`
			Avatar   string `json:"avatar"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatal(err)
	}
	if raw.Data.ID != "wps-user-1" || raw.Data.UserName != "张三" {
		t.Fatalf("unexpected parse: %+v", raw.Data)
	}
}
