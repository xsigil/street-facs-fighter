package terminal

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/mattn/go-sixel"
	"golang.org/x/image/draw"
)

func RenderImageSixel(img image.Image, targetWidth int, offsetX int) {
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

func CreateRageImage(base image.Image) image.Image {
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

			rRage := uint8(math.Min(255, float64(r8)*1.45+40))
			gRage := uint8(float64(g8) * 0.45)
			bRage := uint8(float64(b8) * 0.45)
			rage.Set(x, y, color.RGBA{R: rRage, G: gRage, B: bRage, A: uint8(a >> 8)})
		}
	}
	return rage
}

func EnemyAttackFlash(rageImg, origImg image.Image, width int, banner, atkMsg string) {
	fmt.Print("\x1b[2J\x1b[H")
	fmt.Println("\x1b[1;31m" + banner + "\x1b[0m")
	fmt.Printf("\x1b[1;33m%s\x1b[0m\n", atkMsg)
	if rageImg != nil {
		RenderImageSixel(rageImg, width, 0)
	} else if origImg != nil {
		RenderImageSixel(origImg, width, 0)
	}
	fmt.Print("\n\x1b[5m⚡ IMPACT! ⚡\x1b[0m\n")
}

func WrapText(text string, width int, prefix string) {
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
