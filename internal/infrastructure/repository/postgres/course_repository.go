package postgres

import (
	"context"
	"encoding/json"

	dom "github.com/example/learngo/internal/domain/course"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CourseModel struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	Slug             string    `gorm:"size:255;uniqueIndex;not null;default:''"`
	Title            string    `gorm:"size:255;not null"`
	Description      string    `gorm:"type:text;not null"`
	Summary          string    `gorm:"type:text;not null;default:''"`
	Language         string    `gorm:"size:32;not null;default:'go'"`
	Difficulty       string    `gorm:"size:32;not null;default:'beginner'"` // beginner, intermediate, advanced
	DurationHours    int       `gorm:"not null;default:0"`                  // длительность в часах
	DurationMin      int       `gorm:"not null;default:0"`                  // длительность в минутах (для обратной совместимости)
	TagsJSON         string    `gorm:"type:text;not null;default:'[]'"`
	ThumbnailURL     string    `gorm:"size:512;not null;default:''"` // обложка курса
	ImageURL         string    `gorm:"size:512;not null;default:''"` // для обратной совместимости
	ObjectivesJSON   string    `gorm:"type:text;not null;default:'[]'"`
	RequirementsJSON string    `gorm:"type:text;not null;default:'[]'"`
	IsFree           bool      `gorm:"not null;default:true"` // бесплатный курс
	Price            *float64  `gorm:"type:decimal(10,2)"`    // цена в рублях (null для бесплатных)
	PriceCents       int       `gorm:"not null;default:0"`    // цена в копейках (для обратной совместимости)
	Rating           float64   `gorm:"not null;default:0"`
	Popularity       int       `gorm:"not null;default:0"`
}

func (CourseModel) TableName() string { return "courses" }

func toModel(c dom.Course) CourseModel {
	tags, _ := json.Marshal(c.Tags)
	objectives, _ := json.Marshal(c.Objectives)
	reqs, _ := json.Marshal(c.Requirements)

	// Определяем thumbnail_url (приоритет новому полю)
	thumbnailURL := c.ThumbnailURL
	if thumbnailURL == "" {
		thumbnailURL = c.ImageURL
	}

	// Определяем price из PriceCents если Price не задан
	price := c.Price
	if price == nil && c.PriceCents > 0 {
		p := float64(c.PriceCents) / 100.0
		price = &p
	}

	// Определяем is_free
	isFree := c.IsFree
	if !isFree && c.PriceCents == 0 && (price == nil || *price == 0) {
		isFree = true
	}

	return CourseModel{
		ID:               c.ID,
		Slug:             c.Slug,
		Title:            c.Title,
		Description:      c.Description,
		Summary:          c.Summary,
		Language:         c.Language,
		Difficulty:       c.Difficulty,
		DurationHours:    c.DurationHours,
		DurationMin:      c.DurationMin,
		TagsJSON:         string(tags),
		ThumbnailURL:     thumbnailURL,
		ImageURL:         c.ImageURL,
		ObjectivesJSON:   string(objectives),
		RequirementsJSON: string(reqs),
		IsFree:           isFree,
		Price:            price,
		PriceCents:       c.PriceCents,
		Rating:           c.Rating,
		Popularity:       c.Popularity,
	}
}

func toDomain(m CourseModel) dom.Course {
	var tags []string
	var objectives []string
	var reqs []string
	_ = json.Unmarshal([]byte(m.TagsJSON), &tags)
	_ = json.Unmarshal([]byte(m.ObjectivesJSON), &objectives)
	_ = json.Unmarshal([]byte(m.RequirementsJSON), &reqs)

	// Определяем imageUrl для обратной совместимости
	imageURL := m.ImageURL
	if imageURL == "" {
		imageURL = m.ThumbnailURL
	}

	return dom.Course{
		ID:            m.ID,
		Slug:          m.Slug,
		Title:         m.Title,
		Description:   m.Description,
		Summary:       m.Summary,
		Language:      m.Language,
		Difficulty:    m.Difficulty,
		DurationHours: m.DurationHours,
		DurationMin:   m.DurationMin,
		Tags:          tags,
		ThumbnailURL:  m.ThumbnailURL,
		ImageURL:      imageURL,
		Objectives:    objectives,
		Requirements:  reqs,
		IsFree:        m.IsFree,
		Price:         m.Price,
		PriceCents:    m.PriceCents,
		Rating:        m.Rating,
		Popularity:    m.Popularity,
	}
}

type CourseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

