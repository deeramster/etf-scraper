const { useState, useEffect } = React;

const ETFDashboard = () => {
    // Состояние данных
    const [etfData, setEtfData] = useState([]);
    const [filteredData, setFilteredData] = useState([]);

    // Состояние фильтров
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedAssetClass, setSelectedAssetClass] = useState('Все');
    const [assetClasses, setAssetClasses] = useState(['Все']);
    const [sortBy, setSortBy] = useState('nav_million_rub');
    const [sortOrder, setSortOrder] = useState('desc');

    // Состояние загрузки и ошибок
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    // Статистика
    const [globalStats, setGlobalStats] = useState({
        totalFunds: 0,
        totalNAV: 0,
        avgTER: 0,
        lastUpdate: ''
    });
    const [displayStats, setDisplayStats] = useState({
        totalFunds: 0,
        totalNAV: 0,
        avgTER: 0
    });

    // Загрузка данных при монтировании
    useEffect(() => {
        loadData();
        loadAssetClasses();
    }, []);

    // Пересчет статистики при изменении отфильтрованных данных
    useEffect(() => {
        const stats = calculateStats(filteredData);
        setDisplayStats(stats);
    }, [filteredData]);

    // Перезагрузка данных при изменении сортировки
    useEffect(() => {
        if (!loading) {
            loadData();
        }
    }, [sortBy, sortOrder]);

    // Применение фильтра по классу активов
    useEffect(() => {
        setSearchTerm('');
        applyFilters();
    }, [selectedAssetClass, etfData]);

    // Загрузка данных ETF
    const loadData = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await apiService.getETFs({
                sortBy,
                order: sortOrder.toUpperCase()
            });
            setEtfData(data || []);
            applyFilters(data || []);
            await loadStats();
        } catch (err) {
            setError(err.message);
            console.error('Ошибка загрузки данных:', err);
        } finally {
            setLoading(false);
        }
    };

    // Загрузка статистики
    const loadStats = async () => {
        try {
            const data = await apiService.getStats();
            setGlobalStats({
                totalFunds: data.uniqueTickers || 0,
                totalNAV: data.totalNAV || 0,
                avgTER: data.avgTER || 0,
                lastUpdate: formatDate(data.lastUpdate)
            });
        } catch (err) {
            console.error('Ошибка загрузки статистики:', err);
        }
    };

    // Загрузка классов активов
    const loadAssetClasses = async () => {
        try {
            const data = await apiService.getAssetClasses();
            setAssetClasses(data.assetClasses || ['Все']);
        } catch (err) {
            console.error('Ошибка загрузки классов активов:', err);
        }
    };

    // Применение фильтров
    const applyFilters = (data = etfData) => {
        if (selectedAssetClass === 'Все') {
            setFilteredData(data);
        } else {
            const filtered = data.filter(etf => etf.assetClass === selectedAssetClass);
            setFilteredData(filtered);
        }
    };

    // Обработка поиска
    const handleSearch = async (term) => {
        setSearchTerm(term);
        if (!term.trim()) {
            applyFilters();
            return;
        }

        try {
            const data = await apiService.search(term);
            setFilteredData(data || []);
        } catch (err) {
            console.error('Ошибка поиска:', err);
        }
    };

    // Использование debounce для поиска
    const debouncedSearch = debounce(handleSearch, 300);

    // Экспорт в CSV
    const handleExport = () => {
        exportToCSV(filteredData, 'etf_data');
    };

    // Экран загрузки
    if (loading) {
        return React.createElement('div', {
                className: 'min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 flex items-center justify-center'
            },
            React.createElement('div', { className: 'text-center' },
                React.createElement(RefreshCw, {
                    className: 'w-12 h-12 text-blue-600 animate-spin mx-auto mb-4'
                }),
                React.createElement('p', {
                    className: 'text-slate-600 text-lg'
                }, 'Загрузка данных...')
            )
        );
    }

    // Экран ошибки
    if (error) {
        return React.createElement('div', {
                className: 'min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 flex items-center justify-center p-6'
            },
            React.createElement('div', {
                    className: 'bg-white rounded-xl shadow-lg p-8 max-w-md w-full'
                },
                React.createElement(AlertCircle, {
                    className: 'w-16 h-16 text-red-500 mx-auto mb-4'
                }),
                React.createElement('h2', {
                    className: 'text-2xl font-bold text-slate-800 text-center mb-2'
                }, 'Ошибка подключения'),
                React.createElement('p', {
                    className: 'text-slate-600 text-center mb-6'
                }, error),
                React.createElement('button', {
                        onClick: loadData,
                        className: 'w-full flex items-center justify-center gap-2 px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition'
                    },
                    React.createElement(RefreshCw, { className: 'w-5 h-5' }),
                    'Попробовать снова'
                )
            )
        );
    }

    // Основной интерфейс
    return React.createElement('div', {
            className: 'min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 p-6'
        },
        React.createElement('div', { className: 'max-w-7xl mx-auto' },
            // Заголовок
            React.createElement('div', { className: 'mb-8' },
                React.createElement('h1', {
                    className: 'text-4xl font-bold text-slate-800 mb-2'
                }, 'ETF Dashboard'),
                React.createElement('p', {
                    className: 'text-slate-600'
                }, 'Мониторинг биржевых фондов на Мосбирже')
            ),

            // Карточки статистики
            React.createElement(StatsCards, {
                globalStats,
                displayStats
            }),

            // Панель фильтров (БЕЗ обработчика onScrape)
            React.createElement(FilterPanel, {
                searchTerm,
                onSearchChange: (term) => {
                    setSearchTerm(term);
                    debouncedSearch(term);
                },
                assetClasses,
                selectedAssetClass,
                onAssetClassChange: setSelectedAssetClass,
                sortBy,
                onSortByChange: setSortBy,
                sortOrder,
                onSortOrderToggle: () => setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc'),
                onRefresh: loadData,
                onExport: handleExport
            }),

            // Таблица ETF
            React.createElement(ETFTable, { data: filteredData }),

            // Футер
            React.createElement('div', {
                    className: 'mt-6 text-center text-slate-500 text-sm'
                },
                React.createElement('p', {},
                    `Последнее обновление: ${globalStats.lastUpdate}`
                ),
                React.createElement('p', { className: 'mt-1' },
                    `Найдено записей: ${filteredData.length}`
                )
            )
        )
    );
};

// Рендер приложения
const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(React.createElement(ETFDashboard));