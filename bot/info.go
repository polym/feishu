package bot

import (
	"github.com/polym/feishu/internal/transport"
)

type BotStatus int

const (
	BotStatusDisabled BotStatus = iota
	BotStatusEnabled
	BotStatusInstalledNotEnabled
	BotStatusUpdatedNotEnabled
	BotStatusLicenseExpiredDisabled
	BotStatusPkgExpired
)

type BotInfoSpec struct {
	ActivateStatus BotStatus `json:"activate_status"`
	AppName        string    `json:"app_name"`
	AvatarUrl      string    `json:"avatar_url"`
	IPWhiteList    []string  `json:"ip_white_list"`
	OpenId         string    `json:"open_id"`
}

func GetBotInfo(p transport.Transport) (*BotInfoSpec, error) {
	req := &transport.Request{
		Method: "GET",
		Url:    "https://open.feishu.cn/open-apis/bot/v3/info",
	}
	wp := struct {
		Info *BotInfoSpec `json:"bot"`
	}{}
	err := p.Do(req, &wp, transport.AuthKindTenant)
	if err != nil {
		return nil, err
	}
	return wp.Info, nil
}