func (r *CourseRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&CourseModel{})
}

func (r *CourseRepository) List(ctx context.Context) ([]dom.Course, error) {
	var rows []CourseModel
	if err := r.db.WithContext(ctx).Order("title asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dom.Course, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomain(row))
	}
	return out, nil
}

// Search с фильтрами и пагинацией
func (r *CourseRepository) Search(ctx context.Context, f dom.ListFilter) (dom.ListResult, error) {
	q := r.db.WithContext(ctx).Model(&CourseModel{})
	if f.Query != "" {
		like := "%" + f.Query + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}
	if f.Language != "" {
		q = q.Where("language = ?", f.Language)
	}
	if f.Difficulty != "" {
		q = q.Where("difficulty = ?", f.Difficulty)
	}
	if f.MinPrice > 0 {
		q = q.Where("price_cents >= ?", f.MinPrice)
	}
	if f.MaxPrice > 0 {
		q = q.Where("price_cents <= ?", f.MaxPrice)
	}
	if len(f.Tags) > 0 {
		// простая фильтрация по JSON-строке (contains любой из тегов)
		for _, t := range f.Tags {
			q = q.Where("tags_json ILIKE ?", "%\""+t+"\"%")
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return dom.ListResult{}, err
	}
	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.PageSize
	if size <= 0 || size > 100 {
		size = 12
	}
	switch f.Sort {
	case "title_desc":
		q = q.Order("title desc")
	case "popularity_desc":
		q = q.Order("popularity desc")
	case "rating_desc":
		q = q.Order("rating desc")
	case "newest":
		q = q.Order("created_at desc")
	default:
		q = q.Order("title asc")
	}
	var rows []CourseModel
	if err := q.Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return dom.ListResult{}, err
	}
	items := make([]dom.Course, 0, len(rows))
	for _, r0 := range rows {
		items = append(items, toDomain(r0))
	}
	return dom.ListResult{Items: items, Total: total}, nil
}

func (r *CourseRepository) Create(ctx context.Context, course dom.Course) (dom.Course, error) {
	row := toModel(course)
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	if row.Slug == "" {
		row.Slug = generateSlug(row.Title)
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return dom.Course{}, err
	}
	return toDomain(row), nil
}

func (r *CourseRepository) Get(ctx context.Context, id uuid.UUID) (dom.Course, error) {
	var row CourseModel
	if err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return dom.Course{}, nil
		}
		return dom.Course{}, err
	}
	return toDomain(row), nil
}

func (r *CourseRepository) GetBySlug(ctx context.Context, slug string) (dom.Course, error) {
	var row CourseModel
	if err := r.db.WithContext(ctx).First(&row, "slug = ?", slug).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return dom.Course{}, nil
		}
		return dom.Course{}, err
	}
	return toDomain(row), nil
}

func (r *CourseRepository) Update(ctx context.Context, id uuid.UUID, updated dom.Course) (dom.Course, error) {
	var row CourseModel
	if err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return dom.Course{}, nil
		}
		return dom.Course{}, err
	}
	row.Title = updated.Title
	if updated.Slug != "" {
		row.Slug = updated.Slug
	}
	row.Description = updated.Description
	row.Summary = updated.Summary
	if updated.Language != "" {
		row.Language = updated.Language
	}
	row.Difficulty = updated.Difficulty
	row.DurationHours = updated.DurationHours
	row.DurationMin = updated.DurationMin
	// serialize tags/objectives/requirements
	tg, _ := json.Marshal(updated.Tags)
	obj, _ := json.Marshal(updated.Objectives)
	reqs, _ := json.Marshal(updated.Requirements)
	row.TagsJSON = string(tg)
	row.ObjectivesJSON = string(obj)
	row.RequirementsJSON = string(reqs)
	row.ThumbnailURL = updated.ThumbnailURL
	if updated.ThumbnailURL == "" {
		row.ThumbnailURL = updated.ImageURL
	}
	row.ImageURL = updated.ImageURL
	row.IsFree = updated.IsFree
	row.Price = updated.Price
	row.PriceCents = updated.PriceCents
	row.Rating = updated.Rating
	row.Popularity = updated.Popularity
	if err := r.db.WithContext(ctx).Save(&row).Error; err != nil {
		return dom.Course{}, err
	}
	return toDomain(row), nil
}

func generateSlug(s string) string {
	// очень простой slugifier; для прод заменить
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

func (r *CourseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&CourseModel{}, "id = ?", id).Error
}
