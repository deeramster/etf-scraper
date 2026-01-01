const FilterPanel = ({
                         searchTerm,
                         onSearchChange,
                         assetClasses,
                         selectedAssetClass,
                         onAssetClassChange,
                         sortBy,
                         onSortByChange,
                         sortOrder,
                         onSortOrderToggle,
                         onRefresh,
                         onExport
                     }) => {
    return React.createElement('div', {
            className: 'bg-white rounded-xl shadow-md p-6 mb-6'
        },
        // Фильтры
        React.createElement('div', {
                className: 'grid grid-cols-1 md:grid-cols-4 gap-4'
            },
            // Поиск
            React.createElement('div', { className: 'md:col-span-2' },
                React.createElement('label', {
                    className: 'block text-sm font-medium text-slate-700 mb-2'
                }, 'Поиск'),
                React.createElement('div', { className: 'relative' },
                    React.createElement(Search, {
                        className: 'absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400 w-5 h-5'
                    }),
                    React.createElement('input', {
                        type: 'text',
                        placeholder: 'Тикер, название или УК...',
                        value: searchTerm,
                        onChange: (e) => onSearchChange(e.target.value),
                        className: 'w-full pl-10 pr-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent'
                    })
                )
            ),

            // Класс активов
            React.createElement('div', {},
                React.createElement('label', {
                    className: 'block text-sm font-medium text-slate-700 mb-2'
                }, 'Класс активов'),
                React.createElement('select', {
                        value: selectedAssetClass,
                        onChange: (e) => onAssetClassChange(e.target.value),
                        className: 'w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent'
                    },
                    assetClasses.map(ac =>
                        React.createElement('option', { key: ac, value: ac }, ac)
                    )
                )
            ),

            // Сортировка
            React.createElement('div', {},
                React.createElement('label', {
                    className: 'block text-sm font-medium text-slate-700 mb-2'
                }, 'Сортировка'),
                React.createElement('div', { className: 'flex gap-2' },
                    React.createElement('select', {
                            value: sortBy,
                            onChange: (e) => onSortByChange(e.target.value),
                            className: 'flex-1 px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent'
                        },
                        React.createElement('option', { value: 'nav_million_rub' }, 'По СЧА'),
                        React.createElement('option', { value: 'ter_percent' }, 'По TER'),
                        React.createElement('option', { value: 'price_change_2024' }, 'По изм. 2024'),
                        React.createElement('option', { value: 'ticker' }, 'По тикеру')
                    ),
                    React.createElement('button', {
                        onClick: onSortOrderToggle,
                        className: 'px-3 py-2 bg-slate-100 border border-slate-300 rounded-lg hover:bg-slate-200 transition',
                        title: sortOrder === 'asc' ? 'По возрастанию' : 'По убыванию'
                    }, sortOrder === 'asc' ? '↑' : '↓')
                )
            )
        ),

        // Кнопки действий (БЕЗ кнопки скрейпинга)
        React.createElement('div', { className: 'flex gap-2 mt-4' },
            React.createElement('button', {
                    onClick: onRefresh,
                    className: 'flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition'
                },
                React.createElement(RefreshCw, { className: 'w-4 h-4' }),
                'Обновить данные'
            ),
            React.createElement('button', {
                    onClick: onExport,
                    className: 'flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition'
                },
                React.createElement(Download, { className: 'w-4 h-4' }),
                'Экспорт в CSV'
            )
        )
    );
};