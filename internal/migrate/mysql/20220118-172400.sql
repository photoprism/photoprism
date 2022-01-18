ALTER TABLE albums MODIFY album_filter VARBINARY(767) DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_albums_album_filter ON albums (album_filter);