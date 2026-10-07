package sqlite3

import (
	"context"
	"database/sql"
	"fmt"

	"street-facs-fighter/internal/domain/entity"

	"github.com/jmoiron/sqlx"
)

type questionModel struct {
	ID        string         `db:"id"`
	ItemID    sql.NullString `db:"item_id"`
	Split     sql.NullString `db:"split"`
	MediaType sql.NullString `db:"media_type"`
	RawScore  sql.NullString `db:"raw_score"`
	FilePath  string         `db:"file_path"`
	FileName  string         `db:"file_name"`
	Rationale sql.NullString `db:"rationale"`
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

	qQuery := `INSERT INTO questions (id, item_id, split, media_type, raw_score, file_path, file_name, rationale)
              VALUES (:id, :item_id, :split, :media_type, :raw_score, :file_path, :file_name, :rationale)
              ON CONFLICT(id) DO UPDATE SET 
                  item_id = excluded.item_id,
                  split = excluded.split,
                  media_type = excluded.media_type,
                  raw_score = excluded.raw_score,
                  file_path = excluded.file_path,
                  file_name = excluded.file_name,
                  rationale = excluded.rationale`

	qm := questionModel{
		ID:        q.ID,
		ItemID:    sql.NullString{String: q.ItemID, Valid: q.ItemID != ""},
		Split:     sql.NullString{String: q.Split, Valid: q.Split != ""},
		MediaType: sql.NullString{String: q.MediaType, Valid: q.MediaType != ""},
		RawScore:  sql.NullString{String: q.RawScore, Valid: q.RawScore != ""},
		FilePath:  q.FilePath,
		FileName:  q.FileName,
		Rationale: sql.NullString{String: q.Rationale, Valid: q.Rationale != ""},
	}

	if _, err := sqlx.NamedExecContext(ctx, conn, qQuery, qm); err != nil {
		return fmt.Errorf("failed to save question: %w", err)
	}

	// 既存ターゲットの再構成
	if _, err := conn.ExecContext(ctx, `DELETE FROM question_targets WHERE question_id = ?`, q.ID); err != nil {
		return fmt.Errorf("failed to clean up old targets: %w", err)
	}

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

func (r *QuestionRepository) FindRandom(ctx context.Context, limit int) ([]*entity.Question, error) {
	conn := GetExtContext(ctx, r.db)

	var qModels []questionModel
	// 静止画かつ正解AUが存在するものを優先抽出
	query := `SELECT id, item_id, split, media_type, raw_score, file_path, file_name, rationale 
              FROM questions 
              WHERE media_type = 'image'
              ORDER BY RANDOM() LIMIT ?`
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
			ID:        qm.ID,
			ItemID:    qm.ItemID.String,
			Split:     qm.Split.String,
			MediaType: qm.MediaType.String,
			RawScore:  qm.RawScore.String,
			FilePath:  qm.FilePath,
			FileName:  qm.FileName,
			Rationale: qm.Rationale.String,
			Targets:   targets,
		})
	}

	return questions, nil
}

// FindAll でも同様に SELECT & マッピングするように統一
func (r *QuestionRepository) FindAll(ctx context.Context) ([]*entity.Question, error) {
	conn := GetExtContext(ctx, r.db)

	var qModels []questionModel
	query := `SELECT id, item_id, split, media_type, raw_score, file_path, file_name, rationale FROM questions`
	if err := sqlx.SelectContext(ctx, conn, &qModels, query); err != nil {
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
			ID:        qm.ID,
			ItemID:    qm.ItemID.String,
			Split:     qm.Split.String,
			MediaType: qm.MediaType.String,
			RawScore:  qm.RawScore.String,
			FilePath:  qm.FilePath,
			FileName:  qm.FileName,
			Rationale: qm.Rationale.String,
			Targets:   targets,
		})
	}

	return questions, nil
}
