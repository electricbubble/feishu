package feishu

import (
	"context"
	"encoding/json"
	"fmt"
)

// 名称: [消息与群组] 发送消息
// Func: [api_messenger.go] SendMessage
//
// 描述: 给指定用户或者会话发送消息，支持文本、富文本、可交互的消息卡片、群名片、个人名片、图片、视频、音频、文件、表情包
// Info: 需要开启机器人能力
// Info: 给用户发送消息，需要机器人对用户有可用性
// Info: 给群组发送消息，需要机器人在群中
// Info: 该接口不支持给部门成员发消息，请使用 批量发送消息
// Info: 文本消息请求体最大不能超过150KB
// Info: 卡片及富文本消息请求体最大不能超过30KB
// Info: 消息卡片的 update_multi（是否为共享卡片）字段在卡片内容的config结构体中设置。详细参考文档配置卡片属性
//
// Doc: https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/create
//
// 自建应用: true
// 商店应用: true
//
// HTTP URL: /open-apis/im/v1/messages
// HTTP Method: POST
//
// 请求头: Authorization=Bearer {{AppAccessToken}}
// 请求头: Content-Type=application/json; charset=utf-8
//
type sendMessageRequest struct {
	// 依据 receive_id_type 的值，填写对应的消息接收者id
	ReceiveID string `json:"receive_id,omitempty"`

	// 消息内容，json结构序列化后的字符串。不同msg_type对应不同内容。
	//
	// 消息类型 包括：text、post、image、file、audio、media、sticker、interactive、share_chat、share_user等
	Content string `json:"content"`

	// 消息类型 包括：text、post、image、file、audio、media、sticker、interactive、share_chat、share_user等
	MsgType string `json:"msg_type"`
}

type sendMessageResponse struct {
	fsResponse
	Data MessageDetail `json:"data"`
}

type MessageDetail struct {
	MessageID      string      `json:"message_id"`       // 消息id
	RootID         string      `json:"root_id"`          // 根消息id
	ParentID       string      `json:"parent_id"`        // 父消息的id
	MsgType        string      `json:"msg_type"`         // 消息类型 包括：text、post、image、file、audio、media、sticker、interactive、share_chat、share_user等
	CreateTime     string      `json:"create_time"`      // 消息生成的时间戳（毫秒）
	UpdateTime     string      `json:"update_time"`      // 消息更新的时间戳（毫秒）
	Deleted        bool        `json:"deleted"`          // 消息是否被撤回
	Updated        bool        `json:"updated"`          // 消息是否被更新
	ChatID         string      `json:"chat_id"`          // 所属的群
	Sender         Sender      `json:"sender"`           // 发送者，可以是用户或应用
	Body           MessageBody `json:"body"`             // 消息内容
	Mentions       []Mention   `json:"mentions"`         // 被@的用户或机器人的id列表
	UpperMessageID string      `json:"upper_message_id"` // 合并转发消息中，上一层级的消息id message_id
}

type Sender struct {
	ID         string `json:"id"`          // 发送者的id
	IDType     string `json:"id_type"`     // 发送者的id类型
	SenderType string `json:"sender_type"` // 发送者的类型
	TenantKey  string `json:"tenant_key"`  // 为租户在飞书上的唯一标识，用来换取对应的tenant_access_token，也可以用作租户在应用里面的唯一标识
}

type MessageBody struct {
	// json结构序列化后的字符串。不同msg_type对应不同内容。
	//
	// 消息类型 包括：text、post、image、file、audio、media、sticker、interactive、share_chat、share_user等
	Content string `json:"content"`
}

type Mention struct {
	Key       string `json:"key"`        // 被@的用户或机器人的序号。例如，第3个被@到的成员，值为“@_user_3”
	ID        string `json:"id"`         // 被@的用户或者机器人的open_id
	IDType    string `json:"id_type"`    // 被@的用户或机器人 id 类型，目前仅支持 open_id
	Name      string `json:"name"`       // 被@的用户或机器人的姓名
	TenantKey string `json:"tenant_key"` // 为租户在飞书上的唯一标识，用来换取对应的tenant_access_token，也可以用作租户在应用里面的唯一标识
}

