package database

import (
	"database/sql"
	"fmt"
	"log"

	"etf-scraper/internal/models"
)

// Repository предоставляет методы для работы с данными ETF
type Repository struct {
	db *Database
}

// NewRepository создает новый репозиторий
func NewRepository(db *Database) *Repository {
	return &Repository{db: db}
}

// SaveETFs сохраняет массив ETF данных в БД
func (r *Repository) SaveETFs(data []models.ETFData) error {
	if len(data) == 0 {
		return fmt.Errorf("нет данных для сохранения")
	}

	tx, err := r.db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO etf_data (
			date_scraped, ticker, trade_status, management_company, asset_class,
			ter_percent, ter_direction, fund_name, management_style, target_index,
			currency, start_date, info_icon, price_change_6m, price_change_2024,
			price_change_2023, price_change_2022, price_change_2021, price_change_2020,
			nav_million_rub, last_update_date
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	savedCount := 0
	for _, etf := range data {
		_, err := stmt.Exec(
			etf.DateScraped, etf.Ticker, etf.TradeStatus, etf.ManagementCo, etf.AssetClass,
			etf.TERPercent, etf.TERDirection, etf.FundName, etf.ManagementStyle, etf.TargetIndex,
			etf.Currency, etf.StartDate, etf.InfoIcon, etf.PriceChange6M, etf.PriceChange2024,
			etf.PriceChange2023, etf.PriceChange2022, etf.PriceChange2021, etf.PriceChange2020,
			etf.NAVMillionRub, etf.LastUpdateDate,
		)
		if err != nil {
			log.Printf("Ошибка сохранения записи %s: %v", etf.Ticker, err)
			continue
		}
		savedCount++
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Сохранено записей в БД: %d из %d", savedCount, len(data))
	return nil
}

// GetLatestETFs возвращает последние данные ETF
func (r *Repository) GetLatestETFs(ticker string) ([]models.ETFData, error) {
	var query string
	var args []interface{}

	if ticker != "" {
		query = `
			SELECT * FROM etf_data 
			WHERE ticker = ? 
			ORDER BY date_scraped DESC 
			LIMIT 1
		`
		args = append(args, ticker)
	} else {
		query = `
			SELECT * FROM etf_data 
			WHERE date_scraped = (SELECT MAX(date_scraped) FROM etf_data)
			ORDER BY ticker
		`
	}

	rows, err := r.db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanETFRows(rows)
}

// GetTopByNAV возвращает топ ETF по размеру СЧА
func (r *Repository) GetTopByNAV(limit int) ([]models.ETFData, error) {
	query := `
		SELECT * FROM etf_data 
		WHERE date_scraped = (SELECT MAX(date_scraped) FROM etf_data)
		AND nav_million_rub IS NOT NULL
		ORDER BY nav_million_rub DESC 
		LIMIT ?
	`

	rows, err := r.db.DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanETFRows(rows)
}

// GetStats возвращает статистику по БД
func (r *Repository) GetStats() (totalRecords, uniqueTickers, scrapeSessions int, err error) {
	err = r.db.DB.QueryRow("SELECT COUNT(*) FROM etf_data").Scan(&totalRecords)
	if err != nil {
		return
	}

	err = r.db.DB.QueryRow("SELECT COUNT(DISTINCT ticker) FROM etf_data").Scan(&uniqueTickers)
	if err != nil {
		return
	}

	err = r.db.DB.QueryRow("SELECT COUNT(DISTINCT date_scraped) FROM etf_data").Scan(&scrapeSessions)
	return
}

// scanETFRows сканирует строки БД в срез ETFData
func (r *Repository) scanETFRows(rows *sql.Rows) ([]models.ETFData, error) {
	var data []models.ETFData
	for rows.Next() {
		var etf models.ETFData
		var id int

		err := rows.Scan(
			&id, &etf.DateScraped, &etf.Ticker, &etf.TradeStatus, &etf.ManagementCo,
			&etf.AssetClass, &etf.TERPercent, &etf.TERDirection, &etf.FundName,
			&etf.ManagementStyle, &etf.TargetIndex, &etf.Currency, &etf.StartDate,
			&etf.InfoIcon, &etf.PriceChange6M, &etf.PriceChange2024, &etf.PriceChange2023,
			&etf.PriceChange2022, &etf.PriceChange2021, &etf.PriceChange2020,
			&etf.NAVMillionRub, &etf.LastUpdateDate,
		)
		if err != nil {
			return nil, err
		}
		data = append(data, etf)
	}

	return data, nil
}
