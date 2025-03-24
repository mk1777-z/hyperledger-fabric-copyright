// 页面加载时初始化
window.onload = function() {
    const username = localStorage.getItem('username');
    document.getElementById("username").textContent = username || "未登录";
    const token = localStorage.getItem('token');
    if (!token) {
        alert('请登录');
        return (window.location.href = '/');
    }

    // 初始化layui组件
    layui.use(['form', 'laydate', 'table', 'element'], function() {
        const form = layui.form;
        const laydate = layui.laydate;
        const table = layui.table;
        const element = layui.element;

        // 初始化日期选择器
        laydate.render({
            elem: '#date-range',
            range: true
        });

        // 初始化表格
        initTables(table);

        // 渲染表单
        form.render();
    });

    // 初始化图表
    initCharts();

    // 初始化地图
    initMap();

    // 绑定筛选按钮事件
    document.getElementById('filter-btn').addEventListener('click', function() {
        applyFilters();
    });

    // 绑定导出按钮事件
    document.getElementById('export-excel').addEventListener('click', function() {
        exportData('excel');
    });

    document.getElementById('export-pdf').addEventListener('click', function() {
        exportData('pdf');
    });
};

// 退出功能
function logout() {
    localStorage.clear();
    window.location.href = '/';
}

// 初始化表格
function initTables(table) {
    // 交易表格
    table.render({
        elem: '#transaction-table',
        url: '/statistics/transactionData',
        method: 'post',
        headers: {
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        },
        where: { timeRange: '', category: '' },
        page: true,
        cols: [[
            {field: 'id', title: 'ID', width: 80, sort: true},
            {field: 'name', title: '版权名称', width: 120},
            {field: 'category', title: '分类', width: 100},
            {field: 'seller', title: '卖家', width: 120},
            {field: 'buyer', title: '买家', width: 120},
            {field: 'price', title: '价格', width: 100, sort: true},
            {field: 'time', title: '交易时间', width: 180, sort: true},
            {field: 'location', title: '交易地点', width: 150}
        ]],
        limit: 10,
        limits: [10, 20, 50, 100]
    });

    // 用户表格
    table.render({
        elem: '#user-table',
        url: '/statistics/userData',
        method: 'post',
        headers: {
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        },
        where: { timeRange: '' },
        page: true,
        cols: [[
            {field: 'username', title: '用户名', width: 120},
            {field: 'buyCount', title: '购买数', width: 100, sort: true},
            {field: 'sellCount', title: '出售数', width: 100, sort: true},
            {field: 'totalAmount', title: '交易总额', width: 120, sort: true},
            {field: 'lastActiveTime', title: '最后活跃时间', width: 180, sort: true}
        ]],
        limit: 10,
        limits: [10, 20, 50, 100]
    });

    // 收益表格
    table.render({
        elem: '#revenue-table',
        url: '/statistics/revenueData',
        method: 'post',
        headers: {
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        },
        where: { timeRange: '' },
        page: true,
        cols: [[
            {field: 'period', title: '时间段', width: 120},
            {field: 'totalRevenue', title: '总收益', width: 120, sort: true},
            {field: 'transactionCount', title: '交易数量', width: 120, sort: true},
            {field: 'avgPrice', title: '平均价格', width: 120, sort: true},
            {field: 'growth', title: '增长率', width: 120, sort: true, templet: function(d) {
                return d.growth + '%';
            }}
        ]],
        limit: 10,
        limits: [10, 20, 50, 100]
    });
}

// 初始化图表
function initCharts() {
    // 获取统计数据
    fetch('/statistics/chartData', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        },
        body: JSON.stringify({
            timeRange: '',
            category: ''
        })
    })
    .then(res => res.json())
    .then(data => {
        // 渲染分类统计图表
        renderCategoryChart(data.categoryData);
        
        // 渲染价格分布图表
        renderPriceChart(data.priceData);
        
        // 渲染交易趋势图表
        renderTrendChart(data.trendData);
        
        // 渲染用户活跃度图表
        renderActivityChart(data.activityData);
    })
    .catch(error => {
        console.error('获取统计数据失败:', error);
    });
}

