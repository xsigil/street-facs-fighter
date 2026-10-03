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

// 各ステージの戦闘履歴を記録する構造体
type StageRecord struct {
	StageNum   int
	FilePath   string
	Targets    []entity.TargetAU
	HitAUs     []string
	MissInputs []string
	Cleared    bool
}

// BGM再生プロセスを管理する構造体
type BGMController struct {
	cmd *exec.Cmd
	mu  sync.Mutex
}

// BGMをバックグラウンドでループ再生する
func startBGM(candidates []string) *BGMController {
	var bgmPath string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			bgmPath = p
			break
		}
	}
	if bgmPath == "" {
		return nil
	}

	ctrl := &BGMController{}

	go func() {
		ctrl.mu.Lock()
		if path, err := exec.LookPath("mpv"); err == nil {
			ctrl.cmd = exec.Command(path, "--no-video", "--loop=inf", "--volume=65", bgmPath)
			ctrl.mu.Unlock()
			_ = ctrl.cmd.Run()
			return
		}
		if path, err := exec.LookPath("ffplay"); err == nil {
			ctrl.cmd = exec.Command(path, "-nodisp", "-loop", "0", "-volume", "65", bgmPath)
			ctrl.mu.Unlock()
			_ = ctrl.cmd.Run()
			return
		}
		player := "pw-play"
		if _, err := exec.LookPath(player); err != nil {
			player = "paplay"
		}
		loopScript := fmt.Sprintf("while true; do %s %q; done", player, bgmPath)
		ctrl.cmd = exec.Command("bash", "-c", loopScript)
		ctrl.mu.Unlock()
		_ = ctrl.cmd.Run()
	}()

	return ctrl
}

// BGMを停止する
func (b *BGMController) Stop() {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cmd != nil && b.cmd.Process != nil {
		_ = b.cmd.Process.Kill()
		_ = b.cmd.Wait()
	}
}

// 非同期で効果音・ジングルを再生する
func playSound(soundPath string) {
	if _, err := os.Stat(soundPath); os.IsNotExist(err) {
		return
	}
	go func() {
		if path, err := exec.LookPath("mpv"); err == nil {
			cmd := exec.Command(path, "--no-video", "--volume=85", soundPath)
			_ = cmd.Run()
			return
		}
		if path, err := exec.LookPath("ffplay"); err == nil {
			cmd := exec.Command(path, "-nodisp", "-autoexit", "-volume", "85", soundPath)
			_ = cmd.Run()
			return
		}
		players := []string{"pw-play", "paplay", "aplay"}
		for _, player := range players {
			if path, err := exec.LookPath(player); err == nil {
				cmd := exec.Command(path, soundPath)
				_ = cmd.Run()
				return
			}
		}
	}()
}

// 多言語メッセージ定義
type Localization struct {
	Lang            string
	AUDictionary    map[string]string
	EnemyAttacks    []string
	TitleInstruct   []string
	TitleLangPrompt string
	StartPrompt     string
	CheatSheetTitle string
	CheatSheetExit  string
	TargetHP        string
	EnemyATB        string
	DestroyedHeader string
	StrikePrompt    string
	RageAttackTitle string
	HitBanner       string
	MissBanner      string
	KoTitle         string
	KoSubtitle      string
	DefeatMsg       string
	VictoryMsg      string
	ReportTitle     string
	ReportHitBadge  string
	ReportLoseBadge string
	ReportTargetAU  string
	ReportHit       string
	ReportMiss      string
	ReportPerfect   string
	ReportNone      string
	ReportColItem   string
	ReportColData   string
	ReportStages    string
	ReportTargets   string
	ReportAccuracy  string
	ReportScore     string
	ReportRankTitle string
	Ranks           [4]string
	ExitPrompt      string
}

