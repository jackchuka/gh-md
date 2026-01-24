package executil

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func OpenInEditor(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			editor = "vim"
		}
	}

	editorArgs, err := splitCommandLine(editor)
	if err != nil {
		return fmt.Errorf("invalid editor command %q: %w", editor, err)
	}
	if len(editorArgs) == 0 {
		return fmt.Errorf("invalid editor command %q", editor)
	}

	editorCmd := editorArgs[0]
	editorArgs = append(editorArgs[1:], filePath)

	cmd := exec.Command(editorCmd, editorArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func OpenInBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		if _, err := exec.LookPath("xdg-open"); err == nil {
			cmd = exec.Command("xdg-open", url)
		} else if _, err := exec.LookPath("gio"); err == nil {
			cmd = exec.Command("gio", "open", url)
		} else {
			return fmt.Errorf("no browser opener found (install xdg-utils)")
		}
	}
	return cmd.Run()
}

func CopyToClipboard(text string) error {
	switch runtime.GOOS {
	case "darwin":
		return pipeToCommand("pbcopy", nil, text)
	case "windows":
		return pipeToCommand("cmd", []string{"/c", "clip"}, text)
	default:
		candidates := []struct {
			name string
			args []string
		}{
			{"wl-copy", nil},
			{"xclip", []string{"-selection", "clipboard"}},
			{"xsel", []string{"--clipboard", "--input"}},
		}

		var lastErr error
		for _, c := range candidates {
			if _, err := exec.LookPath(c.name); err != nil {
				lastErr = err
				continue
			}
			if err := pipeToCommand(c.name, c.args, text); err != nil {
				return err
			}
			return nil
		}
		if lastErr != nil {
			return fmt.Errorf("no clipboard tool found (install wl-clipboard or xclip or xsel)")
		}
		return nil
	}
}

func pipeToCommand(name string, args []string, stdin string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// splitCommandLine splits a command line into argv-ish tokens, supporting simple
// single/double quotes and backslash escaping.
func splitCommandLine(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty command")
	}

	var parts []string
	var cur strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	flush := func() {
		if cur.Len() > 0 {
			parts = append(parts, cur.String())
			cur.Reset()
		}
	}

	for _, r := range s {
		if escaped {
			cur.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' && !inSingle {
			escaped = true
			continue
		}

		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
				continue
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
				continue
			}
		}

		if !inSingle && !inDouble && (r == ' ' || r == '\t' || r == '\n' || r == '\r') {
			flush()
			continue
		}

		cur.WriteRune(r)
	}

	if escaped {
		return nil, errors.New("dangling escape")
	}
	if inSingle || inDouble {
		return nil, errors.New("unterminated quote")
	}
	flush()

	return parts, nil
}
