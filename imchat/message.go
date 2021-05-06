package imchat

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/polym/feishu/internal/transport"
)

type Time string

func (t Time) Time() time.Time {
	v, _ := strconv.ParseInt(string(t), 10, 63)
	return time.Unix(v, 0)
}

type MessageType string

const (
	MessageTypeText MessageType = "text"
	MessageTypePost MessageType = "post"
)

type MessageTag string

const (
	MessageTagText  MessageTag = "text"
	MessageTagLink  MessageTag = "a"
	MessageTagAt    MessageTag = "at"
	MessageTagImage MessageTag = "img"
)

type ReceiverKind string

const (
	ReceiverKindOpenId  ReceiverKind = "open_id"
	ReceiverKindUserId  ReceiverKind = "user_id"
	ReceiverKindUnionId ReceiverKind = "union_id"
	ReceiverKindEmail   ReceiverKind = "email"
	ReceiverKindChatId  ReceiverKind = "chat_id"
)

type MessageWord struct {
	Tag      MessageTag `json:"tag"`
	Text     string     `json:"text"`
	Href     string     `json:"href"`
	UserId   string     `json:"user_id"`
	ImageKey string     `json:"image_key"`
	Width    int        `json:"width"`
	Height   int        `json:"height"`
}

type MessageContent interface {
	String() (string, error)
}

type MessagePostContent struct {
	Title   string          `json:"title"`
	Content [][]MessageWord `json:"content"`
}

func (mpc MessagePostContent) String() (string, error) {
	content, err := json.Marshal(map[string]interface{}{"zh_cn": mpc})
	if err != nil {
		return "", err
	}
	return string(content), nil
}

type MessageTextContent struct {
	Text string `json:"text"`
}

func (mtc MessageTextContent) String() (string, error) {
	content, err := json.Marshal(mtc)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

type Message struct {
	ReceiverKind ReceiverKind
	ReceiveId    string
	Type         MessageType
	Content      MessageContent
}

func (m Message) MarshalJSON() ([]byte, error) {
	if _, ok := m.Content.(MessageTextContent); ok {
		m.Type = MessageTypeText
	}
	if _, ok := m.Content.(MessagePostContent); ok {
		m.Type = MessageTypePost
	}
	content, err := m.Content.String()
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]interface{}{
		"receive_id": m.ReceiveId,
		"msg_type":   m.Type,
		"content":    content,
	})
}

func SendMessage(p transport.Transport, m Message) error {
	req := &transport.Request{
		Method:      "POST",
		Url:         "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=" + string(m.ReceiverKind),
		ContentType: "application/json",
		BodySpec:    m,
	}
	return p.Do(req, nil, transport.AuthKindTenant)
}

type GetChatMessagesOptions struct {
	ContainerIdType string
	ContainerId     string
	StartTime       int64
	EndTime         int64
	PageSize        int
	PageToken       string
}

type MessageDetail struct {
	MessageId  string `json:"message_id"`
	RootId     string `json:"root_id"`
	ParentId   string `json:"parent_id"`
	MsgType    string `json:"msg_type"`
	CreateTime Time   `json:"create_time"`
	UpdateTime Time   `json:"update_time"`
	Deleted    bool   `json:"deleted"`
	Updated    bool   `json:"updated"`
	ChatId     string `json:"chat_id"`
	Sender     struct {
		Id         string `json:"id"`
		IdType     string `json:"id_type"`
		SenderType string `json:"sender_type"`
	} `json:"sender"`
	Body struct {
		Content string `json:"content"`
	} `json:"body"`
	Mentions []struct {
		Key    string `json:"key"`
		Id     string `json:"id"`
		IdType string `json:"id_type"`
		Name   string `json:"name"`
	} `json:"mentions"`
	UpperMessageId string `json:"upper_message_id"`
}

type MessageList struct {
	transport.LabelEmbedData
	HasMore   bool            `json:"has_more"`
	PageToken string          `json:"page_token"`
	Items     []MessageDetail `json:"items"`
}

func GetChatMessages(p transport.Transport, opt GetChatMessagesOptions) (*MessageList, error) {
	req := &transport.Request{
		Method: "GET",
		Url: fmt.Sprintf("https://open.feishu.cn/open-apis/im/v1/messages?container_id_type=%s&container_id=%s&start_time=%d&end_time=%d&page_size=%d&page_token=%s",
			opt.ContainerIdType, opt.ContainerId, opt.StartTime, opt.EndTime, opt.PageSize, opt.PageToken),
	}
	spec := &MessageList{}
	err := p.Do(req, spec, transport.AuthKindTenant)
	if err != nil {
		return nil, err
	}
	return spec, nil
}

type GetMentionMessagesOptions struct {
	GetChatMessagesOptions
	MentionOpenId string
}

func GetMentionMessages(p transport.Transport, opt GetMentionMessagesOptions) (*MessageList, error) {
	msgList, err := GetChatMessages(p, opt.GetChatMessagesOptions)
	if err != nil {
		return nil, err
	}

	items := make([]MessageDetail, 0, len(msgList.Items))
	for _, item := range msgList.Items {
		found := false
		for _, m := range item.Mentions {
			if m.IdType == "open_id" && m.Id == opt.MentionOpenId {
				found = true
				item.Body.Content = strings.Replace(item.Body.Content, m.Key, "", -1)
				break
			}
		}
		if found {
			items = append(items, item)
		}
	}
	msgList.Items = items
	return msgList, nil
}

// https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/reply
func ReplyMessage(p transport.Transport, msgId string, m Message) error {
	req := &transport.Request{
		Method:      "POST",
		Url:         fmt.Sprintf("https://open.feishu.cn/open-apis/im/v1/messages/%s/reply", msgId),
		ContentType: "application/json",
		BodySpec:    m,
	}
	return p.Do(req, nil, transport.AuthKindTenant)
}
