package usecase

import (
	"context"
	"fmt"
	"os"

	"street-facs-fighter/internal/domain/entity"
	"street-facs-fighter/internal/domain/repository"
)

type BattleUsecase struct {
	qRepo repository.QuestionRepository
}

func NewBattleUsecase(qRepo repository.QuestionRepository) *BattleUsecase {
	return &BattleUsecase{qRepo: qRepo}
}

// GetStageQuestions は実際に assets ディレクトリに画像が存在する有効な問題のみを抽出する
func (u *BattleUsecase) GetStageQuestions(ctx context.Context, count int) ([]*entity.Question, error) {
	// 余裕を持って多めに取得し、ファイル存在チェックをパスしたものを count 件採用
	candidates, err := u.qRepo.FindRandom(ctx, count*3)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stage targets: %w", err)
	}

	var validQuestions []*entity.Question
	for _, q := range candidates {
		if len(q.Targets) == 0 {
			continue
		}
		if _, err := os.Stat(q.FilePath); err == nil {
			validQuestions = append(validQuestions, q)
			if len(validQuestions) >= count {
				break
			}
		}
	}

	if len(validQuestions) == 0 {
		return nil, fmt.Errorf("no playable stage questions found with existing image files")
	}

	return validQuestions, nil
}
