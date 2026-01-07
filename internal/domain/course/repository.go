package course

import (
	"context"

	"github.com/google/uuid"
)

// Repository контракт хранилища курсов.
type Repository interface {
	List(ctx context.Context) ([]Course, error)
	Create(ctx context.Context, course Course) (Course, error)
	Get(ctx context.Context, id uuid.UUID) (Course, error)
	GetBySlug(ctx context.Context, slug string) (Course, error)
	Update(ctx context.Context, id uuid.UUID, updated Course) (Course, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, f ListFilter) (ListResult, error)
}

// ListFilter параметры фильтрации/пагинации списка курсов.
type ListFilter struct {
	Query      string
	Language   string
	Difficulty string // beginner, intermediate, advanced
	MinPrice   int
	MaxPrice   int
	Tags       []string
	Page       int
	PageSize   int
	Limit      int    // альтернатива PageSize
	Sort       string // e.g. "title_asc", "popularity_desc", "rating_desc", "newest"
}

// ListResult результат поиска с пагинацией.
type ListResult struct {
	Items []Course
	Total int64
}
