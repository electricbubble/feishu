package feishu

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// 名称: [消息与群组] 上传图片
// Func: [api_messenger.go] UploadImage
//
// 描述: 上传图片接口，可以上传 JPEG、PNG、WEBP、GIF、TIFF、BMP、ICO格式图片
// Info: 需要开启机器人能力
// Info: 上传的图片大小不能超过10MB
//
// Doc: https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/image/create
//
// 自建应用: true
// 商店应用: true
//
// HTTP URL: /open-apis/im/v1/images
// HTTP Method: POST
//
// 请求头: Authorization=Bearer {{AppAccessToken}}
// 请求头: Content-Type=multipart/form-data; boundary=---7MA4YWxkTrZu0gW
//
type uploadImageResponse struct {
	fsResponse

	Data struct {
		ImageKey string `json:"image_key"`
	} `json:"data"`
}

type UploadImageOption = doOption

func WithUploadImage(filename string) UploadImageOption {
	return withDoUploadFormFile("image", filename, true, nil)
}

func WithUploadImageViaReader(filename string, src io.Reader) UploadImageOption {
	return withDoUploadFormFile("image", filename, false, src)
}

func (a *app) UploadImage(src UploadImageOption) (imageKey string, err error) {
	return a.UploadImageWithContext(context.Background(), src)
}

func (a *app) UploadImageWithContext(ctx context.Context, src UploadImageOption) (imageKey string, err error) {
	apiDomain := "消息与群组"
	apiName := "上传图片"
	urlSuffix := "/open-apis/im/v1/images"

	if !a.isSupported(true, true) {
		return "", fmt.Errorf(_fmtErrNotSupported, apiDomain, apiName)
	}

	header := map[string]string{
		"Authorization": "Bearer ",
	}
	if accessToken, err := a.getAppAccessTokenWithContext(ctx); err != nil {
		return "", err
	} else {
		header["Authorization"] = fmt.Sprintf("Bearer %s", accessToken)
	}
	doOpts := a.buildOpts(apiDomain, apiName, header,
		withDoUploadFormData("image_type", strings.NewReader("message")),
		src,
	)
	reqID, reader, err := a._doUploadWithContext(ctx, urlSuffix, doOpts...)
	if err != nil {
		return "", err
	}

	resp := new(uploadImageResponse)
	if err = a._decodeResp(apiDomain, apiName, reader, resp); err != nil {
		return "", err
	}

	if err = resp.check(reqID, apiDomain, apiName); err != nil {
		return "", err
	}

	return resp.Data.ImageKey, nil
}
