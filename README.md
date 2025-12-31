# ETF Dashboard

## Установка

```bash
go mod download
```

## Структура проекта

```
etf-scraper/
├── main.go
├── server.go
├── go.mod
├── go.sum
├── etf_data.db (создается автоматически)
└── static/
    └── index.html
```

## Запуск

### Скрейпинг данных
```bash
go run . scrape
```

### Запуск веб-сервера
```bash
go run . serve
```

### Сборка
```bash
go build -o etf-scraper
./etf-scraper serve
```

## API Endpoints

- `GET /api/etfs` - Все ETF
- `GET /api/etfs/{ticker}` - ETF по тикеру
- `GET /api/stats` - Статистика
- `GET /api/asset-classes` - Классы активов
- `GET /api/top-by-nav?limit=10` - Топ по СЧА
- `GET /api/search?q=term` - Поиск
- `POST /api/scrape` - Запуск скрейпинга

## Использование

1. Первый запуск - скрейпинг:
```bash
./etf-scraper scrape
```

2. Запуск сервера:
```bash
./etf-scraper serve
```

3. Открыть в браузере:
```
http://localhost:8080
```