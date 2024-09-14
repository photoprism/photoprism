DROP INDEX IF EXISTS idx_albums_album_filter ON albums;
ALTER TABLE albums MODIFY album_filter VARBINARY(2048) DEFAULT '';
CREATE OR REPLACE INDEX idx_albums_album_filter ON albums (album_filter(512));