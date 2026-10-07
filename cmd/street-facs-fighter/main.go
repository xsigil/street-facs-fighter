package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"street-facs-fighter/internal/domain/entity"
	"street-facs-fighter/internal/infrastructure/sqlite3"
	"street-facs-fighter/internal/usecase"

	"github.com/mattn/go-sixel"
	"golang.org/x/image/draw"
)

type Localization struct {
	TitleBanner       string
	LangPrompt        string
	StartPrompt       string
	TargetHP          string
	EnemyATB          string
	StrikePrompt      string
	HitBanner         string
	MissBanner        string
	KoTitle           string
	KoSubtitle        string
	RageAttackTitle   string
	DestroyedHeader   string
	CheatSheetTitle   string
	CheatSheetExit    string
	ReportTitle       string
	ReportSubtitle    string
	StageLabel        string
	TargetLabel       string
	HitsLabel         string
	MissesLabel       string
	ClearedStatus     string
	FailedStatus      string
	TotalStrikesLabel string
	AccuracyLabel     string
	FinalScoreLabel   string
	RankLabel         string
	ExitPrompt        string
	EnemyAttacks      []string
	AUDictionary      map[string]string
}

var locJA = Localization{
	TitleBanner: `
===========================================================
  🥊 STREET FACS FIGHTER - ターミナル表情筋格闘ゲーム 🥊
===========================================================`,
	LangPrompt:        "言語切替 (Switch Lang): [J: 日本語] / [E: English] (切替は 'e' または 'j' を入力してEnter)",
	StartPrompt:       ">>> [Enter]キーを押して FIGHT! <<<",
	TargetHP:          "TARGET HP",
	EnemyATB:          "ENEMY TENSION",
	StrikePrompt:      "急所AUを撃て！ (例: 1 4 / ヒント: '?'): ",
	HitBanner:         "💥 クリティカルHIT！急所AUを破壊！",
	MissBanner:        "❌ MISS！相手のテンションが急上昇！",
	KoTitle:           "\n🌟🌟🌟 TARGET K.O.!! 🌟🌟🌟",
	KoSubtitle:        "全ターゲットAUの破壊に成功！",
	RageAttackTitle:   "⚡⚡⚡ 敵の怒り攻撃が炸裂！ ⚡⚡⚡",
	DestroyedHeader:   "【破壊済みAU部位】:",
	CheatSheetTitle:   "【FACS アクション・ユニット (AU) あんちょこ速見表】",
	CheatSheetExit:    "[Enter]キーを押して戦闘に戻る...",
	ReportTitle:       "\n=================== 戦闘結果解析レポート ===================",
	ReportSubtitle:    "各ステージで出現した表情写真と解析結果一覧:",
	StageLabel:        "ステージ",
	TargetLabel:       "正解AU",
	HitsLabel:         "命中AU",
	MissesLabel:       "誤入力",
	ClearedStatus:     "【撃破 K.O.】",
	FailedStatus:      "【被弾敗北】",
	TotalStrikesLabel: "総打撃数",
	AccuracyLabel:     "命中精度 (正答率)",
	FinalScoreLabel:   "最終獲得スコア",
	RankLabel:         "FACS解析官ランク",
	ExitPrompt:        "[Enter]キーを押してゲームを終了します...",
	EnemyAttacks: []string{
		"「フッ…貴様に私の真の感情が見抜けるか！」",
		"「筋肉の弛緩が遅い！出直してくるんだな！」",
		"「微表情（マイクロ・エクスプレッション）の嵐を喰らえ！」",
		"「眉間のしわ（AU4）ひとつすら読めぬとはな！」",
	},
	AUDictionary: map[string]string{
		"1":  "内側眉上げ (Frontalis, pars medialis)",
		"2":  "外側眉上げ (Frontalis, pars lateralis)",
		"4":  "眉下げ・眉間の縦ジワ (Corrugator supercilii)",
		"5":  "上まぶた引き上げ (Levator palpebrae superioris)",
		"6":  "頬上げ・目尻のカラスの足跡 (Orbicularis oculi)",
		"7":  "下まぶた緊張・細目 (Orbicularis oculi, pars palpebralis)",
		"9":  "鼻筋のシワ (Levator labii superioris alaeque nasi)",
		"10": "上唇引き上げ (Levator labii superioris)",
		"12": "口角斜め引き上げ・笑顔 (Zygomaticus major)",
		"14": "口角のディンプル・えくぼ (Buccinator)",
		"15": "口角下げ・への字口 (Depressor anguli oris)",
		"16": "下唇引き下げ (Depressor labii inferioris)",
		"17": "オトガイ引き上げ・顎の梅干しジワ (Mentalis)",
		"18": "唇すぼめ (Incisivii labii superioris)",
		"20": "口角水平引き伸ばし (Risorius)",
		"22": "唇の尖らし・キス口 (Orbicularis oris)",
		"23": "唇の引き締め (Orbicularis oris)",
		"24": "唇の圧迫 (Orbicularis oris)",
		"25": "唇の開き・歯の露出なし (Depressor labii)",
		"26": "顎の自然な落下 (Relaxation of masseter)",
		"27": "口を大きく開ける (Pterygoids & Digastric)",
		"43": "閉眼・目を閉じる (Relaxation of levator)",
	},
}

