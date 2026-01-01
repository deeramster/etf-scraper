// API клиент для работы с backend
const API_BASE_URL = 'http://localhost:8080/api';

class APIService {
    /**
     * Получить все ETF с фильтрацией и сортировкой
     */
    async getETFs(params = {}) {
        const { sortBy = 'nav_million_rub', order = 'DESC', assetClass } = params;

        let url = `${API_BASE_URL}/etfs?sortBy=${sortBy}&order=${order}`;
        if (assetClass && assetClass !== 'Все') {
            url += `&assetClass=${encodeURIComponent(assetClass)}`;
        }

        const response = await fetch(url);
        if (!response.ok) {
            throw new Error('Ошибка загрузки данных ETF');
        }
        return await response.json();
    }

    /**
     * Получить ETF по тикеру
     */
    async getETFByTicker(ticker) {
        const response = await fetch(`${API_BASE_URL}/etfs/${ticker}`);
        if (!response.ok) {
            if (response.status === 404) {
                throw new Error('ETF не найден');
            }
            throw new Error('Ошибка загрузки данных ETF');
        }
        return await response.json();
    }

    /**
     * Получить статистику
     */
    async getStats() {
        const response = await fetch(`${API_BASE_URL}/stats`);
        if (!response.ok) {
            throw new Error('Ошибка загрузки статистики');
        }
        return await response.json();
    }

    /**
     * Получить список классов активов
     */
    async getAssetClasses() {
        const response = await fetch(`${API_BASE_URL}/asset-classes`);
        if (!response.ok) {
            throw new Error('Ошибка загрузки классов активов');
        }
        return await response.json();
    }

    /**
     * Получить топ ETF по размеру СЧА
     */
    async getTopByNAV(limit = 10) {
        const response = await fetch(`${API_BASE_URL}/top-by-nav?limit=${limit}`);
        if (!response.ok) {
            throw new Error('Ошибка загрузки топ ETF');
        }
        return await response.json();
    }

    /**
     * Поиск ETF
     */
    async search(query) {
        const response = await fetch(`${API_BASE_URL}/search?q=${encodeURIComponent(query)}`);
        if (!response.ok) {
            throw new Error('Ошибка поиска');
        }
        return await response.json();
    }

    /**
     * Запустить скрейпинг
     */
    async startScraping() {
        const response = await fetch(`${API_BASE_URL}/scrape`, {
            method: 'POST',
        });
        if (!response.ok) {
            throw new Error('Ошибка запуска скрейпинга');
        }
        return await response.json();
    }
}

// Экспортируем singleton
const apiService = new APIService();