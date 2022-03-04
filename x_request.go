package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

var DefaultHTTPClient = &http.Client{}

type _doOpt struct {
	apiDomain      string
	apiName        string
	header         map[string]string
	query          map[string]string
	uploadFormData []*_doFormData
	httpCli        *http.Client
	debug          bool
	logger         Logger
}

func _newDoOpt(opts ...doOption) *_doOpt {
	opt := new(_doOpt)
	for _, fn := range opts {
		if fn == nil {
			continue
		}
		fn(opt)
	}
	return opt
}

type doOption func(*_doOpt)

func withDoAPIDomain(domain string) doOption {
	return func(opt *_doOpt) {
		opt.apiDomain = domain
	}
}

func withDoAPIName(name string) doOption {
	return func(opt *_doOpt) {
		opt.apiName = name
	}
}

func withDoHeader(header map[string]string) doOption {
	return func(opt *_doOpt) {
		opt.header = header
	}
}

func withDoQuery(query map[string]string) doOption {
	return func(opt *_doOpt) {
		if len(opt.query) == 0 {
			opt.query = query
			return
		}
		for k, v := range query {
			opt.query[k] = v
		}
	}
}

func withDoQueryKV(k, v string) doOption {
	return func(opt *_doOpt) {
		if len(opt.query) == 0 {
			opt.query = map[string]string{k: v}
			return
		}
		opt.query[k] = v
	}
}

func withDoUploadFormData(name string, value io.Reader) doOption {
	return func(opt *_doOpt) {
		if opt.uploadFormData == nil {
			opt.uploadFormData = make([]*_doFormData, 0, 8)
		}
		opt.uploadFormData = append(opt.uploadFormData, newDoFormData(name, value))
	}
}

func withDoUploadFormFile(name, filename string, needToReadFile bool, data io.Reader) doOption {
	return func(opt *_doOpt) {
		if opt.uploadFormData == nil {
			opt.uploadFormData = make([]*_doFormData, 0, 8)
		}
		opt.uploadFormData = append(opt.uploadFormData, newDoFormDataWithFile(name, filename, needToReadFile, data))
	}
}

func withDoHTTPCli(cli *http.Client) doOption {
	return func(opt *_doOpt) {
		opt.httpCli = cli
	}
}

func withDoDebug(b bool) doOption {
	return func(opt *_doOpt) {
		opt.debug = b
	}
}

func withDoLogger(logger Logger) doOption {
	return func(opt *_doOpt) {
		opt.logger = logger
	}
}

func (opt *_doOpt) debugLog(msg string) {
	if opt.logger == nil || !opt.debug {
		return
	}
	opt.logger.Debug("[FEISHU-DEBUG] " + msg)
}

func (opt *_doOpt) NewRequest(ctx context.Context, method, rawURL string, data interface{}) (req *http.Request, err error) {
	if len(opt.query) != 0 {
		tmp, err := url.Parse(rawURL)
		if err != nil {
			return nil, err
		}
		q := tmp.Query()
		for k, v := range opt.query {
			q.Set(k, v)
		}
		tmp.RawQuery = q.Encode()
		rawURL = tmp.String()
	}

	if len(opt.uploadFormData) == 0 {
		// 普通请求

		buf := new(bytes.Buffer)
		if err = json.NewEncoder(buf).Encode(data); err != nil {
			return nil, err
		}
		if data == nil {
			opt.debugLog(fmt.Sprintf("--> [%s - %s] %s %s\n", opt.apiDomain, opt.apiName, method, rawURL))
		} else {
			opt.debugLog(fmt.Sprintf("--> [%s - %s] %s %s\n%s", opt.apiDomain, opt.apiName, method, rawURL, buf.String()))
		}

		req, err = http.NewRequestWithContext(ctx, method, rawURL, buf)
	} else {
		// 上传请求
		req, err = opt.newUploadReq(ctx, method, rawURL)
	}
	if err != nil {
		return nil, err
	}

	for k, v := range opt.header {
		req.Header.Set(k, v)
	}

	return
}

