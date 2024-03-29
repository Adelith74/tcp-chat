package postgres

import (
	"context"
	"fmt"
	"go_chat/internal/core/interface/repository"
	"go_chat/internal/lib/db"
	"go_chat/internal/model"
	"go_chat/internal/repository/dbModel"
)

type _chatRepository struct {
	db *db.Db
}

func NewChatRepo(db *db.Db) repository.ChatRepository {
	return _chatRepository{db}
}

func (chatRepository _chatRepository) CreateChat(ctx context.Context, chat model.Chat) (int, error) {
	chatDb := dbModel.Chat(chat)
	var id int

	var str_format = fmt.Sprintf("INSERT INTO public.chat(chat_name, chat_id, creator, is_open) values (%s,%s,%s,%s) RETURNING chat_id",
		chatDb.Chat_name,
		chatDb.Chat_id,
		chatDb.Creator,
		chatDb.IsOpen)

	fmt.Println(str_format)

	err := chatRepository.db.PgConn.QueryRow(ctx,
		`INSERT INTO public.chat(chat_name, creator, is_open) values ($1,$2,$3) RETURNING chat_id`,
		chatDb.Chat_name,
		chatDb.Creator,
		chatDb.IsOpen).Scan(&id)

	chatRepository.db.PgConn.QueryRow(ctx, `COMMIT`)

	if err != nil {
		fmt.Println(err.Error())
	}

	return id, err
}

// returns available Id
func (postRepository _chatRepository) GetId(ctx context.Context) (int, error) {
	id := 0
	err := postRepository.db.PgConn.QueryRow(ctx,
		`SELECT MAX(c.chat_id) FROM public.chat as c`).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка получения чата: %s", err.Error())
	}

	return id, nil
}

func (postRepository _chatRepository) GetChat(ctx context.Context, chatId int) (model.Chat, error) {
	var chat dbModel.Chat

	err := postRepository.db.PgConn.QueryRow(ctx,
		`SELECT c.chat_name, c.chat_id, c.creator, c.is_open FROM public.chat c WHERE c.id=$1`,
		chatId).Scan(&chat.Chat_name, &chat.Chat_id, &chat.Creator, &chat.IsOpen)

	if err != nil {
		return model.Chat{}, fmt.Errorf("ошибка получения чата: %s", err.Error())
	}

	return model.Chat(chat), nil
}

func (postRepository _chatRepository) DeleteChat(ctx context.Context, chatId int) error {
	var chat dbModel.Chat

	err := postRepository.db.PgConn.QueryRow(ctx,
		`DELETE FROM public.chat c WHERE c.id=$1`,
		chatId).Scan(&chat.Chat_id)

	if err != nil {
		return fmt.Errorf("ошибка получения чата: %s", err.Error())
	}

	return nil
}
