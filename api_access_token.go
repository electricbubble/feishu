package feishu

import (
	"context"
	"fmt"
	"time"
)

// 名称: [访问凭证] 获取 app_access_token（企业自建应用）
// Func: [api_access_token.go] GetAppAccessTokenInternal
//
// 描述: 企业自建应用通过此接口获取 app_access_token，调用接口获取应用资源时，需要使用 app_access_token 作为授权凭证
// Info: token 有效期为 2 小时，在此期间调用该接口 token 不会改变。当 token 有效期小于 30 分的时候，再次请求获取 token 的时候，会生成一个新的 token，与此同时老的 token 依然有效
//
// Doc: https://open.feishu.cn/document/ukTMukTMukTM/ukDNz4SO0MjL5QzM/auth-v3/auth/app_access_token_internal
//
// 自建应用: true
// 商店应用: false
//
// HTTP URL: /open-apis/auth/v3/app_access_token/internal
// HTTP Method: POST
//
// 请求头: Content-Type=application/json; charset=utf-8
//
type appAccessTokenInternalRequest struct {
	AppID     string `json:"app_id"`     // 应用唯一标识，创建应用后获得
	AppSecret string `json:"app_secret"` // 应用秘钥，创建应用后获得
}

type appAccessTokenInternalResponse struct {
	fsResponse
	AppAccessTokenInternal
}

type AppAccessTokenInternal struct {
	AppAccessToken string `json:"app_access_token"` // 访问 token
	Expire         int    `json:"expire"`           // app_access_token 过期时间，单位：秒
}

func (a *app) GetAppAccessTokenInternal() (AppAccessTokenInternal, error) {
	return a.GetAppAccessTokenInternalWithContext(context.Background())
}

func (a *app) GetAppAccessTokenInternalWithContext(ctx context.Context) (AppAccessTokenInternal, error) {
	apiDomain := "访问凭证"
	apiName := "获取 app_access_token（企业自建应用）"
	urlSuffix := "/open-apis/auth/v3/app_access_token/internal"

	if !a.isSupported(true, false) {
		return AppAccessTokenInternal{}, fmt.Errorf(_fmtErrNotSupported, apiDomain, apiName)
	}

	data := &appAccessTokenInternalRequest{
		AppID:     a.id,
		AppSecret: a.secret,
	}
	header := map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}

	reqID, reader, err := a._postWithContext(ctx, urlSuffix, data, a.buildOpts(apiDomain, apiName, header)...)
	if err != nil {
		return AppAccessTokenInternal{}, err
	}

	resp := new(appAccessTokenInternalResponse)
	if err = a._decodeResp(apiDomain, apiName, reader, resp); err != nil {
		return AppAccessTokenInternal{}, err
	}

	if err = resp.check(reqID, apiDomain, apiName); err != nil {
		return AppAccessTokenInternal{}, err
	}

	a.appAccess.Lock()
	defer a.appAccess.Unlock()
	a.appAccess.set(resp.AppAccessToken, time.Duration(resp.Expire)*time.Second, 30*time.Minute)

	return resp.AppAccessTokenInternal, nil
}

// 名称: [访问凭证] 获取 tenant_access_token（企业自建应用）
// Func: [api_access_token.go] GetTenantAccessTokenInternal
//
// 描述: 企业自建应用通过此接口获取 tenant_access_token，调用接口获取企业资源时，需要使用 tenant_access_token 作为授权凭证
// Info: token 有效期为 2 小时，在此期间调用该接口 token 不会改变。当 token 有效期小于 30 分的时候，再次请求获取 token 的时候，会生成一个新的 token，与此同时老的 token 依然有效
//
// Doc: https://open.feishu.cn/document/ukTMukTMukTM/ukDNz4SO0MjL5QzM/auth-v3/auth/tenant_access_token_internal
//
// 自建应用: true
// 商店应用: false
//
// HTTP URL: /open-apis/auth/v3/tenant_access_token/internal
// HTTP Method: POST
// 权限要求: 无
//
// 请求头: Content-Type=application/json; charset=utf-8
//
type tenantAccessTokenInternalRequest struct {
	AppID     string `json:"app_id"`     // 应用唯一标识，创建应用后获得
	AppSecret string `json:"app_secret"` // 应用秘钥，创建应用后获得
}

type tenantAccessTokenInternalResponse struct {
	fsResponse
	TenantAccessTokenInternal
}

type TenantAccessTokenInternal struct {
	TenantAccessToken string `json:"tenant_access_token"` // 访问 token
	Expire            int    `json:"expire"`              // token 过期时间，单位: 秒
}

func (a *app) GetTenantAccessTokenInternal() (TenantAccessTokenInternal, error) {
	return a.GetTenantAccessTokenInternalWithContext(context.Background())
}

func (a *app) GetTenantAccessTokenInternalWithContext(ctx context.Context) (TenantAccessTokenInternal, error) {
	apiDomain := "访问凭证"
	apiName := "获取 tenant_access_token（企业自建应用）"
	urlSuffix := "/open-apis/auth/v3/tenant_access_token/internal"

	if !a.isSupported(true, false) {
		return TenantAccessTokenInternal{}, fmt.Errorf(_fmtErrNotSupported, apiDomain, apiName)
	}

	data := &tenantAccessTokenInternalRequest{
		AppID:     a.id,
		AppSecret: a.secret,
	}
	header := map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}

	reqID, reader, err := a._postWithContext(ctx, urlSuffix, data, a.buildOpts(apiDomain, apiName, header)...)
	if err != nil {
		return TenantAccessTokenInternal{}, err
	}

	resp := new(tenantAccessTokenInternalResponse)
	if err = a._decodeResp(apiDomain, apiName, reader, resp); err != nil {
		return TenantAccessTokenInternal{}, err
	}

	if err = resp.check(reqID, apiDomain, apiName); err != nil {
		return TenantAccessTokenInternal{}, err
	}

	a.tenantAccess.Lock()
	defer a.tenantAccess.Unlock()
	a.tenantAccess.set(resp.TenantAccessToken, time.Duration(resp.Expire)*time.Second, 30*time.Minute)

	return resp.TenantAccessTokenInternal, nil

}