func (opt *_doOpt) newUploadReq(ctx context.Context, method, rawURL string) (req *http.Request, err error) {
	bufBody := new(bytes.Buffer)
	writer := multipart.NewWriter(bufBody)

	var uploadedName string
	for _, fd := range opt.uploadFormData {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fd.headerValue)
		part, err := writer.CreatePart(h)
		if err != nil {
			return nil, err
		}
		if fd.filename != "" {
			uploadedName = fd.filename
		}
		if fd.needToReadFile {
			// 通过文件路径上传
			imgFile, err := os.Open(fd.filename)
			if err != nil {
				return nil, err
			}

			if _, err = io.Copy(part, imgFile); err != nil {
				return nil, err
			}

			if err := imgFile.Close(); err != nil {
				opt.debugLog(fmt.Sprintf("--> [%s - %s] %s %s\nfilename: %s: unexpected close: %s", opt.apiDomain, opt.apiName, method, rawURL, fd.filename, err))
			}

		} else {
			if _, err = io.Copy(part, fd.data); err != nil {
				return nil, err
			}
		}
	}

	if err = writer.Close(); err != nil {
		return nil, err
	}

	opt.debugLog(fmt.Sprintf("--> [%s - %s] %s %s\nfilename: %s", opt.apiDomain, opt.apiName, method, rawURL, uploadedName))

	if req, err = http.NewRequestWithContext(ctx, method, rawURL, bufBody); err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return
}

func _doWithContext(ctx context.Context, method, rawURL string, data interface{}, opts ...doOption) (reqID string, respReader io.Reader, err error) {
	doOpt := _newDoOpt(opts...)

	var req *http.Request
	if req, err = doOpt.NewRequest(ctx, method, rawURL, data); err != nil {
		return "", nil, err
	}

	tmpCli := DefaultHTTPClient
	if doOpt.httpCli != nil {
		tmpCli = doOpt.httpCli
	}
	tmpCli.Timeout = 0

	start := time.Now()

	var resp *http.Response
	if resp, err = tmpCli.Do(req); err != nil {
		return "", nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	reqID = resp.Header.Get("X-Request-Id")

	respBuf := new(bytes.Buffer)
	_, err = io.Copy(respBuf, resp.Body)
	doOpt.debugLog(fmt.Sprintf("<-- [%s - %s] %s %s %d %s\n%s\n", doOpt.apiDomain, doOpt.apiName, method, rawURL, resp.StatusCode, time.Since(start), respBuf.String()))
	if err != nil {
		return reqID, nil, err
	}

	return reqID, respBuf, nil
}

type _doFormData struct {
	headerValue    string
	data           io.Reader
	needToReadFile bool
	filename       string
}

func newDoFormData(name string, data io.Reader) *_doFormData {
	return &_doFormData{
		headerValue:    fmt.Sprintf(`form-data; name="%s"`, name),
		data:           data,
		needToReadFile: false,
	}
}

func newDoFormDataWithFile(name, filename string, needToReadFile bool, data io.Reader) *_doFormData {
	return &_doFormData{
		headerValue:    fmt.Sprintf(`form-data; name="%s"; filename="%s"`, name, escapeQuotes(path.Base(filename))),
		data:           data,
		needToReadFile: needToReadFile,
		filename:       filename,
	}
}

// func (dd *_doFormData) AppendFilename(needToReadFile bool) {
// 	dd.needToReadFile = needToReadFile
// 	// form-data; name="%s"; filename="%s"
// 	dd.headerValue += fmt.Sprintf(`; filename="%s"`, escapeQuotes(path.Base(dd.filename)))
// }

// func newDoFormDataWithName(name string, data io.Reader) *_doFormData {
// 	return &_doFormData{
// 		headerValue: fmt.Sprintf(`form-data; name="%s"`, name),
// 		data:   data,
// 	}
// }
//
// func newDoFormDataWithNameAndFilename(name, filename string, data io.Reader) *_doFormData {
// 	return &_doFormData{
// 		headerValue: fmt.Sprintf(`form-data; name="%s"; filename="%s"`, name, escapeQuotes(path.Base(filename))),
// 		data:   data,
// 	}
// }

// func (dd *_doFormData) SetName(name string, data io.Reader) {
// 	dd.headerValue = fmt.Sprintf(`form-data; name="%s"`, name)
// 	dd.data = data
// }
//
// func (dd *_doFormData) SetNameFilename(name, filename string, data io.Reader) {
// 	dd.headerValue = fmt.Sprintf(`form-data; name="%s"; filename="%s"`, name, escapeQuotes(path.Base(filename)))
// 	dd.data = data
// }

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}
