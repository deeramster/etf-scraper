// Компонент карточек статистики

const StatsCards = ({ globalStats, displayStats }) => {
    return React.createElement('div', { className: 'grid grid-cols-1 md:grid-cols-4 gap-4 mb-6' },
        // Карточка: Количество фондов
        React.createElement('div', {
                className: 'bg-white rounded-xl shadow-md p-6 border-l-4 border-blue-500'
            },
            React.createElement('div', { className: 'flex items-center justify-between' },
                React.createElement('div', {},
                    React.createElement('p', {
                        className: 'text-slate-600 text-sm mb-1'
                    }, 'Фондов в выборке'),
                    React.createElement('p', {
                        className: 'text-3xl font-bold text-slate-800'
                    }, displayStats.totalFunds),
                    React.createElement('p', {
                        className: 'text-xs text-slate-400 mt-1'
                    }, `из ${globalStats.totalFunds} всего`)
                ),
                React.createElement(DollarSign, {
                    className: 'w-12 h-12 text-blue-500 opacity-20'
                })
            )
        ),

        // Карточка: СЧА
        React.createElement('div', {
                className: 'bg-white rounded-xl shadow-md p-6 border-l-4 border-green-500'
            },
            React.createElement('div', { className: 'flex items-center justify-between' },
                React.createElement('div', {},
                    React.createElement('p', {
                        className: 'text-slate-600 text-sm mb-1'
                    }, 'СЧА в выборке'),
                    React.createElement('p', {
                        className: 'text-2xl font-bold text-slate-800'
                    }, `${(displayStats.totalNAV / 1000).toFixed(1)}B ₽`),
                    React.createElement('p', {
                        className: 'text-xs text-slate-400 mt-1'
                    }, `из ${(globalStats.totalNAV / 1000).toFixed(1)}B ₽`)
                ),
                React.createElement(TrendingUp, {
                    className: 'w-12 h-12 text-green-500 opacity-20'
                })
            )
        ),

        // Карточка: Средний TER
        React.createElement('div', {
                className: 'bg-white rounded-xl shadow-md p-6 border-l-4 border-purple-500'
            },
            React.createElement('div', { className: 'flex items-center justify-between' },
                React.createElement('div', {},
                    React.createElement('p', {
                        className: 'text-slate-600 text-sm mb-1'
                    }, 'Средний TER'),
                    React.createElement('p', {
                        className: 'text-3xl font-bold text-slate-800'
                    }, `${displayStats.avgTER.toFixed(2)}%`),
                    React.createElement('p', {
                        className: 'text-xs text-slate-400 mt-1'
                    }, 'в выборке')
                ),
                React.createElement(Filter, {
                    className: 'w-12 h-12 text-purple-500 opacity-20'
                })
            )
        ),

        // Карточка: Дата обновления
        React.createElement('div', {
                className: 'bg-white rounded-xl shadow-md p-6 border-l-4 border-orange-500'
            },
            React.createElement('div', { className: 'flex items-center justify-between' },
                React.createElement('div', {},
                    React.createElement('p', {
                        className: 'text-slate-600 text-sm mb-1'
                    }, 'Обновлено'),
                    React.createElement('p', {
                        className: 'text-lg font-bold text-slate-800'
                    }, globalStats.lastUpdate)
                ),
                React.createElement(Calendar, {
                    className: 'w-12 h-12 text-orange-500 opacity-20'
                })
            )
        )
    );
};