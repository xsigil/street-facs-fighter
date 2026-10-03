package entity

type TargetAU struct {
	Side      string // "L", "R", ""
	Code      string // "1", "12" など
	Intensity string // "A"〜"E", または "X", "Y", "Z"
}

type Question struct {
	ID       string
	FilePath string
	FileName string
	Targets  []TargetAU
}
