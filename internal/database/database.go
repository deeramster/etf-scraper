package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Database представляет подключение к базе данных
type Database struct {
	DB *sql.DB
}

// NewDatabase создает новое подключение к БД и инициализирует схему
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия БД: %w", err)
	}

	database := &Database{DB: db}
	if err := database.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return database, nil
}

// initSchema создает таблицы и индексы в БД
func (d *Database) initSchema() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS etf_data (
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

	CREATE INDEX IF NOT EXISTS idx_date_ticker 
	ON etf_data(date_scraped, ticker);
	`

	_, err := d.DB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("ошибка создания схемы: %w", err)
	}

	return nil
}

// Close закрывает подключение к БД
func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}
