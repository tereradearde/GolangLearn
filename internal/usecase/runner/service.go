package runner

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/example/learngo/pkg/utils"
)

// Result результат выполнения кода.
type Result struct {
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExitCode  int    `json:"exitCode"`
	ElapsedMs int64  `json:"elapsedMs"`
}

// Service интерфейс раннера.
type Service interface {
	Execute(ctx context.Context, code string) (Result, error)
}

// service простая dev-реализация через локальный go run (не для продакшена).
type service struct {
	logger *utils.Logger
	// таймаут на выполнение
	timeout time.Duration
}

// NewService создаёт dev-реализацию.
func NewService(logger *utils.Logger) Service {
	return &service{logger: logger, timeout: 5 * time.Second}
}

func (s *service) Execute(ctx context.Context, code string) (Result, error) {
	if code == "" {
		return Result{}, errors.New("empty code")
	}

	// Контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	workDir, err := os.MkdirTemp("", "go-run-*")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(workDir)

	mainPath := filepath.Join(workDir, "main.go")
	if err := os.WriteFile(mainPath, []byte(code), 0o600); err != nil {
		return Result{}, err
	}

	cmd := exec.CommandContext(ctx, "go", "run", "main.go")
	cmd.Dir = workDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start).Milliseconds()

	exitCode := 0
	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	res := Result{
		Stdout:    stdout.String(),
		Stderr:    stderr.String(),
		ExitCode:  exitCode,
		ElapsedMs: elapsed,
	}
	return res, nil
}
