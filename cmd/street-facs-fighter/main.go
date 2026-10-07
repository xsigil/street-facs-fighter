package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"street-facs-fighter/internal/audio"
	"street-facs-fighter/internal/domain/entity"
	"street-facs-fighter/internal/domain/locale"
	"street-facs-fighter/internal/infrastructure/sqlite3"
	"street-facs-fighter/internal/ui/input"
	"street-facs-fighter/internal/ui/terminal"
	"street-facs-fighter/internal/usecase"
)

// 厳格なFACS入力正規表現 (AUxx, ADxx, Mxx, LAUxx, RAUxx, Lxx, Rxx)
var strictFACSPattern = regexp.MustCompile(`^(?i)(L|R|U)?(AU|AD|M)?([0-9]{1,2})$`)

type StageRecord struct {
	StageNum   int
	FilePath   string
	RawScore   string
	Rationale  string
	Targets    []entity.TargetAU
	HitAUs     []string
	MissInputs []string
	Cleared    bool
}

// 長大な解説文を指定行数で安全に打ち切るヘルパー（スクロール防止用）
func truncateRationale(text string, maxLines int, width int, prefix string) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}

	lines := 0
	line := prefix
	for _, w := range words {
		if len(line)+len(w)+1 > width {
			fmt.Println(line + "\x1b[K")
			lines++
			if lines >= maxLines {
				fmt.Printf("%s\x1b[90m... (全文解説は撃破後の診断レポートまたは '?' で確認)\x1b[0m\x1b[K\n", prefix)
				return
			}
			line = prefix + w
		} else {
			if line == prefix {
				line += w
			} else {
				line += " " + w
			}
		}
	}
	if line != prefix && lines < maxLines {
		fmt.Println(line + "\x1b[K")
	}
}

func showCheatSheet(inputCtrl *input.InputController, loc locale.Localization) {
	fmt.Print("\x1b[2J\x1b[H")
	fmt.Printf("\x1b[1;33m%s\x1b[0m\n\n", loc.CheatSheetTitle)

	order := []string{"1", "2", "4", "5", "6", "7", "9", "10", "12", "14", "15", "16", "17", "18", "20", "22", "23", "24", "25", "26", "27", "43"}
	for _, code := range order {
		if desc, ok := loc.AUDictionary[code]; ok {
			fmt.Printf(" \x1b[1;36mAU%-2s\x1b[0m : %s\n", code, desc)
		}
	}
	fmt.Printf("\n\x1b[1;32m%s\x1b[0m\n", loc.CheatSheetExit)
	<-inputCtrl.CommitChan()
}

func showTitleScreen(inputCtrl *input.InputController, currentLoc *locale.Localization, isEN *bool, isTraining *bool) {
	inputCtrl.SetOnKey(func(curr string) {
		fmt.Printf("\r\x1b[K> \x1b[1;37m%s\x1b[0m", curr)
	})
	defer inputCtrl.SetOnKey(nil)

	for {
		fmt.Print("\x1b[2J\x1b[H")
		fmt.Println("\x1b[1;31m" + currentLoc.TitleBanner + "\x1b[0m")
		fmt.Printf("\n\x1b[1;33m%s\x1b[0m\n", currentLoc.LangPrompt)

		modePrompt := "\x1b[1;32m>>> [Enter]キー: 通常バトル / [T]キー: カンニング付きトレーニングモード <<<\x1b[0m"
		if *isEN {
			modePrompt = "\x1b[1;32m>>> [Enter]: Battle Mode / [T]: Training Mode (with Answers) <<<\x1b[0m"
		}
		fmt.Printf("\n%s\n", modePrompt)
		fmt.Print("\n> ")

		choice := strings.ToLower(<-inputCtrl.CommitChan())
		if choice == "e" || choice == "en" {
			*isEN = true
			*currentLoc = locale.LocEN
			continue
		} else if choice == "j" || choice == "ja" {
			*isEN = false
			*currentLoc = locale.LocJA
			continue
		} else if choice == "t" || choice == "train" || choice == "practice" {
			*isTraining = true
			break
		} else {
			break
		}
	}
}