var jaLocale = Localization{
	Lang: "ja",
	AUDictionary: map[string]string{
		"1":  "Inner Brow Raiser (前頭筋 内側部) - 眉の内側が上がり額中央に水平ジワ",
		"2":  "Outer Brow Raiser (前頭筋 外側部) - 眉の外側が持ち上がる",
		"4":  "Brow Lowerer (皺眉筋/鼻根筋) - 眉が下がり眉間に縦ジワ",
		"5":  "Upper Lid Raiser (上眼瞼挙筋) - 上まぶたが上がり強膜露出",
		"6":  "Cheek Raiser (眼輪筋 眼窩部) - 目尻にカラスの足跡シワ",
		"7":  "Lid Tightener (眼輪筋 眼瞼部) - 下まぶた緊張・細目",
		"9":  "Nose Wrinkler (上唇鼻翼挙筋) - 鼻背にシワ・上唇引き上げ",
		"10": "Upper Lip Raiser (上唇挙筋) - 上前歯が見える引き上げ",
		"12": "Lip Corner Puller (大頬骨筋) - 口角斜め上（笑顔）",
		"14": "Dimpler (頬筋) - 口角奥へエクボ状",
		"15": "Lip Corner Depressor (口角下制筋) - への字口",
		"16": "Lower Lip Depressor (下唇下制筋) - 下唇が引き下げられる",
		"17": "Chin Raiser (オトガイ筋) - 顎の梅干しジワ",
		"18": "Lip Puckerer (口輪筋 切歯部) - 唇をすぼめる",
		"20": "Lip Stretcher (笑筋) - 口角水平引き伸ばし",
		"22": "Lip Funneler (口輪筋) - 唇を筒状に突き出す",
		"23": "Lip Tightener (口輪筋) - 唇引き締め",
		"24": "Lip Pressor (口輪筋) - 唇を押しつぶす",
		"25": "Lips Part (下顎下制) - 唇開き",
		"26": "Jaw Drop (咬筋弛緩) - 顎の自然落下",
		"27": "Mouth Stretch (外側翼突筋等) - 口の最大開口",
	},
	EnemyAttacks: []string{
		"【睨みつけ】 眉間の圧迫感で精神にダメージ！",
		"【冷笑の波動】 プレイヤーのメンタルが削られた！",
		"【怒りの咆哮】 激しい表情圧に圧倒された！",
		"【不気味な微笑】 精神的プレッシャーを受けた！",
	},
	TitleInstruct: []string{
		"・敵写真の表情筋（AU）をキーボードから撃ち込んでHPを削れ！",
		"・複数部位はスペース区切り（例: '1 4'）でまとめ撃ち（コンボ）可能！",
		"・相手のATBゲージが100%になると表情攻撃を被弾します！",
		"・迷ったら '?' であんちょこ閲覧可能！",
	},
	TitleLangPrompt: "\x1b[36m言語切替 (Switch Lang):\x1b[0m [\x1b[1;32mJ: 日本語\x1b[0m] / [\x1b[33mE: English\x1b[0m] (切替は 'e' または 'j' を入力してEnter)",
	StartPrompt:     ">>> [Enter]キーを押して FIGHT! <<<",
	CheatSheetTitle: "=========== 📖 FACS 弱点解析あんちょこ (Cheat Sheet) ===========",
	CheatSheetExit:  "\n[Enter]キーで戦闘に戻る...",
	TargetHP:        "TARGET HP",
	EnemyATB:        "ENEMY ATB",
	DestroyedHeader: "【破壊した表情部位】:",
	StrikePrompt:    "急所AUを撃て！ (例: 12 / 複数: 1 4 / ヒント: '?'): > ",
	RageAttackTitle: "   [!] ENEMY RAGE ATTACK: 敵の表情圧が炸裂！ [AU4+AU5]",
	HitBanner:       "      ⚡⚡⚡ CRITICAL HIT!! ⚡⚡⚡",
	MissBanner:      "\n\x1b[33m空振り！ 反動で敵のチャージが加速した！\x1b[0m",
	KoTitle:         "\n\x1b[1;33m🌟 TARGET K.O.!! STAGE CLEAR! 🌟\x1b[0m\n",
	KoSubtitle:      "\n【完全解読完了】",
	DefeatMsg:       "\x1b[1;31m💀 DEFEAT... 表情の圧力に屈してしまった... 💀\x1b[0m",
	VictoryMsg:      "\x1b[1;32m🏆 VICTORY! すべての表情ボスを撃破した！ 🏆\x1b[0m",
	ReportTitle:     "        📊 STREET FACS FIGHTER - 戦闘結果解析レポート (REVIEW) 📊       ",
	ReportHitBadge:  "[ K.O. 撃破 ]",
	ReportLoseBadge: "[ 敗北/未クリア ]",
	ReportTargetAU:  "【写真の正解AU】:",
	ReportHit:       "【命中 (HIT)】",
	ReportMiss:      "【誤答 (MISS)】",
	ReportPerfect:   "なし (パーフェクト！)",
	ReportNone:      "なし",
	ReportColItem:   "項目",
	ReportColData:   "戦績データ",
	ReportStages:    "進行ステージ数",
	ReportTargets:   "総ターゲット破壊数",
	ReportAccuracy:  "打撃命中精度 (正答率)",
	ReportScore:     "最終獲得スコア",
	ReportRankTitle: "FACS 解析官ランク",
	Ranks: [4]string{
		"S 級 (FACS 認定マスター・神の眼)",
		"A 級 (一流 FACS ストリートファイター)",
		"B 級 (敏腕 表情筋ハンター)",
		"C 級 (表情観察員 見習い)",
	},
	ExitPrompt: ">>> [Enter]キーを押してゲームを終了します <<<",
}

