ALTER TABLE IF EXISTS photos RENAME COLUMN photo_description TO photo_caption;
ALTER TABLE IF EXISTS photos RENAME COLUMN description_src TO caption_src;