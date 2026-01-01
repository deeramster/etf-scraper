package server

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// AdminConfig содержит настройки для администраторского доступа
type AdminConfig struct {
	AllowedDNs []string // Список разрешенных Distinguished Names
}

// createTLSConfig создает TLS конфигурацию с требованием клиентского сертификата
func createTLSConfig(caCertPath string) (*tls.Config, error) {
	// Загружаем CA сертификат для проверки клиентских сертификатов
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
		MinVersion: tls.VersionTLS12,
	}, nil
}

// adminMiddleware проверяет клиентский сертификат для доступа к админке
func adminMiddleware(allowedDNs []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем наличие TLS соединения
			if r.TLS == nil {
				log.Printf("Admin access denied: no TLS connection from %s", r.RemoteAddr)
				http.Error(w, "TLS required", http.StatusForbidden)
				return
			}

			// Проверяем наличие клиентских сертификатов
			if len(r.TLS.PeerCertificates) == 0 {
				log.Printf("Admin access denied: no client certificate from %s", r.RemoteAddr)
				http.Error(w, "Client certificate required", http.StatusForbidden)
				return
			}

			// Получаем первый сертификат (клиентский)
			cert := r.TLS.PeerCertificates[0]
			clientDN := cert.Subject.String()

			// Проверяем DN в списке разрешенных
			if !isDNAllowed(clientDN, allowedDNs) {
				log.Printf("Admin access denied: unauthorized DN '%s' from %s", clientDN, r.RemoteAddr)
				http.Error(w, "Unauthorized", http.StatusForbidden)
				return
			}

			log.Printf("Admin access granted: DN '%s' from %s", clientDN, r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
	}
}

// isDNAllowed проверяет, разрешен ли данный DN
func isDNAllowed(clientDN string, allowedDNs []string) bool {
	clientDN = normalizeDN(clientDN)

	for _, allowedDN := range allowedDNs {
		if normalizeDN(allowedDN) == clientDN {
			return true
		}
	}
	return false
}

// normalizeDN нормализует DN для сравнения
func normalizeDN(dn string) string {
	// Удаляем лишние пробелы и приводим к нижнему регистру
	dn = strings.TrimSpace(dn)
	dn = strings.ToLower(dn)
	// Удаляем пробелы вокруг запятых
	dn = strings.ReplaceAll(dn, " ,", ",")
	dn = strings.ReplaceAll(dn, ", ", ",")
	return dn
}

// parseCertInfo извлекает информацию из сертификата для логирования
func parseCertInfo(cert *x509.Certificate) map[string]string {
	return map[string]string{
		"CN":           cert.Subject.CommonName,
		"O":            strings.Join(cert.Subject.Organization, ","),
		"OU":           strings.Join(cert.Subject.OrganizationalUnit, ","),
		"SerialNumber": cert.SerialNumber.String(),
		"Issuer":       cert.Issuer.String(),
	}
}
