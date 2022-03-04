package feishu

import (
	"fmt"
)

const _openBaseURL = "https://open.feishu.cn"

type IDType string

const (
	OpenID  IDType = "open_id"
	UnionID IDType = "union_id"
	UserID  IDType = "user_id"
	Email   IDType = "email"
	ChatID  IDType = "chat_id"
)

type fsResponse struct {
	Code int    `json:"code"` // 错误码，非 0 表示失败
	Msg  string `json:"msg"`  // 错误描述
}

func (resp fsResponse) check(reqID, domain, apiName string) error {
	if resp.Code == 0 {
		return nil
	}
	if reqID != "" {
		return fmt.Errorf(_fmtErrResp, domain, apiName, reqID, resp.Code, resp.Msg)
	}
	return fmt.Errorf(_fmtErrRespNoID, domain, apiName, resp.Code, resp.Msg)
}

const (
	_fmtErrNotSupported = "[%s] %s: not supported"
	_fmtErrNoReqID      = "[%s] %s: %w"
	_fmtErrReq          = "[%s] %s (X-Request-ID: %s): %w"
	_fmtErrResp         = "[%s] %s (X-Request-ID: %s): %d: %s"
	_fmtErrRespNoID     = "[%s] %s: %d: %s"
)