var enLocale = Localization{
	Lang: "en",
	AUDictionary: map[string]string{
		"1":  "Inner Brow Raiser (Frontalis, pars medialis) - Pulls inner eyebrows up, horizontal forehead wrinkles",
		"2":  "Outer Brow Raiser (Frontalis, pars lateralis) - Pulls outer eyebrows up",
		"4":  "Brow Lowerer (Corrugator/Depressor supercilii) - Lowers & draws eyebrows together, vertical furrows",
		"5":  "Upper Lid Raiser (Levator palpebrae superioris) - Widens eye aperture, exposes upper sclera",
		"6":  "Cheek Raiser (Orbicularis oculi, pars orbitalis) - Crow's feet wrinkles, elevates cheeks",
		"7":  "Lid Tightener (Orbicularis oculi, pars palpebralis) - Narrows eye opening, tightens lower lid",
		"9":  "Nose Wrinkler (Levator labii superioris alaeque nasi) - Wrinkles nose bridge, elevates upper lip",
		"10": "Upper Lip Raiser (Levator labii superioris) - Elevates center of upper lip, exposes teeth",
		"12": "Lip Corner Puller (Zygomaticus major) - Pulls lip corners up & back (Smile)",
		"14": "Dimpler (Buccinator) - Compresses and dimples lip corners inward",
		"15": "Lip Corner Depressor (Depressor anguli oris) - Depresses lip corners downward (Frown)",
		"16": "Lower Lip Depressor (Depressor labii inferioris) - Pulls lower lip downward",
		"17": "Chin Raiser (Mentalis) - Pushes chin boss and lower lip upward, chin dimpling",
		"18": "Lip Puckerer (Incisivii labii) - Purses and pushes lips forward",
		"20": "Lip Stretcher (Risorius) - Stretches lip corners horizontally",
		"22": "Lip Funneler (Orbicularis oris) - Funnels and rounds lips outward",
		"23": "Lip Tightener (Orbicularis oris) - Tightens and narrows lips together",
		"24": "Lip Pressor (Orbicularis oris) - Presses lips tightly against each other",
		"25": "Lips Part (Mandibular depression) - Lips parted slightly, dental separation",
		"26": "Jaw Drop (Masseter relaxation) - Natural lowering of mandible",
		"27": "Mouth Stretch (Lateral pterygoid) - Maximum mouth opening",
	},
	EnemyAttacks: []string{
		"【Glower Strike】 Mental damage from furrowed brow pressure!",
		"【Scornful Smirk】 Player's mental fortitude was drained!",
		"【Furious Roar】 Overwhelmed by sheer facial intensity!",
		"【Eerie Grin】 High psychological pressure inflicted!",
	},
	TitleInstruct: []string{
		"• Type the target face's Action Units (AUs) to drain HP!",
		"• Combo attacks supported with space separation (e.g. '1 4')!",
		"• Enemy unleashes an expression strike when ATB reaches 100%!",
		"• Stuck? Type '?' to open the FACS Cheat Sheet!",
	},
	TitleLangPrompt: "\x1b[36mSwitch Language:\x1b[0m [\x1b[33mJ: 日本語\x1b[0m] / [\x1b[1;32mE: English\x1b[0m] (Type 'j' or 'e' then press Enter)",
	StartPrompt:     ">>> Press [Enter] to FIGHT! <<<",
	CheatSheetTitle: "=========== 📖 FACS WEAKNESS CHEAT SHEET ===========",
	CheatSheetExit:  "\nPress [Enter] to return to battle...",
	TargetHP:        "TARGET HP",
	EnemyATB:        "ENEMY ATB",
	DestroyedHeader: "【DESTROYED ACTION UNITS】:",
	StrikePrompt:    "Strike weak AU! (e.g. 12 / combo: 1 4 / hint: '?'): > ",
	RageAttackTitle: "   [!] ENEMY RAGE ATTACK: Intense Expression Wave! [AU4+AU5]",
	HitBanner:       "      ⚡⚡⚡ CRITICAL HIT!! ⚡⚡⚡",
	MissBanner:      "\n\x1b[33mMISS! The recoil accelerated enemy ATB charge!\x1b[0m",
	KoTitle:         "\n\x1b[1;33m🌟 TARGET K.O.!! STAGE CLEAR! 🌟\x1b[0m\n",
	KoSubtitle:      "\n【FULL DECODING COMPLETE】",
	DefeatMsg:       "\x1b[1;31m💀 DEFEAT... Overcome by raw facial pressure... 💀\x1b[0m",
	VictoryMsg:      "\x1b[1;32m🏆 VICTORY! All facial bosses defeated! 🏆\x1b[0m",
	ReportTitle:     "        📊 STREET FACS FIGHTER - BATTLE ANALYSIS REPORT 📊        ",
	ReportHitBadge:  "[ K.O. DEFEATED ]",
	ReportLoseBadge: "[ DEFEAT/FAILED ]",
	ReportTargetAU:  "【TARGET CORRECT AUs】:",
	ReportHit:       "【HITS】",
	ReportMiss:      "【MISSES】",
	ReportPerfect:   "None (PERFECT!)",
	ReportNone:      "None",
	ReportColItem:   "Metric",
	ReportColData:   "Battle Record",
	ReportStages:    "Stages Cleared",
	ReportTargets:   "Total Targets Destroyed",
	ReportAccuracy:  "Hit Accuracy Rate",
	ReportScore:     "Final Score",
	ReportRankTitle: "FACS Analyst Rank",
	Ranks: [4]string{
		"Rank S (Certified FACS Master - Divine Gaze)",
		"Rank A (Top-Tier FACS Street Fighter)",
		"Rank B (Skilled Action Unit Hunter)",
		"Rank C (Apprentice Facial Observer)",
	},
	ExitPrompt: ">>> Press [Enter] to exit the game <<<",
}

