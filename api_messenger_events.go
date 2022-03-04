package feishu

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type EventHandler func(header EventHeaderV2, event json.RawMessage)
type EventHandlerV1 func(header EventHeaderV1, event json.RawMessage)

func (a *app) ListenEventCallback(w http.ResponseWriter, r *http.Request) {
	var (
		opt = _newDoOpt(a.buildOpts("事件订阅", "收到一个新事件", nil)...)
		err error
	)
	defer func() {
		if err == nil {
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}()

	var body bytes.Buffer
	if _, err = io.Copy(&body, r.Body); err != nil {
		opt.debugLog(fmt.Sprintf("[%s - %s] %s\n", opt.apiDomain, opt.apiName, err))
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	eReq := new(eventRequest)
	if err = json.Unmarshal(body.Bytes(), eReq); err != nil {
		opt.debugLog(fmt.Sprintf("[%s - %s] unmarshal event: %s\n", opt.apiDomain, opt.apiName, err))
		return
	}

	if eReq.Encrypt != "" {
		if body, err = eReq.decrypt(a.encryptKey); err != nil {
			opt.debugLog(fmt.Sprintf("[%s - %s] decrypt event: %s\n", opt.apiDomain, opt.apiName, err))
			return
		}
	}

	switch {
	case eReq.Type == EventTypeURLVerification:
		v := new(eventURLVerificationRequest)
		if err = json.Unmarshal(body.Bytes(), v); err != nil {
			opt.debugLog(fmt.Sprintf("[%s - %s] unmarshal event(%s): %s\n", opt.apiDomain, opt.apiName, EventTypeURLVerification, err))
			return
		}

		if v.Token != a.verificationToken {
			opt.debugLog(fmt.Sprintf("[%s - %s] unexpected event callback token (app: %s): %s\n", opt.apiDomain, opt.apiName, a.verificationToken, v.Token))
			return
		}

		resp := &eventURLVerificationResponse{Challenge: v.Challenge}
		var bs []byte
		bs, err = json.Marshal(resp)
		if err != nil {
			opt.debugLog(fmt.Sprintf("[%s - %s] marshal response event(%s): %s\n", opt.apiDomain, opt.apiName, EventTypeURLVerification, err))
			return
		}

		if _, err = w.Write(bs); err != nil {
			opt.debugLog(fmt.Sprintf("[%s - %s] write response event(%s): %s\n", opt.apiDomain, opt.apiName, EventTypeURLVerification, err))
			return
		}
		opt.debugLog(fmt.Sprintf("[%s - %s] %s successful\n", opt.apiDomain, opt.apiName, EventTypeURLVerification))
	case eReq.Schema == "2.0":
		if eReq.Header == nil {
			opt.debugLog(fmt.Sprintf("[%s - %s] get event(2.0), but header is nil\n", opt.apiDomain, opt.apiName))
			return
		}
		if eReq.Header.Token != a.verificationToken {
			opt.debugLog(fmt.Sprintf("[%s - %s] unexpected event callback token (app: %s): %s\n", opt.apiDomain, opt.apiName, a.verificationToken, eReq.Header.Token))
			return
		}
		handler, ok := a.eventHandler[eReq.Header.EventType]
		if ok {
			go handler(*eReq.Header, eReq.Event)
		} else {
			opt.debugLog(fmt.Sprintf("[%s - %s] unregistered event: %s\n", opt.apiDomain, opt.apiName, eReq.Header.EventType))
		}
		w.WriteHeader(http.StatusOK)
	case eReq.EventHeaderV1.UUID != "":
		if eReq.EventHeaderV1.Token != a.verificationToken {
			opt.debugLog(fmt.Sprintf("[%s - %s] unexpected event callback token (app: %s): %s\n", opt.apiDomain, opt.apiName, a.verificationToken, eReq.EventHeaderV1.Token))
			return
		}
		handler, ok := a.eventHandlerV1[eReq.EventHeaderV1.Type]
		if ok {
			go handler(eReq.EventHeaderV1, eReq.Event)
		} else {
			opt.debugLog(fmt.Sprintf("[%s - %s] unregistered event(v1.0): %s\n", opt.apiDomain, opt.apiName, eReq.Header.EventType))
		}
		w.WriteHeader(http.StatusOK)
	}

}

func (a *app) RegisterEventCallback(eventType EventType, handler EventHandler) {
	a.eventHandler[eventType] = handler
}

func (a *app) RegisterEventCallbackV1(eventType EventType, handler EventHandlerV1) {
	a.eventHandlerV1[eventType] = handler
}

type EventType string

const (
	EventTypeURLVerification EventType = "url_verification"
	EventTypeMessageReceived EventType = "im.message.receive_v1" // 接收消息 v2.0
)

type eventRequest struct {
	Encrypt string `json:"encrypt,omitempty"` // 加密字符串, 如果设置了 EncryptKey 会需要进行解密

	Schema string         `json:"schema,omitempty"` // 2.0 版本字段
	Header *EventHeaderV2 `json:"header,omitempty"` // 2.0 版本字段

	EventHeaderV1 // 1.0 版本字段

	Event json.RawMessage `json:"event,omitempty"` // 不同事件此处数据不同
}

type EventHeaderV2 struct {
	EventID    string    `json:"event_id,omitempty"`    // 事件 ID
	EventType  EventType `json:"event_type,omitempty"`  // 事件类型
	CreateTime string    `json:"create_time,omitempty"` // 事件创建时间戳（单位：毫秒）
	Token      string    `json:"token,omitempty"`       // 事件 Token
	AppID      string    `json:"app_id,omitempty"`      // 应用 ID
	TenantKey  string    `json:"tenant_key,omitempty"`  // 租户 Key
}

type EventHeaderV1 struct {
	Timestamp string    `json:"ts,omitempty"`    // 事件发送的时间，一般近似于事件发生的时间
	UUID      string    `json:"uuid,omitempty"`  // 事件的唯一标识
	Token     string    `json:"token,omitempty"` // 即 Verification Token
	Type      EventType `json:"type,omitempty"`  // 事件类型
}

type eventURLVerificationRequest struct {
	Challenge string    `json:"challenge"` // 应用需要原样返回的值
	Token     string    `json:"token"`     // Token的使用可参考文档“通过Token验证事件来源”
	Type      EventType `json:"type"`      // 表示这是一个验证请求
}

type eventURLVerificationResponse struct {
	Challenge string `json:"challenge"`
}

func (eq *eventRequest) decrypt(encryptKey string) (raw bytes.Buffer, err error) {
	ciphertext, err := base64.StdEncoding.DecodeString(eq.Encrypt)
	if err != nil {
		return bytes.Buffer{}, err
	}
	if len(ciphertext) < aes.BlockSize {
		return bytes.Buffer{}, errors.New("ciphertext too short")
	}

	key := make([]byte, 0, sha256.Size)
	{
		h := sha256.New()
		h.Write([]byte(encryptKey))
		key = h.Sum(nil)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return bytes.Buffer{}, err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	// CBC mode always works in whole blocks.
	if len(ciphertext)%aes.BlockSize != 0 {
		return bytes.Buffer{}, errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	left := bytes.IndexByte(ciphertext, '{')
	if left == -1 {
		left = 0
	}
	right := bytes.LastIndexByte(ciphertext, '}')
	if right == -1 {
		right = len(ciphertext) - 1
	}

	ciphertext = ciphertext[left : right+1]
	raw.Write(ciphertext)

	if err = json.Unmarshal(ciphertext, eq); err != nil {
		return raw, err
	}

	return
}

type EventMessageReceived struct {
	Sender  EventSender  `json:"sender"`  // 事件的发送者
	Message EventMessage `json:"message"` // 事件中包含的消息内容
}

type EventSender struct {
	SenderID   EventSenderID `json:"sender_id"`   // 用户 ID
	SenderType string        `json:"sender_type"` // 消息发送者类型。目前只支持用户(user)发送的消息。
	TenantKey  string        `json:"tenant_key"`  // tenant key
}

type EventSenderID struct {
	UnionID string `json:"union_id"` // 用户的 union id
	UserID  string `json:"user_id"`  // 用户的 user id
	OpenID  string `json:"open_id"`  // 用户的 open id
}

type EventUserID struct {
	UnionID string `json:"union_id"` // 用户的 union id
	UserID  string `json:"user_id"`  // 用户的 user id
	OpenID  string `json:"open_id"`  // 用户的 open id
}

type EventMention struct {
	Key       string      `json:"key"`        // mention key
	ID        EventUserID `json:"id"`         // 用户 ID
	Name      string      `json:"name"`       // 用户姓名
	TenantKey string      `json:"tenant_key"` // tenant key
}

type EventMessage struct {
	MessageID   string         `json:"message_id"`   // 消息的 open_message_id
	RootID      string         `json:"root_id"`      // 回复消息 根 id
	ParentID    string         `json:"parent_id"`    // 回复消息 父 id
	CreateTime  string         `json:"create_time"`  // 消息发送时间 毫秒
	ChatID      string         `json:"chat_id"`      // 消息所在的群组 id
	ChatType    string         `json:"chat_type"`    // 消息所在的群组类型，单聊（p2p）或群聊（group）
	MessageType string         `json:"message_type"` // 消息类型
	Content     string         `json:"content"`      // 消息内容, json 格式各类型消息Content
	Mentions    []EventMention `json:"mentions"`     // 被提及用户的信息
}
