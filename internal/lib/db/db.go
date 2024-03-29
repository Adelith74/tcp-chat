package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

const connection = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s application_name=%s"

type Db struct {
	PgConn *pgxpool.Pool
}

func New(ctx context.Context) *Db {

	connectionString := fmt.Sprintf(connection, "localhost", "5432", "postgres", "root", "tcp-chat", "disable")

	conn, err := pgxpool.New(ctx, connectionString)

	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	fmt.Println("Создалось")

	return &Db{PgConn: conn}
}
