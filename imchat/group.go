package imchat

import (
	"github.com/polym/feishu/internal/transport"
)

func GetGroupList(p transport.Transport) error {
	req := &transport.Request{
		Method:      "GET",
		Url:         "https://open.feishu.cn/open-apis/im/v1/chats",
		ContentType: "application/json",
	}
	return p.Do(req, nil, transport.AuthKindTenant)
}
