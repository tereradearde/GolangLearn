package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type jobPayload struct {
	ID       string `json:"id"`
	Language string `json:"language"`
	Code     string `json:"code"`
}

func main() {
	rabbitURL := getenv("RABBIT_URL", "amqp://guest:guest@rabbitmq:5672/")
	queue := getenv("RUNNER_QUEUE", "runner.jobs")
	redisAddr := getenv("REDIS_ADDR", "redis:6379")
	redisPass := os.Getenv("REDIS_PASSWORD")

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("rabbit dial: %v", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("rabbit channel: %v", err)
	}
	defer ch.Close()
	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		log.Fatalf("queue declare: %v", err)
	}
	// QoS чтобы не перегружать воркер
	_ = ch.Qos(1, 0, false)

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr, Password: redisPass})

	msgs, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("consume: %v", err)
	}
	log.Println("runner consumer started")
	for m := range msgs {
		var p jobPayload
		if err := json.Unmarshal(m.Body, &p); err != nil {
			log.Printf("bad payload: %v", err)
			_ = m.Nack(false, false)
			continue
		}
		// Помечаем в Redis статус processing
		_ = rdb.HSet(context.Background(), p.ID, "status", "processing").Err()
		started := time.Now()
		res := runGo(p.Code)
		dur := time.Since(started).Milliseconds()
		// Сохраняем результат
		_ = rdb.HSet(context.Background(), p.ID,
			"status", "done",
			"stdout", res.Stdout,
			"stderr", res.Stderr,
			"exitCode", res.ExitCode,
			"elapsedMs", dur,
		).Err()
		_ = m.Ack(false)
	}
}

type runResult struct {
	Stdout, Stderr string
	ExitCode       int
}

func runGo(code string) runResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	workDir, _ := os.MkdirTemp("", "runner-go-*")
	defer os.RemoveAll(workDir)
	mainPath := filepath.Join(workDir, "main.go")
	_ = os.WriteFile(mainPath, []byte(code), 0o600)
	cmd := exec.CommandContext(ctx, "go", "run", "main.go")
	cmd.Dir = workDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return runResult{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: exitCode}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
