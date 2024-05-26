package postgres

import (
	"context"
	"fmt"
	"go_chat/internal/core/interface/repository"
	"go_chat/internal/lib/db"
	"golang.org/x/crypto/bcrypt"
)

type userDB struct {
	Login    string `db:"login"`
	Password string `db:"password"`
}

type _authRepo struct {
	*db.Db
}

func NewRepo(db *db.Db) repository.AuthRepository {
	return _authRepo{db}
}

func (repo _authRepo) GetUser(ctx context.Context, login, hashPassword string) (string, error) {
	var user userDB

	row := repo.PgConn.QueryRow(ctx, `SELECT * FROM public.users WHERE login=$1 AND pas=$2`, login, hashPassword)

	if err := row.Scan(&user); err != nil {
		return "", fmt.Errorf("couldn't get user: %x", err)
	}

	return login, nil

}

func (repo _authRepo) GetUserByLogin(ctx context.Context, login string) (string, error) {
	var user userDB

	row := repo.PgConn.QueryRow(ctx, `SELECT * FROM public.users WHERE login=$`, login)

	if err := row.Scan(&user); err != nil {
		return "", fmt.Errorf("couldn't get user: %x", err)
	}

	return login, nil

}

func (repo _authRepo) EncodePassword(ctx context.Context, password string) (string, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (repo _authRepo) Register(ctx context.Context, login, hashPassword string) (string, error) {

	_, err := repo.GetUserByLogin(ctx, login)

	if err != nil {
		return "", fmt.Errorf("user with this login already exists")
	}

	_, err = repo.PgConn.Exec(
		ctx,
		`INSERT INTO public.users(username, pass) values ($1, $2)`,
		login, hashPassword,
	)

	if err != nil {
		return "", fmt.Errorf("couldn't create: %x", err)
	}

	return login, nil
}
