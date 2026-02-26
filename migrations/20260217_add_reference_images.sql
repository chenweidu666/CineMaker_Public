-- 添加参考图片字段
-- 为 characters 表添加 reference_images 列
ALTER TABLE characters ADD COLUMN reference_images TEXT;

-- 为 scenes 表添加 reference_images 列
ALTER TABLE scenes ADD COLUMN reference_images TEXT;

-- 为 props 表添加 reference_images 列
ALTER TABLE props ADD COLUMN reference_images TEXT;