// 彩度と赤みパルス（殺気の脈動）を適用するフィルター
func applySaturationPulse(src image.Image, pulseIntensity float64) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			origColor := src.At(x, y)
			r, g, b, a := origColor.RGBA()

			r8 := float64(r >> 8)
			g8 := float64(g >> 8)
			b8 := float64(b >> 8)

			gray := 0.299*r8 + 0.587*g8 + 0.114*b8
			newR := gray + (r8-gray)*(1.0+pulseIntensity*0.8) + (pulseIntensity * 70.0)
			newG := gray + (g8-gray)*(1.0-pulseIntensity*0.3) - (pulseIntensity * 20.0)
			newB := gray + (b8-gray)*(1.0-pulseIntensity*0.3) - (pulseIntensity * 20.0)

			clamp := func(v float64) uint8 {
				if v < 0 {
					return 0
				}
				if v > 255 {
					return 255
				}
				return uint8(v)
			}

			dst.Set(x, y, color.RGBA{
				R: clamp(newR),
				G: clamp(newG),
				B: clamp(newB),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

// 画像に赤暗いシャドー（被弾エフェクト）をかける
func applyRedShadow(src image.Image, shadowIntensity float64) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			origColor := src.At(x, y)
			r, g, b, a := origColor.RGBA()

			r8 := float64(r >> 8)
			g8 := float64(g >> 8)
			b8 := float64(b >> 8)

			newR := uint8(r8 * (1.0 - shadowIntensity*0.15))
			newG := uint8(g8 * (1.0 - shadowIntensity*0.75))
			newB := uint8(b8 * (1.0 - shadowIntensity*0.75))

			dst.Set(x, y, color.RGBA{R: newR, G: newG, B: newB, A: uint8(a >> 8)})
		}
	}
	return dst
}

// Sixelでターミナルに画像を描画
func renderImageSixel(img image.Image, targetWidth int, offsetX int) {
	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	targetHeight := (origH * targetWidth) / origW
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.BiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	var buf bytes.Buffer
	enc := sixel.NewEncoder(&buf)
	if err := enc.Encode(dst); err != nil {
		fmt.Printf("Sixel encode error: %v\n", err)
		return
	}

	if offsetX > 0 {
		fmt.Print(strings.Repeat(" ", offsetX))
	}
	os.Stdout.Write(buf.Bytes())
	fmt.Print("\x1b[K\n")
}

