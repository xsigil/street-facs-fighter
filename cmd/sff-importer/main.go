package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"street-facs-fighter/internal/infrastructure/sqlite3"
	"street-facs-fighter/internal/usecase"
)

func main() {
	srcDir := flag.String("src", "", "FACSメディアまたは展開先パス (例: /mnt/cdrom)")
	destDir := flag.String("dest", "./assets", "画像保存先ディレクトリ")
	dbPath := flag.String("db", "app.db", "SQLiteデータベースパス")
	flag.Parse()

	if *srcDir == "" {
		fmt.Println("使用方法: go run ./cmd/sff-importer -src <FACS教材のパス>")
		os.Exit(1)
	}

	ctx := context.Background()
	db, err := sqlite3.NewDB(ctx, *dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB初期化失敗: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	qRepo := sqlite3.NewQuestionRepository(db)
	txManager := sqlite3.NewTxManager(db)

	importer := usecase.NewImportUsecase(qRepo, txManager)
	count, err := importer.Execute(ctx, *srcDir, *destDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "インポート失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n[+] インポート成功: %d 件のFACS静止画ターゲットを登録しました！\n", count)
	fmt.Println("[+] 'go run ./cmd/street-facs-fighter' でゲームを開始できます。")
}
