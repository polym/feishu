package feishu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"

	"github.com/polym/feishu/accesstoken"
	"github.com/polym/feishu/internal/transport"
)

type codeMessageWrap struct {
	Code    int             `json:"code"`
	Message string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
}

type App struct {
	appId     string
	appSecret string
	http      *http.Client
	tenant    string
}

func NewApp(appId, appSecret string) (*App, error) {
	return &App{
		appId:     appId,
		appSecret: appSecret,
		http:      http.DefaultClient,
	}, nil
}

func (app *App) GetTenantAccessToken() (string, error) {
	if app.tenant == "" {
		res, err := accesstoken.GetTenantAccessToken(app, accesstoken.GetTenantAccessTokenOptions{
			AppId: app.appId, AppSecret: app.appSecret,
		})
		if err != nil {
			return "", err
		}
		app.tenant = res.AccessToken
	}
	return app.tenant, nil
}

func (app *App) Do(req *transport.Request, v interface{}, kind transport.AuthKind) error {
	var (
		body    io.Reader
		headers = make(map[string]string)
	)
	switch req.ContentType {
	case "application/json":
		if req.BodySpec != nil {
			content, err := json.Marshal(req.BodySpec)
			if err != nil {
				return err
			}
			body = bytes.NewReader(content)
		}
	case "multipart/form-data":
		if len(req.Params) > 0 {
			formBody := &bytes.Buffer{}
			formWriter := multipart.NewWriter(formBody)
			defer formWriter.Close()

			fieldname, filename := "", ""
			for k, v := range req.Params {
				if strings.HasPrefix(k, "&") {
					filename = filepath.Base(v)
					fieldname = strings.TrimPrefix(k, "&")
					continue
				}
				formWriter.WriteField(k, v)
			}

			if fieldname == "" {
				body = formBody
				break
			}

			boundary := formWriter.Boundary()
			bdBuf := bytes.NewBufferString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

			_, err := formWriter.CreateFormFile(fieldname, filename)
			if err != nil {
				return err
			}

			headers["Content-Type"] = "multipart/form-data; boundary=" + boundary
			body = io.MultiReader(formBody, req.Body, bdBuf)
		}
	}

	r, err := http.NewRequest(req.Method, req.Url, body)
	if err != nil {
		return fmt.Errorf("%s %s: %w", req.Method, req.Url, err)
	}
	for k, v := range headers {
		r.Header.Set(k, v)
	}

	if err := app.do(r, v, kind); err != nil {
		return fmt.Errorf("%s %s: %w", req.Method, req.Url, err)
	}

	return nil
}

func (app *App) do(req *http.Request, v interface{}, kind transport.AuthKind) error {
	switch kind {
	case transport.AuthKindTenant:
		token, err := app.GetTenantAccessToken()
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	req.Header.Set("User-Agent", "feishu/go")

	if isDebug() {
		x, _ := httputil.DumpRequest(req, true)
		log.Println(string(x))
	}

	resp, err := app.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if isDebug() {
		x, _ := httputil.DumpResponse(resp, true)
		log.Println(string(x))
	}

	contentType := resp.Header.Get("Content-Type")
	switch {
	case strings.Contains(contentType, "application/json"):
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		c := &codeMessageWrap{}
		if err = json.Unmarshal(content, c); err != nil {
			return err
		}
		if c.Code != 0 {
			return fmt.Errorf("%d: %s", c.Code, c.Message)
		}
		if v != nil {
			if _, ok := v.(transport.EmbedData); ok {
				err = json.Unmarshal(c.Data, v)
			} else {
				err = json.Unmarshal(content, v)
			}
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupport decode %s", contentType)
	}

	return nil
}

func isDebug() bool {
	return strings.Contains(os.Getenv("DEBUG"), "sdk")
}
