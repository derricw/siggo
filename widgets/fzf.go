package widgets

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

// FZFFile opens up FZF and fuzzy-searches for a file
func FZFFile() (string, error) {
	cmd := exec.Command("fzf", "--prompt=attach: ", "--margin=10%,10%,10%,10%", "--border")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	buf := bytes.NewBuffer([]byte{})
	cmd.Stdout = buf
	usr, err := user.Current()
	if err != nil {
		// if we can find home, we run fzf from there
		return "", err
	}
	cmd.Dir = usr.HomeDir

	if err = cmd.Run(); cmd.ProcessState.ExitCode() == 130 {
		// exit code 130 is when we cancel FZF
		// not an error
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("failed to find a file: %s", err)
	}

	f := strings.TrimSpace(buf.String())
	path := filepath.Join(usr.HomeDir, f)
	return path, err
}

// FZFList opens up FZF and fuzzy-searches from a list of strings
func FZFList(items []string) (string, error) {
	cmd := exec.Command("fzf", "--prompt=contact: ", "--margin=10%,10%,10%,10%", "--border")
	//cmd.Stdin = os.Stdin
	inputBuffer := strings.NewReader(strings.Join(items, "\n"))
	cmd.Stdin = inputBuffer
	cmd.Stderr = os.Stderr
	buf := bytes.NewBuffer([]byte{})
	cmd.Stdout = buf
	if err := cmd.Run(); cmd.ProcessState.ExitCode() == 130 {
		// exit code 130 is when we cancel FZF
		// not an error
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("failed to call fzf: %s", err)
	}

	result := strings.TrimSpace(buf.String())
	return result, nil
}
