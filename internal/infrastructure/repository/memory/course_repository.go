package memory

import (
	"context"
	"sort"
	"strings"
	"sync"

	dom "github.com/example/learngo/internal/domain/course"
	"github.com/google/uuid"
)

// InMemoryCourseRepository простая потокобезопасная in-memory реализация Repository.
type InMemoryCourseRepository struct {
	mu      sync.RWMutex
	storage map[uuid.UUID]dom.Course
}

func NewInMemoryCourseRepository() *InMemoryCourseRepository {
	r := &InMemoryCourseRepository{storage: make(map[uuid.UUID]dom.Course)}
	// Демо-данные
	c1 := dom.Course{ID: uuid.New(), Slug: "go-basics", Title: "Введение в Go", Description: "Основы синтаксиса, типы, пакеты", Summary: "Быстрый старт по Go", Language: "go", Difficulty: "beginner", DurationMin: 240, Tags: []string{"go", "basics"}, ImageURL: "", Objectives: []string{"Понять основы Go", "Освоить пакеты"}}
	c2 := dom.Course{ID: uuid.New(), Slug: "go-concurrency", Title: "Go: конкуррентность", Description: "Горутины, каналы, контексты", Summary: "Параллелизм и каналы", Language: "go", Difficulty: "intermediate", DurationMin: 300, Tags: []string{"go", "concurrency"}, ImageURL: "", Objectives: []string{"Горутины", "Каналы"}}
	r.storage[c1.ID] = c1
	r.storage[c2.ID] = c2
	return r
}

func (r *InMemoryCourseRepository) List(ctx context.Context) ([]dom.Course, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]dom.Course, 0, len(r.storage))
	for _, c := range r.storage {
		result = append(result, c)
	}
	return result, nil
}

// Search (простая фильтрация/пагинация в памяти)
func (r *InMemoryCourseRepository) Search(ctx context.Context, f dom.ListFilter) (dom.ListResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]dom.Course, 0, len(r.storage))
	for _, c := range r.storage {
		if f.Query != "" {
			if !(containsFold(c.Title, f.Query) || containsFold(c.Description, f.Query)) {
				continue
			}
		}
		if f.Language != "" && !strings.EqualFold(c.Language, f.Language) {
			continue
		}
		if f.Difficulty != "" && c.Difficulty != f.Difficulty {
			continue
		}
		if f.MinPrice > 0 && c.PriceCents < f.MinPrice {
			continue
		}
		if f.MaxPrice > 0 && c.PriceCents > f.MaxPrice {
			continue
		}
		if len(f.Tags) > 0 {
			match := true
			for _, t := range f.Tags {
				if !sliceContains(c.Tags, t) {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}
		items = append(items, c)
	}
	// sort
	switch f.Sort {
	case "title_desc":
		sort.Slice(items, func(i, j int) bool { return items[i].Title > items[j].Title })
	case "popularity_desc":
		sort.Slice(items, func(i, j int) bool { return items[i].Popularity > items[j].Popularity })
	case "rating_desc":
		sort.Slice(items, func(i, j int) bool { return items[i].Rating > items[j].Rating })
	default:
		sort.Slice(items, func(i, j int) bool { return items[i].Title < items[j].Title })
	}
	total := int64(len(items))
	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.PageSize
	if size <= 0 || size > 100 {
		size = 12
	}
	start := (page - 1) * size
	if start > len(items) {
		return dom.ListResult{Items: []dom.Course{}, Total: total}, nil
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return dom.ListResult{Items: items[start:end], Total: total}, nil
}

func containsFold(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
func sliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func (r *InMemoryCourseRepository) Create(ctx context.Context, course dom.Course) (dom.Course, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if course.ID == uuid.Nil {
		course.ID = uuid.New()
	}
	if course.Slug == "" {
		course.Slug = simpleSlug(course.Title)
	}
	r.storage[course.ID] = course
	return course, nil
}

func (r *InMemoryCourseRepository) Get(ctx context.Context, id uuid.UUID) (dom.Course, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if c, ok := r.storage[id]; ok {
		return c, nil
	}
	return dom.Course{}, nil
}

func (r *InMemoryCourseRepository) Update(ctx context.Context, id uuid.UUID, updated dom.Course) (dom.Course, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.storage[id]
	if !ok {
		return dom.Course{}, nil
	}
	if updated.Slug != "" {
		c.Slug = updated.Slug
	}
	c.Title = updated.Title
	c.Description = updated.Description
	c.Summary = updated.Summary
	if updated.Language != "" {
		c.Language = updated.Language
	}
	c.Difficulty = updated.Difficulty
	c.DurationMin = updated.DurationMin
	c.DurationHours = updated.DurationHours
	c.Tags = updated.Tags
	c.ImageURL = updated.ImageURL
	c.Objectives = updated.Objectives
	r.storage[id] = c
	return c, nil
}

func (r *InMemoryCourseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.storage, id)
	return nil
}

func (r *InMemoryCourseRepository) GetBySlug(ctx context.Context, slug string) (dom.Course, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.storage {
		if c.Slug == slug {
			return c, nil
		}
	}
	return dom.Course{}, nil
}

func simpleSlug(s string) string {
	b := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b = append(b, r)
		case r >= 'A' && r <= 'Z':
			b = append(b, r+32)
		case r == ' ' || r == '-' || r == '_':
			b = append(b, '-')
		}
	}
	if len(b) == 0 {
		return "course"
	}
	return string(b)
}
