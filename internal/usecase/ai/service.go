package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	aidom "github.com/example/learngo/internal/domain/ai"
	"github.com/example/learngo/pkg/ai"
	"github.com/example/learngo/pkg/utils"
)

// Service интерфейс ИИ-сервиса
type Service interface {
	Chat(ctx context.Context, userID string, messages []aidom.ChatMessage, context aidom.ChatContext) (string, int, error)
	ChatStream(ctx context.Context, userID string, messages []aidom.ChatMessage, context aidom.ChatContext, onChunk func(content string) error) (int, error)
	CodeReview(ctx context.Context, req aidom.CodeReviewRequest) (*aidom.CodeReviewResponse, error)
	ExplainError(ctx context.Context, req aidom.ExplainErrorRequest) (*aidom.ExplainErrorResponse, error)
	GetHints(ctx context.Context, req aidom.HintsRequest) (*aidom.HintsResponse, error)
}

type service struct {
	client *ai.Client
	logger *utils.Logger
}

func NewService(client *ai.Client, logger *utils.Logger) Service {
	return &service{
		client: client,
		logger: logger,
	}
}

func (s *service) Chat(ctx context.Context, userID string, messages []aidom.ChatMessage, chatContext aidom.ChatContext) (string, int, error) {
	// Формируем системный промпт
	systemPrompt := s.buildSystemPrompt(chatContext)

	// Преобразуем сообщения в формат OpenAI
	openAIMessages := make([]ai.Message, 0, len(messages)+1)
	openAIMessages = append(openAIMessages, ai.Message{
		Role:    "system",
		Content: systemPrompt,
	})

	for _, msg := range messages {
		openAIMessages = append(openAIMessages, ai.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Выполняем запрос
	resp, err := s.client.Chat(ctx, openAIMessages)
	if err != nil {
		return "", 0, fmt.Errorf("openai chat: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", 0, fmt.Errorf("no choices in response")
	}

	content := resp.Choices[0].Message.Content
	tokensUsed := resp.Usage.TotalTokens

	return content, tokensUsed, nil
}

func (s *service) ChatStream(ctx context.Context, userID string, messages []aidom.ChatMessage, chatContext aidom.ChatContext, onChunk func(content string) error) (int, error) {
	// Формируем системный промпт
	systemPrompt := s.buildSystemPrompt(chatContext)

	// Преобразуем сообщения в формат OpenAI
	openAIMessages := make([]ai.Message, 0, len(messages)+1)
	openAIMessages = append(openAIMessages, ai.Message{
		Role:    "system",
		Content: systemPrompt,
	})

	for _, msg := range messages {
		openAIMessages = append(openAIMessages, ai.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Выполняем streaming запрос
	// Подсчитываем токены приблизительно (1 токен ≈ 0.75 слова)
	tokensUsed := 0
	err := s.client.ChatStream(ctx, openAIMessages, func(content string) error {
		// Приблизительный подсчет: слова * 1.33 (обратное от 0.75)
		tokensUsed += int(float64(len(strings.Fields(content))) * 1.33)
		return onChunk(content)
	})

	if err != nil {
		return tokensUsed, fmt.Errorf("openai chat stream: %w", err)
	}

	// Добавляем токены промпта (приблизительно)
	promptTokens := len(strings.Fields(systemPrompt))
	for _, msg := range messages {
		promptTokens += len(strings.Fields(msg.Content))
	}
	tokensUsed += int(float64(promptTokens) * 1.33)

	return tokensUsed, nil
}

func (s *service) CodeReview(ctx context.Context, req aidom.CodeReviewRequest) (*aidom.CodeReviewResponse, error) {
	prompt := s.buildCodeReviewPrompt(req)

	messages := []ai.Message{
		{
			Role:    "system",
			Content: "Ты - опытный преподаватель программирования на русском языке. Твоя задача - проверять код студентов, находить ошибки, давать конструктивную обратную связь и оценивать соответствие целям урока.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	resp, err := s.client.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("openai code review: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Парсим JSON ответ
	content := resp.Choices[0].Message.Content
	// Извлекаем JSON из markdown code blocks если есть
	jsonContent := extractJSONFromMarkdown(content)

	var reviewResp aidom.CodeReviewResponse
	if err := json.Unmarshal([]byte(jsonContent), &reviewResp); err != nil {
		// Если не JSON, создаем простой ответ
		reviewResp = aidom.CodeReviewResponse{
			Review: aidom.CodeReview{
				HasErrors:         false,
				Errors:            []aidom.CodeError{},
				Suggestions:       []aidom.Suggestion{},
				PositiveFeedback:  []string{content},
				MatchesObjectives: true,
				OverallScore:      80,
				Explanation:       content,
			},
		}
	}

	return &reviewResp, nil
}

func (s *service) ExplainError(ctx context.Context, req aidom.ExplainErrorRequest) (*aidom.ExplainErrorResponse, error) {
	prompt := s.buildExplainErrorPrompt(req)

	messages := []ai.Message{
		{
			Role:    "system",
			Content: "Ты - опытный преподаватель программирования на русском языке. Твоя задача - объяснять ошибки студентам простым и понятным языком, показывать как их исправить и давать полезные советы для обучения.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	resp, err := s.client.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("openai explain error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Парсим JSON ответ
	content := resp.Choices[0].Message.Content
	jsonContent := extractJSONFromMarkdown(content)

	var explainResp aidom.ExplainErrorResponse
	if err := json.Unmarshal([]byte(jsonContent), &explainResp); err != nil {
		// Если не JSON, создаем простой ответ
		explainResp = aidom.ExplainErrorResponse{
			Explanation: aidom.ErrorExplanation{
				SimpleExplanation:   content,
				DetailedExplanation: content,
				FixSteps:            []string{"Изучите сообщение об ошибке", "Проверьте синтаксис", "Исправьте ошибку"},
				CorrectedCode:       req.Code,
				LearningTip:         "Внимательно читайте сообщения об ошибках - они подсказывают, что не так",
			},
		}
	}

	return &explainResp, nil
}

func (s *service) GetHints(ctx context.Context, req aidom.HintsRequest) (*aidom.HintsResponse, error) {
	prompt := s.buildHintsPrompt(req)

	messages := []ai.Message{
		{
			Role:    "system",
			Content: "Ты - опытный преподаватель программирования на русском языке. Твоя задача - давать подсказки студентам, которые застряли на задаче. Подсказки должны быть постепенными: сначала мягкие (gentle), потом умеренные (moderate), и наконец прямые (direct).",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	resp, err := s.client.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("openai hints: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Парсим JSON ответ
	content := resp.Choices[0].Message.Content
	jsonContent := extractJSONFromMarkdown(content)

	var hintsResp aidom.HintsResponse
	if err := json.Unmarshal([]byte(jsonContent), &hintsResp); err != nil {
		// Если не JSON, создаем простой ответ
		hintsResp = aidom.HintsResponse{
			Hints: []aidom.Hint{
				{Level: "gentle", Text: content},
			},
			CurrentHintLevel:    "gentle",
			NextHintAvailableIn: 60,
		}
	}

	return &hintsResp, nil
}

func (s *service) buildSystemPrompt(context aidom.ChatContext) string {
	var parts []string
	parts = append(parts, "Ты - опытный преподаватель программирования на русском языке.")

	if context.LessonID != "" {
		parts = append(parts, fmt.Sprintf("Текущий урок: %s", context.LessonID))
	}
	if context.UserCode != "" {
		parts = append(parts, fmt.Sprintf("Код студента:\n```\n%s\n```", context.UserCode))
	}

	parts = append(parts, "Отвечай кратко, понятно и по-дружески.")

	return strings.Join(parts, "\n")
}

func (s *service) buildCodeReviewPrompt(req aidom.CodeReviewRequest) string {
	var parts []string
	parts = append(parts, "Проверь следующий код:")
	parts = append(parts, fmt.Sprintf("```%s\n%s\n```", req.Language, req.Code))

	if len(req.LessonObjectives) > 0 {
		parts = append(parts, "\nЦели урока:")
		for _, obj := range req.LessonObjectives {
			parts = append(parts, fmt.Sprintf("- %s", obj))
		}
	}

	parts = append(parts, "\nВерни JSON с результатом проверки в формате:")
	parts = append(parts, `{
  "has_errors": true/false,
  "errors": [{"line": 1, "type": "syntax_error", "message": "...", "severity": "error", "suggestion": "..."}],
  "suggestions": [{"type": "style", "message": "...", "severity": "info"}],
  "positive_feedback": ["..."],
  "matches_objectives": true/false,
  "overall_score": 75,
  "explanation": "..."
}`)

	return strings.Join(parts, "\n")
}

func (s *service) buildExplainErrorPrompt(req aidom.ExplainErrorRequest) string {
	var parts []string
	parts = append(parts, "Объясни следующую ошибку:")
	parts = append(parts, fmt.Sprintf("Код:\n```%s\n%s\n```", req.Language, req.Code))
	parts = append(parts, fmt.Sprintf("Ошибка: %s", req.Error))

	parts = append(parts, "\nВерни JSON с объяснением в формате:")
	parts = append(parts, `{
  "simple_explanation": "...",
  "detailed_explanation": "...",
  "fix_steps": ["шаг 1", "шаг 2"],
  "corrected_code": "...",
  "learning_tip": "..."
}`)

	return strings.Join(parts, "\n")
}

func (s *service) buildHintsPrompt(req aidom.HintsRequest) string {
	var parts []string
	parts = append(parts, "Студент застрял на задаче.")
	parts = append(parts, fmt.Sprintf("Текущий код:\n```\n%s\n```", req.UserCode))
	parts = append(parts, fmt.Sprintf("Время застревания: %d секунд", req.StuckTimeSeconds))

	parts = append(parts, "\nДай подсказки трех уровней (gentle, moderate, direct) в формате JSON:")
	parts = append(parts, `{
  "hints": [
    {"level": "gentle", "text": "..."},
    {"level": "moderate", "text": "..."},
    {"level": "direct", "text": "..."}
  ],
  "current_hint_level": "gentle",
  "next_hint_available_in": 60
}`)

	return strings.Join(parts, "\n")
}

// extractJSONFromMarkdown извлекает JSON из markdown code blocks
func extractJSONFromMarkdown(content string) string {
	// Ищем JSON в markdown code blocks (```json ... ```)
	start := strings.Index(content, "```json")
	if start != -1 {
		start += 7 // длина "```json"
		end := strings.Index(content[start:], "```")
		if end != -1 {
			return strings.TrimSpace(content[start : start+end])
		}
	}

	// Ищем JSON в обычных code blocks (``` ... ```)
	start = strings.Index(content, "```")
	if start != -1 {
		start += 3 // длина "```"
		end := strings.Index(content[start:], "```")
		if end != -1 {
			jsonStr := strings.TrimSpace(content[start : start+end])
			// Проверяем, что это похоже на JSON
			if strings.HasPrefix(jsonStr, "{") || strings.HasPrefix(jsonStr, "[") {
				return jsonStr
			}
		}
	}

	// Ищем JSON объект напрямую
	start = strings.Index(content, "{")
	if start != -1 {
		end := strings.LastIndex(content, "}")
		if end > start {
			return content[start : end+1]
		}
	}

	return content
}
