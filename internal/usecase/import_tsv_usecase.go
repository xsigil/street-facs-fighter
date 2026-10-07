package usecase

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"street-facs-fighter/internal/domain/entity"
	"street-facs-fighter/internal/domain/repository"
)

//go:embed facs_master_dataset.tsv
var EmbeddedMasterTSV string

type ImportTSVUsecase struct {
	qRepo     repository.QuestionRepository
	txManager repository.TxManager
}

func NewImportTSVUsecase(qRepo repository.QuestionRepository, txManager repository.TxManager) *ImportTSVUsecase {
	return &ImportTSVUsecase{qRepo: qRepo, txManager: txManager}
}

var auPattern = regexp.MustCompile(`\b([LR])?([1-9]|1[0-9]|2[0-9]|3[0-9]|4[0-6]|5[0-9]|6[0-4])([A-Ea-eX-Zx-z])?\b`)

func ParseScoreToTargets(score string) []entity.TargetAU {
	matches := auPattern.FindAllStringSubmatch(score, -1)
	var targets []entity.TargetAU
	seen := make(map[string]bool)

	for _, m := range matches {
		side := strings.ToUpper(m[1])
		code := m[2]
		intensity := strings.ToUpper(m[3])

		key := fmt.Sprintf("%s_%s", side, code)
		if seen[key] {
			continue
		}
		seen[key] = true

		targets = append(targets, entity.TargetAU{
			Side:      side,
			Code:      code,
			Intensity: intensity,
		})
	}
	return targets
}

func (u *ImportTSVUsecase) ExecuteEmbedded(ctx context.Context, assetsDir string) (int, error) {
	scanner := bufio.NewScanner(strings.NewReader(EmbeddedMasterTSV))
	// 判定根拠（rationale）の長文に耐えられるようバッファを拡大
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	lineNum := 0
	importedCount := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if lineNum == 1 && strings.HasPrefix(line, "split\t") {
			continue // ヘッダー行スキップ
		}
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 5 {
			continue
		}

		split := parts[0]
		itemID := parts[1]
		mediaPath := parts[2]
		rawScore := parts[3]
		rationale := parts[4]

		fileName := filepath.Base(mediaPath)
		ext := strings.ToLower(filepath.Ext(fileName))

		mediaType := "image"
		if ext == ".mov" || ext == ".mp4" {
			mediaType = "video"
		}

		targets := ParseScoreToTargets(rawScore)
		resolvedPath := filepath.Join(assetsDir, fileName)

		q := &entity.Question{
			ID:        fileName,
			ItemID:    itemID,
			Split:     split,
			MediaType: mediaType,
			RawScore:  rawScore,
			FilePath:  resolvedPath,
			FileName:  fileName,
			Rationale: rationale,
			Targets:   targets,
		}

		err := u.txManager.Do(ctx, func(txCtx context.Context) error {
			return u.qRepo.Save(txCtx, q)
		})
		if err != nil {
			return importedCount, fmt.Errorf("line %d: db error: %w", lineNum, err)
		}
		importedCount++
	}

	return importedCount, scanner.Err()
}
