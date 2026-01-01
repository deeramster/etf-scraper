package models

// ETFData представляет данные о ETF фонде
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

// ETFResponse представляет ответ API для ETF
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

// StatsResponse представляет статистику по ETF
type StatsResponse struct {
	TotalRecords   int     `json:"totalRecords"`
	UniqueTickers  int     `json:"uniqueTickers"`
	ScrapeSessions int     `json:"scrapeSessions"`
	TotalNAV       float64 `json:"totalNAV"`
	AvgTER         float64 `json:"avgTER"`
	LastUpdate     string  `json:"lastUpdate"`
}

// AssetClassResponse представляет список классов активов
type AssetClassResponse struct {
	AssetClasses []string `json:"assetClasses"`
}

// ScrapeResponse представляет ответ на запрос скрейпинга
type ScrapeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
