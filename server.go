package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type APIServer struct {
	DB     *sql.DB
	Router *mux.Router
	Port   string
}

type ETFResponse struct {
	ID              int      `json:"id"`
	DateScraped     string   `json:"dateScraped"`
	Ticker          string   `json:"ticker"`
	TradeStatus     string   `json:"tradeStatus"`
	ManagementCo    string   `json:"managementCo"`
	AssetClass      string   `json:"assetClass"`
	TERPercent      *float64 `json:"terPercent"`
	TERDirection    string   `json:"terDirection"`
	FundName        string   `json:"fundName"`
	ManagementStyle string   `json:"managementStyle"`
	TargetIndex     string   `json:"targetIndex"`
	Currency        string   `json:"currency"`
	StartDate       string   `json:"startDate"`
	InfoIcon        string   `json:"infoIcon"`
	PriceChange6M   *float64 `json:"priceChange6M"`
	PriceChange2024 *float64 `json:"priceChange2024"`
	PriceChange2023 *float64 `json:"priceChange2023"`
	PriceChange2022 *float64 `json:"priceChange2022"`
	PriceChange2021 *float64 `json:"priceChange2021"`
	PriceChange2020 *float64 `json:"priceChange2020"`
	NAVMillionRub   *float64 `json:"navMillionRub"`
	LastUpdateDate  string   `json:"lastUpdateDate"`
}

type StatsResponse struct {
	TotalRecords   int     `json:"totalRecords"`
	UniqueTickers  int     `json:"uniqueTickers"`
	ScrapeSessions int     `json:"scrapeSessions"`
	TotalNAV       float64 `json:"totalNAV"`
	AvgTER         float64 `json:"avgTER"`
	LastUpdate     string  `json:"lastUpdate"`
}

type AssetClassResponse struct {
	AssetClasses []string `json:"assetClasses"`
}

func NewAPIServer(dbPath string, port string) (*APIServer, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	server := &APIServer{
		DB:     db,
		Router: mux.NewRouter(),
		Port:   port,
	}

	server.setupRoutes()
	return server, nil
}

