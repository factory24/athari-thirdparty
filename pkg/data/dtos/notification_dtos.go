package dtos

// PushNotification represents the data needed to send a push notification
type PushNotification struct {
	To       string
	Title    string
	Content  string
	SubTitle string
	Data     map[string]interface{}
}
