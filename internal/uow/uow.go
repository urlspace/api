package uow

import (
	"context"

	"github.com/urlspace/api/internal/collection"
	"github.com/urlspace/api/internal/link"
	"github.com/urlspace/api/internal/subscription"
	"github.com/urlspace/api/internal/tag"
	"github.com/urlspace/api/internal/user"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("github.com/urlspace/api/internal/uow")

// Repositories groups all repositories available within a transaction. It lives here
// because the uow package coordinates across multiple domain repositories.
// No domain package imports this package.
type Repositories struct {
	Links         link.Repository
	Tags          tag.Repository
	Collections   collection.Repository
	Users         user.Repository
	Subscriptions subscription.Repository
}

// UnitOfWork runs fn inside a single database transaction. Every repository in
// the Repositories value passed to fn executes against that transaction.
type UnitOfWork interface {
	RunInTx(ctx context.Context, fn func(Repositories) error) error
}

type Service struct {
	// repos holds repositories for operations that don't need transactional
	// guarantees.
	repos Repositories
	// uow wraps multi-repo operations in a transaction. It is the only place
	// in the codebase that runs transactions.
	uow UnitOfWork
}

func NewService(repos Repositories, uow UnitOfWork) *Service {
	return &Service{repos: repos, uow: uow}
}
