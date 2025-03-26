// 页面加载时初始化
window.onload = function() {
    console.log("统计页面开始加载...");
    const username = localStorage.getItem('username');
    document.getElementById("username").textContent = username || "未登录";
    const token = localStorage.getItem('token');
    if (!token) {
        alert('请登录');
        return (window.location.href = '/');
    }

    console.log("准备初始化组件...");
    // 初始化layui组件
    layui.use(['form', 'laydate', 'table', 'element'], function() {
        console.log("Layui组件加载完成");
        const table = layui.table;
        const laydate = layui.laydate;
        
        // 初始化日期选择器 - 确保其正确渲染
        laydate.render({
            elem: '#date-range',
            type: 'date',
            range: true,
            value: getDefaultDateRange(),
            // 添加触发方式，确保点击时显示
            trigger: 'click',
            // 添加回调确认选择有效
            done: function(value){
                console.log('日期选择完成:', value);
                // 当选择日期后，可选择自动触发筛选
                // setTimeout(() => document.getElementById('filter-btn').click(), 100);
            }
        });
        
        // 初始化表格 - 不依赖于筛选按钮
        initTables(table);
    });

    // 初始化图表（这里也直接调用，不只依赖DOM事件中的调用）
    console.log("准备初始化图表...");
    initCharts();

    // 初始化地图
    console.log("准备初始化地图...");
    initMap();

    // 绑定筛选按钮事件
    document.getElementById('filter-btn').addEventListener('click', function() {
        console.log("筛选按钮被点击");
        // 显示筛选中的状态
        const loadingIndex = layui.layer.msg('数据加载中...', {
            icon: 16,
            time: 0,
            shade: 0.1
        });
        
        // 执行筛选
        applyFilters(loadingIndex);
    });

    // 绑定导出按钮事件
    document.getElementById('export-excel').addEventListener('click', function() {
        exportData('excel');
    });

    document.getElementById('export-pdf').addEventListener('click', function() {
        exportData('pdf');
    });
};

// 获取默认日期范围（过去30天）
function getDefaultDateRange() {
    const endDate = new Date();
    const startDate = new Date();
    startDate.setDate(startDate.getDate() - 30);
    
    // 格式化日期
    const formatDate = (date) => {
        const year = date.getFullYear();
        const month = (date.getMonth() + 1).toString().padStart(2, '0');
        const day = date.getDate().toString().padStart(2, '0');
        return `${year}-${month}-${day}`;
    };
    
    return formatDate(startDate) + ' - ' + formatDate(endDate);
}

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
    console.log("开始初始化图表...");
    // 获取统计数据
    const dateRange = document.getElementById('date-range').value || '';
    const category = document.getElementById('category-filter').value || '';
    
    console.log("发送图表数据请求，参数:", {timeRange: dateRange, category: category});
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
    .then(res => {
        console.log("图表数据请求状态:", res.status);
        if (!res.ok) {
            throw new Error(`服务器返回错误: ${res.status}`);
        }
        return res.json();
    })
    .then(data => {
        console.log("获取到图表数据:", data);
        // 确保数据存在，如果后端没有返回数据则使用空数组/对象
        const categoryData = data.categoryData || [];
        const priceData = data.priceData || [];
        const trendData = data.trendData || {months:[], counts:[], amounts:[]};
        const activityData = data.activityData || {dates:[], newUsers:[], activeUsers:[], tradingUsers:[]};
        
        // 渲染各种图表
        renderCategoryChart(categoryData);
        renderPriceChart(priceData);
        renderTrendChart(trendData);
        renderActivityChart(activityData);
    })
    .catch(error => {
        console.error('获取统计数据失败:', error);
        // 显示友好的错误提示
        layui.layer.msg('初始化图表失败: ' + error.message, {icon: 2});
        
        // 使用空数据渲染图表
        renderCategoryChart([]);
        renderPriceChart([]);
        renderTrendChart({months:[], counts:[], amounts:[]});
        renderActivityChart({dates:[], newUsers:[], activeUsers:[], tradingUsers:[]});
    });
}

// 渲染分类统计图表
function renderCategoryChart(data) {
    const categoryChart = echarts.init(document.getElementById('category-chart'));
    
    // 处理空数据情况
    if (!data || data.length === 0) {
        categoryChart.setOption({
            title: {
                text: '暂无数据',
                left: 'center',
                top: 'center'
            }
        });
        return;
    }
    
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
    
    // 处理空数据情况
    if (!data || data.length === 0) {
        priceChart.setOption({
            title: {
                text: '暂无数据',
                left: 'center',
                top: 'center'
            }
        });
        return;
    }
    
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
                left: 'center',
                top: 'center'
            }
        ]
    };
    
    priceChart.setOption(option);
    window.addEventListener('resize', function() {
        priceChart.resize();
    });
}

