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
	config      *config.Config
	db          *database.Database
	repo        *database.Repository
	router      *mux.Router
	adminRouter *mux.Router
	handlers    *Handlers
}

// NewServer создает новый HTTP сервер
func NewServer(cfg *config.Config, db *database.Database, repo *database.Repository) *Server {
	router := mux.NewRouter()
	adminRouter := mux.NewRouter()

	handlers := NewHandlers(cfg, db, repo)

	server := &Server{
		config:      cfg,
		db:          db,
		repo:        repo,
		router:      router,
		adminRouter: adminRouter,
		handlers:    handlers,
	}

	server.setupRoutes()
	server.setupAdminRoutes()
	return server
}

// setupRoutes настраивает публичные маршруты API
func (s *Server) setupRoutes() {
	s.router.Use(corsMiddleware)
	s.router.Use(loggingMiddleware)

	// API endpoints (публичные)
	api := s.router.PathPrefix("/api").Subrouter()

	api.HandleFunc("/etfs", s.handlers.HandleGetAllETFs).Methods("GET", "OPTIONS")
	api.HandleFunc("/etfs/{ticker}", s.handlers.HandleGetETFByTicker).Methods("GET", "OPTIONS")
	api.HandleFunc("/stats", s.handlers.HandleGetStats).Methods("GET", "OPTIONS")
	api.HandleFunc("/asset-classes", s.handlers.HandleGetAssetClasses).Methods("GET", "OPTIONS")
	api.HandleFunc("/top-by-nav", s.handlers.HandleGetTopByNAV).Methods("GET", "OPTIONS")
	api.HandleFunc("/search", s.handlers.HandleSearch).Methods("GET", "OPTIONS")

	// Статические файлы
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir(s.config.StaticDir)))
}

// setupAdminRoutes настраивает защищенные административные маршруты
func (s *Server) setupAdminRoutes() {
	s.adminRouter.Use(loggingMiddleware)
	s.adminRouter.Use(adminMiddleware(s.config.AdminAllowedDNs))

	// Admin API endpoints
	admin := s.adminRouter.PathPrefix("/admin").Subrouter()

	admin.HandleFunc("/scrape", s.handlers.HandleAdminScrape).Methods("POST")
	admin.HandleFunc("/status", s.handlers.HandleAdminStatus).Methods("GET")
	admin.HandleFunc("/info", s.handlers.HandleAdminInfo).Methods("GET")

	// Статическая страница админки
	s.adminRouter.PathPrefix("/").Handler(http.FileServer(http.Dir(s.config.StaticDir + "/admin")))
}

// Start запускает HTTP серверы
func (s *Server) Start() error {
	s.printServerInfo()

	// Запускаем публичный сервер (HTTP)
	go func() {
		publicServer := &http.Server{
			Handler:      s.router,
			Addr:         ":" + s.config.ServerPort,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		log.Printf("🌍 Public server listening on http://localhost:%s", s.config.ServerPort)
		if err := publicServer.ListenAndServe(); err != nil {
			log.Fatalf("Public server error: %v", err)
		}
	}()

	// Запускаем админский сервер (HTTPS с mTLS)
	tlsConfig, err := createTLSConfig(s.config.CACertPath)
	if err != nil {
		return err
	}

	adminServer := &http.Server{
		Handler:      s.adminRouter,
		Addr:         ":" + s.config.AdminPort,
		TLSConfig:    tlsConfig,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("🔒 Admin server listening on https://localhost:%s (mTLS required)", s.config.AdminPort)
	log.Println()

	return adminServer.ListenAndServeTLS(s.config.ServerCertPath, s.config.ServerKeyPath)
}

// printServerInfo выводит информацию о сервере
func (s *Server) printServerInfo() {
	log.Println("==================================================")
	log.Println("🚀 ETF Scraper Server")
	log.Println("==================================================")
	log.Printf("📊 Public API: http://localhost:%s", s.config.ServerPort)
	log.Printf("   GET  /api/etfs                - All ETFs")
	log.Printf("   GET  /api/etfs/{ticker}       - ETF by ticker")
	log.Printf("   GET  /api/stats               - Statistics")
	log.Printf("   GET  /api/asset-classes       - Asset classes")
	log.Printf("   GET  /api/top-by-nav?limit=10 - Top by NAV")
	log.Printf("   GET  /api/search?q=term       - Search")
	log.Println()
	log.Printf("🔒 Admin API: https://localhost:%s (mTLS)", s.config.AdminPort)
	log.Printf("   POST /admin/scrape            - Start scraping")
	log.Printf("   GET  /admin/status            - System status")
	log.Printf("   GET  /admin/info              - Certificate info")
	log.Println()
	log.Printf("📝 Allowed admin DNs:")
	if len(s.config.AdminAllowedDNs) == 0 {
		log.Printf("   ⚠️  WARNING: No admin DNs configured!")
	} else {
		for _, dn := range s.config.AdminAllowedDNs {
			log.Printf("   ✓ %s", dn)
		}
	}
	log.Println("==================================================")
	log.Println()
}
