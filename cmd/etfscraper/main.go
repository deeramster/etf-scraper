package main

import (
	"fmt"
	"log"
	"os"

	"etf-scraper/internal/config"
	"etf-scraper/internal/database"
	"etf-scraper/internal/scraper"
	"etf-scraper/internal/server"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.NewConfig()

	// Обрабатываем аргументы командной строки
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "serve":
			runServer(cfg)
			return
		case "scrape":
			runScraper(cfg)
			return
		case "help":
			printHelp()
			return
		default:
			fmt.Printf("Неизвестная команда: %s\n", os.Args[1])
			printHelp()
			os.Exit(1)
		}
	}

	// По умолчанию запускаем скрейпер
	runScraper(cfg)
}

func runServer(cfg *config.Config) {
	log.Println("Запуск API сервера...")

	// Инициализируем БД
	db, err := database.NewDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	defer db.Close()

	// Создаем репозиторий
	repo := database.NewRepository(db)

	// Запускаем сервер
	srv := server.NewServer(cfg, db, repo)
	if err := srv.Start(); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

func runScraper(cfg *config.Config) {
	log.Println("Запуск скрейпера...")

	// Инициализируем БД
	db, err := database.NewDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	defer db.Close()

	// Создаем репозиторий
	repo := database.NewRepository(db)

	// Создаем скрейпер
	s := scraper.NewScraper(cfg, repo)

	// Выполняем скрейпинг
	if err := s.Run(); err != nil {
		log.Fatalf("Ошибка выполнения скрейпинга: %v", err)
	}

	// Выводим статистику
	if err := s.PrintStats(); err != nil {
		log.Printf("Ошибка получения статистики: %v", err)
	}

	// Выводим топ-10
	if err := s.PrintTopFunds(10); err != nil {
		log.Printf("Ошибка получения топ фондов: %v", err)
	}

	log.Println("\n✓ Все данные сохранены в", cfg.DBPath)
}

func printHelp() {
	fmt.Println(`
ETF Scraper - инструмент для сбора данных о ETF фондах

Использование:
  etfscraper [команда]

Команды:
  scrape    Запустить скрейпинг данных (по умолчанию)
  serve     Запустить API сервер с веб-интерфейсом
  help      Показать эту справку

Переменные окружения:
  DB_PATH       Путь к файлу БД (по умолчанию: etf_data.db)
  SERVER_PORT   Порт сервера (по умолчанию: 8080)
  SCRAPER_URL   URL для скрейпинга (по умолчанию: https://assetallocation.ru/etf/)
  VERBOSE       Подробный вывод (true/false)
  STATIC_DIR    Путь к статическим файлам (по умолчанию: ./static)

Примеры:
  etfscraper scrape              # Запустить скрейпинг
  etfscraper serve               # Запустить веб-сервер
  DB_PATH=data.db etfscraper     # Использовать другую БД
  SERVER_PORT=3000 etfscraper serve  # Запустить на порту 3000
`)
}
