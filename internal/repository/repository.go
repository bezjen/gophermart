//go:generate mockery --name=Repository --output=../mocks --outpkg=mocks --case=underscore
package repository

import (
	"context"
)

type Repository interface {
	Ping(ctx context.Context) error
	Close() error
}
