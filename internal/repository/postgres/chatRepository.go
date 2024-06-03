package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"go_chat/internal/core/interface/repository"
	"go_chat/internal/lib/db"
	"go_chat/internal/model"
	"go_chat/internal/repository/dbModel"
	"strconv"
)

type _chatRepository struct {
	db *db.Db
}

func NewChatRepo(db *db.Db) repository.ChatRepository {
	return &_chatRepository{db}
}

func (chatRepository *_chatRepository) CreateChat(ctx context.Context, chat model.Chat) (int, error) {
	chatDb := dbModel.Chat(chat)
	var id int

	var str_format = fmt.Sprintf("INSERT INTO public.chat(chat_name, chat_id, creator, is_open, creator_id) values ('%s', %v, '%s', %v, %v) RETURNING chat_id",
		chatDb.Chat_name,
		chatDb.Chat_id,
		chatDb.Creator,
		chatDb.IsOpen,
		chatDb.Creator_id)

	fmt.Println(str_format)

	err := chatRepository.db.PgConn.QueryRow(ctx,
		str_format).Scan(&id)

	chatRepository.db.PgConn.QueryRow(ctx, `COMMIT`)

	if err != nil {
		fmt.Println(err.Error())
	}

	return id, err
}

func (chatRepository *_chatRepository) LinkChat(ctx context.Context, tgID int, internalID int) error {
	var find = fmt.Sprintf("SELECT chat_id from chat WHERE chat_id = %v", tgID)
	var str_format = fmt.Sprintf("SELECT tgchatid from chat WHERE chat_id = %v", internalID)
	var id sql.NullInt64
	var chat_id sql.NullInt64
	err := chatRepository.db.PgConn.QueryRow(context.Background(), str_format).Scan(&id)
	if err != nil {
		return errors.New("unable to find chat with such id")
	} else {
		err = chatRepository.db.PgConn.QueryRow(context.Background(), find).Scan(&chat_id)
		if !errors.Is(err, pgx.ErrNoRows) && err != nil {
			return errors.New("one of chats already linked to this tg chat")
		}
		if id.Valid {
			return errors.New("this chat already has linked chat")
		} else {
			str_format = fmt.Sprintf("UPDATE chat set tgchatid = %v where chat_id = %v", tgID, internalID)
			_, er := chatRepository.db.PgConn.Query(context.Background(), str_format)
			if er != nil {
				return errors.New("error during update")
			} else {
				chatRepository.db.PgConn.Query(context.Background(), "COMMIT")
				return nil
			}

		}
	}
}

// GetId returns available Id
func (chatRepository *_chatRepository) GetId(ctx context.Context) (int, error) {
	id := 0
	row := chatRepository.db.PgConn.QueryRow(ctx,
		`SELECT MAX(COALESCE(c.chat_id,0)) FROM public.chat as c`)
	err := row.Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error of getting chat: %s", err.Error())
	}

	return id + 1, nil
}

func (chatRepository *_chatRepository) GetIdWithTgID(ctx context.Context, tgID int64) (model.Chat, error) {
	var chat dbModel.Chat
	var tg_id sql.NullString

	err := chatRepository.db.PgConn.QueryRow(ctx,
		`SELECT c.chat_name, c.chat_id, c.creator_id, c.is_open, c.tgchatid FROM public.chat c WHERE c.tgchatid=$1`,
		strconv.FormatInt(tgID, 10)).Scan(&chat.Chat_name, &chat.Chat_id, &chat.Creator_id, &chat.IsOpen, &tg_id)

	if err != nil {
		return model.Chat{}, fmt.Errorf("error of getting chat: %s", err.Error())
	}

	if tg_id.Valid {
		chat.TgID = tg_id.String
	} else {
		chat.TgID = ""
	}

	return model.Chat(chat), nil
}

func (chatRepository *_chatRepository) GetChats(ctx context.Context) ([]model.Chat, error) {
	var chats []model.Chat

	rows, err := chatRepository.db.PgConn.Query(ctx,
		`SELECT c.chat_name, c.chat_id, c.creator, c.is_open, c.creator_id, c.tgchatid FROM public.chat c `)
	if err != nil {
		return []model.Chat{}, fmt.Errorf("error of getting chat: %s", err.Error())
	}
	defer rows.Close()
	chat := dbModel.Chat{}
	for rows.Next() {
		var tgID sql.NullString
		err := rows.Scan(&chat.Chat_name, &chat.Chat_id, &chat.Creator, &chat.IsOpen, &chat.Creator_id, &tgID)
		if err != nil {
			fmt.Println(err)
		}
		if tgID.Valid {
			chat.TgID = tgID.String
		} else {
			chat.TgID = ""
		}
		chats = append(chats, model.Chat(chat))
	}
	return chats, nil
}

func (chatRepository *_chatRepository) GetChat(ctx context.Context, chatId int) (model.Chat, error) {
	var chat dbModel.Chat
	var tgID sql.NullString

	err := chatRepository.db.PgConn.QueryRow(ctx,
		`SELECT c.chat_name, c.chat_id, c.creator_id, c.is_open, c.tgchatid FROM public.chat c WHERE c.id=$1`,
		chatId).Scan(&chat.Chat_name, &chat.Chat_id, &chat.Creator_id, &chat.IsOpen, &tgID)

	if err != nil {
		return model.Chat{}, fmt.Errorf("error of getting chat: %s", err.Error())
	}

	if tgID.Valid {
		chat.TgID = tgID.String
		return model.Chat(chat), nil
	} else {
		chat.TgID = ""
		return model.Chat(chat), nil
	}

}

func (chatRepository *_chatRepository) DeleteChat(ctx context.Context, chatId int) error {
	var chat dbModel.Chat

	err := chatRepository.db.PgConn.QueryRow(ctx,
		`DELETE FROM public.chat c WHERE c.id=$1`,
		chatId).Scan(&chat.Chat_id)

	if err != nil {
		return fmt.Errorf("error of getting chat: %s", err.Error())
	}

	return nil
}
