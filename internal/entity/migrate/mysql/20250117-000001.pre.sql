ALTER TABLE photos CHANGE COLUMN IF EXISTS photo_description photo_caption VARCHAR(4096);
ALTER TABLE photos CHANGE COLUMN IF EXISTS description_src caption_src VARBINARY(8);