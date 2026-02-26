-- Add image_orientation field to props table
ALTER TABLE props ADD COLUMN image_orientation VARCHAR(20) DEFAULT 'horizontal';
