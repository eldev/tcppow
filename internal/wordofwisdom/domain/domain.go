package domain

import (
	"context"
	"tcppow/internal/wordofwisdom/repository"
)

type Model struct {
	wisdomRepo repository.Repository
}

func New(wisdomRepo repository.Repository) *Model {
	return &Model{
		wisdomRepo: wisdomRepo,
	}
}

func (m *Model) GetWordOfWisdom(ctx context.Context) (string, error) {
	return m.wisdomRepo.GetWordOfWisdom(ctx)
}