// 渲染趋势图表
function renderTrendChart(data) {
    const trendChart = echarts.init(document.getElementById('trend-chart'));
    
    // 处理空数据情况
    if (!data || data.months.length === 0) {
        trendChart.setOption({
            title: {
                text: '暂无数据',
                left: 'center',
                top: 'center'
            }
        });
        return;
    }
    
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
    
    // 处理空数据情况
    if (!data || data.dates.length === 0) {
        activityChart.setOption({
            title: {
                text: '暂无数据',
                left: 'center',
                top: 'center'
            }
        });
        return;
    }
    
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
    console.log("初始化地图...");
    // 显示地图加载中状态
    document.getElementById('map-container').innerHTML = '<div style="text-align:center;line-height:400px;">地图加载中...</div>';
    
    // 加载Leaflet库
    loadScript('https://unpkg.com/leaflet@1.9.4/dist/leaflet.js', function() {
        if (typeof L === 'undefined') {
            console.error("Leaflet库加载失败");
            document.getElementById('map-container').innerHTML = '<div class="map-error">地图加载失败，请刷新页面重试</div>';
            return;
        }
        
        // 加载高德地图插件
        loadScript('/static/js/leaflet-amap.js', function() {
            // 获取交易地理位置数据
            fetchLocationData();
        });
    });
}

// 获取地理位置数据
function fetchLocationData() {
    fetch('/statistics/locationData', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('token')
        }
    })
    .then(res => {
        if (!res.ok) {
            throw new Error('服务器返回错误: ' + res.status);
        }
        return res.json();
    })
    .then(data => {
        renderMap(data);
    })
    .catch(error => {
        console.error('获取地理位置数据失败:', error);
        document.getElementById('map-container').innerHTML = 
            `<div class="map-error">加载地图数据失败: ${error.message}</div>`;
    });
}

// 渲染地图
function renderMap(data) {
    try {
        // 初始化地图
        const map = L.map('map-container').setView([35.86166, 104.195397], 4);
        
        // 使用高德地图图层
        L.tileLayer.amap('https://webrd0{s}.is.autonavi.com/appmaptile?lang=zh_cn&size=1&scale=1&style=8&x={x}&y={y}&z={z}', {
            subdomains: '1234',
            attribution: 'Leaflet | © OpenStreetMap contributors © 高德地图'
        }).addTo(map);
        
        // 添加标记点
        if (data && data.length > 0) {
            data.forEach(item => {
                // 创建一个标记
                const marker = L.marker([item.lat, item.lng]).addTo(map);
                
                // 添加弹出信息
                marker.bindPopup(`
                    <div style="padding:5px;">
                        <h4 style="margin:0;font-size:14px;">用户: ${item.username}</h4>
                        <p style="margin:5px 0;">交易数: ${item.count}</p>
                        <p style="margin:5px 0;">最近交易: ${item.lastTransaction}</p>
                    </div>
                `);
            });
            
            // 如果数据点多于3个，尝试添加热力图
            if (data.length > 3) {
                // 加载热力图插件
                loadScript('https://unpkg.com/leaflet.heat@0.2.0/dist/leaflet-heat.js', function() {
                    if (typeof L.heatLayer !== 'undefined') {
                        // 准备热力图数据
                        const heatData = data.map(item => [
                            item.lat, 
                            item.lng, 
                            item.count / 2 // 强度比例
                        ]);
                        
                        // 创建热力图层
                        L.heatLayer(heatData, {
                            radius: 25,
                            blur: 15,
                            maxZoom: 10,
                            gradient: {0.4: 'blue', 0.6: 'lime', 0.8: 'yellow', 1: 'red'}
                        }).addTo(map);
                    }
                });
            }
        } else {
            // 没有数据时显示提示
            const noDataMarker = L.marker([35.86166, 104.195397]).addTo(map);
            noDataMarker.bindPopup('<b>暂无交易地理数据</b>').openPopup();
        }
        
        // 确保地图正确调整大小
        setTimeout(() => {
            map.invalidateSize();
        }, 100);
    } catch (e) {
        console.error("渲染地图时出错:", e);
        document.getElementById('map-container').innerHTML = 
            `<div class="map-error">地图渲染失败: ${e.message}</div>`;
    }
}

// 应用筛选条件
function applyFilters(loadingIndex) {
    const dateRange = document.getElementById('date-range').value;
    const category = document.getElementById('category-filter').value;
    
    console.log("应用筛选，参数:", {timeRange: dateRange, category: category});
    
    // 如果没有传入loadingIndex，则创建一个新的
    if (!loadingIndex) {
        loadingIndex = layui.layer.load(2);
    }
    
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
    .then(res => {
        if (!res.ok) {
            throw new Error('服务器返回错误: ' + res.status);
        }
        return res.json();
    })
    .then(data => {
        // 渲染各图表
        renderCategoryChart(data.categoryData || []);
        renderPriceChart(data.priceData || []);
        renderTrendChart(data.trendData || {months:[], counts:[], amounts:[]});
        renderActivityChart(data.activityData || {dates:[], newUsers:[], activeUsers:[], tradingUsers:[]});
        
        // 关闭加载层
        layui.layer.close(loadingIndex);
        layui.layer.msg('数据加载成功', {icon: 1, time: 1000});
    })
    .catch(error => {
        // 关闭加载层
        layui.layer.close(loadingIndex);
        console.error('获取统计数据失败:', error);
        layui.layer.msg('获取数据失败，请稍后重试: ' + error.message, {icon: 2});
        
        // 使用默认空数据渲染图表，避免界面空白
        renderCategoryChart([]);
        renderPriceChart([]);
        renderTrendChart({months:[], counts:[], amounts:[]});
        renderActivityChart({dates:[], newUsers:[], activeUsers:[], tradingUsers:[]});
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
