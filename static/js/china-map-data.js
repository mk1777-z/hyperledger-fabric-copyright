// 简化中国地图数据
(function() {
    // 确保echarts已加载
    if (typeof echarts === 'undefined') {
        console.error('echarts未加载，无法注册地图数据');
        return;
    }
    
    console.log('开始加载本地中国地图数据');
    
    // 简化版的中国地图坐标数据 - 主要城市的坐标点
    var chinaGeoCoordMap = {
        '北京': [116.407, 39.904],
        '上海': [121.473, 31.230],
        '广州': [113.280, 23.125],
        '深圳': [114.085, 22.547],
        '成都': [104.066, 30.659],
        '武汉': [114.305, 30.593],
        '南京': [118.778, 32.057],
        '杭州': [120.155, 30.274],
        '重庆': [106.551, 29.563],
        '西安': [108.940, 34.341]
    };
    
    // 注册一个非常简化的中国地图数据
    try {
        var features = [];
        
        for (var city in chinaGeoCoordMap) {
            features.push({
                type: 'Feature',
                properties: {
                    name: city
                },
                geometry: {
                    type: 'Point',
                    coordinates: chinaGeoCoordMap[city]
                }
            });
        }
        
        // 注册一个简单的点地图
        echarts.registerMap('china', {
            type: 'FeatureCollection',
            features: features
        });
        
        console.log('成功注册简化版中国地图');
    } catch(e) {
        console.error('注册中国地图数据失败:', e);
    }
})();
