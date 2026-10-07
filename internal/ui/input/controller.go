package input

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func SetTerminalRawMode() func() {
	cmd := exec.Command("stty", "-F", "/dev/tty", "-icanon", "-echo")
	if err := cmd.Run(); err != nil {
		_ = exec.Command("stty", "-icanon", "-echo").Run()
	}

	return func() {
		cmdReset := exec.Command("stty", "-F", "/dev/tty", "sane")
		if err := cmdReset.Run(); err != nil {
			_ = exec.Command("stty", "sane").Run()
		}
		fmt.Print("\x1b[?25h")
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

				if b == 3 { // Ctrl+C
					p, _ := os.FindProcess(os.Getpid())
					_ = p.Signal(os.Interrupt)
					return
				}

				if b == '\r' || b == '\n' {
					c.mu.Lock()
					line := strings.TrimSpace(c.currentBuf)
					c.currentBuf = ""
					c.mu.Unlock()
					c.commitChan <- line
					continue
				}

				if b == 127 || b == 8 { // Backspace
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

				if b == 27 { // Escape sequences
					if i+2 < n && buf[i+1] == '[' {
						i += 2
					}
					continue
				}

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

func (c *InputController) CommitChan() <-chan string {
	return c.commitChan
}