func printBattleReport(records []StageRecord, totalScore int, inputCtrl *input.InputController, loc locale.Localization) {
	fmt.Print("\x1b[2J\x1b[H")
	fmt.Println("\x1b[1;36m" + loc.ReportTitle + "\x1b[0m")
	fmt.Printf("%s\n\n", loc.ReportSubtitle)

	totalHits := 0
	totalMisses := 0

	for _, rec := range records {
		fmt.Printf("\x1b[1;33m====================================================================\x1b[0m\n")
		statusStr := "\x1b[1;32m" + loc.ClearedStatus + "\x1b[0m"
		if !rec.Cleared {
			statusStr = "\x1b[1;31m" + loc.FailedStatus + "\x1b[0m"
		}
		fmt.Printf("%s %d: %s  %s\n", loc.StageLabel, rec.StageNum, rec.FilePath, statusStr)

		if f, err := os.Open(rec.FilePath); err == nil {
			if thumbImg, _, err := image.Decode(f); err == nil {
				terminal.RenderImageSixel(thumbImg, 160, 0)
			}
			f.Close()
		}

		if rec.RawScore != "" {
			fmt.Printf("\n  \x1b[1;35m公式FACSスコア\x1b[0m : \x1b[1;37m%s\x1b[0m\n", rec.RawScore)
		}

		var targetCodes []string
		for _, t := range rec.Targets {
			desc := loc.AUDictionary[t.Code]
			targetCodes = append(targetCodes, fmt.Sprintf("AU%s (%s)", t.Code, desc))
		}
		fmt.Printf("  \x1b[1;35m%s\x1b[0m: %s\n", loc.TargetLabel, strings.Join(targetCodes, ", "))

		if len(rec.HitAUs) > 0 {
			var hits []string
			for _, h := range rec.HitAUs {
				hits = append(hits, h)
			}
			fmt.Printf("  \x1b[1;32m%s\x1b[0m   : %s\n", loc.HitsLabel, strings.Join(hits, ", "))
			totalHits += len(rec.HitAUs)
		} else {
			fmt.Printf("  \x1b[1;32m%s\x1b[0m   : (None)\n", loc.HitsLabel)
		}

		if len(rec.MissInputs) > 0 {
			fmt.Printf("  \x1b[1;31m%s\x1b[0m   : %s\n", loc.MissesLabel, strings.Join(rec.MissInputs, ", "))
			totalMisses += len(rec.MissInputs)
		}

		if rec.Rationale != "" {
			fmt.Println("\n  \x1b[1;34m【FACS 解剖判定解説 (Rationale)】\x1b[0m:")
			terminal.WrapText(rec.Rationale, 75, "    ")
		}
		fmt.Println()
	}

	totalStrikes := totalHits + totalMisses
	accuracy := 0.0
	if totalStrikes > 0 {
		accuracy = (float64(totalHits) / float64(totalStrikes)) * 100.0
	}

	rank := "C (Novice Apprentice)"
	rankColor := "\x1b[1;37m"
	if accuracy >= 90.0 && totalHits >= 5 {
		rank = "S (Certified FACS Master)"
		rankColor = "\x1b[1;33m"
	} else if accuracy >= 75.0 {
		rank = "A (Senior Facial Analyst)"
		rankColor = "\x1b[1;32m"
	} else if accuracy >= 50.0 {
		rank = "B (Trained Observer)"
		rankColor = "\x1b[1;36m"
	}

	fmt.Printf("\x1b[1;33m==================== 総合評価 / FINAL EVALUATION ====================\x1b[0m\n")
	fmt.Printf(" %s  : %d (Hits: %d / Misses: %d)\n", loc.TotalStrikesLabel, totalStrikes, totalHits, totalMisses)
	fmt.Printf(" %s  : %.1f%%\n", loc.AccuracyLabel, accuracy)
	fmt.Printf(" %s  : %d pts\n", loc.FinalScoreLabel, totalScore)
	fmt.Printf(" %s  : %s%s\x1b[0m\n", loc.RankLabel, rankColor, rank)
	fmt.Printf("\x1b[1;33m====================================================================\x1b[0m\n\n")

	fmt.Printf("\x1b[1;32m%s\x1b[0m\n", loc.ExitPrompt)
	<-inputCtrl.CommitChan()
}

