# 📊 ETF Scraper - Модульная версия

Веб-приложение для мониторинга биржевых фондов (ETF) на Московской бирже с автоматическим сбором данных.

## 🏗️ Архитектура проекта

Проект полностью переработан в модульную структуру для лучшей поддерживаемости и масштабируемости.

### Backend

```
internal/
├── config/         # Конфигурация приложения
├── models/         # Модели данных
├── database/       # Работа с БД (инициализация, репозиторий)
├── scraper/        # Логика скрейпинга и парсинга
└── server/         # HTTP сервер, handlers, middleware
```

### Frontend

```
static/
├── index.html
└── js/
    ├── app.js              # Главное приложение
    ├── components/         # React компоненты
    │   ├── StatsCards.js   # Карточки статистики
    │   ├── FilterPanel.js  # Панель фильтров
    │   ├── ETFTable.js     # Таблица данных
    │   └── Icons.js        # SVG иконки
    ├── services/
    │   └── api.js          # API клиент
    └── utils/
        └── helpers.js      # Вспомогательные функции
```

## 🚀 Быстрый старт

### 1. Клонирование и установка зависимостей

```bash
# Клонировать репозиторий
git clone 
cd etf-scraper

# Установить зависимости Go
make deps
# или
go mod download
```

### 2. Первый запуск

```bash
# Выполнить скрейпинг данных
make run-scraper
# или
go run cmd/etfscraper/main.go scrape

# Запустить веб-сервер
make run-server
# или
go run cmd/etfscraper/main.go serve
```

### 3. Открыть в браузере

Перейдите по адресу: **http://localhost:8080**

## 📦 Makefile команды

```bash
make help          # Показать справку
make deps          # Установить зависимости
make build         # Собрать бинарный файл
make run-scraper   # Запустить скрейпинг
make run-server    # Запустить веб-сервер
make test          # Запустить тесты
make lint          # Проверить код линтером
make format        # Форматировать код
make clean         # Удалить собранные файлы
```

## ⚙️ Конфигурация

Настройка через переменные окружения:

```bash
# Путь к файлу БД
export DB_PATH=etf_data.db

# Порт сервера
export SERVER_PORT=8080

# URL для скрейпинга
export SCRAPER_URL=https://assetallocation.ru/etf/

# Подробный вывод
export VERBOSE=true

# Путь к статическим файлам
export STATIC_DIR=./static
```

## 🔧 Использование

### Скрейпинг данных

```bash
# Базовый запуск
go run cmd/etfscraper/main.go scrape

# С другой БД
DB_PATH=test.db go run cmd/etfscraper/main.go scrape

# С подробным выводом
VERBOSE=true go run cmd/etfscraper/main.go scrape
```

### Запуск сервера

```bash
# Базовый запуск на порту 8080
go run cmd/etfscraper/main.go serve

# На другом порту
SERVER_PORT=3000 go run cmd/etfscraper/main.go serve
```

## 🌐 API Endpoints

### GET /api/etfs
Получить все ETF с фильтрацией и сортировкой

**Параметры:**
- `sortBy` - поле сортировки (nav_million_rub, ter_percent, ticker, price_change_2024)
- `order` - порядок сортировки (ASC, DESC)
- `assetClass` - фильтр по классу активов

**Пример:**
```bash
curl "http://localhost:8080/api/etfs?sortBy=nav_million_rub&order=DESC"
```

### GET /api/etfs/{ticker}
Получить данные конкретного ETF по тикеру

**Пример:**
```bash
curl "http://localhost:8080/api/etfs/TMOS"
```

### GET /api/stats
Получить общую статистику

**Ответ:**
```json
{
  "totalRecords": 1500,
  "uniqueTickers": 150,
  "scrapeSessions": 10,
  "totalNAV": 250000.5,
  "avgTER": 0.85,
  "lastUpdate": "2024-01-15"
}
```

### GET /api/asset-classes
Получить список классов активов

### GET /api/top-by-nav?limit=10
Получить топ ETF по размеру СЧА

### GET /api/search?q=term
Поиск ETF по тикеру, названию или УК

### POST /api/scrape
Запустить скрейпинг в фоновом режиме

## 📊 Структура базы данных

```sql
CREATE TABLE etf_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date_scraped TEXT NOT NULL,
    ticker TEXT NOT NULL,
    trade_status TEXT,
    management_company TEXT,
    asset_class TEXT,
    ter_percent REAL,
    ter_direction TEXT,
    fund_name TEXT,
    management_style TEXT,
    target_index TEXT,
    currency TEXT,
    start_date TEXT,
    info_icon TEXT,
    price_change_6m REAL,
    price_change_2024 REAL,
    price_change_2023 REAL,
    price_change_2022 REAL,
    price_change_2021 REAL,
    price_change_2020 REAL,
    nav_million_rub REAL,
    last_update_date TEXT
);
```
