package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"os"
)

// writeWavHeader writes standard 44.1kHz 16-bit Mono WAV headers
func writeWavHeader(f *os.File, sampleRate, numSamples int) {
	byteRate := sampleRate * 2 // 16-bit mono
	blockAlign := 2
	dataSize := numSamples * 2
	chunkSize := 36 + dataSize

	f.WriteString("RIFF")
	binary.Write(f, binary.LittleEndian, int32(chunkSize))
	f.WriteString("WAVEfmt ")
	binary.Write(f, binary.LittleEndian, int32(16))
	binary.Write(f, binary.LittleEndian, int16(1)) // PCM
	binary.Write(f, binary.LittleEndian, int16(1)) // Mono
	binary.Write(f, binary.LittleEndian, int32(sampleRate))
	binary.Write(f, binary.LittleEndian, int32(byteRate))
	binary.Write(f, binary.LittleEndian, int16(blockAlign))
	binary.Write(f, binary.LittleEndian, int16(16)) // 16-bit
	f.WriteString("data")
	binary.Write(f, binary.LittleEndian, int32(dataSize))
}

// 1. クリティカルHIT音 (鋭いパンチ・打撃音)
func generateHitSE(path string) error {
	sampleRate := 44100
	duration := 0.20
	numSamples := int(float64(sampleRate) * duration)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	writeWavHeader(f, sampleRate, numSamples)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		env := math.Exp(-t * 24.0)

		// 周波数が750Hzから70Hzへ急降下
		freq := 750.0*math.Exp(-t*26.0) + 70.0
		sine := math.Sin(2.0 * math.Pi * freq * t)

		// アタック瞬間のホワイトノイズ
		noise := (rand.Float64()*2.0 - 1.0) * math.Exp(-t*60.0)

		val := (sine*0.75 + noise*0.6) * env
		val = math.Max(-1.0, math.Min(1.0, val))
		binary.Write(f, binary.LittleEndian, int16(val*32767.0))
	}
	return nil
}

// 2. 敵の怒り攻撃被弾音 (重低音クラッシュ)
func generateDamageSE(path string) error {
	sampleRate := 44100
	duration := 0.40
	numSamples := int(float64(sampleRate) * duration)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	writeWavHeader(f, sampleRate, numSamples)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		env := math.Exp(-t * 8.5)

		// 歪んだ低音矩形波
		freq := 110.0*math.Exp(-t*5.0) + 38.0
		sq := 0.0
		if math.Sin(2.0*math.Pi*freq*t) > 0 {
			sq = 0.6
		} else {
			sq = -0.6
		}
		noise := (rand.Float64()*2.0 - 1.0) * 0.65

		val := (sq*0.5 + noise*0.65) * env
		val = math.Max(-1.0, math.Min(1.0, val))
		binary.Write(f, binary.LittleEndian, int16(val*32767.0))
	}
	return nil
}

// 3. TARGET K.O. 勝利ファンファーレ
func generateKOSE(path string) error {
	sampleRate := 44100
	notes := []float64{392.0, 523.25, 659.25, 783.99} // G4, C5, E5, G5
	noteDuration := 0.12
	totalDuration := noteDuration * float64(len(notes))
	numSamples := int(float64(sampleRate) * totalDuration)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	writeWavHeader(f, sampleRate, numSamples)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		noteIdx := int(t / noteDuration)
		if noteIdx >= len(notes) {
			noteIdx = len(notes) - 1
		}
		freq := notes[noteIdx]
		noteT := math.Mod(t, noteDuration)
		env := math.Exp(-noteT * 7.0)

		// ファミコン風の矩形波
		sq := 0.0
		if math.Sin(2.0*math.Pi*freq*noteT) > 0 {
			sq = 0.5
		} else {
			sq = -0.5
		}

		val := sq * env
		val = math.Max(-1.0, math.Min(1.0, val))
		binary.Write(f, binary.LittleEndian, int16(val*32767.0))
	}
	return nil
}

// 4. DEFEAT 敗北ジングル (哀愁・脱力の「たららららん...」)
func generateLoseSE(path string) error {
	sampleRate := 44100

	// 脱力感あふれる下降フレーズ (G4 -> F4 -> Eb4 -> D4 -> C4)
	notes := []struct {
		freq     float64
		duration float64
	}{
		{freq: 392.00, duration: 0.13}, // G4 (タ)
		{freq: 349.23, duration: 0.13}, // F4 (ラ)
		{freq: 311.13, duration: 0.13}, // Eb4 (ラ)
		{freq: 293.66, duration: 0.14}, // D4 (ラ)
		{freq: 261.63, duration: 0.65}, // C4 (ラ〜〜ン...最後ピッチがポヨンと下がる)
	}

	totalDuration := 0.0
	for _, n := range notes {
		totalDuration += n.duration
	}
	numSamples := int(float64(sampleRate) * totalDuration)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	writeWavHeader(f, sampleRate, numSamples)

	elapsed := 0.0
	for noteIdx, n := range notes {
		noteSamples := int(float64(sampleRate) * n.duration)
		for s := 0; s < noteSamples; s++ {
			t := float64(s) / float64(sampleRate)
			env := math.Exp(-t * 4.5)

			freq := n.freq
			// 最後の「ラ〜ン」はため息のようにピッチが脱力して下降
			if noteIdx == len(notes)-1 {
				vibrato := math.Sin(2.0*math.Pi*6.0*t) * 4.0
				freq = (n.freq*math.Exp(-t*1.8) + vibrato)
			}

			// ファミコン風の矩形波（パルス波）
			sq := 0.0
			phase := math.Mod(t*freq, 1.0)
			if phase < 0.35 {
				sq = 0.55
			} else {
				sq = -0.55
			}

			val := sq * env * 0.85
			val = math.Max(-1.0, math.Min(1.0, val))
			binary.Write(f, binary.LittleEndian, int16(val*32767.0))
		}
		elapsed += n.duration
	}
	return nil
}

func main() {
	outDir := "assets/sounds"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "ディレクトリ作成失敗: %v\n", err)
		os.Exit(1)
	}

	if err := generateHitSE(outDir + "/hit.wav"); err != nil {
		fmt.Printf("HIT音生成失敗: %v\n", err)
	}
	if err := generateDamageSE(outDir + "/damage.wav"); err != nil {
		fmt.Printf("DAMAGE音生成失敗: %v\n", err)
	}
	if err := generateKOSE(outDir + "/ko.wav"); err != nil {
		fmt.Printf("KO音生成失敗: %v\n", err)
	}
	if err := generateLoseSE(outDir + "/lose.wav"); err != nil {
		fmt.Printf("LOSE音生成失敗: %v\n", err)
	}

	fmt.Println("[+] assets/sounds/ に hit.wav, damage.wav, ko.wav, lose.wav を生成しました！")
}