package ai

import "time"

// ChatContext контекст для ИИ-чата
type ChatContext struct {
	LessonID string `json:"lesson_id,omitempty"`
	CourseID string `json:"course_id,omitempty"`
	UserCode string `json:"user_code,omitempty"`
}

// ChatMessage сообщение в чате
type ChatMessage struct {
	Role    string `json:"role"` // user, assistant
	Content string `json:"content"`
}

// CodeReviewRequest запрос на проверку кода
type CodeReviewRequest struct {
	Code             string   `json:"code"`
	Language         string   `json:"language"`
	LessonID         string   `json:"lesson_id"`
	LessonObjectives []string `json:"lesson_objectives"`
}

// CodeReviewResponse ответ проверки кода
type CodeReviewResponse struct {
	Review CodeReview `json:"review"`
}

// CodeReview результат проверки кода
type CodeReview struct {
	HasErrors         bool         `json:"has_errors"`
	Errors            []CodeError  `json:"errors"`
	Suggestions       []Suggestion `json:"suggestions"`
	PositiveFeedback  []string     `json:"positive_feedback"`
	MatchesObjectives bool         `json:"matches_objectives"`
	OverallScore      int          `json:"overall_score"`
	Explanation       string       `json:"explanation"`
}

// CodeError ошибка в коде
type CodeError struct {
	Line       int    `json:"line"`
	Type       string `json:"type"` // syntax_error, runtime_error, logic_error
	Message    string `json:"message"`
	Severity   string `json:"severity"` // error, warning, info
	Suggestion string `json:"suggestion"`
}

// Suggestion предложение по улучшению
type Suggestion struct {
	Type     string `json:"type"` // style, performance, best_practice
	Message  string `json:"message"`
	Severity string `json:"severity"` // error, warning, info
}

// ExplainErrorRequest запрос на объяснение ошибки
type ExplainErrorRequest struct {
	Code     string `json:"code"`
	Error    string `json:"error"`
	Language string `json:"language"`
	LessonID string `json:"lesson_id"`
}

// ExplainErrorResponse ответ объяснения ошибки
type ExplainErrorResponse struct {
	Explanation ErrorExplanation `json:"explanation"`
}

// ErrorExplanation объяснение ошибки
type ErrorExplanation struct {
	SimpleExplanation   string   `json:"simple_explanation"`
	DetailedExplanation string   `json:"detailed_explanation"`
	FixSteps            []string `json:"fix_steps"`
	CorrectedCode       string   `json:"corrected_code"`
	LearningTip         string   `json:"learning_tip"`
}

// HintsRequest запрос на подсказки
type HintsRequest struct {
	LessonID         string `json:"lesson_id"`
	UserCode         string `json:"user_code"`
	StuckTimeSeconds int    `json:"stuck_time_seconds"`
}

// HintsResponse ответ с подсказками
type HintsResponse struct {
	Hints               []Hint `json:"hints"`
	CurrentHintLevel    string `json:"current_hint_level"`     // gentle, moderate, direct
	NextHintAvailableIn int    `json:"next_hint_available_in"` // seconds
}

// Hint подсказка
type Hint struct {
	Level string `json:"level"` // gentle, moderate, direct
	Text  string `json:"text"`
}

// ChatHistory история чата для сохранения
type ChatHistory struct {
	ID         string        `json:"id"`
	UserID     string        `json:"user_id"`
	LessonID   string        `json:"lesson_id,omitempty"`
	Messages   []ChatMessage `json:"messages"`
	TokensUsed int           `json:"tokens_used"`
	CreatedAt  time.Time     `json:"created_at"`
}
