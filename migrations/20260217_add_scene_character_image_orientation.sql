-- 为 scenes 和 characters 表添加 image_orientation 字段
ALTER TABLE scenes ADD COLUMN image_orientation VARCHAR(20) DEFAULT 'horizontal';
ALTER TABLE characters ADD COLUMN image_orientation VARCHAR(20) DEFAULT 'horizontal';
