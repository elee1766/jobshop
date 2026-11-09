package hooks

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/elee1766/jobshop/pkg/execution/automation"
)

// Digest identifies the accepted source and authority manifest.
func (p *Program) Digest() string { return p.digest }

// Prompt resolves prompt files within the accepted program's asset root.
func (p *Program) Prompt(action automation.Action) (string, error) {
	if action.PromptFile == "" {
		if strings.TrimSpace(action.Prompt) == "" {
			return "", errors.New("step prompt must not be blank")
		}
		return action.Prompt, nil
	}
	root, err := os.OpenRoot(p.root)
	if err != nil {
		return "", err
	}
	defer func() { _ = root.Close() }()
	file, err := root.Open(action.PromptFile)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("prompt_file must be a regular file")
	}
	data, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if err != nil {
		return "", err
	}
	if len(data) > 1<<20 {
		return "", errors.New("prompt_file exceeds 1 MiB")
	}
	return string(data), nil
}
