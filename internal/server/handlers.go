package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"etf-scraper/internal/config"
	"etf-scraper/internal/database"
	"etf-scraper/internal/models"
	"etf-scraper/internal/scraper"

	"github.com/gorilla/mux"
)

// Handlers содержит все HTTP обработчики
type Handlers struct {
	config *config.Config
	db     *database.Database
	repo   *database.Repository
}

// NewHandlers создает новый набор обработчиков
func NewHandlers(cfg *config.Config, db *database.Database, repo *database.Repository) *Handlers {
	return &Handlers{
		config: cfg,
		db:     db,
		repo:   repo,
	}
}

// HandleGetAllETFs возвращает все ETF с возможностью фильтрации и сортировки
func (h *Handlers) HandleGetAllETFs(w http.ResponseWriter, r *http.Request) {
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

	rows, err := h.db.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	etfs := []models.ETFResponse{}
	for rows.Next() {
		var etf models.ETFResponse
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

// HandleGetETFByTicker возвращает данные конкретного ETF по тикеру
func (h *Handlers) HandleGetETFByTicker(w http.ResponseWriter, r *http.Request) {
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

	var etf models.ETFResponse
	err := h.db.DB.QueryRow(query, ticker).Scan(
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

// HandleGetStats возвращает статистику по ETF
func (h *Handlers) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	var stats models.StatsResponse

	err := h.db.DB.QueryRow("SELECT COUNT(*) FROM etf_data").Scan(&stats.TotalRecords)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.db.DB.QueryRow("SELECT COUNT(DISTINCT ticker) FROM etf_data").Scan(&stats.UniqueTickers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.db.DB.QueryRow("SELECT COUNT(DISTINCT date_scraped) FROM etf_data").Scan(&stats.ScrapeSessions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.db.DB.QueryRow(`
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

	var rawDate string
	err = h.db.DB.QueryRow(`
		SELECT last_update_date 
		FROM etf_data 
		WHERE date_scraped = (SELECT MAX(date_scraped) FROM etf_data)
		LIMIT 1
	`).Scan(&rawDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rawDate == "" {
		stats.LastUpdate = "Нет данных"
	} else {
		stats.LastUpdate = rawDate
	}

	respondJSON(w, stats)
}

// HandleGetAssetClasses возвращает список классов активов
func (h *Handlers) HandleGetAssetClasses(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT DISTINCT asset_class 
		FROM etf_data 
		WHERE date_scraped = (SELECT MAX(date_scraped) FROM etf_data)
		ORDER BY asset_class
	`

	rows, err := h.db.DB.Query(query)
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

	respondJSON(w, models.AssetClassResponse{AssetClasses: assetClasses})
}

// HandleGetTopByNAV возвращает топ ETF по размеру СЧА
func (h *Handlers) HandleGetTopByNAV(w http.ResponseWriter, r *http.Request) {
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

	rows, err := h.db.DB.Query(query, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	etfs := []models.ETFResponse{}
	for rows.Next() {
		var etf models.ETFResponse
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

// HandleSearch выполняет поиск ETF
func (h *Handlers) HandleSearch(w http.ResponseWriter, r *http.Request) {
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
	rows, err := h.db.DB.Query(query, searchPattern, searchPattern, searchPattern)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	etfs := []models.ETFResponse{}
	for rows.Next() {
		var etf models.ETFResponse
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

// HandleScrape запускает скрейпинг в фоновом режиме
func (h *Handlers) HandleScrape(w http.ResponseWriter, r *http.Request) {
	go func() {
		s := scraper.NewScraper(h.config, h.repo)
		if err := s.Run(); err != nil {
			log.Printf("Ошибка скрейпинга: %v", err)
			return
		}
		log.Println("Скрейпинг завершен успешно")
	}()

	respondJSON(w, models.ScrapeResponse{
		Status:  "started",
		Message: "Scraping started in background",
	})
}

// respondJSON отправляет JSON ответ
func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
