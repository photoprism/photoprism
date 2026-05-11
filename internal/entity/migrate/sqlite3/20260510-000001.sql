ALTER TABLE photos ADD COLUMN photo_rating SMALLINT DEFAULT 0;
ALTER TABLE photos ADD COLUMN rating_src VARBINARY(8);
UPDATE photos SET photo_rating = 0 WHERE photo_rating IS NULL;
