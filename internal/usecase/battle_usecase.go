package usecase

import (
	"context"
	"fmt"

	"street-facs-fighter/internal/domain/entity"
	"street-facs-fighter/internal/domain/repository"
)

type BattleUsecase struct {
	qRepo repository.QuestionRepository
}

func NewBattleUsecase(qRepo repository.QuestionRepository) *BattleUsecase {
	return &BattleUsecase{qRepo: qRepo}
}

func (u *BattleUsecase) GetStageQuestions(ctx context.Context, count int) ([]*entity.Question, error) {
	questions, err := u.qRepo.FindRandom(ctx, count)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stage targets: %w", err)
	}
	return questions, nil
}
