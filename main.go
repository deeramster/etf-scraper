package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	_ "github.com/mattn/go-sqlite3"
)

type ETFData struct {
	DateScraped     string
	Ticker          string
	TradeStatus     string
	ManagementCo    string
	AssetClass      string
	TERPercent      *float64
	TERDirection    string
	FundName        string
	ManagementStyle string
	TargetIndex     string
	Currency        string
	StartDate       string
	InfoIcon        string
	PriceChange6M   *float64
	PriceChange2024 *float64
	PriceChange2023 *float64
	PriceChange2022 *float64
	PriceChange2021 *float64
	PriceChange2020 *float64
	NAVMillionRub   *float64
	LastUpdateDate  string
}

type ETFScraper struct {
	URL     string
	DBPath  string
	DB      *sql.DB
	Verbose bool
}

func NewETFScraper(dbPath string, verbose bool) (*ETFScraper, error) {
	scraper := &ETFScraper{
		URL:     "https://assetallocation.ru/etf/",
		DBPath:  dbPath,
		Verbose: verbose,
	}

	if err := scraper.initDatabase(); err != nil {
		return nil, fmt.Errorf("ошибка инициализации БД: %v", err)
	}

	return scraper, nil
}

func (s *ETFScraper) initDatabase() error {
	db, err := sql.Open("sqlite3", s.DBPath)
	if err != nil {
		return err
	}
	s.DB = db

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

	_, err = db.Exec(createTableSQL)
	return err
}

func cleanText(text string) string {
	text = strings.TrimSpace(text)
	space := regexp.MustCompile(`\s+`)
	return space.ReplaceAllString(text, " ")
}

func parseNumber(text string) *float64 {
	text = cleanText(text)

	if text == "" || text == "—" || text == "*—*" || text == "—*" ||
		strings.Contains(text, "⸗️") || strings.Contains(text, "ℹ️") {
		return nil
	}

	original := text

	text = strings.ReplaceAll(text, "*", "")
	text = strings.ReplaceAll(text, "'", "")
	text = strings.ReplaceAll(text, "'", "")
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, "\u00a0", "")
	text = strings.ReplaceAll(text, "%", "")
	text = strings.ReplaceAll(text, "₽", "")
	text = strings.ReplaceAll(text, ",", ".")

	re := regexp.MustCompile(`[^\d.\-]`)
	text = re.ReplaceAllString(text, "")

	if strings.Count(text, ".") > 1 {
		parts := strings.Split(text, ".")
		text = parts[0] + "." + strings.Join(parts[1:], "")
	}

	text = strings.TrimSpace(text)

	if text == "" || text == "." || text == "-" {
		return nil
	}

	val, err := strconv.ParseFloat(text, 64)
	if err != nil {
		log.Printf("Не удалось распарсить число: '%s' -> '%s', ошибка: %v", original, text, err)
		return nil
	}
	return &val
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseRussianMonth(month string) string {
	months := map[string]string{
		"января":   "01",
		"февраля":  "02",
		"марта":    "03",
		"апреля":   "04",
		"мая":      "05",
		"июня":     "06",
		"июля":     "07",
		"августа":  "08",
		"сентября": "09",
		"октября":  "10",
		"ноября":   "11",
		"декабря":  "12",
	}

	monthLower := strings.ToLower(month)
	if num, ok := months[monthLower]; ok {
		return num
	}
	return "00"
}

func formatRussianDate(day, month, year string) string {
	monthNum := parseRussianMonth(month)

	if len(day) == 1 {
		day = "0" + day
	}

	return year + "-" + monthNum + "-" + day
}