// 渲染分类统计图表
function renderCategoryChart(data) {
    const categoryChart = echarts.init(document.getElementById('category-chart'));
    
    const option = {
        tooltip: {
            trigger: 'item',
            formatter: '{a} <br/>{b}: {c} ({d}%)'
        },
        legend: {
            orient: 'vertical',
            right: 10,
            top: 'center',
            data: data.map(item => item.name)
        },
        series: [
            {
                name: '版权分类',
                type: 'pie',
                radius: ['50%', '70%'],
                avoidLabelOverlap: false,
                itemStyle: {
                    borderRadius: 10,
                    borderColor: '#fff',
                    borderWidth: 2
                },
                label: {
                    show: false,
                    position: 'center'
                },
                emphasis: {
                    label: {
                        show: true,
                        fontSize: '16',
                        fontWeight: 'bold'
                    }
                },
                labelLine: {
                    show: false
                },
                data: data
            }
        ]
    };
    
    categoryChart.setOption(option);
    window.addEventListener('resize', function() {
        categoryChart.resize();
    });
}

// 渲染价格分布图表
function renderPriceChart(data) {
    const priceChart = echarts.init(document.getElementById('price-chart'));
    
    const option = {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'shadow'
            }
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true
        },
        xAxis: [
            {
                type: 'category',
                data: data.map(item => item.range),
                axisTick: {
                    alignWithLabel: true
                }
            }
        ],
        yAxis: [
            {
                type: 'value'
            }
        ],
        series: [
            {
                name: '版权数量',
                type: 'bar',
                barWidth: '60%',
                data: data.map(item => item.count),
                itemStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        { offset: 0, color: '#83bff6' },
                        { offset: 0.5, color: '#188df0' },
                        { offset: 1, color: '#006d77' }
                    ])
                }
            }
        ]
    };
    
    priceChart.setOption(option);
    window.addEventListener('resize', function() {
        priceChart.resize();
    });
}

// 渲染交易趋势图表
function renderTrendChart(data) {
    const trendChart = echarts.init(document.getElementById('trend-chart'));
    
    const option = {
        tooltip: {
            trigger: 'axis'
        },
        legend: {
            data: ['交易数量', '交易金额']
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true
        },
        xAxis: {
            type: 'category',
            boundaryGap: false,
            data: data.months
        },
        yAxis: [
            {
                type: 'value',
                name: '交易数量',
                position: 'left'
            },
            {
                type: 'value',
                name: '交易金额',
                position: 'right'
            }
        ],
        series: [
            {
                name: '交易数量',
                type: 'line',
                data: data.counts,
                smooth: true
            },
            {
                name: '交易金额',
                type: 'line',
                yAxisIndex: 1,
                data: data.amounts,
                smooth: true,
                lineStyle: {
                    width: 0
                },
                areaStyle: {
                    opacity: 0.5,
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        { offset: 0, color: 'rgba(0, 109, 119, 0.8)' },
                        { offset: 1, color: 'rgba(0, 109, 119, 0.1)' }
                    ])
                }
            }
        ]
    };
    
    trendChart.setOption(option);
    window.addEventListener('resize', function() {
        trendChart.resize();
    });
}