func (s *APIServer) setupRoutes() {
	s.Router.Use(corsMiddleware)

	s.Router.HandleFunc("/api/etfs", s.handleGetAllETFs).Methods("GET", "OPTIONS")
	s.Router.HandleFunc("/api/etfs/{ticker}", s.handleGetETFByTicker).Methods("GET", "OPTIONS")
	s.Router.HandleFunc("/api/stats", s.handleGetStats).Methods("GET", "OPTIONS")
	s.Router.HandleFunc("/api/asset-classes", s.handleGetAssetClasses).Methods("GET", "OPTIONS")
	s.Router.HandleFunc("/api/top-by-nav", s.handleGetTopByNAV).Methods("GET", "OPTIONS")
	s.Router.HandleFunc("/api/search", s.handleSearch).Methods("GET", "OPTIONS")
	s.Router.HandleFunc("/api/scrape", s.handleScrape).Methods("POST", "OPTIONS")

	s.Router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *APIServer) handleGetAllETFs(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	assetClass := queryParams.Get("assetClass")
	sortBy := queryParams.Get("sortBy")
	order := queryParams.Get("order")

	if sortBy == "" {
		sortBy = "nav_million_rub"
	}
	if order == "" {
		order = "DESC"
	}

	query := `
		SELECT 
			id, date_scraped, ticker, trade_status, management_company, 
			asset_class, ter_percent, ter_direction, fund_name, management_style, 
			target_index, currency, start_date, info_icon, price_change_6m, 
			price_change_2024, price_change_2023, price_change_2022, 
			price_change_2021, price_change_2020, nav_million_rub, last_update_date
		FROM etf_data 
		WHERE date_scraped = (SELECT MAX(date_scraped) FROM etf_data)
	`

	if assetClass != "" && assetClass != "Все" {
		query += fmt.Sprintf(" AND asset_class = '%s'", assetClass)
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, order)

	rows, err := s.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	etfs := []ETFResponse{}
	for rows.Next() {
		var etf ETFResponse
		err := rows.Scan(
			&etf.ID, &etf.DateScraped, &etf.Ticker, &etf.TradeStatus,
			&etf.ManagementCo, &etf.AssetClass, &etf.TERPercent, &etf.TERDirection,
			&etf.FundName, &etf.ManagementStyle, &etf.TargetIndex, &etf.Currency,
			&etf.StartDate, &etf.InfoIcon, &etf.PriceChange6M, &etf.PriceChange2024,
			&etf.PriceChange2023, &etf.PriceChange2022, &etf.PriceChange2021,
			&etf.PriceChange2020, &etf.NAVMillionRub, &etf.LastUpdateDate,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		etfs = append(etfs, etf)
	}

	respondJSON(w, etfs)
}

func (s *APIServer) handleGetETFByTicker(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticker := vars["ticker"]

	query := `
		SELECT 
			id, date_scraped, ticker, trade_status, management_company, 
			asset_class, ter_percent, ter_direction, fund_name, management_style, 
			target_index, currency, start_date, info_icon, price_change_6m, 
			price_change_2024, price_change_2023, price_change_2022, 
			price_change_2021, price_change_2020, nav_million_rub, last_update_date
		FROM etf_data 
		WHERE ticker = ? 
		ORDER BY date_scraped DESC 
		LIMIT 1
	`

	var etf ETFResponse
	err := s.DB.QueryRow(query, ticker).Scan(
		&etf.ID, &etf.DateScraped, &etf.Ticker, &etf.TradeStatus,
		&etf.ManagementCo, &etf.AssetClass, &etf.TERPercent, &etf.TERDirection,
		&etf.FundName, &etf.ManagementStyle, &etf.TargetIndex, &etf.Currency,
		&etf.StartDate, &etf.InfoIcon, &etf.PriceChange6M, &etf.PriceChange2024,
		&etf.PriceChange2023, &etf.PriceChange2022, &etf.PriceChange2021,
		&etf.PriceChange2020, &etf.NAVMillionRub, &etf.LastUpdateDate,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "ETF not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, etf)
}

func (s *APIServer) handleGetStats(w http.ResponseWriter, r *http.Request) {
	var stats StatsResponse

	err := s.DB.QueryRow("SELECT COUNT(*) FROM etf_data").Scan(&stats.TotalRecords)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.DB.QueryRow("SELECT COUNT(DISTINCT ticker) FROM etf_data").Scan(&stats.UniqueTickers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.DB.QueryRow("SELECT COUNT(DISTINCT date_scraped) FROM etf_data").Scan(&stats.ScrapeSessions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.DB.QueryRow(`
		SELECT 
			COALESCE(SUM(nav_million_rub), 0),
			COALESCE(AVG(ter_percent), 0)
		FROM etf_data 
		WHERE date_scraped = (SELECT MAX(date_scraped) FROM etf_data)
	`).Scan(&stats.TotalNAV, &stats.AvgTER)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.DB.QueryRow("SELECT MAX(date_scraped) FROM etf_data").Scan(&stats.LastUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, stats)
}

func (s *APIServer) handleGetAssetClasses(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT DISTINCT asset_class 
		FROM etf_data 
		WHERE date_scraped = (SELECT MAX(date_scraped) FROM etf_data)
		ORDER BY asset_class
	`

	rows, err := s.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	assetClasses := []string{"Все"}
	for rows.Next() {
		var ac string
		if err := rows.Scan(&ac); err != nil {
			continue
		}
		assetClasses = append(assetClasses, ac)
	}

	respondJSON(w, AssetClassResponse{AssetClasses: assetClasses})
}

func (s *APIServer) handleGetTopByNAV(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	query := `
		SELECT 
			id, date_scraped, ticker, trade_status, management_company, 
			asset_class, ter_percent, ter_direction, fund_name, management_style, 
			target_index, currency, start_date, info_icon, price_change_6m, 
			price_change_2024, price_change_2023, price_change_2022, 
			price_change_2021, price_change_2020, nav_million_rub, last_update_date
		FROM etf_data 
		WHERE date_scraped = (SELECT MAX(date_scraped) FROM etf_data)
		AND nav_million_rub IS NOT NULL
		ORDER BY nav_million_rub DESC 
		LIMIT ?
	`

	rows, err := s.DB.Query(query, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	etfs := []ETFResponse{}
	for rows.Next() {
		var etf ETFResponse
		err := rows.Scan(
			&etf.ID, &etf.DateScraped, &etf.Ticker, &etf.TradeStatus,
			&etf.ManagementCo, &etf.AssetClass, &etf.TERPercent, &etf.TERDirection,
			&etf.FundName, &etf.ManagementStyle, &etf.TargetIndex, &etf.Currency,
			&etf.StartDate, &etf.InfoIcon, &etf.PriceChange6M, &etf.PriceChange2024,
			&etf.PriceChange2023, &etf.PriceChange2022, &etf.PriceChange2021,
			&etf.PriceChange2020, &etf.NAVMillionRub, &etf.LastUpdateDate,
		)
		if err != nil {
			continue
		}
		etfs = append(etfs, etf)
	}

	respondJSON(w, etfs)
}

func (s *APIServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		http.Error(w, "search term required", http.StatusBadRequest)
		return
	}

	query := `
		SELECT 
			id, date_scraped, ticker, trade_status, management_company, 
			asset_class, ter_percent, ter_direction, fund_name, management_style, 
			target_index, currency, start_date, info_icon, price_change_6m, 
			price_change_2024, price_change_2023, price_change_2022, 
			price_change_2021, price_change_2020, nav_million_rub, last_update_date
		FROM etf_data 
		WHERE date_scraped = (SELECT MAX(date_scraped) FROM etf_data)
		AND (
			ticker LIKE ? OR 
			fund_name LIKE ? OR 
			management_company LIKE ?
		)
		ORDER BY nav_million_rub DESC
	`

	searchPattern := "%" + searchTerm + "%"
	rows, err := s.DB.Query(query, searchPattern, searchPattern, searchPattern)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	etfs := []ETFResponse{}
	for rows.Next() {
		var etf ETFResponse
		err := rows.Scan(
			&etf.ID, &etf.DateScraped, &etf.Ticker, &etf.TradeStatus,
			&etf.ManagementCo, &etf.AssetClass, &etf.TERPercent, &etf.TERDirection,
			&etf.FundName, &etf.ManagementStyle, &etf.TargetIndex, &etf.Currency,
			&etf.StartDate, &etf.InfoIcon, &etf.PriceChange6M, &etf.PriceChange2024,
			&etf.PriceChange2023, &etf.PriceChange2022, &etf.PriceChange2021,
			&etf.PriceChange2020, &etf.NAVMillionRub, &etf.LastUpdateDate,
		)
		if err != nil {
			continue
		}
		etfs = append(etfs, etf)
	}

	respondJSON(w, etfs)
}

func (s *APIServer) handleScrape(w http.ResponseWriter, r *http.Request) {
	go func() {
		scraper, err := NewETFScraper("etf_data.db", false)
		if err != nil {
			log.Printf("Ошибка создания скрейпера: %v", err)
			return
		}
		defer scraper.Close()

		if err := scraper.Run(); err != nil {
			log.Printf("Ошибка скрейпинга: %v", err)
			return
		}
		log.Println("Скрейпинг завершен успешно")
	}()

	respondJSON(w, map[string]string{
		"status":  "started",
		"message": "Scraping started in background",
	})
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *APIServer) Start() error {
	log.Printf("🚀 API сервер запущен на http://localhost:%s", s.Port)
	log.Printf("📊 API endpoints:")
	log.Printf("   GET  /api/etfs                - Все ETF")
	log.Printf("   GET  /api/etfs/{ticker}       - ETF по тикеру")
	log.Printf("   GET  /api/stats               - Статистика")
	log.Printf("   GET  /api/asset-classes       - Классы активов")
	log.Printf("   GET  /api/top-by-nav?limit=10 - Топ по СЧА")
	log.Printf("   GET  /api/search?q=term       - Поиск")
	log.Printf("   POST /api/scrape              - Запуск скрейпинга")
	log.Println()

	srv := &http.Server{
		Handler:      s.Router,
		Addr:         ":" + s.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return srv.ListenAndServe()
}

func (s *APIServer) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}

func RunServer() {
	port := "8080"

	server, err := NewAPIServer("etf_data.db", port)
	if err != nil {
		log.Fatalf("Ошибка создания сервера: %v", err)
	}
	defer server.Close()

	if err := server.Start(); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
