INSERT INTO cameras (id, camera_slug, camera_model, camera_make, camera_type, camera_owner, camera_description, camera_notes, created_at, updated_at, deleted_at) VALUES (1, 'unknown', 'Unknown', '', '', '', '', '', '2020-01-06 02:06:29', '2020-01-06 02:07:26', null);
INSERT INTO cameras (id, camera_slug, camera_model, camera_make, camera_type, camera_owner, camera_description, camera_notes, created_at, updated_at, deleted_at) VALUES (2, 'apple-iphone-se', 'iPhone SE', 'Apple', '', '', '', '', '2020-01-06 02:06:30', '2020-01-06 02:07:28', null);
INSERT INTO cameras (id, camera_slug, camera_model, camera_make, camera_type, camera_owner, camera_description, camera_notes, created_at, updated_at, deleted_at) VALUES (3, 'canon-eos-5d', 'EOS 5D', 'Canon', '', '', '', '', '2020-01-06 02:06:32', '2020-01-06 02:06:32', null);
INSERT INTO cameras (id, camera_slug, camera_model, camera_make, camera_type, camera_owner, camera_description, camera_notes, created_at, updated_at, deleted_at) VALUES (4, 'canon-eos-7d', 'EOS 7D', 'Canon', '', '', '', '', '2020-01-06 02:06:33', '2020-01-06 02:06:33', null);
INSERT INTO cameras (id, camera_slug, camera_model, camera_make, camera_type, camera_owner, camera_description, camera_notes, created_at, updated_at, deleted_at) VALUES (5, 'canon-eos-6d', 'EOS 6D', 'Canon', '', '', '', '', '2020-01-06 02:06:35', '2020-01-06 02:06:54', null);
INSERT INTO cameras (id, camera_slug, camera_model, camera_make, camera_type, camera_owner, camera_description, camera_notes, created_at, updated_at, deleted_at) VALUES (6, 'apple-iphone-6', 'iPhone 6', 'Apple', '', '', '', '', '2020-01-06 02:06:42', '2020-01-06 02:06:42', null);
INSERT INTO cameras (id, camera_slug, camera_model, camera_make, camera_type, camera_owner, camera_description, camera_notes, created_at, updated_at, deleted_at) VALUES (7, 'apple-iphone-7', 'iPhone 7', 'Apple', '', '', '', '', '2020-01-06 02:06:51', '2020-01-06 02:06:51', null);
INSERT INTO countries (id, country_slug, country_name, country_description, country_notes, country_photo_id) VALUES ('de', 'germany', 'Germany', 'Country Description', 'Country Notes', 0);
INSERT INTO albums (id, album_uuid, album_name, album_slug, album_favorite) VALUES ('2', '3', 'Christmas2030', 'christmas2030', 0);
INSERT INTO albums (id, album_uuid, cover_uuid, album_name, album_slug, album_favorite) VALUES ('1', '4', '654', 'Holiday2030', 'holiday-2030', 1);
INSERT INTO photos_albums (album_uuid, photo_uuid) VALUES ('4', '654');
INSERT INTO files (id, photo_id, photo_uuid, file_name, file_primary, file_hash, file_missing) VALUES ('1', '1', '654', 'exampleFileName.jpg', 1, '123xxx', 0);
INSERT INTO files (id, photo_id, photo_uuid, file_name, file_primary, file_hash, file_missing) VALUES ('2', '2', '655', 'exampleDNGFile.dng', 1, '124xxx', 0);
INSERT INTO files (id, photo_id, photo_uuid, file_name, file_primary, file_hash, file_missing) VALUES ('3', '2', '655', 'exampleXmpFile.xmp', 0, '125xxx', 0);
INSERT INTO files (id, photo_id, photo_uuid, file_name, file_primary, file_hash, file_missing) VALUES ('4', '5', '658', 'bridge.jpg', 1, '126xxx', 0);
INSERT INTO files (id, photo_id, photo_uuid, file_name, file_primary, file_hash, file_missing) VALUES ('5', '6', '659', 'reunion.jpg', 1, '127xxx', 0);
INSERT INTO photos (id, photo_uuid, photo_year, photo_month, photo_lat, photo_lng) VALUES ('1', '654', 2790, 2, '48.519235', '9.057996666666666');
INSERT INTO photos (id, photo_uuid, photo_year, photo_month, photo_lat, photo_lng) VALUES ('2', '655', 2790, 2, '48.519235', '9.057996666666666');
INSERT INTO photos (id, photo_uuid, photo_year, photo_month, photo_lat, photo_lng) VALUES ('3', '656', 1990, 3, '48.519235', '9.057996666666666');
INSERT INTO photos (id, photo_uuid, photo_year, photo_month, photo_lat, photo_lng) VALUES ('4', '657', 1990, 4, '48.519235', '9.057996666666666');
INSERT INTO photos (id, photo_uuid, taken_at, photo_lat, photo_lng, photo_title) VALUES ('5', '658', '2014-07-17 15:42:12', '48.519235', '9.057996666666666', 'Neckarbrücke');
INSERT INTO photos (id, photo_uuid, taken_at, photo_lat, photo_lng, photo_title) VALUES ('6', '659', '2015-11-11 09:07:18', '-21.34263611111111', '55.466944444444444', 'Reunion');
INSERT INTO keywords (id, keyword, skip) VALUES (1, 'bridge', 0);
INSERT INTO keywords (id, keyword, skip) VALUES (2, 'beach', 0);
INSERT INTO photos_keywords (photo_id, keyword_id) VALUES (5, 1);
INSERT INTO categories (label_id, category_id) VALUES ('1', '1');
INSERT INTO labels (id, label_uuid, label_slug, label_name, label_priority, label_favorite) VALUES ('1', '12', 'flower', 'Flower', 1, 1);
INSERT INTO labels (id, label_uuid, label_slug, label_name, label_priority, label_favorite) VALUES ('2', '13', 'cake', 'Cake', 5, 0);
INSERT INTO labels (id, label_uuid, label_slug, label_name, label_priority, label_favorite) VALUES ('3', '14', 'cow', 'COW', -1, 1);
INSERT INTO photos_labels (photo_id, label_id, label_uncertainty, label_source) VALUES ('1', '1', '38', 'image');
INSERT INTO photos_labels (photo_id, label_id, label_uncertainty, label_source) VALUES ('1', '2', '10', 'image');