func runManualViewer(entries []*entity.Question, inputCtrl *input.InputController, loc locale.Localization) {
	if len(entries) == 0 {
		fmt.Println("No entries found in dataset.")
		return
	}

	currIdx := 0
	for {
		q := entries[currIdx]
		fmt.Print("\x1b[2J\x1b[H")

		fmt.Printf("\x1b[1;36m📖 STREET FACS MANUAL / ATLAS [%d/%d]\x1b[0m\n", currIdx+1, len(entries))
		fmt.Printf("Item ID   : \x1b[1;33m%s\x1b[0m (Split: %s, Media: %s)\n", q.ItemID, q.Split, q.MediaType)
		fmt.Printf("Raw Score : \x1b[1;32m%s\x1b[0m\n", q.RawScore)

		var targetDetails []string
		for _, t := range q.Targets {
			desc := loc.AUDictionary[t.Code]
			targetDetails = append(targetDetails, fmt.Sprintf("AU%s (%s)", t.Code, desc))
		}
		fmt.Printf("Target AUs: %s\n", strings.Join(targetDetails, " + "))
		fmt.Println(strings.Repeat("-", 60))

		if f, err := os.Open(q.FilePath); err == nil {
			if img, _, err := image.Decode(f); err == nil {
				terminal.RenderImageSixel(img, 280, 0)
			}
			f.Close()
		} else {
			fmt.Printf("\x1b[90m[Image not loaded: %s]\x1b[0m\n", q.FilePath)
		}

		if q.Rationale != "" {
			fmt.Println("\n\x1b[1;34m【Ekman Rationale / 解剖学的判定理由】\x1b[0m:")
			terminal.WrapText(q.Rationale, 75, "  ")
		}

		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("\x1b[1;37m[N/Enter] 次へ  |  [P] 前へ  |  [Q] 終了\x1b[0m")
		fmt.Print("> ")

		cmd := strings.ToLower(<-inputCtrl.CommitChan())
		switch cmd {
		case "q", "exit", "quit":
			return
		case "p", "prev", "k":
			if currIdx > 0 {
				currIdx--
			} else {
				currIdx = len(entries) - 1
			}
		default:
			if currIdx < len(entries)-1 {
				currIdx++
			} else {
				currIdx = 0
			}
		}
	}
}

