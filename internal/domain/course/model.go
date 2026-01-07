package course

import "github.com/google/uuid"

// Course доменная модель курса.
type Course struct {
	ID            uuid.UUID `json:"id"`
	Slug          string    `json:"slug"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Summary       string    `json:"summary"`
	Language      string    `json:"language"`       // python, javascript, java, go, cpp
	Difficulty    string    `json:"difficulty"`     // beginner, intermediate, advanced
	DurationHours int       `json:"duration_hours"` // длительность курса в часах
	DurationMin   int       `json:"durationMin"`    // длительность курса в минутах (для обратной совместимости)
	LessonsCount  int       `json:"lessons_count"`  // количество уроков (вычисляемое)
	StudentsCount int       `json:"students_count"` // количество студентов (вычисляемое)
	Tags          []string  `json:"tags"`           // теги каталога
	ThumbnailURL  string    `json:"thumbnail_url"`  // обложка курса
	ImageURL      string    `json:"imageUrl"`       // обложка курса (для обратной совместимости)
	Objectives    []string  `json:"objectives"`     // цели обучения
	Requirements  []string  `json:"requirements"`   // требования по знаниям/ПО
	IsFree        bool      `json:"is_free"`        // бесплатный курс
	Price         *float64  `json:"price"`          // цена в рублях (null для бесплатных)
	PriceCents    int       `json:"priceCents"`     // цена в копейках (для обратной совместимости)
	Rating        float64   `json:"rating"`         // рейтинг курса
	Popularity    int       `json:"popularity"`     // популярность
}
