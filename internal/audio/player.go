package audio

import (
	"os"
	"os/exec"
)

type BGMPlayer struct {
	cmd *exec.Cmd
}

func (b *BGMPlayer) Stop() {
	if b != nil && b.cmd != nil && b.cmd.Process != nil {
		_ = b.cmd.Process.Kill()
		_ = b.cmd.Wait()
	}
}

func StartBGM(path string) *BGMPlayer {
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

func PlaySound(path string) {
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
