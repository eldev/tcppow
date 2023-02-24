package inmemory

import (
	"context"
	"math/rand"
)

type InmemoryRepo struct {
	Wisdoms []string
}

func New() *InmemoryRepo {
	return &InmemoryRepo{
		Wisdoms: []string{
			// some examples from the Internet
			"Never stop learning, because life never stops teaching.",
			"You create your own opportunities. Success doesn’t just come and find you–you have to go out and get it.",
			"Never break your promises. Keep every promise; it makes you credible.",
			"Happiness is a choice. For every minute you are angry, you lose 60 seconds of your own happiness.",
			"Accept what is, let go of what was, have faith in what will be. Sometimes you have to let go to let new things come in.",
			"When you don’t know, don’t speak as if you do. If you don’t know, simply don’t speak.",
			"Build on your strengths. The struggle you are in today is developing the strength you need for tomorrow.",
			"Your attitude will influence your experience. How you respond is at least as important as what happens to you.",
			"Always do what you are afraid to do",
			"Believe and act as if it were impossible to fail",
		},
	}
}

func (r *InmemoryRepo) GetWordOfWisdom(ctx context.Context) (string, error) {
	randIdx := rand.Int() % len(r.Wisdoms)
	return r.Wisdoms[randIdx], nil
}
