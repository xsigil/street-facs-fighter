package sqlite3

import (
	"context"
	"database/sql"
	"fmt"

	"street-facs-fighter/internal/domain/entity"

	"github.com/jmoiron/sqlx"
)

type questionModel struct {
	ID       string `db:"id"`
	FilePath string `db:"file_path"`
	FileName string `db:"file_name"`
}

type targetModel struct {
	ID         int64          `db:"id"`
	QuestionID string         `db:"question_id"`
	Side       sql.NullString `db:"side"`
	Code       string         `db:"code"`
	Intensity  sql.NullString `db:"intensity"`
}

type QuestionRepository struct {
	db *sqlx.DB
}

func NewQuestionRepository(db *sqlx.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

func (r *QuestionRepository) Save(ctx context.Context, q *entity.Question) error {
	conn := GetExtContext(ctx, r.db)

	// questions テーブルへ INSERT / UPDATE
	qQuery := `INSERT INTO questions (id, file_path, file_name) VALUES (:id, :file_path, :file_name)
              ON CONFLICT(id) DO UPDATE SET file_path = excluded.file_path, file_name = excluded.file_name`
	qm := questionModel{
		ID:       q.ID,
		FilePath: q.FilePath,
		FileName: q.FileName,
	}
	if _, err := sqlx.NamedExecContext(ctx, conn, qQuery, qm); err != nil {
		return fmt.Errorf("failed to save question: %w", err)
	}

	// 既存のターゲット情報を一度削除して再作成
	delQuery := `DELETE FROM question_targets WHERE question_id = ?`
	if _, err := conn.ExecContext(ctx, delQuery, q.ID); err != nil {
		return fmt.Errorf("failed to clean up old targets: %w", err)
	}

	// question_targets の一括/順次 INSERT
	tQuery := `INSERT INTO question_targets (question_id, side, code, intensity) VALUES (?, ?, ?, ?)`
	for _, target := range q.Targets {
		var side sql.NullString
		if target.Side != "" {
			side = sql.NullString{String: target.Side, Valid: true}
		}
		var intensity sql.NullString
		if target.Intensity != "" {
			intensity = sql.NullString{String: target.Intensity, Valid: true}
		}

		if _, err := conn.ExecContext(ctx, tQuery, q.ID, side, target.Code, intensity); err != nil {
			return fmt.Errorf("failed to insert target (%s): %w", target.Code, err)
		}
	}

	return nil
}

func (r *QuestionRepository) FindAll(ctx context.Context) ([]*entity.Question, error) {
	conn := GetExtContext(ctx, r.db)

	var qModels []questionModel
	if err := sqlx.SelectContext(ctx, conn, &qModels, `SELECT id, file_path, file_name FROM questions`); err != nil {
		return nil, fmt.Errorf("failed to fetch questions: %w", err)
	}

	var questions []*entity.Question
	for _, qm := range qModels {
		var tModels []targetModel
		tQuery := `SELECT id, question_id, side, code, intensity FROM question_targets WHERE question_id = ?`
		if err := sqlx.SelectContext(ctx, conn, &tModels, tQuery, qm.ID); err != nil {
			return nil, fmt.Errorf("failed to fetch targets for question %s: %w", qm.ID, err)
		}

		var targets []entity.TargetAU
		for _, tm := range tModels {
			targets = append(targets, entity.TargetAU{
				Side:      tm.Side.String,
				Code:      tm.Code,
				Intensity: tm.Intensity.String,
			})
		}

		questions = append(questions, &entity.Question{
			ID:       qm.ID,
			FilePath: qm.FilePath,
			FileName: qm.FileName,
			Targets:  targets,
		})
	}

	return questions, nil
}

func (r *QuestionRepository) FindRandom(ctx context.Context, limit int) ([]*entity.Question, error) {
	conn := GetExtContext(ctx, r.db)

	var qModels []questionModel
	query := `SELECT id, file_path, file_name FROM questions ORDER BY RANDOM() LIMIT ?`
	if err := sqlx.SelectContext(ctx, conn, &qModels, query, limit); err != nil {
		return nil, fmt.Errorf("failed to fetch random questions: %w", err)
	}

	var questions []*entity.Question
	for _, qm := range qModels {
		var tModels []targetModel
		tQuery := `SELECT id, question_id, side, code, intensity FROM question_targets WHERE question_id = ?`
		if err := sqlx.SelectContext(ctx, conn, &tModels, tQuery, qm.ID); err != nil {
			return nil, fmt.Errorf("failed to fetch targets for question %s: %w", qm.ID, err)
		}

		var targets []entity.TargetAU
		for _, tm := range tModels {
			targets = append(targets, entity.TargetAU{
				Side:      tm.Side.String,
				Code:      tm.Code,
				Intensity: tm.Intensity.String,
			})
		}

		questions = append(questions, &entity.Question{
			ID:       qm.ID,
			FilePath: qm.FilePath,
			FileName: qm.FileName,
			Targets:  targets,
		})
	}

	return questions, nil
}
