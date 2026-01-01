package scraper

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

// cleanText очищает текст от лишних пробелов
func cleanText(text string) string {
	text = strings.TrimSpace(text)
	space := regexp.MustCompile(`\s+`)
	return space.ReplaceAllString(text, " ")
}

// parseNumber парсит строку в число, обрабатывая различные форматы
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

// parseRussianMonth конвертирует название месяца на русском в номер
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

// formatRussianDate форматирует русскую дату в формат YYYY-MM-DD
func formatRussianDate(day, month, year string) string {
	monthNum := parseRussianMonth(month)

	if len(day) == 1 {
		day = "0" + day
	}

	return year + "-" + monthNum + "-" + day
}

// min возвращает минимум из двух чисел
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
