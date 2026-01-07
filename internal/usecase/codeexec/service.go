package codeexec

import (
	"context"
	"fmt"
	"strings"
	"time"

	codedom "github.com/example/learngo/internal/domain/code"
	"github.com/example/learngo/pkg/codeexec"
	"github.com/example/learngo/pkg/utils"
)

// Service интерфейс сервиса выполнения кода
type Service interface {
	Execute(ctx context.Context, req codedom.ExecuteRequest) (*codedom.ExecuteResponse, error)
}

type service struct {
	judge0Client *codeexec.Client
	logger       *utils.Logger
	timeout      time.Duration
	memoryLimit  int // KB
}

func NewService(judge0Client *codeexec.Client, logger *utils.Logger, timeout time.Duration, memoryLimitKB int) Service {
	return &service{
		judge0Client: judge0Client,
		logger:       logger,
		timeout:      timeout,
		memoryLimit:  memoryLimitKB,
	}
}

func (s *service) Execute(ctx context.Context, req codedom.ExecuteRequest) (*codedom.ExecuteResponse, error) {
	startTime := time.Now()

	// Если есть тесты, выполняем каждый тест отдельно
	if len(req.TestCases) > 0 {
		return s.executeWithTests(ctx, req, startTime)
	}

	// Обычное выполнение без тестов
	return s.executeSimple(ctx, req, startTime)
}

func (s *service) executeSimple(ctx context.Context, req codedom.ExecuteRequest, startTime time.Time) (*codedom.ExecuteResponse, error) {
	languageID, ok := codeexec.LanguageID[req.Language]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", req.Language)
	}

	submissionReq := codeexec.SubmissionRequest{
		SourceCode:   req.Code,
		LanguageID:   languageID,
		Stdin:        req.Stdin,
		CPUTimeLimit: 5, // 5 seconds
		MemoryLimit:  s.memoryLimit,
	}

	var result *codeexec.SubmissionResult
	var err error

	if s.judge0Client != nil && s.judge0Client.IsAvailable() {
		// Используем Judge0
		result, err = s.judge0Client.SubmitAndWait(ctx, submissionReq, s.timeout)
		if err != nil {
			return nil, fmt.Errorf("judge0 execution: %w", err)
		}
	} else {
		// Fallback: локальное выполнение (только для Go в dev режиме)
		// В продакшене должен использоваться только Judge0
		if req.Language != "go" {
			return nil, fmt.Errorf("judge0 not available and language %s not supported locally. Please configure Judge0", req.Language)
		}
		result, err = s.executeLocalGo(ctx, req.Code, req.Stdin)
		if err != nil {
			return nil, fmt.Errorf("local execution: %w", err)
		}
	}

	elapsed := time.Since(startTime).Milliseconds()

	// Определяем статус
	passed := result.Status.ID == 3 && result.ExitCode == 0
	errorMsg := ""
	if result.Stderr != "" {
		errorMsg = result.Stderr
	} else if result.CompileOutput != "" {
		errorMsg = result.CompileOutput
	} else if result.Status.ID != 3 {
		errorMsg = result.Status.Description
	}

	return &codedom.ExecuteResponse{
		Output:          result.Stdout,
		Error:           errorMsg,
		Passed:          passed,
		ExecutionTimeMs: elapsed,
		ExitCode:        result.ExitCode,
	}, nil
}

func (s *service) executeWithTests(ctx context.Context, req codedom.ExecuteRequest, startTime time.Time) (*codedom.ExecuteResponse, error) {
	testResults := make([]codedom.TestResult, 0, len(req.TestCases))
	allPassed := true

	for _, testCase := range req.TestCases {
		testStart := time.Now()

		// Выполняем код с входными данными теста
		testReq := codedom.ExecuteRequest{
			Code:     req.Code,
			Language: req.Language,
			Stdin:    testCase.Input,
		}

		result, err := s.executeSimple(ctx, testReq, testStart)
		if err != nil {
			testResults = append(testResults, codedom.TestResult{
				TestCase:        testCase,
				ActualOutput:    "",
				Passed:          false,
				ErrorMessage:    err.Error(),
				ExecutionTimeMs: time.Since(testStart).Milliseconds(),
			})
			allPassed = false
			continue
		}

		// Сравниваем вывод (нормализуем пробелы)
		actualNormalized := strings.TrimSpace(result.Output)
		expectedNormalized := strings.TrimSpace(testCase.ExpectedOutput)
		passed := actualNormalized == expectedNormalized

		if !passed {
			allPassed = false
		}

		testResults = append(testResults, codedom.TestResult{
			TestCase:        testCase,
			ActualOutput:    result.Output,
			Passed:          passed,
			ErrorMessage:    result.Error,
			ExecutionTimeMs: result.ExecutionTimeMs,
		})
	}

	elapsed := time.Since(startTime).Milliseconds()

	return &codedom.ExecuteResponse{
		Output:          "",
		Error:           "",
		Passed:          allPassed,
		TestResults:     testResults,
		ExecutionTimeMs: elapsed,
	}, nil
}

// executeLocalGo выполняет Go код локально (только для dev)
func (s *service) executeLocalGo(ctx context.Context, code, stdin string) (*codeexec.SubmissionResult, error) {
	// Это упрощенная версия для dev режима
	// В продакшене должен использоваться только Judge0
	result := &codeexec.SubmissionResult{
		Status: struct {
			ID          int    `json:"id"`
			Description string `json:"description"`
		}{
			ID:          3,
			Description: "Accepted",
		},
		Stdout:   "Local execution not fully implemented",
		Stderr:   "",
		ExitCode: 0,
	}

	return result, nil
}
