package imchat

import (
	"fmt"
	"io"
	"os"

	"github.com/polym/feishu/internal/transport"
)

type UploadImageOptions struct {
	LocalPath string
	Reader    io.Reader
}

type UploadImageResp struct {
	transport.LabelEmbedData
	ImageKey string `json:"image_key"`
}

// https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/image/create
func UploadImage(p transport.Transport, opt UploadImageOptions) (*UploadImageResp, error) {
	if opt.LocalPath != "" {
		fd, err := os.Open(opt.LocalPath)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", opt.LocalPath, err)
		}
		defer fd.Close()
		opt.Reader = fd
	}
	if opt.Reader == nil {
		return nil, fmt.Errorf("upload no content")
	}

	req := &transport.Request{
		Method: "POST",
		Url:    "https://open.feishu.cn/open-apis/im/v1/images",
		Params: map[string]string{
			"image_type": "message",
			"&image":     opt.LocalPath,
		},
		ContentType: "multipart/form-data",
		Body:        opt.Reader,
	}

	resp := new(UploadImageResp)
	err := p.Do(req, resp, transport.AuthKindTenant)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
