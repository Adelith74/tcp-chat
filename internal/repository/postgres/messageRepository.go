package postgres

import (
	"context"
	"fmt"
	"go_chat/internal/core/interface/repository"
	"go_chat/internal/lib/db"
	"go_chat/internal/model"
	"go_chat/internal/repository/dbModel"
	"log"
)

type _msgRepository struct {
	db *db.Db
}

func NewMsgRepo(db *db.Db) repository.MsgRepository {
	return _msgRepository{db}
}

// adds a message to db
func (msgRepository _msgRepository) CreateMsg(ctx context.Context, msg model.Message) (int, error) {
	msgDb := dbModel.Message(msg)
	var id int

	err := msgRepository.db.PgConn.QueryRow(ctx,
		`INSERT INTO public.message(chat_id, message, username, time) values ($1,$2,$3,$4) RETURNING id`,
		msgDb.Chat_id,
		msgDb.Message,
		msgDb.Username,
		msgDb.Time).Scan(&id)

	if err != nil {
		log.Fatal("Error during logging messages into DB")
	} else {
		msgRepository.db.PgConn.Query(ctx, `COMMIT`)
	}

	return id, err
}

// gets all messages from all chats
func (msgRepository _msgRepository) GetMsgs(ctx context.Context) ([]model.Message, error) {

	var messages []model.Message

	var rows, _ = msgRepository.db.PgConn.Query(ctx,
		`SELECT m.chat_id, m,message, m.username, m.time FROM public.message m`)

	for rows.Next() {
		var msg = dbModel.Message{}
		err := rows.
			Scan(&msg.Chat_id, &msg.Message, &msg.Username, &msg.Time)
		if err != nil {
			return []model.Message{}, fmt.Errorf("can't get messages: %s", err.Error())
		}
		messages = append(messages, model.Message(msg))
	}

	return messages, nil
}

// gets chosen amount of messages from db with specified chat_id
func (msgRepository _msgRepository) GetMsgsByChatId(ctx context.Context, chatId int, amount int) ([]model.Message, error) {
	var messages []model.Message

	var rows, _ = msgRepository.db.PgConn.Query(ctx,
		`SELECT m.chat_id, m,message, m.username, m.time FROM public.message m WHERE m.chat_id=$1 order by desc limit $2`,
		chatId, amount)

	for rows.Next() {
		var msg = dbModel.Message{}
		err := rows.
			Scan(&msg.Chat_id, &msg.Message, &msg.Username, &msg.Time)
		if err != nil {
			return []model.Message{}, fmt.Errorf("can't get messages: %s", err.Error())
		}
		messages = append(messages, model.Message(msg))
	}

	return messages, nil
}
