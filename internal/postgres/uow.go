package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/urlspace/api/internal/db"
	"github.com/urlspace/api/internal/uow"
)

type unitOfWork struct {
	pool *pgxpool.Pool
}

func NewUnitOfWork(pool *pgxpool.Pool) uow.UnitOfWork {
	return &unitOfWork{pool: pool}
}

func (u *unitOfWork) RunInTx(ctx context.Context, fn func(uow.Repositories) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txQueries := db.New(tx)
	repos := uow.Repositories{
		Links:       NewLinkRepository(txQueries),
		Tags:        NewTagRepository(txQueries),
		Collections: NewCollectionRepository(txQueries),
	}

	if err := fn(repos); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
