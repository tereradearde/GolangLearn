package code

// ExecuteRequest запрос на выполнение кода
type ExecuteRequest struct {
	Code     string     `json:"code" binding:"required"`
	Language string     `json:"language" binding:"required,oneof=python javascript java go cpp"`
	Stdin    string     `json:"stdin,omitempty"`
	TestCases []TestCase `json:"test_cases,omitempty"`
}

// TestCase тестовый случай
type TestCase struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	Description    string `json:"description,omitempty"`
}

// ExecuteResponse ответ выполнения кода
type ExecuteResponse struct {
	Output          string       `json:"output"`
	Error           string       `json:"error,omitempty"`
	Passed          bool         `json:"passed"`
	TestResults     []TestResult `json:"test_results,omitempty"`
	ExecutionTimeMs int64        `json:"execution_time_ms"`
	ExitCode        int          `json:"exit_code,omitempty"`
}

// TestResult результат теста
type TestResult struct {
	TestCase      TestCase `json:"test_case"`
	ActualOutput  string   `json:"actual_output"`
	Passed        bool     `json:"passed"`
	ErrorMessage  string   `json:"error_message,omitempty"`
	ExecutionTimeMs int64   `json:"execution_time_ms,omitempty"`
}

