/**
 * Leaflet 高德地图插件
 * 将高德地图瓦片服务适配到Leaflet中使用
 */
(function() {
    // 确保Leaflet已加载
    if (typeof L === 'undefined') {
        console.error('Leaflet未加载，无法初始化高德地图插件');
        return;
    }
    
    // 扩展L.TileLayer以支持高德地图
    L.TileLayer.AMap = L.TileLayer.extend({
        options: {
            subdomains: '1234',
            attribution: 'Leaflet | © OpenStreetMap contributors © 高德地图'
        },

        initialize: function(url, options) {
            L.TileLayer.prototype.initialize.call(this, url, options);
        }
    });
    
    // 添加工厂方法
    L.tileLayer.amap = function(url, options) {
        return new L.TileLayer.AMap(url, options);
    };
    
    console.log('高德地图插件已加载');
})();
