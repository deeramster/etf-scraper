// Компонент таблицы ETF данных

const ETFTable = ({ data }) => {
    if (!data || data.length === 0) {
        return React.createElement('div', {
                className: 'bg-white rounded-xl shadow-md overflow-hidden'
            },
            React.createElement('div', { className: 'text-center py-12' },
                React.createElement('p', {
                    className: 'text-slate-500 text-lg'
                }, 'Нет данных для отображения'),
                React.createElement('p', {
                    className: 'text-slate-400 text-sm mt-2'
                }, 'Попробуйте изменить параметры фильтрации')
            )
        );
    }

    return React.createElement('div', {
            className: 'bg-white rounded-xl shadow-md overflow-hidden'
        },
        React.createElement('div', { className: 'overflow-x-auto' },
            React.createElement('table', { className: 'w-full' },
                // Заголовок таблицы
                React.createElement('thead', {
                        className: 'bg-slate-50 border-b border-slate-200'
                    },
                    React.createElement('tr', {},
                        React.createElement('th', {
                            className: 'px-6 py-4 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider'
                        }, 'Тикер'),
                        React.createElement('th', {
                            className: 'px-6 py-4 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider'
                        }, 'Название'),
                        React.createElement('th', {
                            className: 'px-6 py-4 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider'
                        }, 'УК'),
                        React.createElement('th', {
                            className: 'px-6 py-4 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider'
                        }, 'Класс активов'),
                        React.createElement('th', {
                            className: 'px-6 py-4 text-right text-xs font-semibold text-slate-600 uppercase tracking-wider'
                        }, 'TER %'),
                        React.createElement('th', {
                            className: 'px-6 py-4 text-right text-xs font-semibold text-slate-600 uppercase tracking-wider'
                        }, 'СЧА (млн ₽)'),
                        React.createElement('th', {
                            className: 'px-6 py-4 text-right text-xs font-semibold text-slate-600 uppercase tracking-wider'
                        }, '6М %'),
                        React.createElement('th', {
                            className: 'px-6 py-4 text-right text-xs font-semibold text-slate-600 uppercase tracking-wider'
                        }, '2024 %'),
                        React.createElement('th', {
                            className: 'px-6 py-4 text-center text-xs font-semibold text-slate-600 uppercase tracking-wider'
                        }, 'Статус')
                    )
                ),

                // Тело таблицы
                React.createElement('tbody', {
                        className: 'divide-y divide-slate-200'
                    },
                    data.map((etf, index) =>
                        React.createElement('tr', {
                                key: index,
                                className: 'hover:bg-slate-50 transition'
                            },
                            // Тикер
                            React.createElement('td', { className: 'px-6 py-4' },
                                React.createElement('span', {
                                    className: 'font-bold text-blue-600'
                                }, etf.ticker)
                            ),

                            // Название
                            React.createElement('td', {
                                className: 'px-6 py-4 text-slate-800 max-w-xs truncate',
                                title: etf.fundName
                            }, etf.fundName),

                            // УК
                            React.createElement('td', {
                                className: 'px-6 py-4 text-slate-600 text-sm'
                            }, etf.managementCo),

                            // Класс активов
                            React.createElement('td', {
                                className: 'px-6 py-4 text-slate-600 text-sm max-w-xs truncate',
                                title: etf.assetClass
                            }, etf.assetClass),

                            // TER
                            React.createElement('td', {
                                className: 'px-6 py-4 text-right font-medium text-slate-800'
                            }, etf.terPercent ? etf.terPercent.toFixed(2) : '-'),

                            // СЧА
                            React.createElement('td', {
                                className: 'px-6 py-4 text-right font-medium text-slate-800'
                            }, etf.navMillionRub ? formatNumber(etf.navMillionRub) : '-'),

                            // Изменение за 6М
                            React.createElement('td', { className: 'px-6 py-4 text-right' },
                                etf.priceChange6M ?
                                    React.createElement('span', {
                                        className: `font-medium ${getPriceChangeColor(etf.priceChange6M)}`
                                    }, formatPercent(etf.priceChange6M))
                                    : '-'
                            ),

                            // Изменение за 2024
                            React.createElement('td', { className: 'px-6 py-4 text-right' },
                                etf.priceChange2024 ?
                                    React.createElement('span', {
                                        className: `font-medium ${getPriceChangeColor(etf.priceChange2024)}`
                                    }, formatPercent(etf.priceChange2024))
                                    : '-'
                            ),

                            // Статус торговли
                            React.createElement('td', { className: 'px-6 py-4 text-center' },
                                React.createElement('span', {
                                    className: `inline-flex px-3 py-1 text-xs font-medium rounded-full ${getTradeStatusClass(etf.tradeStatus)}`
                                }, etf.tradeStatus)
                            )
                        )
                    )
                )
            )
        )
    );
};