var locEN = Localization{
	TitleBanner: `
===========================================================
  🥊 STREET FACS FIGHTER - TERMINAL FACIAL ACTION BATTLE 🥊
===========================================================`,
	LangPrompt:        "Switch Language: [J: Japanese] / [E: English] (Type 'e' or 'j' and press Enter)",
	StartPrompt:       ">>> Press [Enter] to FIGHT! <<<",
	TargetHP:          "TARGET HP",
	EnemyATB:          "ENEMY TENSION",
	StrikePrompt:      "Strike Target AU! (e.g. 1 4 / Hint: '?'): ",
	HitBanner:         "💥 CRITICAL HIT! Target Action Unit Destroyed!",
	MissBanner:        "❌ MISS! Opponent Tension Spiked!",
	KoTitle:           "\n🌟🌟🌟 TARGET K.O.!! 🌟🌟🌟",
	KoSubtitle:        "All target Action Units successfully neutralized!",
	RageAttackTitle:   "⚡⚡⚡ ENEMY RAGE ATTACK DETONATED! ⚡⚡⚡",
	DestroyedHeader:   "【Destroyed Action Units】:",
	CheatSheetTitle:   "【FACS Action Unit (AU) Reference Matrix】",
	CheatSheetExit:    "Press [Enter] to resume combat...",
	ReportTitle:       "\n================ COMBAT DIAGNOSTIC REPORT ================",
	ReportSubtitle:    "Target Facial Expressions and Performance History:",
	StageLabel:        "STAGE",
	TargetLabel:       "Target AU",
	HitsLabel:         "Hit AUs",
	MissesLabel:       "Misses",
	ClearedStatus:     "【CLEARED K.O.】",
	FailedStatus:      "【DEFEATED】",
	TotalStrikesLabel: "Total Strikes",
	AccuracyLabel:     "Strike Accuracy",
	FinalScoreLabel:   "Final Combat Score",
	RankLabel:         "FACS Analyst Rank",
	ExitPrompt:        "Press [Enter] to exit the game...",
	EnemyAttacks: []string{
		"\"Ha! Can you decode my genuine micro-expression?!\"",
		"\"Too slow! Your ocular calibration is lagging!\"",
		"\"Take this rapid sub-200ms Action Unit burst!\"",
		"\"You cannot even read a Corrugator brow contraction (AU4)!\"",
	},
	AUDictionary: map[string]string{
		"1":  "Inner Brow Raiser (Frontalis, pars medialis)",
		"2":  "Outer Brow Raiser (Frontalis, pars lateralis)",
		"4":  "Brow Lowerer (Corrugator supercilii / Depressor)",
		"5":  "Upper Lid Raiser (Levator palpebrae superioris)",
		"6":  "Cheek Raiser (Orbicularis oculi, pars orbitalis)",
		"7":  "Lid Tightener (Orbicularis oculi, pars palpebralis)",
		"9":  "Nose Wrinkler (Levator labii superioris alaeque nasi)",
		"10": "Upper Lip Raiser (Levator labii superioris)",
		"12": "Lip Corner Puller (Zygomaticus major)",
		"14": "Dimpler (Buccinator)",
		"15": "Lip Corner Depressor (Depressor anguli oris)",
		"16": "Lower Lip Depressor (Depressor labii inferioris)",
		"17": "Chin Raiser (Mentalis)",
		"18": "Lip Pucker (Incisivii labii superioris)",
		"20": "Lip Stretcher (Risorius)",
		"22": "Lip Funneler (Orbicularis oris)",
		"23": "Lip Tightener (Orbicularis oris)",
		"24": "Lip Pressor (Orbicularis oris)",
		"25": "Lips Part (Depressor labii / Jaw relaxation)",
		"26": "Jaw Drop (Masseter relaxation)",
		"27": "Mouth Stretch (Pterygoids & Digastric)",
		"43": "Eyes Closed (Levator relaxation)",
	},
}

