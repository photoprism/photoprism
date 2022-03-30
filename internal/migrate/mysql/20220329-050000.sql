UPDATE photos SET photo_description = SUBSTR(photo_description, 0, 4096) WHERE 1;
ALTER TABLE photos MODIFY photo_description VARCHAR(4096);
