package feishu

import (
	"context"
	"fmt"
	"strconv"
)

// 名称: [消息与群组] 获取用户或机器人所在的群列表
// Func: [api_messenger_group.go] GetAllGroupChats
//
// 描述: 获取用户或者机器人所在群列表
// Info: 需要开启机器人能力
// Info: 查询参数 user_id_type 用于控制响应体中 owner_id 的类型，如果是获取机器人所在群列表该值可以不填
// Info: 请注意区分本接口和获取群信息的请求 URL
//
// Doc: https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/chat/list
//
// 自建应用: true
// 商店应用: true
//
// HTTP URL: /open-apis/im/v1/chats
// HTTP Method: GET
//
// 请求头: Authorization=Bearer {{AppAccessToken}}
//
type groupChatsResponse struct {
	fsResponse
	Data GroupChatsResponse `json:"data"`
}

type GroupChat struct {
	ChatID      string `json:"chat_id"`       // 群组 ID
	Avatar      string `json:"avatar"`        // 群头像 URL
	Name        string `json:"name"`          // 群名称
	Description string `json:"description"`   // 群描述
	OwnerID     string `json:"owner_id"`      // 群主 ID (查询参数 user_id_type 将影响该值的类型)
	OwnerIDType string `json:"owner_id_type"` // 群主 ID 类型
	External    bool   `json:"external"`      // 是否是外部群
	TenantKey   string `json:"tenant_key"`    // tenant key
}

type GroupChatsResponse struct {
	Items     []GroupChat `json:"items"`      // chat 列表
	PageToken string      `json:"page_token"` // 分页标记，当 has_more 为 true 时，会同时返回新的 page_token，否则为空字符串
	HasMore   bool        `json:"has_more"`   // 是否还有更多项
}

type GetAllGroupChatsOption = doOption

// WithGetAllGroupChatsOwnerIDType 控制 GroupChat.OwnerIDType, GroupChat.OwnerID 的值
//  仅支持 OpenID, UnionID, UserID
func WithGetAllGroupChatsOwnerIDType(idType IDType) GetAllGroupChatsOption {
	return withDoQueryKV("user_id_type", string(idType))
}

func WithGetAllGroupChatsPageSize(pageSize int) GetAllGroupChatsOption {
	return withDoQueryKV("page_size", strconv.Itoa(pageSize))
}

func WithGetAllGroupChatsPageToken(pageToken string) GetAllGroupChatsOption {
	return withDoQueryKV("page_token", pageToken)
}

func WithGetAllGroupChatsNextPage(lastResp GroupChatsResponse) GetAllGroupChatsOption {
	if !lastResp.HasMore {
		return nil
	}
	m := map[string]string{
		"page_token": lastResp.PageToken,
	}
	if len(lastResp.Items) > 0 && lastResp.Items[0].OwnerIDType != "" {
		m["user_id_type"] = lastResp.Items[0].OwnerIDType
	}
	return withDoQuery(m)
}

func (a *app) GetAllGroupChats(opts ...GetAllGroupChatsOption) (GroupChatsResponse, error) {
	return a.GetAllGroupChatsWithContext(context.Background(), opts...)
}

func (a *app) GetAllGroupChatsWithContext(ctx context.Context, opts ...GetAllGroupChatsOption) (GroupChatsResponse, error) {
	apiDomain := "消息与群组"
	apiName := "获取用户或机器人所在的群列表"
	urlSuffix := "/open-apis/im/v1/chats"

	if !a.isSupported(true, true) {
		return GroupChatsResponse{}, fmt.Errorf(_fmtErrNotSupported, apiDomain, apiName)
	}

	header := map[string]string{
		"Authorization": "Bearer ",
	}
	if accessToken, err := a.getAppAccessTokenWithContext(ctx); err != nil {
		return GroupChatsResponse{}, err
	} else {
		header["Authorization"] = fmt.Sprintf("Bearer %s", accessToken)
	}
	doOpts := a.buildOpts(apiDomain, apiName, header, opts...)
	reqID, reader, err := a._getWithContext(ctx, urlSuffix, doOpts...)
	if err != nil {
		return GroupChatsResponse{}, err
	}

	resp := new(groupChatsResponse)
	if err = a._decodeResp(apiDomain, apiName, reader, resp); err != nil {
		return GroupChatsResponse{}, err
	}

	if err = resp.check(reqID, apiDomain, apiName); err != nil {
		return GroupChatsResponse{}, err
	}

	return resp.Data, nil
}