type MessageReceiver struct {
	IDType IDType // 接收者的 🆔 类型
	ID     string // 接收者 🆔
}

type Message struct {
	msgType string
	content interface{}
}

func NewMessageText(content string) *Message {
	msgType := "text"
	return &Message{
		msgType: msgType,
		content: map[string]string{
			msgType: content,
		},
	}
}

func NewMessagePost(p Post, more ...Post) *Message {
	more = append([]Post{p}, more...)
	i18nPosts := make([]i18nPost, 0, len(more))
	for _, fn := range more {
		i18nPosts = append(i18nPosts, fn())
	}
	post := make(map[string]interface{}, 3)
	for _, p := range i18nPosts {
		post[p.lang] = map[string]interface{}{
			"title":   p.title,
			"content": p.elements,
		}
	}

	msgType := "post"
	return &Message{
		msgType: msgType,
		content: post,
	}
}

func NewMessageImage(imageKey string) *Message {
	return &Message{
		msgType: "image",
		content: map[string]interface{}{
			"image_key": imageKey,
		},
	}
}

func NewMessageCard(bgColor CardTitleBgColor, cfg CardConfig, c Card, more ...Card) *Message {
	more = append([]Card{c}, more...)
	cards := make([]i18nCard, 0, len(more))
	for _, fn := range more {
		cards = append(cards, fn())
	}

	i18nTitle := make(map[string]string, 3)
	i18nElements := make(map[string]interface{}, 3)
	for _, c := range cards {
		i18nTitle[c.lang] = c.title
		i18nElements[c.lang] = c.elements
	}

	sub := map[string]interface{}{
		"header":        buildCardHeader(bgColor, i18nTitle),
		"i18n_elements": i18nElements,
	}

	if cfg != nil {
		_cfg := cfg()
		if _cfg.mCfg != nil {
			sub["config"] = _cfg.mCfg
		}
		if _cfg.mCardLink != nil {
			sub["card_link"] = _cfg.mCardLink
		}
	}

	return &Message{
		msgType: "interactive",
		content: sub,
	}
}

func NewMessageShareChat(chatID string) *Message {
	return &Message{
		msgType: "share_chat",
		content: map[string]interface{}{
			"chat_id": chatID,
		},
	}
}

func NewMessageShareUser(openID string) *Message {
	return &Message{
		msgType: "share_user",
		content: map[string]interface{}{
			"user_id": openID,
		},
	}
}

func NewMessageAudio(fileKey string) *Message {
	return &Message{
		msgType: "audio",
		content: map[string]interface{}{
			"file_key": fileKey,
		},
	}
}

func NewMessageMedia(fileKey, imageKey string) *Message {
	return &Message{
		msgType: "media",
		content: map[string]interface{}{
			"file_key":  fileKey,
			"image_key": imageKey,
		},
	}
}

func NewMessageFile(fileKey string) *Message {
	return &Message{
		msgType: "file",
		content: map[string]interface{}{
			"file_key": fileKey,
		},
	}
}

func NewMessageSticker(fileKey string) *Message {
	return &Message{
		msgType: "sticker",
		content: map[string]interface{}{
			"file_key": fileKey,
		},
	}
}

func (a *app) SendMessage(receiver MessageReceiver, msg *Message) (MessageDetail, error) {
	return a.SendMessageWithContext(context.Background(), receiver, msg)
}