func (s *ETFScraper) ScrapeData() ([]ETFData, error) {
	log.Printf("Начинаем скрейпинг %s", s.URL)

	var data []ETFData
	var lastUpdateDate string
	var rowCount int
	var errorCount int

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
	)

	c.SetRequestTimeout(30 * time.Second)

	c.OnHTML("body", func(e *colly.HTMLElement) {
		text := e.Text

		idx := strings.Index(text, "Последнее обновление:")
		if idx != -1 {
			fragment := text[idx+len("Последнее обновление:"):]
			if len(fragment) > 100 {
				fragment = fragment[:100]
			}

			fragment = strings.TrimSpace(fragment)

			re := regexp.MustCompile(`(\d{1,2})\s+(\p{L}+)\s+(\d{4})`)
			matches := re.FindStringSubmatch(fragment)

			if len(matches) >= 4 {
				day := matches[1]
				month := matches[2]
				year := matches[3]

				lastUpdateDate = formatRussianDate(day, month, year)
				log.Printf("Найдена дата обновления: '%s' (исходная: %s %s %s)", lastUpdateDate, day, month, year)
			} else {
				log.Printf("ПРЕДУПРЕЖДЕНИЕ: Не удалось распарсить дату")
				log.Printf("Фрагмент: '%s'", fragment[:min(50, len(fragment))])
			}
		} else {
			log.Printf("ПРЕДУПРЕЖДЕНИЕ: Текст 'Последнее обновление:' не найден на странице")
		}
	})

	c.OnHTML("table", func(e *colly.HTMLElement) {
		// Проверяем, что дата уже найдена
		if lastUpdateDate == "" {
			log.Printf("ПРЕДУПРЕЖДЕНИЕ: Дата обновления не найдена!")
		}

		e.DOM.Find("tr").Each(func(i int, row *goquery.Selection) {
			rowCount++

			if i == 0 {
				if s.Verbose {
					log.Printf("Строка %d: заголовок (пропущена)", i)
				}
				return
			}

			var cols []string
			row.Find("td").Each(func(j int, cell *goquery.Selection) {
				cols = append(cols, cell.Text())
			})

			if len(cols) < 20 {
				if s.Verbose {
					log.Printf("Строка %d: недостаточно колонок (%d), пропущена", i, len(cols))
				}
				errorCount++
				return
			}

			if i <= 3 {
				log.Printf("\n=== Строка %d ===", i)
				log.Printf("Колонок всего: %d", len(cols))
				log.Printf("Дата обновления с сайта: '%s'", lastUpdateDate)
				for idx := 0; idx < len(cols) && idx < 21; idx++ {
					log.Printf("  [%d] = '%s'", idx, cols[idx])
				}
			}

			now := time.Now().Format("2006-01-02 15:04:05")

			etf := ETFData{
				DateScraped:     now,
				Ticker:          cleanText(cols[0]),
				TradeStatus:     cleanText(cols[1]),
				ManagementCo:    cleanText(cols[2]),
				AssetClass:      cleanText(cols[4]),
				TERPercent:      parseNumber(cols[5]),
				TERDirection:    cleanText(cols[6]),
				FundName:        cleanText(cols[7]),
				ManagementStyle: cleanText(cols[8]),
				TargetIndex:     cleanText(cols[9]),
				Currency:        cleanText(cols[10]),
				StartDate:       cleanText(cols[11]),
				InfoIcon:        cleanText(cols[12]),
				PriceChange6M:   parseNumber(cols[13]),
				PriceChange2024: parseNumber(cols[14]),
				PriceChange2023: parseNumber(cols[15]),
				PriceChange2022: parseNumber(cols[16]),
				PriceChange2021: parseNumber(cols[17]),
				PriceChange2020: parseNumber(cols[18]),
				NAVMillionRub:   parseNumber(cols[19]),
				LastUpdateDate:  lastUpdateDate,
			}

			if s.Verbose {
				ter := "nil"
				if etf.TERPercent != nil {
					ter = fmt.Sprintf("%.3f", *etf.TERPercent)
				}
				nav := "nil"
				if etf.NAVMillionRub != nil {
					nav = fmt.Sprintf("%.0f", *etf.NAVMillionRub)
				}
				log.Printf("Строка %d: Тикер=%s, TER_raw='%s', TER=%s, NAV_raw='%s', NAV=%s, UpdateDate=%s",
					i, etf.Ticker, cols[5], ter, cols[19], nav, etf.LastUpdateDate)
			}

			data = append(data, etf)
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Ошибка при запросе: %v", err)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("Получен ответ: %d байт, статус: %d", len(r.Body), r.StatusCode)
	})

	err := c.Visit(s.URL)
	if err != nil {
		return nil, err
	}

	log.Printf("Обработано строк: %d", rowCount)
	log.Printf("Ошибок парсинга: %d", errorCount)
	log.Printf("Успешно извлечено записей: %d", len(data))
	log.Printf("Дата обновления с сайта: %s", lastUpdateDate)

	if len(data) == 0 {
		return nil, fmt.Errorf("не удалось извлечь данные из таблицы")
	}

	if lastUpdateDate == "" {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: Дата обновления не была найдена на странице!")
	}

	return data, nil
}

func (s *ETFScraper) SaveToDatabase(data []ETFData) error {
	if len(data) == 0 {
		return fmt.Errorf("нет данных для сохранения")
	}

	tx, err := s.DB.Begin()
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

func (s *ETFScraper) GetLatestData(ticker string) ([]ETFData, error) {
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

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []ETFData
	for rows.Next() {
		var etf ETFData
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

func (s *ETFScraper) GetStats() error {
	var totalRecords, uniqueTickers, scrapeSessions int

	err := s.DB.QueryRow("SELECT COUNT(*) FROM etf_data").Scan(&totalRecords)
	if err != nil {
		return err
	}

	err = s.DB.QueryRow("SELECT COUNT(DISTINCT ticker) FROM etf_data").Scan(&uniqueTickers)
	if err != nil {
		return err
	}

	err = s.DB.QueryRow("SELECT COUNT(DISTINCT date_scraped) FROM etf_data").Scan(&scrapeSessions)
	if err != nil {
		return err
	}

	log.Println("\n==================================================")
	log.Println("Статистика базы данных:")
	log.Println("==================================================")
	log.Printf("Всего записей: %d", totalRecords)
	log.Printf("Уникальных тикеров: %d", uniqueTickers)
	log.Printf("Сеансов скрейпинга: %d", scrapeSessions)

	return nil
}

func (s *ETFScraper) GetTopByNAV(limit int) ([]ETFData, error) {
	query := `
		SELECT * FROM etf_data 
		WHERE date_scraped = (SELECT MAX(date_scraped) FROM etf_data)
		AND nav_million_rub IS NOT NULL
		ORDER BY nav_million_rub DESC 
		LIMIT ?
	`

	rows, err := s.DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []ETFData
	for rows.Next() {
		var etf ETFData
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

func (s *ETFScraper) Run() error {
	log.Println("==================================================")
	log.Printf("Запуск скрейпера: %s", time.Now().Format("2006-01-02 15:04:05"))
	log.Println("==================================================")

	data, err := s.ScrapeData()
	if err != nil {
		log.Printf("✗ Ошибка при скрейпинге: %v", err)
		return err
	}

	if err := s.SaveToDatabase(data); err != nil {
		log.Printf("✗ Ошибка при сохранении: %v", err)
		return err
	}

	log.Println("✓ Скрейпинг успешно завершен")
	return nil
}

func (s *ETFScraper) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func runScraper() {
	verbose := false
	scraper, err := NewETFScraper("etf_data.db", verbose)
	if err != nil {
		log.Fatalf("Ошибка создания скрейпера: %v", err)
	}
	defer scraper.Close()

	if err := scraper.Run(); err != nil {
		log.Fatalf("Ошибка выполнения: %v", err)
	}

	if err := scraper.GetStats(); err != nil {
		log.Printf("Ошибка получения статистики: %v", err)
	}

	log.Println("\n==================================================")
	log.Println("Топ-10 фондов по размеру СЧА:")
	log.Println("==================================================")

	topFunds, err := scraper.GetTopByNAV(10)
	if err != nil {
		log.Printf("Ошибка получения топа: %v", err)
	} else {
		for i, etf := range topFunds {
			terVal := 0.0
			if etf.TERPercent != nil {
				terVal = *etf.TERPercent
			}
			navVal := 0.0
			if etf.NAVMillionRub != nil {
				navVal = *etf.NAVMillionRub
			}

			fmt.Printf("%2d. %-8s | %-40s | TER: %5.2f%% | СЧА: %10.0f млн ₽\n",
				i+1, etf.Ticker, truncate(etf.FundName, 40), terVal, navVal)
		}
	}

	log.Println("\n✓ Все данные сохранены в etf_data.db")
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "serve":
			RunServer()
			return
		case "scrape":
			runScraper()
			return
		}
	}

	runScraper()
}