type StageRecord struct {
	StageNum   int
	FilePath   string
	RawScore   string // 原典スコア表記 (例: "7D+9D+17B")
	Rationale  string // 解剖学的解説
	Targets    []entity.TargetAU
	HitAUs     []string
	MissInputs []string
	Cleared    bool
}

type BGMPlayer struct {
	cmd *exec.Cmd
}

func (b *BGMPlayer) Stop() {
	if b != nil && b.cmd != nil && b.cmd.Process != nil {
		_ = b.cmd.Process.Kill()
		_ = b.cmd.Wait()
	}
}

func startBGM(path string) *BGMPlayer {
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	players := []struct {
		name string
		args []string
	}{
		{"mpv", []string{"--loop=inf", "--no-video", "--really-quiet", path}},
		{"ffplay", []string{"-loop", "0", "-nodisp", "-autoexit", "-loglevel", "quiet", path}},
		{"pw-play", []string{path}},
		{"paplay", []string{path}},
	}
	for _, p := range players {
		if _, err := exec.LookPath(p.name); err == nil {
			cmd := exec.Command(p.name, p.args...)
			if err := cmd.Start(); err == nil {
				return &BGMPlayer{cmd: cmd}
			}
		}
	}
	return nil
}

func playSound(path string) {
	if _, err := os.Stat(path); err != nil {
		return
	}
	players := []struct {
		name string
		args []string
	}{
		{"mpv", []string{"--no-video", "--really-quiet", path}},
		{"ffplay", []string{"-nodisp", "-autoexit", "-loglevel", "quiet", path}},
		{"pw-play", []string{path}},
		{"paplay", []string{path}},
		{"aplay", []string{"-q", path}},
	}
	for _, p := range players {
		if _, err := exec.LookPath(p.name); err == nil {
			cmd := exec.Command(p.name, p.args...)
			_ = cmd.Start()
			go func(c *exec.Cmd) {
				_ = c.Wait()
			}(cmd)
			return
		}
	}
}

func setTerminalRawMode() func() {
	// Directly configure the controlling terminal device to prevent SIGTTIN/SIGTTOU hangs in subshells
	cmd := exec.Command("stty", "-F", "/dev/tty", "-icanon", "-echo")
	if err := cmd.Run(); err != nil {
		// Fallback for macOS / non-Linux terminals
		_ = exec.Command("stty", "-icanon", "-echo").Run()
	}

	return func() {
		cmdReset := exec.Command("stty", "-F", "/dev/tty", "sane")
		if err := cmdReset.Run(); err != nil {
			_ = exec.Command("stty", "sane").Run()
		}
		fmt.Print("\x1b[?25h") // Ensure cursor is visible
	}
}

type InputController struct {
	mu         sync.Mutex
	currentBuf string
	commitChan chan string
	onKey      func(string)
}

func NewInputController(ctx context.Context) *InputController {
	c := &InputController{
		commitChan: make(chan string, 32),
	}

	go func() {
		buf := make([]byte, 32)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			for i := 0; i < n; i++ {
				b := buf[i]

				// Ctrl+C (Interrupt)
				if b == 3 {
					p, _ := os.FindProcess(os.Getpid())
					_ = p.Signal(os.Interrupt)
					return
				}

				// Enter (CR or LF)
				if b == '\r' || b == '\n' {
					c.mu.Lock()
					line := strings.TrimSpace(c.currentBuf)
					c.currentBuf = ""
					c.mu.Unlock()

					c.commitChan <- line
					continue
				}

				// Backspace (127: DEL, 8: BS)
				if b == 127 || b == 8 {
					c.mu.Lock()
					if len(c.currentBuf) > 0 {
						c.currentBuf = c.currentBuf[:len(c.currentBuf)-1]
					}
					curr := c.currentBuf
					cb := c.onKey
					c.mu.Unlock()
					if cb != nil {
						cb(curr)
					}
					continue
				}

				// Escape sequences (e.g. arrow keys)
				if b == 27 {
					if i+2 < n && buf[i+1] == '[' {
						i += 2
					}
					continue
				}

				// Printable characters
				if b >= 32 && b <= 126 {
					c.mu.Lock()
					c.currentBuf += string(b)
					curr := c.currentBuf
					cb := c.onKey
					c.mu.Unlock()
					if cb != nil {
						cb(curr)
					}
				}
			}
		}
	}()

	return c
}

