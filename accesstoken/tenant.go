package accesstoken

import (
	"github.com/polym/feishu/internal/transport"
)

type GetTenantAccessTokenOptions struct {
	AppId     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

type GetTenantAccessTokenResp struct {
	AccessToken string `json:"tenant_access_token"`
	Expire      int    `json:"expire"`
}

// https://open.feishu.cn/document/ukTMukTMukTM/uIjNz4iM2MjLyYzM
func GetTenantAccessToken(p transport.Transport, opt GetTenantAccessTokenOptions) (*GetTenantAccessTokenResp, error) {
	req := &transport.Request{
		Method:      "POST",
		Url:         "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal/",
		BodySpec:    opt,
		ContentType: "application/json",
	}
	resp := &GetTenantAccessTokenResp{}
	err := p.Do(req, resp, transport.AuthKindNone)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
