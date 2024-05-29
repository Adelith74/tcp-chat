package dbModel

type Chat struct {
	Chat_name  string `db:"chat_name"`
	Chat_id    int    `db:"chat_id"`
	Creator    string `db:"creator"`
	IsOpen     bool   `db:"is_open"`
	Creator_id int    `db:"creator_id"`
	TgID       string `db:"tgchatid"`
}
