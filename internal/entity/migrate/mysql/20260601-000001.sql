DROP INDEX IF EXISTS idx_albums_album_path ON albums;
ALTER TABLE albums MODIFY album_path VARBINARY(1024);
CREATE OR REPLACE INDEX idx_albums_album_path ON albums (album_path(512));
