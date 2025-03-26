-- 插入一些默认分类
INSERT INTO categories (category_name, description) VALUES 
('音乐', '音乐作品版权'),
('图片', '图片和图像版权'),
('视频', '视频和影视作品版权'),
('文档', '文字和文档版权'),
('其他', '其他类型版权');

-- 随机更新item表的category
UPDATE item SET 
    category = ELT(FLOOR(1 + RAND() * 5), '音乐', '图片', '视频', '文档', '其他'),
    created_at = DATE_SUB(NOW(), INTERVAL FLOOR(RAND() * 365) DAY),
    updated_at = DATE_SUB(NOW(), INTERVAL FLOOR(RAND() * 30) DAY);

-- 随机更新user表的字段
UPDATE user SET 
    registration_time = DATE_SUB(NOW(), INTERVAL FLOOR(RAND() * 365) DAY),
    last_active_time = DATE_SUB(NOW(), INTERVAL FLOOR(RAND() * 30) DAY),
    user_level = FLOOR(RAND() * 3),
    location = ELT(FLOOR(1 + RAND() * 5), '北京', '上海', '广州', '深圳', '成都'),
    latitude = CASE 
        WHEN FLOOR(RAND() * 5) = 0 THEN 39.904
        WHEN FLOOR(RAND() * 5) = 1 THEN 31.230
        WHEN FLOOR(RAND() * 5) = 2 THEN 23.125
        WHEN FLOOR(RAND() * 5) = 3 THEN 22.547
        ELSE 30.659
    END,
    longitude = CASE 
        WHEN FLOOR(RAND() * 5) = 0 THEN 116.407
        WHEN FLOOR(RAND() * 5) = 1 THEN 121.473
        WHEN FLOOR(RAND() * 5) = 2 THEN 113.280
        WHEN FLOOR(RAND() * 5) = 3 THEN 114.085
        ELSE 104.066
    END;

-- 创建一些模拟交易记录
INSERT INTO transactions (item_id, seller_id, buyer_id, price, transaction_time)
SELECT 
    i.id,
    (SELECT id FROM user WHERE username = i.owner) as seller_id,
    (SELECT id FROM user ORDER BY RAND() LIMIT 1) as buyer_id,
    i.price,
    DATE_SUB(NOW(), INTERVAL FLOOR(RAND() * 180) DAY)
FROM item i
WHERE i.on_sale = 1
LIMIT 50;

-- 创建一些用户活动记录
INSERT INTO user_activities (user_id, activity_type, last_login)
SELECT 
    id,
    ELT(FLOOR(1 + RAND() * 3), '登录', '浏览', '交易'),
    DATE_SUB(NOW(), INTERVAL FLOOR(RAND() * 30) DAY)
FROM user;
