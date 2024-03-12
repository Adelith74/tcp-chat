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
}

func NewChat(name string) *Chat {
	return &Chat{
		Chat_name:        name,
		Alive:            make(map[*User]string),
		Connections:      make(chan *User),
		Dead_connections: make(chan *User),
		Messages:         make(chan string),
		IsOpen:           true,
		Creator:          &User{Username: "admin", Connection: nil},
	}
}
