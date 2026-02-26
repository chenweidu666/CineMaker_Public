-- 添加 generate_audio 字段到 video_generations 表
ALTER TABLE video_generations ADD COLUMN generate_audio BOOLEAN DEFAULT TRUE;
