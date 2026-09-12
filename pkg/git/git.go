package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func Exec(args ...string) (string, error) {
	cmd := exec.Command("git", args...)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git: %s: %w", strings.Join(args, " "), err)
	}

	res := strings.TrimSpace(out.String())

	return res, nil
}

func OriginHead() (string, error) {
	return Exec("rev-parse", "--abbrev-ref", "origin/HEAD")
}

func Diff(branch string) (string, error) {
	return Exec("diff", branch, "--", "*.go")
}
