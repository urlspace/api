package store

import (
	"database/sql"

	"github.com/hreftools/api/internal/db"
)

type Store struct {
	Resources ResourceStore
	Users     UserStore
}

func New(pool *sql.DB) *Store {
	queries := db.New(pool)

	return &Store{
		Resources: NewResourceStore(queries),
		Users:     NewUserStore(queries),
	}
}