func (c *InputController) SetOnKey(fn func(string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onKey = fn
}

func (c *InputController) GetCurrentInput() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.currentBuf
}

func (c *InputController) ClearInput() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentBuf = ""
}

func renderImageSixel(img image.Image, targetWidth int, offsetX int) {
	if img == nil {
		return
	}
	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	if srcW == 0 || srcH == 0 {
		return
	}

	targetHeight := (srcH * targetWidth) / srcW
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.BiLinear.Scale(dst, dst.Bounds(), img, srcBounds, draw.Over, nil)

	var buf bytes.Buffer
	enc := sixel.NewEncoder(&buf)
	enc.Width = targetWidth
	enc.Height = targetHeight
	if err := enc.Encode(dst); err == nil {
		if offsetX > 0 {
			fmt.Print(strings.Repeat(" ", offsetX))
		}
		fmt.Print(buf.String())
	}
}

func createRageImage(base image.Image) image.Image {
	if base == nil {
		return nil
	}
	bounds := base.Bounds()
	rage := image.NewRGBA(bounds)
	draw.Draw(rage, bounds, base, bounds.Min, draw.Src)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := rage.At(x, y)
			r, g, b, a := c.RGBA()
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// Shift toward saturated hostile red
			rRage := uint8(math.Min(255, float64(r8)*1.45+40))
			gRage := uint8(float64(g8) * 0.45)
			bRage := uint8(float64(b8) * 0.45)
			rage.Set(x, y, color.RGBA{R: rRage, G: gRage, B: bRage, A: uint8(a >> 8)})
		}
	}
	return rage
}

func enemyAttackFlash(rageImg, origImg image.Image, width int, banner, atkMsg string) {
	fmt.Print("\x1b[2J\x1b[H")
	fmt.Println("\x1b[1;31m" + banner + "\x1b[0m")
	fmt.Printf("\x1b[1;33m%s\x1b[0m\n", atkMsg)
	if rageImg != nil {
		renderImageSixel(rageImg, width, 0)
	} else if origImg != nil {
		renderImageSixel(origImg, width, 0)
	}
	fmt.Print("\n\x1b[5m⚡ IMPACT! ⚡\x1b[0m\n")
}

func showCheatSheet(inputCtrl *InputController, loc Localization) {
	fmt.Print("\x1b[2J\x1b[H")
	fmt.Printf("\x1b[1;33m%s\x1b[0m\n\n", loc.CheatSheetTitle)

	order := []string{"1", "2", "4", "5", "6", "7", "9", "10", "12", "14", "15", "16", "17", "18", "20", "22", "23", "24", "25", "26", "27", "43"}
	for _, code := range order {
		if desc, ok := loc.AUDictionary[code]; ok {
			fmt.Printf(" \x1b[1;36mAU%-2s\x1b[0m : %s\n", code, desc)
		}
	}
	fmt.Printf("\n\x1b[1;32m%s\x1b[0m\n", loc.CheatSheetExit)
	<-inputCtrl.commitChan
}

func showTitleScreen(inputCtrl *InputController, currentLoc *Localization, isEN *bool) {
	inputCtrl.SetOnKey(func(curr string) {
		fmt.Printf("\r\x1b[K> \x1b[1;37m%s\x1b[0m", curr)
	})
	defer inputCtrl.SetOnKey(nil)

	for {
		fmt.Print("\x1b[2J\x1b[H")
		fmt.Println("\x1b[1;31m" + currentLoc.TitleBanner + "\x1b[0m")
		fmt.Printf("\n\x1b[1;33m%s\x1b[0m\n", currentLoc.LangPrompt)
		fmt.Printf("\n\x1b[1;32m%s\x1b[0m\n", currentLoc.StartPrompt)
		fmt.Print("\n> ")

		choice := strings.ToLower(<-inputCtrl.commitChan)
		if choice == "e" || choice == "en" {
			*isEN = true
			*currentLoc = locEN
			continue
		} else if choice == "j" || choice == "ja" {
			*isEN = false
			*currentLoc = locJA
			continue
		} else {
			break
		}
	}
}

