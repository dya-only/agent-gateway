package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"agent-gateway/internal/config"
	"agent-gateway/internal/openai"
)

type cliStyle string

const (
	cliStyleClaude cliStyle = "claude"
	cliStyleCodex  cliStyle = "codex"
)

type CLIProvider struct {
	bin     string
	args    []string
	workdir string
	style   cliStyle
}

func NewCLIProvider(bin string, args []string, workdir string, style cliStyle) *CLIProvider {
	return &CLIProvider{bin: bin, args: args, workdir: workdir, style: style}
}

func (p *CLIProvider) Complete(ctx context.Context, model config.Model, messages []openai.ChatMessage) (string, error) {
	cmd := p.command(ctx, model, messages)
	if p.workdir != "" {
		cmd.Dir = p.workdir
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("%s failed: %s", p.bin, message)
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (p *CLIProvider) StreamComplete(ctx context.Context, model config.Model, messages []openai.ChatMessage) (<-chan string, <-chan error, error) {
	cmd := p.command(ctx, model, messages)
	if p.workdir != "" {
		cmd.Dir = p.workdir
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	chunks := make(chan string)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		buf := make([]byte, 4096)
		var streamErr error
		for {
			n, readErr := stdout.Read(buf)
			if n > 0 {
				chunks <- string(buf[:n])
			}
			if readErr != nil {
				if readErr != io.EOF {
					streamErr = readErr
				}
				break
			}
		}

		if err := cmd.Wait(); err != nil {
			message := strings.TrimSpace(stderr.String())
			if message == "" {
				message = err.Error()
			}
			streamErr = fmt.Errorf("%s failed: %s", p.bin, message)
		}

		if streamErr != nil {
			errs <- streamErr
		}
	}()

	return chunks, errs, nil
}

func (p *CLIProvider) command(ctx context.Context, model config.Model, messages []openai.ChatMessage) *exec.Cmd {
	prompt := renderPrompt(messages)
	args := append([]string{}, p.args...)

	switch p.style {
	case cliStyleClaude:
		if model.Model != "" {
			args = append(args, "--model", model.Model)
		}
		args = append(args, prompt)
	case cliStyleCodex:
		if model.Model != "" {
			args = append(args, "--model", model.Model)
		}
		args = append(args, prompt)
	default:
		args = append(args, prompt)
	}

	return exec.CommandContext(ctx, p.bin, args...)
}

func renderPrompt(messages []openai.ChatMessage) string {
	var builder strings.Builder
	for _, message := range messages {
		role := strings.TrimSpace(message.Role)
		if role == "" {
			role = "user"
		}
		content := strings.TrimSpace(message.Content)
		if content == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(role))
		builder.WriteString(":\n")
		builder.WriteString(content)
		builder.WriteString("\n\n")
	}
	return strings.TrimSpace(builder.String())
}
