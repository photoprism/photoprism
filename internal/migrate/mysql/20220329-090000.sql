ALTER TABLE files MODIFY IF EXISTS time_index VARBINARY(64) AFTER photo_taken_at;
ALTER TABLE files ADD IF NOT EXISTS time_index VARBINARY(64) AFTER photo_taken_at;