func printBattleReport(records []StageRecord, totalScore int, inputCtrl *InputController, loc Localization) {
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

		// サムネイル表示 (Sixel: 横幅 160px)
		if f, err := os.Open(rec.FilePath); err == nil {
			if thumbImg, _, err := image.Decode(f); err == nil {
				renderImageSixel(thumbImg, 160, 0)
			}
			f.Close()
		}

		// 公式正解表記
		if rec.RawScore != "" {
			fmt.Printf("\n  \x1b[1;35m公式FACSスコア\x1b[0m : \x1b[1;37m%s\x1b[0m\n", rec.RawScore)
		}

		// 正解AUリストと部位名
		var targetCodes []string
		for _, t := range rec.Targets {
			desc := loc.AUDictionary[t.Code]
			targetCodes = append(targetCodes, fmt.Sprintf("AU%s (%s)", t.Code, desc))
		}
		fmt.Printf("  \x1b[1;35m%s\x1b[0m: %s\n", loc.TargetLabel, strings.Join(targetCodes, ", "))

		// ヒットしたAU
		if len(rec.HitAUs) > 0 {
			var hits []string
			for _, h := range rec.HitAUs {
				hits = append(hits, fmt.Sprintf("AU%s", h))
			}
			fmt.Printf("  \x1b[1;32m%s\x1b[0m   : %s\n", loc.HitsLabel, strings.Join(hits, ", "))
			totalHits += len(rec.HitAUs)
		} else {
			fmt.Printf("  \x1b[1;32m%s\x1b[0m   : (None)\n", loc.HitsLabel)
		}

		// ミス入力
		if len(rec.MissInputs) > 0 {
			fmt.Printf("  \x1b[1;31m%s\x1b[0m   : %s\n", loc.MissesLabel, strings.Join(rec.MissInputs, ", "))
			totalMisses += len(rec.MissInputs)
		}

		// ★ ポール・エクマン公式の解剖学的判定理由 (Rationale)
		if rec.Rationale != "" {
			fmt.Println("\n  \x1b[1;34m【FACS 解剖判定解説 (Rationale)】\x1b[0m:")
			// 長文解説を行儀よく折り返して表示
			wrapRationale(rec.Rationale, 70, "    ")
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
	<-inputCtrl.commitChan
}

// ターミナル幅に合わせて長文を綺麗に改行するヘルパー関数
func wrapRationale(text string, width int, prefix string) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}

	line := prefix
	for _, w := range words {
		if len(line)+len(w)+1 > width {
			fmt.Println(line)
			line = prefix + w
		} else {
			if line == prefix {
				line += w
			} else {
				line += " " + w
			}
		}
	}
	if line != prefix {
		fmt.Println(line)
	}
}

