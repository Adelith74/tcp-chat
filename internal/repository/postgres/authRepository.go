package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"go_chat/internal/core/interface/repository"
	"go_chat/internal/lib/db"
)

type userDB struct {
	Login    string `db:"username"`
	Password string `db:"hash_pass"`
}

type _authRepo struct {
	*db.Db
}

func NewRepo(db *db.Db) repository.AuthRepository {
	return _authRepo{db}
}

func (repo _authRepo) GetUser(ctx context.Context, login, hashPassword string) (string, error) {
	var user userDB

	err := repo.PgConn.QueryRow(ctx, `SELECT username, hash_pass FROM public.users WHERE username=$1 AND hash_pass=$2`, login, hashPassword).Scan(&user.Login, &user.Password)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("couldn't get user: %x", err)
	} else if err != nil {
		return "", fmt.Errorf("couldn't get user: %x", err)
	} else if user.Login == "" && user.Password == "" {
		return "", fmt.Errorf("couldn't get user: %x", err)
	}
	return login, nil
}

func (repo _authRepo) GetUserByLogin(ctx context.Context, login string) (string, error) {
	var user userDB
	row := repo.PgConn.QueryRow(ctx, `SELECT * FROM public.users WHERE username=$1`, login)

	err := row.Scan(&user)
	if errors.Is(err, pgx.ErrNoRows) {
		return login, nil
	} else {
		return "", fmt.Errorf("couldn't get user: %x", err)
	}

}

func (repo _authRepo) EncodePassword(ctx context.Context, password string) (string, error) {

	//hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	//if err != nil {
	//	return "", err
	//}
	return password, nil
}

func (repo _authRepo) Register(ctx context.Context, login, hashPassword string) (string, error) {

	_, err := repo.GetUserByLogin(ctx, login)

	if err != nil {
		return "", fmt.Errorf("user with this login already exists")
	}

	_, err = repo.PgConn.Exec(
		ctx,
		`INSERT INTO public.users(username, hash_pass) values ($1, $2)`,
		login, hashPassword,
	)

	if err != nil {
		return "", fmt.Errorf("couldn't create: %x", err)
	}

	return login, nil
}
