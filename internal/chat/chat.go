package chat

type Chat struct {
	Chat_name        string
	Chat_id          int
	Creator          *User
	Connections      chan *User
	Messages         chan string
	Dead_connections chan *User
	Alive            map[*User]string
	IsOpen           bool
	Id               int
	//TgID is a telegram linked chat ID, which can be nil, that's why TgID is a string, I know, that sucks
	TgID string
}

func NewChat(name string, id int) *Chat {
	return &Chat{
		Chat_name:        name,
		Alive:            make(map[*User]string),
		Connections:      make(chan *User),
		Chat_id:          id,
		Dead_connections: make(chan *User),
		Messages:         make(chan string),
		IsOpen:           true,
		Creator:          &User{Username: "admin", Connection: nil},
	}
}
