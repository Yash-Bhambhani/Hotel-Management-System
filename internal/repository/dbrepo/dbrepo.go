package dbrepo

import (
	"database/sql"
	"hotel_management_system/internal/config"
	"hotel_management_system/internal/repository"
)

type PostgresDBRepo struct {
	app *config.AppConfig
	DB  *sql.DB
}

func NewPostGresRepo(conn *sql.DB, app *config.AppConfig) repository.DatabaseRepo {
	return &PostgresDBRepo{
		app: app,
		DB:  conn,
	}
}
