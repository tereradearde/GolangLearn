package assignment

import "github.com/google/uuid"

type Assignment struct {
    ID          uuid.UUID `json:"id"`
    LessonID    uuid.UUID `json:"lessonId"`
    Title       string    `json:"title"`
    Prompt      string    `json:"prompt"`
    StarterCode string    `json:"starterCode"`
    Tests       string    `json:"tests"` // JSON тестов (упрощённо)
    Order       int       `json:"order"`
}


