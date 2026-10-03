package repository

import (
	"context"
	"street-facs-fighter/internal/domain/entity"
)

type QuestionRepository interface {
	Save(ctx context.Context, q *entity.Question) error
	FindAll(ctx context.Context) ([]*entity.Question, error)
	FindRandom(ctx context.Context, limit int) ([]*entity.Question, error)
}
