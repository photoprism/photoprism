DELETE FROM auth_users_details WHERE user_uid NOT IN (SELECT user_uid FROM auth_users);
DELETE FROM auth_users_settings WHERE user_uid NOT IN (SELECT user_uid FROM auth_users);
DELETE FROM auth_users_shares WHERE user_uid NOT IN (SELECT user_uid FROM auth_users);
DELETE FROM categories WHERE label_id NOT IN (SELECT id FROM labels) OR category_id NOT IN (SELECT id FROM labels);
UPDATE cells SET place_id = 'zz' WHERE place_id NOT IN (SELECT id FROM places);
UPDATE countries SET country_photo_id = NULL WHERE country_photo_id NOT IN (SELECT id FROM photos) AND country_photo_id IS NOT NULL;
DELETE FROM details WHERE photo_id NOT IN (SELECT id FROM photos);
UPDATE files SET photo_id = NULL WHERE photo_id NOT IN (SELECT id FROM photos) AND photo_id IS NOT NULL;
UPDATE files, photos SET files.photo_id = photos.id WHERE files.photo_uid = photos.photo_uid AND files.photo_id IS NULL AND files.photo_uid IS NOT NULL;
UPDATE files SET photo_uid = NULL WHERE photo_id IS NULL AND photo_uid IS NOT NULL;
DELETE FROM files_share WHERE file_id NOT IN (SELECT id FROM files) OR service_id NOT IN (SELECT id FROM services);
UPDATE files_sync SET file_id = NULL WHERE file_id NOT IN (SELECT id FROM files);
DELETE FROM files_sync WHERE service_id NOT IN (SELECT id FROM services);
UPDATE photos, cameras SET photos.camera_id = cameras.id WHERE cameras.camera_slug = 'zz' AND photos.camera_id NOT IN (SELECT id FROM cameras) AND photos.camera_id IS NOT NULL;
UPDATE photos SET cell_id = 'zz' WHERE cell_id NOT IN (SELECT id FROM cells) AND cell_id IS NOT NULL;
UPDATE photos, lenses SET photos.lens_id = lenses.id WHERE lenses.lens_slug = 'zz' AND photos.lens_id NOT IN (SELECT id FROM lenses) AND photos.lens_id IS NOT NULL;
UPDATE photos SET place_id = 'zz' WHERE place_id NOT IN (SELECT id FROM places) AND place_id IS NOT NULL;
DELETE FROM photos_albums WHERE photo_uid NOT IN (SELECT photo_uid FROM photos) OR album_uid NOT IN (SELECT album_uid FROM albums);
DELETE FROM photos_keywords WHERE photo_id NOT IN (SELECT id FROM photos) OR keyword_id NOT IN (SELECT id FROM keywords);
DELETE FROM photos_labels WHERE photo_id NOT IN (SELECT id FROM photos) OR label_id NOT IN (SELECT id FROM labels);