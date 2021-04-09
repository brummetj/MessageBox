// Data models for the MessageBox service

package model

type User struct {
	UserName string 	`json:"username"`
	Messages []Message  `json:"messages"`
	Groups []string		`json:"groups"`
}

type UserRegistration struct {
	UserName string 	`json:"username" validate:"optional, alphanum, min=1, max=24"`
}

type GroupRecipient struct {
	Groupname string 	`json:"groupname"`
}

type MessageRecipient struct {
	UserName  string 	`json:"username"`
	GroupName string 	`json:"groupname"`
}

type GroupCreation struct {
	Groupname string 	`json:"groupname"`
	Usernames []string 	`json:"usernames"`
}

type ComposedMessage struct {
	Sender    string           `json:"sender"`
	Recipient MessageRecipient `json:"recipient"`
	Subject   string           `json:"subject"`
	Body      string           `json:"body"`
	Id		  int		   	   `json:"id"`
}

type ReplyMessage struct {
	Sender  string		`json:"sender"`
	Subject string		`json:"subject"`
	Body    string		`json:"body"`
}

type Message struct {
	Id		   int					`json:"id"`
	Re 		   int					`json:"re"`
	Sender     string				`json:"sender"`
	Recipient  MessageRecipient		`json:"recipient"`
	Subject	   string				`json:"subject"`
	Body	   string				`json:"body"`
	SentAt	   string				`json:"sentAt"`
}