// プレイヤー命中時のシェイク演出
func shakePlayerHit(img image.Image, width int, banner string) {
	offsets := []int{2, 0, 1, 0}
	for _, offset := range offsets {
		fmt.Printf("\x1b[H\n\x1b[1;32m%s\x1b[0m\n\n", banner)
		renderImageSixel(img, width, offset)
		time.Sleep(35 * time.Millisecond)
	}
}

// 敵攻撃時の被弾演出
func enemyAttackFlash(rageImg image.Image, defaultImg image.Image, width int, title string, msg string) {
	displayImg := defaultImg
	if rageImg != nil {
		displayImg = rageImg
	}

	shadowImg := applyRedShadow(displayImg, 0.70)
	offsets := []int{2, 0, 1, 0}
	for i, offset := range offsets {
		fmt.Print("\x1b[H\n")
		fmt.Printf("\x1b[1;31m%s\x1b[0m\n", title)
		fmt.Printf("\x1b[1;33m   %s\x1b[0m\n", msg)

		if i < 2 {
			renderImageSixel(shadowImg, width, offset)
		} else {
			renderImageSixel(displayImg, width, offset)
		}
		time.Sleep(75 * time.Millisecond)
	}
}

// 独立スレッドでキーボード入力を受けるコントローラ
type InputController struct {
	mu         sync.Mutex
	currentBuf string
	commitChan chan string
}

func newTerminalInputController(ctx context.Context) *InputController {
	c := &InputController{
		commitChan: make(chan string, 16),
	}

	_ = exec.Command("stty", "-F", "/dev/tty", "-icanon", "-echo").Run()

	go func() {
		defer close(c.commitChan)
		buf := make([]byte, 1)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				time.Sleep(20 * time.Millisecond)
				continue
			}

			b := buf[0]

			if b == 3 { // Ctrl+C
				p, _ := os.FindProcess(os.Getpid())
				_ = p.Signal(os.Interrupt)
				return
			}

			c.mu.Lock()
			if b == '\r' || b == '\n' {
				line := strings.TrimSpace(c.currentBuf)
				c.currentBuf = ""
				c.mu.Unlock()
				c.commitChan <- strings.ToUpper(line)
				continue
			}

			if b == 127 || b == 8 {
				if len(c.currentBuf) > 0 {
					c.currentBuf = c.currentBuf[:len(c.currentBuf)-1]
				}
				c.mu.Unlock()
				continue
			}

			if b >= 32 && b <= 126 {
				c.currentBuf += string(b)
			}
			c.mu.Unlock()
		}
	}()

	return c
}

func (c *InputController) GetCurrentInput() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.currentBuf
}

// あんちょこ閲覧
func showCheatSheet(inputCtrl *InputController, loc Localization) {
	fmt.Print("\x1b[2J\x1b[H")
	fmt.Printf("\x1b[1;33m%s\x1b[0m\n", loc.CheatSheetTitle)
	for code, desc := range loc.AUDictionary {
		fmt.Printf("AU%-4s: %s\n", code, desc)
	}
	fmt.Println(strings.Repeat("-", 70))
	fmt.Print(loc.CheatSheetExit)
	for range inputCtrl.commitChan {
		break
	}
}

// 怒り顔画像のロード
func loadRageImage() image.Image {
	candidates := []string{
		"assets/s4.gif",
		"assets/s4a.gif",
		"assets/s4b.gif",
		"assets/sW4.gif",
		"assets/s1_4a.gif",
		"assets/s4_5x.gif",
	}

	for _, path := range candidates {
		f, err := os.Open(path)
		if err == nil {
			img, _, err := image.Decode(f)
			f.Close()
			if err == nil {
				return img
			}
		}
	}
	return nil
}

