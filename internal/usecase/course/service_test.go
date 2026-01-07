package course

import (
	"context"
	"testing"

	mem "github.com/example/learngo/internal/infrastructure/repository/memory"
	"github.com/example/learngo/pkg/utils"
)

func TestCreateAndGetCourse(t *testing.T) {
	repo := mem.NewInMemoryCourseRepository()
	logger := utils.NewLogger("test")
	svc := NewService(repo, logger)

	created, err := svc.CreateCourse(context.Background(), "Test", "Desc")
	if err != nil {
		t.Fatalf("CreateCourse error: %v", err)
	}
	if created.Title != "Test" {
		t.Fatalf("unexpected title: %s", created.Title)
	}

	got, err := svc.GetCourse(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetCourse error: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("unexpected id: %v", got.ID)
	}
}

func TestListCourses(t *testing.T) {
	repo := mem.NewInMemoryCourseRepository()
	logger := utils.NewLogger("test")
	svc := NewService(repo, logger)

	list, err := svc.ListCourses(context.Background())
	if err != nil {
		t.Fatalf("ListCourses error: %v", err)
	}
	if len(list) == 0 {
		t.Fatalf("expected seed courses in memory repo")
	}

	// sanity create one more and ensure it appears
	_, _ = svc.CreateCourse(context.Background(), "X", "Y")
	list2, err := svc.ListCourses(context.Background())
	if err != nil {
		t.Fatalf("ListCourses2 error: %v", err)
	}
	if len(list2) < len(list)+1 {
		t.Fatalf("expected >= %d, got %d", len(list)+1, len(list2))
	}
}