func main() {
	langFlag := flag.String("lang", "ja", "Initial language: ja or en")
	manualMode := flag.Bool("manual", false, "Start in FACS Manual / Atlas browser mode")
	trainingMode := flag.Bool("training", false, "Start in Training mode (Answers displayed, no enemy attacks)")
	queryAU := flag.String("au", "", "Filter manual by Action Unit (e.g. 12)")
	flag.Parse()

	isEN := strings.ToLower(*langFlag) == "en"
	currentLoc := locale.LocJA
	if isEN {
		currentLoc = locale.LocEN
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := sqlite3.NewDB(ctx, "app.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	qRepo := sqlite3.NewQuestionRepository(db)

	if *manualMode || (len(os.Args) > 1 && os.Args[1] == "manual") {
		all, err := qRepo.FindAll(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching manual entries: %v\n", err)
			return
		}

		var entries []*entity.Question
		if *queryAU != "" {
			cleanAU := strings.TrimSpace(strings.ReplaceAll(strings.ToUpper(*queryAU), "AU", ""))
			for _, q := range all {
				for _, t := range q.Targets {
					if t.Code == cleanAU {
						entries = append(entries, q)
						break
					}
				}
			}
		} else {
			entries = all
		}

		restoreTerm := input.SetTerminalRawMode()
		defer restoreTerm()

		inputCtrl := input.NewInputController(ctx)
		runManualViewer(entries, inputCtrl, currentLoc)
		return
	}

	battleUsecase := usecase.NewBattleUsecase(qRepo)

	questions, err := battleUsecase.GetStageQuestions(ctx, 5)
	if err != nil || len(questions) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No playable questions found in app.db. Please check assets/ and run importer!\n")
		return
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})

	restoreTerm := input.SetTerminalRawMode()
	defer restoreTerm()

	isTraining := *trainingMode || (len(os.Args) > 1 && (os.Args[1] == "practice" || os.Args[1] == "training"))
	inputCtrl := input.NewInputController(ctx)
	showTitleScreen(inputCtrl, &currentLoc, &isEN, &isTraining)

	bgm := audio.StartBGM("assets/sounds/stage1.mp3")
	defer func() {
		if bgm != nil {
			bgm.Stop()
		}
	}()

	playerLife := 5
	score := 0
	// 縦スクロールを防ぐため、端末高さ約12〜13行に収まる幅に設定
	imageWidth := 210
	var battleHistory []StageRecord
	var mu sync.Mutex

	for stageIdx, q := range questions {
		if playerLife <= 0 {
			break
		}

		file, err := os.Open(q.FilePath)
		if err != nil {
			continue
		}
		img, _, err := image.Decode(file)
		file.Close()
		if err != nil {
			continue
		}

		rageImg := terminal.CreateRageImage(img)

		remaining := make(map[string]entity.TargetAU)
		for _, t := range q.Targets {
			remaining[t.Code] = t
		}
		maxHP := len(remaining)
		currentHP := maxHP
		var hitLog []string
		bannerMsg := ""

		currentRecord := StageRecord{
			StageNum:  stageIdx + 1,
			FilePath:  q.FilePath,
			RawScore:  q.RawScore,
			Rationale: q.Rationale,
			Targets:   q.Targets,
		}

		enemyATB := 0
		stageOver := false
		ticker := time.NewTicker(120 * time.Millisecond)

		redrawEntireScreen := func(customMsg string) {
			fmt.Print("\x1b[2J\x1b[H")

			if isTraining {
				fmt.Printf("\x1b[1;36m[TRAINING] STAGE %d/%d | HP: %d/%d\x1b[0m\x1b[K\n", stageIdx+1, len(questions), currentHP, maxHP)
				var answers []string
				for _, t := range q.Targets {
					prefix := "AU"
					if t.Side != "" {
						prefix = t.Side + "AU"
					}
					desc := currentLoc.AUDictionary[t.Code]
					if _, alive := remaining[t.Code]; !alive {
						answers = append(answers, fmt.Sprintf("\x1b[90m%s%s(済)\x1b[0m", prefix, t.Code))
					} else {
						answers = append(answers, fmt.Sprintf("\x1b[1;33m%s%s\x1b[0m(\x1b[37m%s\x1b[0m)", prefix, t.Code, desc))
					}
				}
				fmt.Printf("💡 \x1b[1;32m急所:\x1b[0m %s\x1b[K\n", strings.Join(answers, " | "))
				fmt.Println(strings.Repeat("-", 50) + "\x1b[K")
			} else {
				pLifeBar := strings.Repeat("❤️ ", playerLife)
				fmt.Printf("STAGE %d/%d  |  PLAYER HP: %-15s |  SCORE: %d\x1b[K\n", stageIdx+1, len(questions), pLifeBar, score)

				hpClamped := currentHP
				if hpClamped < 0 {
					hpClamped = 0
				}
				if hpClamped > maxHP {
					hpClamped = maxHP
				}
				bossHpBar := strings.Repeat("■", hpClamped) + strings.Repeat("░", maxHP-hpClamped)
				fmt.Printf("\x1b[1;35m%s: [%s] (%d/%d)\x1b[0m\x1b[K\n", currentLoc.TargetHP, bossHpBar, currentHP, maxHP)

				atbClamped := enemyATB
				if atbClamped < 0 {
					atbClamped = 0
				}
				if atbClamped > 100 {
					atbClamped = 100
				}
				atbCount := atbClamped / 10
				if atbCount > 10 {
					atbCount = 10
				}
				atbBar := strings.Repeat("🔥", atbCount) + strings.Repeat("・", 10-atbCount)
				fmt.Printf("\x1b[1;31m%s: [%s] %3d%%\x1b[0m\x1b[K\n", currentLoc.EnemyATB, atbBar, enemyATB)
				fmt.Println(strings.Repeat("-", 50) + "\x1b[K")
			}

			if customMsg != "" {
				fmt.Println(customMsg + "\x1b[K")
			}

			// Sixel 画像描画
			terminal.RenderImageSixel(img, imageWidth, 0)

			// トレーニングモード時の解説（最大2〜3行に抑制してスクロールアウトを完全防止）
			if isTraining && q.Rationale != "" {
				fmt.Printf("\n\x1b[1;34m【Ekman Guide】\x1b[0m\x1b[K\n")
				truncateRationale(q.Rationale, 2, 75, "  ")
			}

			// 破壊ログ（横並びで省スペース化）
			if len(hitLog) > 0 {
				fmt.Printf("\x1b[32m%s\x1b[0m ", currentLoc.DestroyedHeader)
				fmt.Println(strings.Join(hitLog, " / ") + "\x1b[K")
			}

			// プロンプトと現在の入力文字
			fmt.Printf("\n%s\x1b[1;37m%s\x1b[0m", currentLoc.StrikePrompt, inputCtrl.GetCurrentInput())
		}

		// ATBバーの差分更新
		renderHUDDifferential := func() {
			if isTraining {
				return
			}
			fmt.Print("\x1b[s") // カーソル保存
			fmt.Print("\x1b[H") // 先頭へ

			pLifeBar := strings.Repeat("❤️ ", playerLife)
			fmt.Printf("STAGE %d/%d  |  PLAYER HP: %-15s |  SCORE: %d\x1b[K\n", stageIdx+1, len(questions), pLifeBar, score)

			hpClamped := currentHP
			if hpClamped < 0 {
				hpClamped = 0
			}
			if hpClamped > maxHP {
				hpClamped = maxHP
			}
			bossHpBar := strings.Repeat("■", hpClamped) + strings.Repeat("░", maxHP-hpClamped)
			fmt.Printf("\x1b[1;35m%s: [%s] (%d/%d)\x1b[0m\x1b[K\n", currentLoc.TargetHP, bossHpBar, currentHP, maxHP)

			atbClamped := enemyATB
			if atbClamped < 0 {
				atbClamped = 0
			}
			if atbClamped > 100 {
				atbClamped = 100
			}
			atbCount := atbClamped / 10
			if atbCount > 10 {
				atbCount = 10
			}
			atbBar := strings.Repeat("🔥", atbCount) + strings.Repeat("・", 10-atbCount)
			fmt.Printf("\x1b[1;31m%s: [%s] %3d%%\x1b[0m\x1b[K\n", currentLoc.EnemyATB, atbBar, enemyATB)

			fmt.Print("\x1b[u") // カーソル復元
		}

		redrawEntireScreen("")

		// 入力エコー（行内のみ更新し画像を一切汚染しない）
		inputCtrl.SetOnKey(func(curr string) {
			mu.Lock()
			fmt.Printf("\r\x1b[K%s\x1b[1;37m%s\x1b[0m", currentLoc.StrikePrompt, curr)
			mu.Unlock()
		})

		for !stageOver && playerLife > 0 {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				mu.Lock()
				if !isTraining {
					enemyATB += rng.Intn(4) + 2
					if enemyATB >= 100 {
						playerLife--
						enemyATB = 0
						atkMsg := currentLoc.EnemyAttacks[rng.Intn(len(currentLoc.EnemyAttacks))]

						audio.PlaySound("assets/sounds/damage.wav")
						terminal.EnemyAttackFlash(rageImg, img, imageWidth, currentLoc.RageAttackTitle, atkMsg)

						time.Sleep(700 * time.Millisecond)
						if playerLife <= 0 {
							stageOver = true
						}
						redrawEntireScreen("")
					} else {
						renderHUDDifferential()
					}
				}
				mu.Unlock()

			case committedLine := <-inputCtrl.CommitChan():
				if committedLine == "" {
					continue
				}

				if committedLine == "?" || committedLine == "HINT" {
					showCheatSheet(inputCtrl, currentLoc)
					redrawEntireScreen("")
					continue
				}

				tokens := strings.FieldsFunc(committedLine, func(r rune) bool {
					return r == ' ' || r == ',' || r == ';'
				})

				mu.Lock()
				hitAny := false
				allValidFormat := true

				for _, rawToken := range tokens {
					tok := strings.TrimSpace(rawToken)
					if tok == "" {
						continue
					}

					sub := strictFACSPattern.FindStringSubmatch(tok)
					if len(sub) == 0 {
						allValidFormat = false
						currentRecord.MissInputs = append(currentRecord.MissInputs, fmt.Sprintf("%s (規格外: 'AU%s')", tok, tok))
						continue
					}

					side := strings.ToUpper(sub[1])
					prefix := strings.ToUpper(sub[2])
					num := sub[3]

					if prefix == "" && side == "" {
						allValidFormat = false
						currentRecord.MissInputs = append(currentRecord.MissInputs, fmt.Sprintf("%s (プレフィックス必須: 'AU%s')", tok, num))
						continue
					}

					matched := false
					for code, target := range remaining {
						if target.Side != "" && target.Side != side {
							continue
						}
						if target.Code == num {
							matched = true
							hitAny = true
							audio.PlaySound("assets/sounds/hit.wav")

							delete(remaining, code)
							currentRecord.HitAUs = append(currentRecord.HitAUs, tok)
							currentHP--
							score += 300

							if enemyATB > 40 {
								enemyATB -= 40
							} else {
								enemyATB = 0
							}

							desc := currentLoc.AUDictionary[target.Code]
							hitLog = append(hitLog, fmt.Sprintf("💥 %s:%s", tok, desc))

							if currentHP == 0 {
								stageOver = true
								currentRecord.Cleared = true
								audio.PlaySound("assets/sounds/ko.wav")

								fmt.Print("\x1b[2J\x1b[H")
								fmt.Println(currentLoc.KoTitle)
								terminal.RenderImageSixel(img, imageWidth, 0)
								fmt.Println(currentLoc.KoSubtitle)
								for _, l := range hitLog {
									fmt.Println("  " + l)
								}
								time.Sleep(1500 * time.Millisecond)
								break
							}
						}
					}

					if !matched {
						currentRecord.MissInputs = append(currentRecord.MissInputs, tok)
					}
				}

				if !stageOver {
					if !hitAny || !allValidFormat {
						if !allValidFormat {
							bannerMsg = "\x1b[1;31m❌ 規格外！'AU12' や 'LAU12' のように入力してください！\x1b[0m"
						} else {
							bannerMsg = currentLoc.MissBanner
						}
						if !isTraining {
							enemyATB += 15
						}
					} else {
						bannerMsg = currentLoc.HitBanner
					}
					redrawEntireScreen(bannerMsg)
				}
				mu.Unlock()
			}
		}

		inputCtrl.SetOnKey(nil)
		ticker.Stop()
		battleHistory = append(battleHistory, currentRecord)

		// 倒した直後に画面をクリアして次のステージへ引き渡す
		fmt.Print("\x1b[2J\x1b[H")
	}

	if bgm != nil {
		bgm.Stop()
	}

	loseBgm := audio.StartBGM("assets/sounds/lose.mp3")
	if loseBgm == nil {
		audio.PlaySound("assets/sounds/lose.wav")
	}
	defer func() {
		if loseBgm != nil {
			loseBgm.Stop()
		}
	}()

	printBattleReport(battleHistory, score, inputCtrl, currentLoc)
}