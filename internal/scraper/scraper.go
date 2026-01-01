package scraper

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"etf-scraper/internal/config"
	"etf-scraper/internal/database"
	"etf-scraper/internal/models"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

// Scraper представляет скрейпер ETF данных
type Scraper struct {
	config *config.Config
	repo   *database.Repository
}

// NewScraper создает новый скрейпер
func NewScraper(cfg *config.Config, repo *database.Repository) *Scraper {
	return &Scraper{
		config: cfg,
		repo:   repo,
	}
}

// Run выполняет скрейпинг и сохранение данных
func (s *Scraper) Run() error {
	log.Println("==================================================")
	log.Printf("Запуск скрейпера: %s", time.Now().Format("2006-01-02 15:04:05"))
	log.Println("==================================================")

	data, err := s.ScrapeData()
	if err != nil {
		log.Printf("✗ Ошибка при скрейпинге: %v", err)
		return err
	}

	if err := s.repo.SaveETFs(data); err != nil {
		log.Printf("✗ Ошибка при сохранении: %v", err)
		return err
	}

	log.Println("✓ Скрейпинг успешно завершен")
	return nil
}

// ScrapeData выполняет скрейпинг данных с сайта
func (s *Scraper) ScrapeData() ([]models.ETFData, error) {
	log.Printf("Начинаем скрейпинг %s", s.config.ScraperURL)

	var data []models.ETFData
	var lastUpdateDate string
	var rowCount int
	var errorCount int

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
	)

	c.SetRequestTimeout(30 * time.Second)

	// Парсим дату обновления
	c.OnHTML("body", func(e *colly.HTMLElement) {
		lastUpdateDate = s.parseUpdateDate(e.Text)
	})

	// Парсим таблицу с данными
	c.OnHTML("table", func(e *colly.HTMLElement) {
		if lastUpdateDate == "" {
			log.Printf("ПРЕДУПРЕЖДЕНИЕ: Дата обновления не найдена!")
		}

		e.DOM.Find("tr").Each(func(i int, row *goquery.Selection) {
			rowCount++

			// Пропускаем заголовок
			if i == 0 {
				if s.config.Verbose {
					log.Printf("Строка %d: заголовок (пропущена)", i)
				}
				return
			}

			// Извлекаем данные из строки
			etf, err := s.parseTableRow(row, i, lastUpdateDate)
			if err != nil {
				if s.config.Verbose {
					log.Printf("Строка %d: %v", i, err)
				}
				errorCount++
				return
			}

			data = append(data, *etf)
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Ошибка при запросе: %v", err)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("Получен ответ: %d байт, статус: %d", len(r.Body), r.StatusCode)
	})

	err := c.Visit(s.config.ScraperURL)
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

// parseUpdateDate извлекает дату обновления из текста страницы
func (s *Scraper) parseUpdateDate(text string) string {
	idx := strings.Index(text, "Последнее обновление:")
	if idx == -1 {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: Текст 'Последнее обновление:' не найден на странице")
		return ""
	}

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

		dateStr := formatRussianDate(day, month, year)
		log.Printf("Найдена дата обновления: '%s' (исходная: %s %s %s)", dateStr, day, month, year)
		return dateStr
	}

	log.Printf("ПРЕДУПРЕЖДЕНИЕ: Не удалось распарсить дату")
	log.Printf("Фрагмент: '%s'", fragment[:min(50, len(fragment))])
	return ""
}

// parseTableRow парсит одну строку таблицы
func (s *Scraper) parseTableRow(row *goquery.Selection, index int, lastUpdateDate string) (*models.ETFData, error) {
	var cols []string
	row.Find("td").Each(func(j int, cell *goquery.Selection) {
		cols = append(cols, cell.Text())
	})

	if len(cols) < 20 {
		return nil, fmt.Errorf("недостаточно колонок (%d)", len(cols))
	}

	// Логируем первые несколько строк для отладки
	if index <= 3 {
		log.Printf("\n=== Строка %d ===", index)
		log.Printf("Колонок всего: %d", len(cols))
		log.Printf("Дата обновления с сайта: '%s'", lastUpdateDate)
		for idx := 0; idx < len(cols) && idx < 21; idx++ {
			log.Printf("  [%d] = '%s'", idx, cols[idx])
		}
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	etf := &models.ETFData{
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

	if s.config.Verbose {
		ter := "nil"
		if etf.TERPercent != nil {
			ter = fmt.Sprintf("%.3f", *etf.TERPercent)
		}
		nav := "nil"
		if etf.NAVMillionRub != nil {
			nav = fmt.Sprintf("%.0f", *etf.NAVMillionRub)
		}
		log.Printf("Строка %d: Тикер=%s, TER_raw='%s', TER=%s, NAV_raw='%s', NAV=%s, UpdateDate=%s",
			index, etf.Ticker, cols[5], ter, cols[19], nav, etf.LastUpdateDate)
	}

	return etf, nil
}

// PrintStats выводит статистику БД
func (s *Scraper) PrintStats() error {
	totalRecords, uniqueTickers, scrapeSessions, err := s.repo.GetStats()
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

// PrintTopFunds выводит топ фондов по СЧА
func (s *Scraper) PrintTopFunds(limit int) error {
	log.Println("\n==================================================")
	log.Printf("Топ-%d фондов по размеру СЧА:", limit)
	log.Println("==================================================")

	topFunds, err := s.repo.GetTopByNAV(limit)
	if err != nil {
		return err
	}

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

	return nil
}

// truncate обрезает строку до заданной длины
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