// 渲染用户活跃度图表
function renderActivityChart(data) {
    const activityChart = echarts.init(document.getElementById('activity-chart'));
    
    const option = {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'cross',
                label: {
                    backgroundColor: '#6a7985'
                }
            }
        },
        legend: {
            data: ['新用户', '活跃用户', '交易用户']
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true
        },
        xAxis: [
            {
                type: 'category',
                boundaryGap: false,
                data: data.dates
            }
        ],
        yAxis: [
            {
                type: 'value'
            }
        ],
        series: [
            {
                name: '新用户',
                type: 'line',
                stack: 'Total',
                areaStyle: {},
                emphasis: {
                    focus: 'series'
                },
                data: data.newUsers
            },
            {
                name: '活跃用户',
                type: 'line',
                stack: 'Total',
                areaStyle: {},
                emphasis: {
                    focus: 'series'
                },
                data: data.activeUsers
            },
            {
                name: '交易用户',
                type: 'line',
                stack: 'Total',
                areaStyle: {},
                emphasis: {
                    focus: 'series'
                },
                data: data.tradingUsers
            }
        ]
    };
    
    activityChart.setOption(option);
    window.addEventListener('resize', function() {
        activityChart.resize();
    });
}

// 初始化地图
function initMap() {
    // 检查百度地图API是否已加载
    if (typeof BMap === 'undefined') {
        console.error('百度地图API未正确加载，请检查API密钥');
        document.getElementById('map-container').innerHTML = '<div class="map-error">地图加载失败，请检查API密钥</div>';
        return;
    }

    // 在适当的时机初始化百度地图
    fetch('/statistics/locationData', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        }
    })
    .then(res => res.json())
    .then(data => {
        // 创建百度地图实例
        const map = new BMap.Map('map-container');
        const point = new BMap.Point(104.195397, 35.86166);
        map.centerAndZoom(point, 5);
        map.enableScrollWheelZoom();
        map.addControl(new BMap.NavigationControl());
        map.addControl(new BMap.ScaleControl());
        
        // 添加标记点
        data.forEach(item => {
            const marker = new BMap.Marker(new BMap.Point(item.lng, item.lat));
            map.addOverlay(marker);
            
            const content = `<div style="padding:10px;">
                <h4 style="margin:0;font-size:14px;">用户: ${item.username}</h4>
                <p style="margin:5px 0;">交易数: ${item.count}</p>
                <p style="margin:5px 0;">最近交易: ${item.lastTransaction}</p>
            </div>`;
            
            const infoWindow = new BMap.InfoWindow(content);
            marker.addEventListener('click', function() {
                this.openInfoWindow(infoWindow);
            });
        });
    })
    .catch(error => {
        console.error('获取地理位置数据失败:', error);
    });
}

// 应用筛选条件
function applyFilters() {
    const dateRange = document.getElementById('date-range').value;
    const category = document.getElementById('category-filter').value;
    
    // 重新加载表格数据
    layui.table.reload('transaction-table', {
        where: {
            timeRange: dateRange,
            category: category
        }
    });
    
    layui.table.reload('user-table', {
        where: {
            timeRange: dateRange
        }
    });
    
    layui.table.reload('revenue-table', {
        where: {
            timeRange: dateRange
        }
    });
    
    // 重新获取图表数据
    fetch('/statistics/chartData', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        },
        body: JSON.stringify({
            timeRange: dateRange,
            category: category
        })
    })
    .then(res => res.json())
    .then(data => {
        // 更新各图表
        renderCategoryChart(data.categoryData);
        renderPriceChart(data.priceData);
        renderTrendChart(data.trendData);
        renderActivityChart(data.activityData);
    })
    .catch(error => {
        console.error('获取统计数据失败:', error);
    });
}

// 导出数据
function exportData(type) {
    const dateRange = document.getElementById('date-range').value;
    const category = document.getElementById('category-filter').value;
    
    fetch('/statistics/exportData', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        },
        body: JSON.stringify({
            type: type,
            timeRange: dateRange,
            category: category
        })
    })
    .then(res => {
        if (type === 'excel') {
            return res.blob();
        } else {
            return res.blob();
        }
    })
    .then(blob => {
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = url;
        a.download = `版权统计报表.${type === 'excel' ? 'xlsx' : 'pdf'}`;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
    })
    .catch(error => {
        console.error('导出数据失败:', error);
        alert('导出失败，请稍后重试');
    });
}
