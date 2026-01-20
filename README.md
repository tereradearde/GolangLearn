## Интерактивная платформа обучения Go

Монорепозиторий: backend (Go, Clean Architecture) + frontend (Next.js, TypeScript, Monaco Editor).

### Быстрый старт (локально)

1. Backend
```
go version   # требуется Go 1.22+
cd cmd/app
go run .
```
Сервер слушает порт по умолчанию `:8080` (переменная `HTTP_PORT`).

2. Frontend
```
cd web
npm i
npm run dev
```
Приложение доступно на `http://localhost:3000`.

### Docker/Compose
```
docker compose up --build
```
Web: `http://localhost:3000`, API: `http://localhost:8080`.

### Структура

```
cmd/
  app/            # входная точка HTTP API
internal/
  domain/         # доменные модели и контракты
  usecase/        # бизнес-логика (application layer)
  delivery/       # адаптеры (HTTP)
  infrastructure/ # хранилища, внешние интеграции
pkg/
  utils/          # общие утилиты (логгер, конфиг)
web/              # Next.js фронтенд
```

