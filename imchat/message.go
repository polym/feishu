package imchat

import (
	"encoding/json"

	"github.com/polym/feishu/internal/transport"
)

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
	Text    string          `json:"text"`
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
