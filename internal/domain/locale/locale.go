package locale

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

var LocJA = Localization{
	TitleBanner: `
===========================================================
  🥊 STREET FACS FIGHTER - ターミナル表情筋格闘ゲーム 🥊
===========================================================`,
	LangPrompt:      "言語切替 (Switch Lang): [J: 日本語] / [E: English] (切替は 'e' または 'j' を入力してEnter)",
	StartPrompt:     ">>> [Enter]キーを押して FIGHT! <<<",
	TargetHP:        "TARGET HP",
	EnemyATB:        "ENEMY TENSION",
	StrikePrompt:    "急所AUを撃て！ (例: 1 4 / ヒント: '?'): ",
	HitBanner:       "💥 クリティカルHIT！急所AUを破壊！",
	MissBanner:      "❌ MISS！相手のテンションが急上昇！",
	KoTitle:         "\n🌟🌟🌟 TARGET K.O.!! 🌟🌟🌟",
	KoSubtitle:      "全ターゲットAUの破壊に成功！",
	RageAttackTitle: "⚡⚡⚡ 敵の怒り攻撃が炸裂！ ⚡⚡⚡",
	DestroyedHeader: "【破壊済みAU部位】:",
	CheatSheetTitle: "【FACS アクション・ユニット (AU) あんちょこ速見表】",
	CheatSheetExit:  "[Enter]キーを押して戦闘に戻る...",
	ReportTitle:     "\n=================== 戦闘結果解析レポート ===================",
	ReportSubtitle:  "各ステージで出現した表情写真と解析結果一覧:",
	StageLabel:      "ステージ",
	TargetLabel:     "正解AU",
	HitsLabel:       "命中AU",
	MissesLabel:     "誤入力",
	ClearedStatus:   "【撃破 K.O.】",
	FailedStatus:    "【被弾敗北】",
	TotalStrikesLabel: "総打撃数",
	AccuracyLabel:   "命中精度 (正答率)",
	FinalScoreLabel: "最終獲得スコア",
	RankLabel:       "FACS解析官ランク",
	ExitPrompt:      "[Enter]キーを押してゲームを終了します...",
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

var LocEN = Localization{
	TitleBanner: `
===========================================================
  🥊 STREET FACS FIGHTER - TERMINAL FACIAL ACTION BATTLE 🥊
===========================================================`,
	LangPrompt:      "Switch Language: [J: Japanese] / [E: English] (Type 'e' or 'j' and press Enter)",
	StartPrompt:     ">>> Press [Enter] to FIGHT! <<<",
	TargetHP:        "TARGET HP",
	EnemyATB:        "ENEMY TENSION",
	StrikePrompt:    "Strike Target AU! (e.g. 1 4 / Hint: '?'): ",
	HitBanner:       "💥 CRITICAL HIT! Target Action Unit Destroyed!",
	MissBanner:      "❌ MISS! Opponent Tension Spiked!",
	KoTitle:         "\n🌟🌟🌟 TARGET K.O.!! 🌟🌟🌟",
	KoSubtitle:      "All target Action Units successfully neutralized!",
	RageAttackTitle: "⚡⚡⚡ ENEMY RAGE ATTACK DETONATED! ⚡⚡⚡",
	DestroyedHeader: "【Destroyed Action Units】:",
	CheatSheetTitle: "【FACS Action Unit (AU) Reference Matrix】",
	CheatSheetExit:  "Press [Enter] to resume combat...",
	ReportTitle:     "\n================ COMBAT DIAGNOSTIC REPORT ================",
	ReportSubtitle:  "Target Facial Expressions and Performance History:",
	StageLabel:      "STAGE",
	TargetLabel:     "Target AU",
	HitsLabel:       "Hit AUs",
	MissesLabel:     "Misses",
	ClearedStatus:   "【CLEARED K.O.】",
	FailedStatus:    "【DEFEATED】",
	TotalStrikesLabel: "Total Strikes",
	AccuracyLabel:   "Strike Accuracy",
	FinalScoreLabel: "Final Combat Score",
	RankLabel:       "FACS Analyst Rank",
	ExitPrompt:      "Press [Enter] to exit the game...",
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
