package repository

import "context"

type Repository interface {
	GetWordOfWisdom(ctx context.Context) (string, error)
}
