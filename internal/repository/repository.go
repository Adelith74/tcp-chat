package repository

import (
	"go_chat/internal/core/interface/repository"
	"go_chat/internal/lib/db"
	"go_chat/internal/repository/postgres"
)

type RepositoryManager struct {
	repository.AuthRepository
	repository.ChatRepository
	repository.MsgRepository
}

func NewRepositoryManager(db *db.Db, host string) RepositoryManager {
	return RepositoryManager{
		postgres.NewRepo(db),
		postgres.NewChatRepo(db),
		postgres.NewMsgRepo(db),
	}
}
