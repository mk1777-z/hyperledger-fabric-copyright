const fs = require('fs');
const path = require('path');
const https = require('https');

// 创建目录
const mkdirSync = (dirPath) => {
  if (!fs.existsSync(dirPath)) {
    fs.mkdirSync(dirPath, { recursive: true });
    console.log(`创建目录: ${dirPath}`);
  }
};

// 下载文件
const downloadFile = (url, destPath) => {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(destPath);
    
    https.get(url, (response) => {
      response.pipe(file);
      
      file.on('finish', () => {
        file.close();
        console.log(`下载完成: ${destPath}`);
        resolve();
      });
    }).on('error', (err) => {
      fs.unlink(destPath, () => {}); // 删除部分下载的文件
      console.error(`下载失败: ${url}`, err.message);
      reject(err);
    });
  });
};

// 主函数
async function main() {
  const BASE_DIR = path.join(__dirname, 'lib', 'echarts');
  const MAP_DIR = path.join(BASE_DIR, 'map');
  
  // 创建目录
  mkdirSync(BASE_DIR);
  mkdirSync(MAP_DIR);
  
  // 下载echarts主库
  await downloadFile(
    'https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js',
    path.join(BASE_DIR, 'echarts.min.js')
  );
  
  // 创建简化的中国地图文件
  const chinaMapContent = `(function(root, factory) {
    if (typeof define === 'function' && define.amd) {
        define(['exports', 'echarts'], factory);
    } else if (typeof exports === 'object' && typeof exports.nodeName !== 'string') {
        factory(exports, require('echarts'));
    } else {
        factory({}, root.echarts);
    }
}(this, function(exports, echarts) {
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
    
    echarts.registerMap('china', {
        type: 'FeatureCollection',
        features: features
    });
}));`;

  // 保存中国地图文件
  fs.writeFileSync(path.join(MAP_DIR, 'china.js'), chinaMapContent);
  console.log(`创建中国地图文件: ${path.join(MAP_DIR, 'china.js')}`);
  
  console.log('所有文件下载完成！');
}

main().catch(console.error);
