// Вспомогательные функции для работы с данными

/**
 * Форматирование числа с разделителями тысяч
 */
function formatNumber(num, decimals = 0) {
    if (num === null || num === undefined) return '-';
    return num.toLocaleString('ru-RU', {
        minimumFractionDigits: decimals,
        maximumFractionDigits: decimals
    });
}

/**
 * Форматирование процента
 */
function formatPercent(num, decimals = 2) {
    if (num === null || num === undefined) return '-';
    const sign = num >= 0 ? '+' : '';
    return `${sign}${num.toFixed(decimals)}`;
}

/**
 * Форматирование даты
 */
function formatDate(dateString) {
    if (!dateString) return 'Нет данных';
    try {
        const date = new Date(dateString);
        return date.toLocaleDateString('ru-RU');
    } catch (e) {
        return dateString;
    }
}

/**
 * Экспорт данных в CSV
 */
function exportToCSV(data, filename = 'etf_data.csv') {
    const headers = [
        'Тикер', 'Название', 'УК', 'Класс активов',
        'TER %', 'СЧА млн ₽', 'Изм. 6М %', 'Изм. 2024 %'
    ];

    const rows = data.map(etf => [
        etf.ticker,
        `"${etf.fundName}"`,
        `"${etf.managementCo}"`,
        `"${etf.assetClass}"`,
        etf.terPercent || '',
        etf.navMillionRub || '',
        etf.priceChange6M || '',
        etf.priceChange2024 || ''
    ]);

    const csv = [headers, ...rows]
        .map(row => row.join(','))
        .join('\n');

    // Добавляем BOM для корректного отображения кириллицы в Excel
    const blob = new Blob(['\ufeff' + csv], {
        type: 'text/csv;charset=utf-8;'
    });

    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = `${filename}_${new Date().toISOString().split('T')[0]}.csv`;
    link.click();
}

/**
 * Вычисление статистики для отфильтрованных данных
 */
function calculateStats(data) {
    if (!data || data.length === 0) {
        return {
            totalFunds: 0,
            totalNAV: 0,
            avgTER: 0
        };
    }

    const totalNAV = data.reduce((sum, etf) => {
        return sum + (etf.navMillionRub || 0);
    }, 0);

    const terValues = data.filter(etf =>
        etf.terPercent !== null && etf.terPercent !== undefined
    );

    const avgTER = terValues.length > 0
        ? terValues.reduce((sum, etf) => sum + etf.terPercent, 0) / terValues.length
        : 0;

    return {
        totalFunds: data.length,
        totalNAV: totalNAV,
        avgTER: avgTER
    };
}

/**
 * Debounce функция для оптимизации поиска
 */
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

/**
 * Получить цвет для изменения цены
 */
function getPriceChangeColor(value) {
    if (value === null || value === undefined) return 'text-slate-600';
    return value >= 0 ? 'text-green-600' : 'text-red-600';
}

/**
 * Получить класс для бейджа статуса торговли
 */
function getTradeStatusClass(status) {
    return status === 'Торгуется'
        ? 'bg-green-100 text-green-800'
        : 'bg-red-100 text-red-800';
}

/**
 * Сокращение длинного текста
 */
function truncate(text, maxLength) {
    if (!text || text.length <= maxLength) return text;
    return text.substring(0, maxLength - 3) + '...';
}