package entity

type TargetAU struct {
	Side      string // "L", "R", ""
	Code      string // "1", "12", "43" など
	Intensity string // "A"〜"E", または "X", "Y", "Z"
}

type Question struct {
	ID        string
	ItemID    string
	Split     string // "practice" | "example"
	MediaType string // "image" | "video"
	RawScore  string
	FilePath  string
	FileName  string
	Rationale string // 解剖学的判定根拠
	Targets   []TargetAU
}