// 戦闘終了後の詳細レポートを描画する
func printBattleReport(records []StageRecord, score int, isVictory bool, loc Localization) {
	fmt.Print("\n\n")
	fmt.Println("\x1b[1;36m========================================================================\x1b[0m")
	fmt.Printf("\x1b[1;37m%s\x1b[0m\n", loc.ReportTitle)
	fmt.Println("\x1b[1;36m========================================================================\x1b[0m\n")

	totalTargets := 0
	totalHits := 0
	totalMisses := 0

	for _, rec := range records {
		totalTargets += len(rec.Targets)
		totalHits += len(rec.HitAUs)
		totalMisses += len(rec.MissInputs)

		resultBadge := fmt.Sprintf("\x1b[1;32m%s\x1b[0m", loc.ReportHitBadge)
		if !rec.Cleared {
			resultBadge = fmt.Sprintf("\x1b[1;31m%s\x1b[0m", loc.ReportLoseBadge)
		}

		fmt.Printf("\x1b[1;33m▶ STAGE %d %s (File: %s)\x1b[0m\n", rec.StageNum, resultBadge, rec.FilePath)
		fmt.Println(strings.Repeat("-", 68))

		if file, err := os.Open(rec.FilePath); err == nil {
			if img, _, err := image.Decode(file); err == nil {
				renderImageSixel(img, 240, 2)
			}
			file.Close()
		}

		fmt.Printf("  \x1b[1m%s\x1b[0m\n", loc.ReportTargetAU)
		for _, t := range rec.Targets {
			desc, ok := loc.AUDictionary[t.Code]
			if !ok {
				desc = "Unknown Action Unit"
			}
			fmt.Printf("    • \x1b[35mAU%-3s\x1b[0m: %s\n", t.Code, desc)
		}

		hitStr := loc.ReportNone
		if len(rec.HitAUs) > 0 {
			var hitsWithPrefix []string
			for _, h := range rec.HitAUs {
				hitsWithPrefix = append(hitsWithPrefix, "AU"+h)
			}
			hitStr = strings.Join(hitsWithPrefix, ", ")
		}
		fmt.Printf("  \x1b[32m%s\x1b[0m : %s\n", loc.ReportHit, hitStr)

		missStr := loc.ReportPerfect
		if len(rec.MissInputs) > 0 {
			var missesWithPrefix []string
			for _, m := range rec.MissInputs {
				missesWithPrefix = append(missesWithPrefix, "AU"+m)
			}
			missStr = strings.Join(missesWithPrefix, ", ")
		}
		fmt.Printf("  \x1b[31m%s\x1b[0m: %s\n\n", loc.ReportMiss, missStr)
	}

	fmt.Println("\x1b[1;36m┌───────────────────┬──────────────────────────────────────────────────┐\x1b[0m")
	fmt.Printf("\x1b[1;36m│\x1b[0m %-17s \x1b[1;36m│\x1b[0m %-48s \x1b[1;36m│\x1b[0m\n", loc.ReportColItem, loc.ReportColData)
	fmt.Println("\x1b[1;36m├───────────────────┼──────────────────────────────────────────────────┤\x1b[0m")

	stageProgress := fmt.Sprintf("%d Stages", len(records))
	if loc.Lang == "ja" {
		stageProgress = fmt.Sprintf("%d ステージ走破", len(records))
	}
	fmt.Printf("\x1b[1;36m│\x1b[0m %-17s \x1b[1;36m│\x1b[0m %-48s \x1b[1;36m│\x1b[0m\n", loc.ReportStages, stageProgress)

	hitStat := fmt.Sprintf("%d / %d units", totalHits, totalTargets)
	if loc.Lang == "ja" {
		hitStat = fmt.Sprintf("%d / %d 部位命中", totalHits, totalTargets)
	}
	fmt.Printf("\x1b[1;36m│\x1b[0m %-17s \x1b[1;36m│\x1b[0m %-48s \x1b[1;36m│\x1b[0m\n", loc.ReportTargets, hitStat)

	accuracy := 0.0
	totalInputs := totalHits + totalMisses
	if totalInputs > 0 {
		accuracy = (float64(totalHits) / float64(totalInputs)) * 100.0
	}
	accStat := fmt.Sprintf("%.1f%% (Total: %d / Misses: %d)", accuracy, totalInputs, totalMisses)
	if loc.Lang == "ja" {
		accStat = fmt.Sprintf("%.1f%% (打撃数: %d / 空振り: %d)", accuracy, totalInputs, totalMisses)
	}
	fmt.Printf("\x1b[1;36m│\x1b[0m %-17s \x1b[1;36m│\x1b[0m %-48s \x1b[1;36m│\x1b[0m\n", loc.ReportAccuracy, accStat)

	scoreStat := fmt.Sprintf("%d pts", score)
	fmt.Printf("\x1b[1;36m│\x1b[0m %-17s \x1b[1;36m│\x1b[0m %-48s \x1b[1;36m│\x1b[0m\n", loc.ReportScore, scoreStat)

	rank := loc.Ranks[3]
	if isVictory && accuracy >= 80.0 {
		rank = loc.Ranks[0]
	} else if isVictory {
		rank = loc.Ranks[1]
	} else if totalHits >= 3 {
		rank = loc.Ranks[2]
	}
	fmt.Printf("\x1b[1;36m│\x1b[0m %-17s \x1b[1;36m│\x1b[0m \x1b[1;33m%-48s\x1b[0m \x1b[1;36m│\x1b[0m\n", loc.ReportRankTitle, rank)
	fmt.Println("\x1b[1;36m└───────────────────┴──────────────────────────────────────────────────┘\x1b[0m\n")
}

