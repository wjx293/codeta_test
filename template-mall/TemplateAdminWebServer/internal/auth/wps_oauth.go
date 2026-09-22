package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"template-mall/TemplateAdminWebServer/internal/config"
)

// WPSOAuthProvider WPS 开放平台 OAuth2 授权码模式实现。
// client_secret 仅在服务端使用，MUST NOT 出现在前端/响应/日志。
type WPSOAuthProvider struct {
	cfg *config.Config
}

// NewWPSOAuthProvider 创建 WPS OAuth2 Provider。
func NewWPSOAuthProvider(cfg *config.Config) *WPSOAuthProvider {
	return &WPSOAuthProvider{cfg: cfg}
}

// GetRedirectURL 返回 WPS 授权页跳转地址。
// 文档: https://365.kdocs.cn/3rd/open/documents/app-integration-dev/wps365/server/certification-authorization/user-authorization/flow
func (p *WPSOAuthProvider) GetRedirectURL(state string) string {
	q := url.Values{}
	q.Set("client_id", p.cfg.WPSClientID)
	q.Set("redirect_uri", p.cfg.WPSRedirectURI)
	q.Set("scope", p.cfg.WPSScope)
	q.Set("response_type", "code")
	q.Set("state", state)
	return p.cfg.WPSAuthURL + "?" + q.Encode()
}

// ExchangeCode 用授权码换 access_token（服务端完成）。
func (p *WPSOAuthProvider) ExchangeCode(code string) (*Token, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", p.cfg.WPSClientID)
	form.Set("client_secret", p.cfg.WPSClientSecret)
	form.Set("redirect_uri", p.cfg.WPSRedirectURI)

	req, err := http.NewRequest(http.MethodPost, p.cfg.WPSTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint status %d: %s", resp.StatusCode, string(body))
	}

	var raw struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		Code         int    `json:"code"`
		Msg          string `json:"msg"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse token response: %w", err)
	}
	if raw.Code != 0 && raw.Msg != "" {
		return nil, fmt.Errorf("token api error: %s", raw.Msg)
	}
	if raw.AccessToken == "" {
		return nil, fmt.Errorf("token response missing access_token")
	}

	return &Token{
		AccessToken:  raw.AccessToken,
		RefreshToken: raw.RefreshToken,
		TokenType:    raw.TokenType,
		ExpiresIn:    raw.ExpiresIn,
	}, nil
}

// RefreshAccessToken 刷新 access_token（服务端完成，可选）。
func (p *WPSOAuthProvider) RefreshAccessToken(refreshToken string) (*Token, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", p.cfg.WPSClientID)
	form.Set("client_secret", p.cfg.WPSClientSecret)

	req, err := http.NewRequest(http.MethodPost, p.cfg.WPSTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh token status %d: %s", resp.StatusCode, string(body))
	}

	var raw struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		Code         int    `json:"code"`
		Msg          string `json:"msg"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if raw.Code != 0 && raw.Msg != "" {
		return nil, fmt.Errorf("refresh token error: %s", raw.Msg)
	}
	return &Token{
		AccessToken:  raw.AccessToken,
		RefreshToken: raw.RefreshToken,
		TokenType:    raw.TokenType,
		ExpiresIn:    raw.ExpiresIn,
	}, nil
}

// GetUserInfo 用 access_token 拉取当前用户信息（服务端完成）。
func (p *WPSOAuthProvider) GetUserInfo(token *Token) (*UserInfo, error) {
	req, err := http.NewRequest(http.MethodGet, p.cfg.WPSUserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	if p.cfg.WPSKSOSignEnabled {
		if err := signKSO1(req, p.cfg.WPSClientID, p.cfg.WPSClientSecret, nil); err != nil {
			return nil, err
		}
	} else {
		req.Header.Set("Content-Type", ksoContentType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get userinfo: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read userinfo response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint status %d: %s", resp.StatusCode, string(body))
	}

	var raw struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			ID        string `json:"id"`
			UserName  string `json:"user_name"`
			Avatar    string `json:"avatar"`
			CompanyID string `json:"company_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse userinfo response: %w", err)
	}
	if raw.Code != 0 {
		return nil, fmt.Errorf("userinfo api error: %s", raw.Msg)
	}
	if raw.Data.ID == "" {
		return nil, fmt.Errorf("userinfo response missing user id")
	}

	return &UserInfo{
		UserID:   raw.Data.ID,
		Username: raw.Data.UserName,
		Avatar:   raw.Data.Avatar,
	}, nil
}
