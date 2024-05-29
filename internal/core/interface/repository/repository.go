package repository

import (
	"context"
	"go_chat/internal/model"
)

type AuthRepository interface {
	GetUser(ctx context.Context, login, hashPassword string) (string, error)
	Register(ctx context.Context, login, hashPassword string) (string, error)
	EncodePassword(ctx context.Context, password string) (string, error)
}

type ChatRepository interface {
	CreateChat(ctx context.Context, chat model.Chat) (int, error)
	GetChat(ctx context.Context, chatId int) (model.Chat, error)
	DeleteChat(ctx context.Context, chatId int) error
	GetChats(ctx context.Context) ([]model.Chat, error)
	GetId(ctx context.Context) (int, error)
	LinkChat(ctx context.Context, tgID int, internalID int) error
	GetIdWithTgID(ctx context.Context, tgID int64) (model.Chat, error)
}

type MsgRepository interface {
	CreateMsg(ctx context.Context, msg model.Message) (int, error)
	GetMsgs(ctx context.Context) ([]model.Message, error)
	GetMsgsByChatId(ctx context.Context, chatId int, amount int) ([]model.Message, error)
}

//type EventRepository interface {
//	SendEvent(ctx context.Context, event model.Event) error
//}