func main() {
	langFlag := flag.String("lang", "ja", "UI Language: 'ja' (Japanese) or 'en' (English)")
	flag.Parse()

	currentLoc := jaLocale
	if strings.ToLower(*langFlag) == "en" {
		currentLoc = enLocale
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	defer func() {
		_ = exec.Command("stty", "-F", "/dev/tty", "sane").Run()
	}()

	db, err := sqlite3.NewDB(ctx, "app.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	qRepo := sqlite3.NewQuestionRepository(db)
	battleUsecase := usecase.NewBattleUsecase(qRepo)

	rageImg := loadRageImage()

	questions, err := battleUsecase.GetStageQuestions(ctx, 5)
	if err != nil || len(questions) == 0 {
		fmt.Println("No question images found. Run 'make import' first.")
		return
	}

	inputCtrl := newTerminalInputController(ctx)

	// タイトル画面のループ（言語切り替え対応）
	for {
		fmt.Print("\x1b[2J\x1b[H")
		fmt.Print(`
\x1b[1;31m
   _____ _                  _     \x1b[1;33m______ _    ____ ____  \x1b[1;36m_____ _       _     _            
  / ____| |                | |   \x1b[1;33m|  ____/ \  / ___/ ___| \x1b[1;36m|  ___(_) __ _| |__ | |_ ___ _ __ 
  \___ \| |_ _ __ ___  ___ | |_  \x1b[1;33m| |__ / _ \| |   \___ \ \x1b[1;36m| |_  | |/ _` + "`" + ` | '_ \| __/ _ \ '__|
   ___) |  _| '__/ _ \/ _ \| __| \x1b[1;33m|  __/ ___ \ |___ ___) |\x1b[1;36m|  _| | | (_| | | | | ||  __/ |   
  |____/ \__|_|  \___/\___/ \__| \x1b[1;33m|_| /_/   \_\____|____/ \x1b[1;36m|_|   |_|\__, |_| |_|\__\___|_|   
                                                                 |___/                   
\x1b[0m
          \x1b[1;37m🥊 THE ULTIMATE ACTION UNIT BATTLE ARENA 🥊\x1b[0m
`)
		for _, inst := range currentLoc.TitleInstruct {
			fmt.Println(inst)
		}
		fmt.Println()
		fmt.Println(currentLoc.TitleLangPrompt)
		fmt.Println()
		fmt.Print(currentLoc.StartPrompt + " ")

		titleInput := <-inputCtrl.commitChan

		if titleInput == "E" || titleInput == "EN" || titleInput == "ENGLISH" {
			currentLoc = enLocale
			continue
		}
		if titleInput == "J" || titleInput == "JA" || titleInput == "JP" || titleInput == "JAPANESE" {
			currentLoc = jaLocale
			continue
		}
		// Enter（空文字）またはその他の入力で戦闘開始
		break
	}

	bgmCandidates := []string{
		"assets/sounds/stage1.mp3",
		"assets/sounds/bgm.mp3",
		"assets/sounds/bgm.wav",
		"assets/sounds/bgm.ogg",
	}
	bgm := startBGM(bgmCandidates)
	defer bgm.Stop()

	playerLife := 5
	score := 0
	imageWidth := 500
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var mu sync.Mutex

	var battleHistory []StageRecord

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

		remaining := make(map[string]entity.TargetAU)
		for _, t := range q.Targets {
			remaining[t.Code] = t
		}
		maxHP := len(remaining)
		currentHP := maxHP
		var hitLog []string

		currentRecord := StageRecord{
			StageNum: stageIdx + 1,
			FilePath: q.FilePath,
			Targets:  q.Targets,
		}

		enemyATB := 0
		stageOver := false
		ticker := time.NewTicker(120 * time.Millisecond)
		tickCount := 0

		fmt.Print("\x1b[2J\x1b[H")

		for !stageOver && playerLife > 0 {
			mu.Lock()
			fmt.Print("\x1b[H")

			pLifeBar := strings.Repeat("❤️ ", playerLife)
			fmt.Printf("STAGE %d/%d  |  PLAYER HP: %s |  SCORE: %d\x1b[K\n", stageIdx+1, len(questions), pLifeBar, score)

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
			fmt.Printf("\x1b[1;31m%s: [%s] %d%%\x1b[0m\x1b[K\n", currentLoc.EnemyATB, atbBar, enemyATB)
			fmt.Println(strings.Repeat("-", 50) + "\x1b[K")

			chargeRatio := float64(atbClamped) / 100.0
			pulseWave := (math.Sin(float64(tickCount)*0.35) + 1.0) / 2.0
			pulseIntensity := pulseWave * (0.15 + chargeRatio*0.65)

			pulsedImg := applySaturationPulse(img, pulseIntensity)
			renderImageSixel(pulsedImg, imageWidth, 0)

			if len(hitLog) > 0 {
				fmt.Printf("\x1b[32m%s\x1b[0m\x1b[K\n", currentLoc.DestroyedHeader)
				for _, l := range hitLog {
					fmt.Println("  " + l + "\x1b[K")
				}
			}

			currentTyping := inputCtrl.GetCurrentInput()
			fmt.Printf("\n%s\x1b[1;37m%s\x1b[0m\x1b[K\n", currentLoc.StrikePrompt, currentTyping)
			mu.Unlock()

			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				tickCount++
				enemyATB += rng.Intn(4) + 2
				if enemyATB >= 100 {
					mu.Lock()
					playerLife--
					enemyATB = 0
					atkMsg := currentLoc.EnemyAttacks[rng.Intn(len(currentLoc.EnemyAttacks))]

					playSound("assets/sounds/damage.wav")
					enemyAttackFlash(rageImg, img, imageWidth, currentLoc.RageAttackTitle, atkMsg)

					mu.Unlock()
					time.Sleep(800 * time.Millisecond)

					if playerLife <= 0 {
						stageOver = true
					}
					fmt.Print("\x1b[2J\x1b[H")
				}

			case committedLine := <-inputCtrl.commitChan:
				if committedLine == "" {
					continue
				}

				if committedLine == "?" || committedLine == "HINT" {
					showCheatSheet(inputCtrl, currentLoc)
					fmt.Print("\x1b[2J\x1b[H")
					continue
				}

				tokens := strings.FieldsFunc(committedLine, func(r rune) bool {
					return r == ' ' || r == ',' || r == ';'
				})

				mu.Lock()
				hitAny := false
				for _, token := range tokens {
					cleaned := strings.TrimSpace(strings.ReplaceAll(token, "AU", ""))
					if cleaned == "" {
						continue
					}

					if target, found := remaining[cleaned]; found {
						hitAny = true
						playSound("assets/sounds/hit.wav")
						shakePlayerHit(img, imageWidth, currentLoc.HitBanner)

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
				mu.Unlock()
			}
		}
		ticker.Stop()
		battleHistory = append(battleHistory, currentRecord)
	}

	bgm.Stop()
	_ = exec.Command("stty", "-F", "/dev/tty", "sane").Run()

	fmt.Print("\x1b[2J\x1b[H")
	isVictory := playerLife > 0
	if !isVictory {
		fmt.Println(currentLoc.DefeatMsg)
	} else {
		fmt.Println(currentLoc.VictoryMsg)
	}

	loseCandidates := []string{
		"assets/sounds/lose.mp3",
		"assets/sounds/lose.wav",
	}
	resultBGM := startBGM(loseCandidates)
	defer func() {
		if resultBGM != nil {
			resultBGM.Stop()
		}
	}()

	printBattleReport(battleHistory, score, isVictory, currentLoc)

	fmt.Printf("\x1b[1;33m%s\x1b[0m\n", currentLoc.ExitPrompt)
	<-inputCtrl.commitChan
}