package backend

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

type Tunnel struct {
	cmd *exec.Cmd
	URL string
}

func StartTunnel(localPort int) (*Tunnel, error) {
	cfPath, err := exec.LookPath("cloudflared")
	if err != nil {
		return nil, fmt.Errorf("cloudflared not found: install with `brew install cloudflared`")
	}

	cmd := exec.Command(cfPath, "tunnel", "--url", fmt.Sprintf("http://localhost:%d", localPort))

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	urlCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if idx := strings.Index(line, "trycloudflare.com"); idx != -1 {
				start := strings.LastIndex(line[:idx], "https://")
				if start == -1 {
					start = strings.LastIndex(line[:idx], "http://")
				}
				if start != -1 {
					urlCh <- line[start : idx+len("trycloudflare.com")]
					return
				}
			}
		}
		errCh <- fmt.Errorf("cloudflared exited without providing a URL")
	}()

	select {
	case url := <-urlCh:
		return &Tunnel{cmd: cmd, URL: url}, nil
	case err := <-errCh:
		cmd.Process.Kill()
		return nil, err
	}
}

func (t *Tunnel) Stop() {
	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Kill()
		t.cmd.Wait()
	}
}
