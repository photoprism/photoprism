ALTER TABLE photos ADD COLUMN photo_rating TINYINT NOT NULL DEFAULT 0;
UPDATE photos SET photo_rating = 0 WHERE photo_rating IS NULL;
