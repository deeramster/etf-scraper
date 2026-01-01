package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"etf-scraper/internal/scraper"
)

// HandleAdminScrape запускает скрейпинг (только для администраторов)
func (h *Handlers) HandleAdminScrape(w http.ResponseWriter, r *http.Request) {
	// Получаем информацию о клиенте из сертификата
	clientDN := ""
	if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
		clientDN = r.TLS.PeerCertificates[0].Subject.String()
	}

	log.Printf("Admin scraping initiated by: %s from %s", clientDN, r.RemoteAddr)

	// Запускаем скрейпинг в фоновом режиме
	go func() {
		startTime := time.Now()
		s := scraper.NewScraper(h.config, h.repo)
		if err := s.Run(); err != nil {
			log.Printf("❌ Scraping error (initiated by %s): %v", clientDN, err)
			return
		}
		duration := time.Since(startTime)
		log.Printf("✅ Scraping completed successfully (initiated by %s, duration: %s)", clientDN, duration)
	}()

	response := map[string]interface{}{
		"status":      "started",
		"message":     "Scraping started in background",
		"initiatedBy": clientDN,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAdminStatus возвращает статус системы
func (h *Handlers) HandleAdminStatus(w http.ResponseWriter, r *http.Request) {
	// Получаем все 4 значения из GetStats()
	totalRecords, uniqueTickers, scrapeSessions, err := h.repo.GetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":         "ok",
		"totalRecords":   totalRecords,
		"uniqueTickers":  uniqueTickers,
		"scrapeSessions": scrapeSessions,
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAdminInfo показывает информацию об администраторе
func (h *Handlers) HandleAdminInfo(w http.ResponseWriter, r *http.Request) {
	clientInfo := map[string]string{
		"message": "No certificate information available",
	}

	if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
		cert := r.TLS.PeerCertificates[0]
		clientInfo = parseCertInfo(cert)
		clientInfo["dn"] = cert.Subject.String()
		clientInfo["validFrom"] = cert.NotBefore.Format(time.RFC3339)
		clientInfo["validUntil"] = cert.NotAfter.Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clientInfo)
}