func (a *app) SendMessageWithContext(ctx context.Context, receiver MessageReceiver, msg *Message) (MessageDetail, error) {
	apiDomain := "消息与群组"
	apiName := "发送消息"
	urlSuffix := "/open-apis/im/v1/messages"

	if !a.isSupported(true, true) {
		return MessageDetail{}, fmt.Errorf(_fmtErrNotSupported, apiDomain, apiName)
	}

	data := &sendMessageRequest{
		ReceiveID: receiver.ID,
		Content:   "",
		MsgType:   msg.msgType,
	}
	if bs, err := json.Marshal(msg.content); err != nil {
		return MessageDetail{}, err
	} else {
		data.Content = string(bs)
	}
	header := map[string]string{
		"Content-Type":  "application/json; charset=utf-8",
		"Authorization": "Bearer ",
	}
	if accessToken, err := a.getAppAccessTokenWithContext(ctx); err != nil {
		return MessageDetail{}, err
	} else {
		header["Authorization"] = fmt.Sprintf("Bearer %s", accessToken)
	}
	doOpts := a.buildOpts(apiDomain, apiName, header,
		withDoQueryKV("receive_id_type", string(receiver.IDType)),
	)
	reqID, reader, err := a._postWithContext(ctx, urlSuffix, data, doOpts...)
	if err != nil {
		return MessageDetail{}, err
	}

	resp := new(sendMessageResponse)
	if err = a._decodeResp(apiDomain, apiName, reader, resp); err != nil {
		return MessageDetail{}, err
	}

	if err = resp.check(reqID, apiDomain, apiName); err != nil {
		return MessageDetail{}, err
	}

	return resp.Data, nil

}

// 名称: [消息与群组] 回复消息
// Func: [api_messenger.go] ReplyMessage
//
// 描述: 回复指定消息，支持文本、富文本、卡片、群名片、个人名片、图片、视频、文件等多种消息类型
// Info: 需要开启机器人能力
// Info: 回复私聊消息，需要机器人对用户有可用性
// Info: 回复群组消息，需要机器人在群中
//
// Doc: https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/reply
//
// 自建应用: true
// 商店应用: true
//
// HTTP URL: /open-apis/im/v1/messages/:message_id/reply
// HTTP Method: POST
//
// 请求头: Authorization=Bearer {{AppAccessToken}}
// 请求头: Content-Type=application/json; charset=utf-8
//

func (a *app) ReplyMessage(messageID string, msg *Message) (MessageDetail, error) {
	return a.ReplyMessageWithContext(context.Background(), messageID, msg)
}

func (a *app) ReplyMessageWithContext(ctx context.Context, messageID string, msg *Message) (MessageDetail, error) {
	apiDomain := "消息与群组"
	apiName := "回复消息"
	urlSuffix := fmt.Sprintf("/open-apis/im/v1/messages/%s/reply", messageID)

	if !a.isSupported(true, true) {
		return MessageDetail{}, fmt.Errorf(_fmtErrNotSupported, apiDomain, apiName)
	}

	data := &sendMessageRequest{
		Content: "",
		MsgType: msg.msgType,
	}
	if bs, err := json.Marshal(msg.content); err != nil {
		return MessageDetail{}, err
	} else {
		data.Content = string(bs)
	}
	header := map[string]string{
		"Content-Type":  "application/json; charset=utf-8",
		"Authorization": "Bearer ",
	}
	if accessToken, err := a.getAppAccessTokenWithContext(ctx); err != nil {
		return MessageDetail{}, err
	} else {
		header["Authorization"] = fmt.Sprintf("Bearer %s", accessToken)
	}
	doOpts := a.buildOpts(apiDomain, apiName, header)
	reqID, reader, err := a._postWithContext(ctx, urlSuffix, data, doOpts...)
	if err != nil {
		return MessageDetail{}, err
	}

	resp := new(sendMessageResponse)
	if err = a._decodeResp(apiDomain, apiName, reader, resp); err != nil {
		return MessageDetail{}, err
	}

	if err = resp.check(reqID, apiDomain, apiName); err != nil {
		return MessageDetail{}, err
	}

	return resp.Data, nil

}
