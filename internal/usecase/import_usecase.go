package usecase

import (
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"street-facs-fighter/internal/domain/entity"
	"street-facs-fighter/internal/domain/repository"
)

type ImportUsecase struct {
	qRepo     repository.QuestionRepository
	txManager repository.TxManager
}

func NewImportUsecase(qRepo repository.QuestionRepository, txManager repository.TxManager) *ImportUsecase {
	return &ImportUsecase{qRepo: qRepo, txManager: txManager}
}

// parseFACSFilename: FACS Examplesのファイル名を解析する
func parseFACSFilename(filename string) []entity.TargetAU {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))

	// 先頭の "s", "sW", "sJ" などを除外
	rePrefix := regexp.MustCompile(`(?i)^s[WJ]?`)
	core := rePrefix.ReplaceAllString(base, "")

	if core == "" || core == "0" {
		return nil
	}

	var results []entity.TargetAU
	tokens := strings.Split(core, "_")
	for _, tok := range tokens {
		subMatches := regexp.MustCompile(`([LR])?([0-9]{1,2})([a-z])?`).FindAllStringSubmatch(tok, -1)
		for _, sm := range subMatches {
			side := strings.ToUpper(sm[1])
			code := sm[2]
			intensity := strings.ToUpper(sm[3])

			results = append(results, entity.TargetAU{
				Side:      side,
				Code:      code,
				Intensity: intensity,
			})
		}
	}

	return results
}

func (u *ImportUsecase) Execute(ctx context.Context, srcDir, destDir string) (int, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create destDir: %w", err)
	}

	// Examples フォルダの探索
	examplesDir := filepath.Join(srcDir, "Manual", "Examples")
	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		examplesDir = srcDir
	}

	importedCount := 0
	err := filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".gif" {
			return nil
		}

		targets := parseFACSFilename(info.Name())
		if len(targets) == 0 {
			return nil
		}

		// 画像ファイルの健全性チェック
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		_, _, err = image.DecodeConfig(f)
		f.Close()
		if err != nil {
			return nil
		}

		destPath := filepath.Join(destDir, info.Name())
		if err := copyFile(path, destPath); err != nil {
			return fmt.Errorf("copy failed %s: %w", info.Name(), err)
		}

		q := &entity.Question{
			ID:       info.Name(),
			FilePath: destPath,
			FileName: info.Name(),
			Targets:  targets,
		}

		err = u.txManager.Do(ctx, func(txCtx context.Context) error {
			return u.qRepo.Save(txCtx, q)
		})
		if err != nil {
			return fmt.Errorf("failed to save question to db: %w", err)
		}

		importedCount++
		return nil
	})

	return importedCount, err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
