package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

type App interface {
	GetAppAccessTokenInternal() (AppAccessTokenInternal, error)
	GetAppAccessTokenInternalWithContext(ctx context.Context) (AppAccessTokenInternal, error)
	GetTenantAccessTokenInternal() (TenantAccessTokenInternal, error)
	GetTenantAccessTokenInternalWithContext(ctx context.Context) (TenantAccessTokenInternal, error)

	SendMessage(receiver MessageReceiver, msg *Message) (MessageDetail, error)
	SendMessageWithContext(ctx context.Context, receiver MessageReceiver, msg *Message) (MessageDetail, error)
	ReplyMessage(messageID string, msg *Message) (MessageDetail, error)
	ReplyMessageWithContext(ctx context.Context, messageID string, msg *Message) (MessageDetail, error)

	UploadImage(src UploadImageOption) (imageKey string, err error)
	UploadImageWithContext(ctx context.Context, src UploadImageOption) (imageKey string, err error)

	GetAllGroupChats(opts ...GetAllGroupChatsOption) (GroupChatsResponse, error)
	GetAllGroupChatsWithContext(ctx context.Context, opts ...GetAllGroupChatsOption) (GroupChatsResponse, error)

	ListenEventCallback(w http.ResponseWriter, r *http.Request)
	RegisterEventCallback(eventType EventType, handler EventHandler)
	RegisterEventCallbackV1(eventType EventType, handler EventHandlerV1)
}

type Logger interface {
	Debug(string)
}

var _ App = (*app)(nil)

type app struct {
	isCustomApp       bool
	isStoreApp        bool
	id                string
	secret            string
	encryptKey        string
	verificationToken string

	openBaseURL string
	opt         struct {
		logger Logger
		debug  bool
		cli    *http.Client
	}

	sync.Mutex
	appAccess      fsToken
	tenantAccess   fsToken
	eventHandler   map[EventType]EventHandler
	eventHandlerV1 map[EventType]EventHandlerV1
}

type AppOption func(*app)

func WithAppEventEncryptKey(encryptKey string) AppOption {
	return func(a *app) {
		a.encryptKey = encryptKey
	}
}

func WithAppEventVerificationToken(verificationToken string) AppOption {
	return func(a *app) {
		a.verificationToken = verificationToken
	}
}

func WithAppOpenBaseURL(urlPrefix string) AppOption {
	return func(a *app) {
		a.openBaseURL = strings.TrimRight(urlPrefix, "/")
	}
}

func WithAppDebug(b bool) AppOption {
	return func(a *app) {
		a.opt.debug = b
	}
}

func WithAppDebugLogger(logger Logger) AppOption {
	return func(a *app) {
		a.opt.logger = logger
	}
}

func NewCustomApp(id, secret string, opts ...AppOption) App {
	a := newApp(id, secret, opts...)
	a.isCustomApp = true
	a.isStoreApp = false
	return a
}

// func NewStoreApp(id, secret string, opts ...AppOption) App {
// 	a := newApp(id, secret, opts...)
// 	a.isCustomApp = false
// 	a.isStoreApp = true
// 	return a
// }

func newApp(appID, appSecret string, opts ...AppOption) *app {
	opts = append([]AppOption{WithAppOpenBaseURL(_openBaseURL)}, opts...)
	a := &app{
		id:             appID,
		secret:         appSecret,
		eventHandler:   make(map[EventType]EventHandler),
		eventHandlerV1: make(map[EventType]EventHandlerV1),
	}
	for _, fn := range opts {
		if fn == nil {
			continue
		}
		fn(a)
	}
	return a
}

func (a *app) isSupported(custom, store bool) bool {
	return (a.isCustomApp == custom && a.isCustomApp) ||
		(a.isStoreApp == store && a.isStoreApp)
}

func (a *app) _getWithContext(ctx context.Context, urlSuffix string, opts ...doOption) (reqID string, resp io.Reader, err error) {
	return a._do(ctx, http.MethodGet, a.openBaseURL+urlSuffix, nil, opts...)
}

func (a *app) _postWithContext(ctx context.Context, urlSuffix string, data interface{}, opts ...doOption) (reqID string, resp io.Reader, err error) {
	return a._do(ctx, http.MethodPost, a.openBaseURL+urlSuffix, data, opts...)
}

func (a *app) _doUploadWithContext(ctx context.Context, urlSuffix string, opts ...doOption) (reqID string, resp io.Reader, err error) {
	return a._do(ctx, http.MethodPost, a.openBaseURL+urlSuffix, nil, opts...)
}

func (a *app) _do(ctx context.Context, method, rawURL string, data interface{}, opts ...doOption) (reqID string, resp io.Reader, err error) {
	reqID, resp, err = _doWithContext(ctx, method, rawURL, data, opts...)
	if err == nil {
		return
	}

	doOpt := _newDoOpt(opts...)
	if reqID != "" {
		return reqID, resp, fmt.Errorf(_fmtErrReq, doOpt.apiDomain, doOpt.apiName, reqID, err)
	}
	return reqID, resp, fmt.Errorf(_fmtErrNoReqID, doOpt.apiDomain, doOpt.apiName, err)
}

func (a *app) _decodeResp(domain, apiName string, reader io.Reader, resp interface{}) (err error) {
	if err = json.NewDecoder(reader).Decode(resp); err != nil {
		return fmt.Errorf(_fmtErrNoReqID, domain, apiName, err)
	}
	return nil
}

func (a *app) buildOpts(apiDomain, apiName string, header map[string]string, opts ...doOption) []doOption {
	nOpts := make([]doOption, 0, 16)
	nOpts = append(nOpts, withDoAPIDomain(apiDomain), withDoAPIName(apiName), withDoHeader(header))
	if a.opt.logger != nil {
		nOpts = append(nOpts, withDoLogger(a.opt.logger))
	}
	if a.opt.debug {
		nOpts = append(nOpts, withDoDebug(a.opt.debug))
	}
	if a.opt.cli != nil {
		nOpts = append(nOpts, withDoHTTPCli(a.opt.cli))
	}
	nOpts = append(nOpts, opts...)
	return nOpts
}

func (a *app) getAppAccessTokenWithContext(ctx context.Context) (accessToken string, err error) {
	a.Lock()
	defer a.Unlock()

	if !a.appAccess.isEmpty() && a.appAccess.notExpired() {
		return a.appAccess.get(), nil
	}

	if a.isCustomApp {
		if _, err = a.GetAppAccessTokenInternalWithContext(ctx); err != nil {
			return "", err
		}
	}

	// if a.isStoreApp {
	// }

	return a.appAccess.get(), nil
}
