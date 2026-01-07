package lesson

import (
	"encoding/json"

	"github.com/google/uuid"
)

// LessonContent структура контента урока согласно документации
type LessonContent struct {
	Theory         string     `json:"theory"`          // теория урока
	Objectives     []string   `json:"objectives"`      // цели урока
	CodeTemplate   string     `json:"code_template"`   // шаблон кода для выполнения
	ExpectedOutput string     `json:"expected_output"` // ожидаемый вывод
	Hints          []string   `json:"hints"`           // подсказки
	TestCases      []TestCase `json:"test_cases"`      // тестовые случаи
}

// TestCase тестовый случай для проверки кода
type TestCase struct {
	Input          string `json:"input"`           // входные данные
	ExpectedOutput string `json:"expected_output"` // ожидаемый вывод
	Description    string `json:"description"`     // описание теста
}

type Lesson struct {
	ID               uuid.UUID       `json:"id"`
	CourseID         uuid.UUID       `json:"course_id"`
	ModuleID         uuid.UUID       `json:"module_id"` // ID модуля (было sectionId)
	SectionID        uuid.UUID       `json:"sectionId"` // для обратной совместимости
	Slug             string          `json:"slug"`
	Title            string          `json:"title"`
	Type             string          `json:"type"`             // theory|video|quiz|task
	Content          json.RawMessage `json:"content"`          // JSON с LessonContent
	DurationMinutes  int             `json:"duration_minutes"` // длительность урока в минутах
	Order            int             `json:"order"`
	IsFree           bool            `json:"is_free"`                      // бесплатный урок
	NextLessonID     *uuid.UUID      `json:"next_lesson_id,omitempty"`     // следующий урок
	PreviousLessonID *uuid.UUID      `json:"previous_lesson_id,omitempty"` // предыдущий урок
}
