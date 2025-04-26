package groupme

type Message struct {
	Id          string        `json:"id"`
	Name        string        `json:"name"`
	AvatarUrl   string        `json:"avatar_url"`
	GroupId     string        `json:"group_id"`
	CreatedAt   int           `json:"created_at"`
	UserId      string        `json:"user_id"`
	Text        string        `json:"text"`
	Attachments []interface{} `json:"attachments"`
	System      bool          `json:"system"`
	SourceGuid  string        `json:"source_guid"`
	SenderId    string        `json:"sender_id"`
	SenderType  string        `json:"sender_type"`
}

type MessageSendRequest struct {
	BotId string `json:"bot_id"`
	Text  string `json:"text"`
}

type GetBotResponse struct {
	Response []Bot `json:"response"`
}

type Bot struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	AvatarUrl       string `json:"avatar_url"`
	GroupId         string `json:"group_id"`
	CallbackUrl     string `json:"callback_url"`
	DMNotifications bool   `json:"dm_notifications"`
	Active          bool   `json:"active"`
}