func main() {
	langFlag := flag.String("lang", "ja", "Initial language: ja or en")
	flag.Parse()

	isEN := strings.ToLower(*langFlag) == "en"
	currentLoc := locJA
	if isEN {
		currentLoc = locEN
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize database and repositories using DDD infrastructure
	db, err := sqlite3.NewDB(ctx, "app.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	qRepo := sqlite3.NewQuestionRepository(db)
	battleUsecase := usecase.NewBattleUsecase(qRepo)

	questions, err := battleUsecase.GetStageQuestions(ctx, 5)
	if err != nil || len(questions) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No questions found in app.db. Please run 'make import' first!\n")
		return
	}

	// Shuffle questions randomly
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})

	// Enter raw terminal mode safely
	restoreTerm := setTerminalRawMode()
	defer restoreTerm()

	inputCtrl := NewInputController(ctx)
	showTitleScreen(inputCtrl, &currentLoc, &isEN)

	bgm := startBGM("assets/sounds/stage1.mp3")
	defer func() {
		if bgm != nil {
			bgm.Stop()
		}
	}()

	playerLife := 5
	score := 0
	imageWidth := 280
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

		rageImg := createRageImage(img)

		remaining := make(map[string]entity.TargetAU)
		for _, t := range q.Targets {
			remaining[t.Code] = t
		}
		maxHP := len(remaining)
		currentHP := maxHP
		var hitLog []string

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

		// Initial full-screen render
		fmt.Print("\x1b[2J\x1b[H")

		renderHUD := func() {
			fmt.Print("\x1b[H") // Return cursor to row 1 without clearing Sixel image pixels below
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

		// Initial display: HUD, Sixel image, and prompt
		renderHUD()
		renderImageSixel(img, imageWidth, 0)
		fmt.Printf("\n%s\x1b[K\n", currentLoc.StrikePrompt)

		redrawFullStage := func(customBanner string) {
			fmt.Print("\x1b[2J\x1b[H")
			renderHUD()
			if customBanner != "" {
				fmt.Println(customBanner)
			}
			renderImageSixel(img, imageWidth, 0)
			if len(hitLog) > 0 {
				fmt.Printf("\x1b[32m%s\x1b[0m\x1b[K\n", currentLoc.DestroyedHeader)
				for _, l := range hitLog {
					fmt.Println("  " + l + "\x1b[K")
				}
			}
			fmt.Printf("\n%s%s\x1b[K", currentLoc.StrikePrompt, inputCtrl.GetCurrentInput())
		}

		// Zero-latency keystroke echo callback
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
				enemyATB += rng.Intn(4) + 2
				if enemyATB >= 100 {
					playerLife--
					enemyATB = 0
					atkMsg := currentLoc.EnemyAttacks[rng.Intn(len(currentLoc.EnemyAttacks))]

					playSound("assets/sounds/damage.wav")
					enemyAttackFlash(rageImg, img, imageWidth, currentLoc.RageAttackTitle, atkMsg)

					time.Sleep(700 * time.Millisecond)
					if playerLife <= 0 {
						stageOver = true
					}
					redrawFullStage("")
				} else {
					// Differential HUD update: no Sixel re-transmission
					renderHUD()
				}
				mu.Unlock()

			case committedLine := <-inputCtrl.commitChan:
				if committedLine == "" {
					continue
				}

				if committedLine == "?" || committedLine == "HINT" {
					showCheatSheet(inputCtrl, currentLoc)
					redrawFullStage("")
					continue
				}

				tokens := strings.FieldsFunc(committedLine, func(r rune) bool {
					return r == ' ' || r == ',' || r == ';'
				})

				mu.Lock()
				hitAny := false
				for _, token := range tokens {
					cleaned := strings.TrimSpace(strings.ReplaceAll(strings.ToUpper(token), "AU", ""))
					if cleaned == "" {
						continue
					}

					if target, found := remaining[cleaned]; found {
						hitAny = true
						playSound("assets/sounds/hit.wav")

						delete(remaining, cleaned)
						currentRecord.HitAUs = append(currentRecord.HitAUs, cleaned)
						currentHP--
						score += 300

						if enemyATB > 40 {
							enemyATB -= 40
						} else {
							enemyATB = 0
						}

						desc := currentLoc.AUDictionary[target.Code]
						hitLog = append(hitLog, fmt.Sprintf("💥 AU%s: %s", target.Code, desc))

						if currentHP == 0 {
							stageOver = true
							currentRecord.Cleared = true
							playSound("assets/sounds/ko.wav")

							fmt.Print("\x1b[2J\x1b[H")
							fmt.Println(currentLoc.KoTitle)
							renderImageSixel(img, imageWidth, 0)
							fmt.Println(currentLoc.KoSubtitle)
							for _, l := range hitLog {
								fmt.Println("  " + l)
							}
							time.Sleep(1800 * time.Millisecond)
							break
						}
					} else {
						currentRecord.MissInputs = append(currentRecord.MissInputs, cleaned)
					}
				}

				if !hitAny && currentHP > 0 {
					fmt.Println(currentLoc.MissBanner)
					enemyATB += 15
					time.Sleep(300 * time.Millisecond)
				}

				if !stageOver {
					redrawFullStage(currentLoc.HitBanner)
				}
				mu.Unlock()
			}
		}

		inputCtrl.SetOnKey(nil)
		ticker.Stop()
		battleHistory = append(battleHistory, currentRecord)
	}

	if bgm != nil {
		bgm.Stop()
	}

	loseBgm := startBGM("assets/sounds/lose.mp3")
	if loseBgm == nil {
		playSound("assets/sounds/lose.wav")
	}
	defer func() {
		if loseBgm != nil {
			loseBgm.Stop()
		}
	}()

	printBattleReport(battleHistory, score, inputCtrl, currentLoc)
}
