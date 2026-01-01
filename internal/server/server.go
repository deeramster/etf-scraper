package server

import (
	"log"
	"net/http"
	"time"

	"etf-scraper/internal/config"
	"etf-scraper/internal/database"

	"github.com/gorilla/mux"
)

// Server представляет HTTP сервер
type Server struct {
	config   *config.Config
	db       *database.Database
	repo     *database.Repository
	router   *mux.Router
	handlers *Handlers
}

// NewServer создает новый HTTP сервер
func NewServer(cfg *config.Config, db *database.Database, repo *database.Repository) *Server {
	router := mux.NewRouter()

	handlers := NewHandlers(cfg, db, repo)

	server := &Server{
		config:   cfg,
		db:       db,
		repo:     repo,
		router:   router,
		handlers: handlers,
	}

	server.setupRoutes()
	return server
}

// setupRoutes настраивает маршруты API
func (s *Server) setupRoutes() {
	// Применяем middleware
	s.router.Use(corsMiddleware)
	s.router.Use(loggingMiddleware)

	// API endpoints
	api := s.router.PathPrefix("/api").Subrouter()

	api.HandleFunc("/etfs", s.handlers.HandleGetAllETFs).Methods("GET", "OPTIONS")
	api.HandleFunc("/etfs/{ticker}", s.handlers.HandleGetETFByTicker).Methods("GET", "OPTIONS")
	api.HandleFunc("/stats", s.handlers.HandleGetStats).Methods("GET", "OPTIONS")
	api.HandleFunc("/asset-classes", s.handlers.HandleGetAssetClasses).Methods("GET", "OPTIONS")
	api.HandleFunc("/top-by-nav", s.handlers.HandleGetTopByNAV).Methods("GET", "OPTIONS")
	api.HandleFunc("/search", s.handlers.HandleSearch).Methods("GET", "OPTIONS")
	api.HandleFunc("/scrape", s.handlers.HandleScrape).Methods("POST", "OPTIONS")

	// Статические файлы
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir(s.config.StaticDir)))
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	s.printServerInfo()

	srv := &http.Server{
		Handler:      s.router,
		Addr:         ":" + s.config.ServerPort,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("🌍 Сервер доступен по адресу: http://localhost:%s", s.config.ServerPort)
	log.Println()

	return srv.ListenAndServe()
}

// printServerInfo выводит информацию о сервере
func (s *Server) printServerInfo() {
	log.Printf("🚀 API сервер запущен на http://localhost:%s", s.config.ServerPort)
	log.Printf("📊 API endpoints:")
	log.Printf("   GET  /api/etfs                - Все ETF")
	log.Printf("   GET  /api/etfs/{ticker}       - ETF по тикеру")
	log.Printf("   GET  /api/stats               - Статистика")
	log.Printf("   GET  /api/asset-classes       - Классы активов")
	log.Printf("   GET  /api/top-by-nav?limit=10 - Топ по СЧА")
	log.Printf("   GET  /api/search?q=term       - Поиск")
	log.Printf("   POST /api/scrape              - Запуск скрейпинга")
	log.Println()
}
