CREATE OR REPLACE INDEX idx_albums_album_filter ON albums (album_filter(512));
CREATE OR REPLACE INDEX idx_albums_album_path ON albums (album_path(512));
