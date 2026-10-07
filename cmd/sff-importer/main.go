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
	assetsDir := flag.String("dest", "./assets", "画像・メディア配置先ディレクトリ")
	dbPath := flag.String("db", "app.db", "SQLiteデータベースパス")
	flag.Parse()

	ctx := context.Background()
	db, err := sqlite3.NewDB(ctx, *dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB初期化失敗: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	qRepo := sqlite3.NewQuestionRepository(db)
	txManager := sqlite3.NewTxManager(db)

	importer := usecase.NewImportTSVUsecase(qRepo, txManager)
	count, err := importer.ExecuteEmbedded(ctx, *assetsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "インポート失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n[+] 内蔵 FACS Master Dataset から %d 件を %s へインポート完了！\n", count, *dbPath)
	fmt.Println("[+] 'go run ./cmd/street-facs-fighter' でゲームを開始できます。")
}
