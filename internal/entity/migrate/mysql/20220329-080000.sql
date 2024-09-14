ALTER TABLE files MODIFY IF EXISTS media_id VARBINARY(32) AFTER photo_taken_at;
ALTER TABLE files ADD IF NOT EXISTS media_id VARBINARY(32) AFTER photo_taken_at;