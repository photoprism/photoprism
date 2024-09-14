/*!999999\- enable the sandbox mode */ 
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `albums` (
  `id` int(10) unsigned NOT NULL,
  `album_uid` varbinary(42) DEFAULT NULL,
  `parent_uid` varbinary(42) DEFAULT '',
  `album_slug` varbinary(160) DEFAULT NULL,
  `album_path` varchar(1024) DEFAULT NULL,
  `album_type` varbinary(8) DEFAULT 'album',
  `album_title` varchar(160) DEFAULT NULL,
  `album_location` varchar(160) DEFAULT NULL,
  `album_category` varchar(100) DEFAULT NULL,
  `album_caption` varchar(1024) DEFAULT NULL,
  `album_description` varchar(2048) DEFAULT NULL,
  `album_notes` varchar(1024) DEFAULT NULL,
  `album_filter` varbinary(2048) DEFAULT '',
  `album_order` varbinary(32) DEFAULT NULL,
  `album_template` varbinary(255) DEFAULT NULL,
  `album_state` varchar(100) DEFAULT NULL,
  `album_country` varbinary(2) DEFAULT 'zz',
  `album_year` int(11) DEFAULT NULL,
  `album_month` int(11) DEFAULT NULL,
  `album_day` int(11) DEFAULT NULL,
  `album_favorite` tinyint(1) DEFAULT NULL,
  `album_private` tinyint(1) DEFAULT NULL,
  `thumb` varbinary(128) DEFAULT '',
  `thumb_src` varbinary(8) DEFAULT '',
  `created_by` varbinary(42) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `published_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uix_albums_album_uid` (`album_uid`),
  KEY `idx_albums_country_year_month` (`album_country`,`album_year`,`album_month`),
  KEY `idx_albums_ymd` (`album_day`),
  KEY `idx_albums_album_slug` (`album_slug`),
  KEY `idx_albums_album_path` (`album_path`(768)),
  KEY `idx_albums_album_category` (`album_category`),
  KEY `idx_albums_album_state` (`album_state`),
  KEY `idx_albums_deleted_at` (`deleted_at`),
  KEY `idx_albums_album_title` (`album_title`),
  KEY `idx_albums_thumb` (`thumb`),
  KEY `idx_albums_created_by` (`created_by`),
  KEY `idx_albums_published_at` (`published_at`),
  KEY `idx_albums_album_filter` (`album_filter`(512))
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `albums_users` (
  `uid` varbinary(42) NOT NULL,
  `user_uid` varbinary(42) NOT NULL,
  `team_uid` varbinary(42) DEFAULT NULL,
  `perm` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`uid`,`user_uid`),
  KEY `idx_albums_users_user_uid` (`user_uid`),
  KEY `idx_albums_users_team_uid` (`team_uid`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `audit_logins` (
  `client_ip` varchar(64) NOT NULL,
  `login_name` varchar(64) NOT NULL,
  `login_realm` varchar(64) NOT NULL,
  `login_status` varchar(32) DEFAULT NULL,
  `error_message` varchar(512) DEFAULT NULL,
  `error_repeated` bigint(20) DEFAULT NULL,
  `client_browser` varchar(512) DEFAULT NULL,
  `login_at` datetime DEFAULT NULL,
  `failed_at` datetime DEFAULT NULL,
  `banned_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`client_ip`,`login_name`,`login_realm`),
  KEY `idx_audit_logins_failed_at` (`failed_at`),
  KEY `idx_audit_logins_banned_at` (`banned_at`),
  KEY `idx_audit_logins_updated_at` (`updated_at`),
  KEY `idx_audit_logins_login_name` (`login_name`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `auth_clients` (
  `client_uid` varbinary(42) NOT NULL,
  `user_uid` varbinary(42) DEFAULT '',
  `user_name` varchar(200) DEFAULT NULL,
  `client_name` varchar(200) DEFAULT NULL,
  `client_role` varchar(64) DEFAULT '',
  `client_type` varbinary(16) DEFAULT NULL,
  `client_url` varbinary(255) DEFAULT '',
  `callback_url` varbinary(255) DEFAULT '',
  `auth_provider` varbinary(128) DEFAULT '',
  `auth_method` varbinary(128) DEFAULT '',
  `auth_scope` varchar(1024) DEFAULT '',
  `auth_expires` bigint(20) DEFAULT NULL,
  `auth_tokens` bigint(20) DEFAULT NULL,
  `auth_enabled` tinyint(1) DEFAULT NULL,
  `last_active` bigint(20) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`client_uid`),
  KEY `idx_auth_clients_user_name` (`user_name`),
  KEY `idx_auth_clients_user_uid` (`user_uid`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `auth_sessions` (
  `id` varbinary(2048) NOT NULL,
  `user_uid` varbinary(42) DEFAULT '',
  `user_name` varchar(200) DEFAULT NULL,
  `client_uid` varbinary(42) DEFAULT '',
  `client_name` varchar(200) DEFAULT '',
  `client_ip` varchar(64) DEFAULT NULL,
  `auth_provider` varbinary(128) DEFAULT '',
  `auth_method` varbinary(128) DEFAULT '',
  `auth_issuer` varbinary(255) DEFAULT '',
  `auth_id` varbinary(255) DEFAULT '',
  `auth_scope` varchar(1024) DEFAULT '',
  `grant_type` varbinary(64) DEFAULT '',
  `last_active` bigint(20) DEFAULT NULL,
  `sess_expires` bigint(20) DEFAULT NULL,
  `sess_timeout` bigint(20) DEFAULT NULL,
  `preview_token` varbinary(64) DEFAULT '',
  `download_token` varbinary(64) DEFAULT '',
  `access_token` varbinary(4096) DEFAULT '',
  `refresh_token` varbinary(2048) DEFAULT NULL,
  `id_token` varbinary(2048) DEFAULT NULL,
  `user_agent` varchar(512) DEFAULT NULL,
  `data_json` varbinary(4096) DEFAULT NULL,
  `ref_id` varbinary(16) DEFAULT '',
  `login_ip` varchar(64) DEFAULT NULL,
  `login_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_auth_sessions_sess_expires` (`sess_expires`),
  KEY `idx_auth_sessions_user_uid` (`user_uid`),
  KEY `idx_auth_sessions_user_name` (`user_name`),
  KEY `idx_auth_sessions_client_uid` (`client_uid`),
  KEY `idx_auth_sessions_client_ip` (`client_ip`),
  KEY `idx_auth_sessions_auth_id` (`auth_id`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `auth_users` (
  `id` int(11) NOT NULL,
  `user_uuid` varbinary(64) DEFAULT NULL,
  `user_uid` varbinary(42) DEFAULT NULL,
  `auth_provider` varbinary(128) DEFAULT '',
  `auth_method` varbinary(128) DEFAULT '',
  `auth_issuer` varbinary(255) DEFAULT '',
  `auth_id` varbinary(255) DEFAULT '',
  `user_name` varchar(200) DEFAULT NULL,
  `display_name` varchar(200) DEFAULT NULL,
  `user_email` varchar(255) DEFAULT NULL,
  `backup_email` varchar(255) DEFAULT NULL,
  `user_role` varchar(64) DEFAULT '',
  `user_attr` varchar(1024) DEFAULT NULL,
  `super_admin` tinyint(1) DEFAULT NULL,
  `can_login` tinyint(1) DEFAULT NULL,
  `login_at` datetime DEFAULT NULL,
  `expires_at` datetime DEFAULT NULL,
  `webdav` tinyint(1) DEFAULT NULL,
  `base_path` varbinary(1024) DEFAULT NULL,
  `upload_path` varbinary(1024) DEFAULT NULL,
  `can_invite` tinyint(1) DEFAULT NULL,
  `invite_token` varbinary(64) DEFAULT NULL,
  `invited_by` varchar(64) DEFAULT NULL,
  `verify_token` varbinary(64) DEFAULT NULL,
  `verified_at` datetime DEFAULT NULL,
  `consent_at` datetime DEFAULT NULL,
  `born_at` datetime DEFAULT NULL,
  `reset_token` varbinary(64) DEFAULT NULL,
  `preview_token` varbinary(64) DEFAULT NULL,
  `download_token` varbinary(64) DEFAULT NULL,
  `thumb` varbinary(128) DEFAULT '',
  `thumb_src` varbinary(8) DEFAULT '',
  `ref_id` varbinary(16) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uix_auth_users_user_uid` (`user_uid`),
  KEY `idx_auth_users_thumb` (`thumb`),
  KEY `idx_auth_users_deleted_at` (`deleted_at`),
  KEY `idx_auth_users_user_name` (`user_name`),
  KEY `idx_auth_users_user_email` (`user_email`),
  KEY `idx_auth_users_expires_at` (`expires_at`),
  KEY `idx_auth_users_invite_token` (`invite_token`),
  KEY `idx_auth_users_born_at` (`born_at`),
  KEY `idx_auth_users_user_uuid` (`user_uuid`),
  KEY `idx_auth_users_auth_id` (`auth_id`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `auth_users_details` (
  `user_uid` varbinary(42) NOT NULL,
  `subj_uid` varbinary(42) DEFAULT NULL,
  `subj_src` varbinary(8) DEFAULT '',
  `place_id` varbinary(42) DEFAULT 'zz',
  `place_src` varbinary(8) DEFAULT NULL,
  `cell_id` varbinary(42) DEFAULT 'zz',
  `birth_year` int(11) DEFAULT NULL,
  `birth_month` int(11) DEFAULT NULL,
  `birth_day` int(11) DEFAULT NULL,
  `name_title` varchar(32) DEFAULT NULL,
  `given_name` varchar(64) DEFAULT NULL,
  `middle_name` varchar(64) DEFAULT NULL,
  `family_name` varchar(64) DEFAULT NULL,
  `name_suffix` varchar(32) DEFAULT NULL,
  `nick_name` varchar(64) DEFAULT NULL,
  `name_src` varbinary(8) DEFAULT NULL,
  `user_gender` varchar(16) DEFAULT NULL,
  `user_about` varchar(512) DEFAULT NULL,
  `user_bio` varchar(2048) DEFAULT NULL,
  `user_location` varchar(512) DEFAULT NULL,
  `user_country` varbinary(2) DEFAULT NULL,
  `user_phone` varchar(32) DEFAULT NULL,
  `site_url` varbinary(512) DEFAULT NULL,
  `profile_url` varbinary(512) DEFAULT NULL,
  `feed_url` varbinary(512) DEFAULT NULL,
  `avatar_url` varbinary(512) DEFAULT NULL,
  `org_title` varchar(64) DEFAULT NULL,
  `org_name` varchar(128) DEFAULT NULL,
  `org_email` varchar(255) DEFAULT NULL,
  `org_phone` varchar(32) DEFAULT NULL,
  `org_url` varbinary(512) DEFAULT NULL,
  `id_url` varbinary(512) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`user_uid`),
  KEY `idx_auth_users_details_subj_uid` (`subj_uid`),
  KEY `idx_auth_users_details_place_id` (`place_id`),
  KEY `idx_auth_users_details_cell_id` (`cell_id`),
  KEY `idx_auth_users_details_org_email` (`org_email`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `auth_users_settings` (
  `user_uid` varbinary(42) NOT NULL,
  `ui_theme` varbinary(32) DEFAULT NULL,
  `ui_language` varbinary(32) DEFAULT NULL,
  `ui_time_zone` varbinary(64) DEFAULT NULL,
  `maps_style` varbinary(32) DEFAULT NULL,
  `maps_animate` int(11) DEFAULT 0,
  `index_path` varbinary(1024) DEFAULT NULL,
  `index_rescan` int(11) DEFAULT 0,
  `import_path` varbinary(1024) DEFAULT NULL,
  `import_move` int(11) DEFAULT 0,
  `download_originals` int(11) DEFAULT 0,
  `download_media_raw` int(11) DEFAULT 0,
  `download_media_sidecar` int(11) DEFAULT 0,
  `upload_path` varbinary(1024) DEFAULT NULL,
  `default_page` varbinary(128) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`user_uid`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `auth_users_shares` (
  `user_uid` varbinary(42) NOT NULL,
  `share_uid` varbinary(42) NOT NULL,
  `link_uid` varbinary(42) DEFAULT NULL,
  `expires_at` datetime DEFAULT NULL,
  `comment` varchar(512) DEFAULT NULL,
  `perm` int(10) unsigned DEFAULT NULL,
  `ref_id` varbinary(16) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`user_uid`,`share_uid`),
  KEY `idx_auth_users_shares_share_uid` (`share_uid`),
  KEY `idx_auth_users_shares_expires_at` (`expires_at`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cameras` (
  `id` int(10) unsigned NOT NULL,
  `camera_slug` varbinary(160) DEFAULT NULL,
  `camera_name` varchar(160) DEFAULT NULL,
  `camera_make` varchar(160) DEFAULT NULL,
  `camera_model` varchar(160) DEFAULT NULL,
  `camera_type` varchar(100) DEFAULT NULL,
  `camera_description` varchar(2048) DEFAULT NULL,
  `camera_notes` varchar(1024) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uix_cameras_camera_slug` (`camera_slug`),
  KEY `idx_cameras_deleted_at` (`deleted_at`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `categories` (
  `label_id` int(10) unsigned NOT NULL,
  `category_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`label_id`,`category_id`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cells` (
  `id` varbinary(42) NOT NULL,
  `cell_name` varchar(200) DEFAULT NULL,
  `cell_street` varchar(100) DEFAULT NULL,
  `cell_postcode` varchar(50) DEFAULT NULL,
  `cell_category` varchar(50) DEFAULT NULL,
  `place_id` varbinary(42) DEFAULT 'zz',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `countries` (
  `id` varbinary(2) NOT NULL,
  `country_slug` varbinary(160) DEFAULT NULL,
  `country_name` varchar(160) DEFAULT NULL,
  `country_description` varchar(2048) DEFAULT NULL,
  `country_notes` varchar(1024) DEFAULT NULL,
  `country_photo_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uix_countries_country_slug` (`country_slug`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `details` (
  `photo_id` int(10) unsigned NOT NULL,
  `keywords` varchar(2048) DEFAULT NULL,
  `keywords_src` varbinary(8) DEFAULT NULL,
  `notes` varchar(2048) DEFAULT NULL,
  `notes_src` varbinary(8) DEFAULT NULL,
  `subject` varchar(1024) DEFAULT NULL,
  `subject_src` varbinary(8) DEFAULT NULL,
  `artist` varchar(1024) DEFAULT NULL,
  `artist_src` varbinary(8) DEFAULT NULL,
  `copyright` varchar(1024) DEFAULT NULL,
  `copyright_src` varbinary(8) DEFAULT NULL,
  `license` varchar(1024) DEFAULT NULL,
  `license_src` varbinary(8) DEFAULT NULL,
  `software` varchar(1024) DEFAULT NULL,
  `software_src` varbinary(8) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`photo_id`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `duplicates` (
  `file_name` varbinary(755) NOT NULL,
  `file_root` varbinary(16) NOT NULL DEFAULT '/',
  `file_hash` varbinary(128) DEFAULT '',
  `file_size` bigint(20) DEFAULT NULL,
  `mod_time` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`file_name`,`file_root`),
  KEY `idx_duplicates_file_hash` (`file_hash`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `errors` (
  `id` int(10) unsigned NOT NULL,
  `error_time` datetime DEFAULT NULL,
  `error_level` varbinary(32) DEFAULT NULL,
  `error_message` varbinary(2048) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_errors_error_time` (`error_time`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `faces` (
  `id` varbinary(64) NOT NULL,
  `face_src` varbinary(8) DEFAULT NULL,
  `face_kind` int(11) DEFAULT NULL,
  `face_hidden` tinyint(1) DEFAULT NULL,
  `subj_uid` varbinary(42) DEFAULT '',
  `samples` int(11) DEFAULT NULL,
  `sample_radius` double DEFAULT NULL,
  `collisions` int(11) DEFAULT NULL,
  `collision_radius` double DEFAULT NULL,
  `embedding_json` mediumblob DEFAULT NULL,
  `matched_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_faces_subj_uid` (`subj_uid`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `files` (
  `id` int(10) unsigned NOT NULL,
  `photo_id` int(10) unsigned DEFAULT NULL,
  `photo_uid` varbinary(42) DEFAULT NULL,
  `photo_taken_at` datetime DEFAULT NULL,
  `time_index` varbinary(64) DEFAULT NULL,
  `media_id` varbinary(32) DEFAULT NULL,
  `media_utc` bigint(20) DEFAULT NULL,
  `instance_id` varbinary(64) DEFAULT NULL,
  `file_uid` varbinary(42) DEFAULT NULL,
  `file_name` varbinary(1024) DEFAULT NULL,
  `file_root` varbinary(16) DEFAULT '/',
  `original_name` varbinary(755) DEFAULT NULL,
  `file_hash` varbinary(128) DEFAULT NULL,
  `file_size` bigint(20) DEFAULT NULL,
  `file_codec` varbinary(32) DEFAULT NULL,
  `file_type` varbinary(16) DEFAULT NULL,
  `media_type` varbinary(16) DEFAULT NULL,
  `file_mime` varbinary(64) DEFAULT NULL,
  `file_primary` tinyint(1) DEFAULT NULL,
  `file_sidecar` tinyint(1) DEFAULT NULL,
  `file_missing` tinyint(1) DEFAULT NULL,
  `file_portrait` tinyint(1) DEFAULT NULL,
  `file_video` tinyint(1) DEFAULT NULL,
  `file_duration` bigint(20) DEFAULT NULL,
  `file_fps` double DEFAULT NULL,
  `file_frames` int(11) DEFAULT NULL,
  `file_width` int(11) DEFAULT NULL,
  `file_height` int(11) DEFAULT NULL,
  `file_orientation` int(11) DEFAULT NULL,
  `file_orientation_src` varbinary(8) DEFAULT '',
  `file_projection` varbinary(64) DEFAULT NULL,
  `file_aspect_ratio` float DEFAULT NULL,
  `file_hdr` tinyint(1) DEFAULT NULL,
  `file_watermark` tinyint(1) DEFAULT NULL,
  `file_color_profile` varbinary(64) DEFAULT NULL,
  `file_main_color` varbinary(16) DEFAULT NULL,
  `file_colors` varbinary(18) DEFAULT NULL,
  `File_luminance` varbinary(18) DEFAULT NULL,
  `file_diff` int(11) DEFAULT -1,
  `file_chroma` smallint(6) DEFAULT -1,
  `file_software` varchar(64) DEFAULT NULL,
  `file_error` varbinary(512) DEFAULT NULL,
  `mod_time` bigint(20) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `created_in` bigint(20) DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `updated_in` bigint(20) DEFAULT NULL,
  `published_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uix_files_file_uid` (`file_uid`),
  UNIQUE KEY `idx_files_name_root` (`file_name`,`file_root`),
  UNIQUE KEY `idx_files_search_media` (`media_id`),
  UNIQUE KEY `idx_files_search_timeline` (`time_index`),
  KEY `idx_files_photo_uid` (`photo_uid`),
  KEY `idx_files_file_hash` (`file_hash`),
  KEY `idx_files_published_at` (`published_at`),
  KEY `idx_files_file_error` (`file_error`),
  KEY `idx_files_deleted_at` (`deleted_at`),
  KEY `idx_files_photo_id` (`photo_id`,`file_primary`),
  KEY `idx_files_photo_taken_at` (`photo_taken_at`),
  KEY `idx_files_media_utc` (`media_utc`),
  KEY `idx_files_instance_id` (`instance_id`),
  KEY `idx_files_missing_root` (`file_missing`,`file_root`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `files_share` (
  `file_id` int(10) unsigned NOT NULL,
  `service_id` int(10) unsigned NOT NULL,
  `remote_name` varbinary(255) NOT NULL,
  `status` varbinary(16) DEFAULT NULL,
  `error` varbinary(512) DEFAULT NULL,
  `errors` int(11) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`file_id`,`service_id`,`remote_name`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `files_sync` (
  `remote_name` varbinary(255) NOT NULL,
  `service_id` int(10) unsigned NOT NULL,
  `file_id` int(10) unsigned DEFAULT NULL,
  `remote_date` datetime DEFAULT NULL,
  `remote_size` bigint(20) DEFAULT NULL,
  `status` varbinary(16) DEFAULT NULL,
  `error` varbinary(512) DEFAULT NULL,
  `errors` int(11) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`remote_name`,`service_id`),
  KEY `idx_files_sync_file_id` (`file_id`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `folders` (
  `path` varbinary(1024) DEFAULT NULL,
  `root` varbinary(16) DEFAULT '',
  `folder_uid` varbinary(42) NOT NULL,
  `folder_type` varbinary(16) DEFAULT NULL,
  `folder_title` varchar(200) DEFAULT NULL,
  `folder_category` varchar(100) DEFAULT NULL,
  `folder_description` varchar(2048) DEFAULT NULL,
  `folder_order` varbinary(32) DEFAULT NULL,
  `folder_country` varbinary(2) DEFAULT 'zz',
  `folder_year` int(11) DEFAULT NULL,
  `folder_month` int(11) DEFAULT NULL,
  `folder_day` int(11) DEFAULT NULL,
  `folder_favorite` tinyint(1) DEFAULT NULL,
  `folder_private` tinyint(1) DEFAULT NULL,
  `folder_ignore` tinyint(1) DEFAULT NULL,
  `folder_watch` tinyint(1) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `modified_at` datetime DEFAULT NULL,
  `published_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`folder_uid`),
  UNIQUE KEY `idx_folders_path_root` (`path`,`root`),
  KEY `idx_folders_folder_category` (`folder_category`),
  KEY `idx_folders_country_year_month` (`folder_country`,`folder_year`,`folder_month`),
  KEY `idx_folders_published_at` (`published_at`),
  KEY `idx_folders_deleted_at` (`deleted_at`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `keywords` (
  `id` int(10) unsigned NOT NULL,
  `keyword` varchar(64) DEFAULT NULL,
  `skip` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_keywords_keyword` (`keyword`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `labels` (
  `id` int(10) unsigned NOT NULL,
  `label_uid` varbinary(42) DEFAULT NULL,
  `label_slug` varbinary(160) DEFAULT NULL,
  `custom_slug` varbinary(160) DEFAULT NULL,
  `label_name` varchar(160) DEFAULT NULL,
  `label_priority` int(11) DEFAULT NULL,
  `label_favorite` tinyint(1) DEFAULT NULL,
  `label_description` varchar(2048) DEFAULT NULL,
  `label_notes` varchar(1024) DEFAULT NULL,
  `photo_count` int(11) DEFAULT 1,
  `thumb` varbinary(128) DEFAULT '',
  `thumb_src` varbinary(8) DEFAULT '',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `published_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uix_labels_label_uid` (`label_uid`),
  UNIQUE KEY `uix_labels_label_slug` (`label_slug`),
  KEY `idx_labels_custom_slug` (`custom_slug`),
  KEY `idx_labels_thumb` (`thumb`),
  KEY `idx_labels_published_at` (`published_at`),
  KEY `idx_labels_deleted_at` (`deleted_at`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `lenses` (
  `id` int(10) unsigned NOT NULL,
  `lens_slug` varbinary(160) DEFAULT NULL,
  `lens_name` varchar(160) DEFAULT NULL,
  `lens_make` varchar(160) DEFAULT NULL,
  `lens_model` varchar(160) DEFAULT NULL,
  `lens_type` varchar(100) DEFAULT NULL,
  `lens_description` varchar(2048) DEFAULT NULL,
  `lens_notes` varchar(1024) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uix_lenses_lens_slug` (`lens_slug`),
  KEY `idx_lenses_deleted_at` (`deleted_at`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `links` (
  `link_uid` varbinary(42) NOT NULL,
  `share_uid` varbinary(42) DEFAULT NULL,
  `share_slug` varbinary(160) DEFAULT NULL,
  `link_token` varbinary(160) DEFAULT NULL,
  `link_expires` int(11) DEFAULT NULL,
  `link_views` int(10) unsigned DEFAULT NULL,
  `max_views` int(10) unsigned DEFAULT NULL,
  `has_password` tinyint(1) DEFAULT NULL,
  `comment` varchar(512) DEFAULT NULL,
  `perm` int(10) unsigned DEFAULT NULL,
  `ref_id` varbinary(16) DEFAULT NULL,
  `created_by` varbinary(42) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `modified_at` datetime DEFAULT NULL,
  PRIMARY KEY (`link_uid`),
  UNIQUE KEY `idx_links_uid_token` (`share_uid`,`link_token`),
  KEY `idx_links_share_slug` (`share_slug`),
  KEY `idx_links_created_by` (`created_by`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `markers` (
  `marker_uid` varbinary(42) NOT NULL,
  `file_uid` varbinary(42) DEFAULT '',
  `marker_type` varbinary(8) DEFAULT '',
  `marker_src` varbinary(8) DEFAULT '',
  `marker_name` varchar(160) DEFAULT NULL,
  `marker_review` tinyint(1) DEFAULT NULL,
  `marker_invalid` tinyint(1) DEFAULT NULL,
  `subj_uid` varbinary(42) DEFAULT NULL,
  `subj_src` varbinary(8) DEFAULT '',
  `face_id` varbinary(64) DEFAULT NULL,
  `face_dist` double DEFAULT -1,
  `embeddings_json` mediumblob DEFAULT NULL,
  `landmarks_json` mediumblob DEFAULT NULL,
  `x` float DEFAULT NULL,
  `y` float DEFAULT NULL,
  `w` float DEFAULT NULL,
  `h` float DEFAULT NULL,
  `q` int(11) DEFAULT NULL,
  `size` int(11) DEFAULT -1,
  `score` smallint(6) DEFAULT NULL,
  `thumb` varbinary(128) DEFAULT '',
  `matched_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`marker_uid`),
  KEY `idx_markers_file_uid` (`file_uid`),
  KEY `idx_markers_subj_uid_src` (`subj_uid`,`subj_src`),
  KEY `idx_markers_face_id` (`face_id`),
  KEY `idx_markers_thumb` (`thumb`),
  KEY `idx_markers_matched_at` (`matched_at`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `migrations` (
  `id` varchar(16) NOT NULL,
  `dialect` varchar(16) DEFAULT NULL,
  `stage` varchar(16) DEFAULT NULL,
  `error` varchar(255) DEFAULT NULL,
  `source` varchar(16) DEFAULT NULL,
  `started_at` datetime DEFAULT NULL,
  `finished_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `passcodes` (
  `uid` varbinary(255) NOT NULL,
  `key_type` varchar(64) NOT NULL DEFAULT '',
  `key_url` varchar(2048) DEFAULT '',
  `recovery_code` varchar(255) DEFAULT '',
  `verified_at` datetime DEFAULT NULL,
  `activated_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`uid`,`key_type`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `passwords` (
  `uid` varbinary(255) NOT NULL,
  `hash` varbinary(255) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`uid`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `photos` (
  `id` int(10) unsigned NOT NULL,
  `uuid` varbinary(64) DEFAULT NULL,
  `taken_at` datetime DEFAULT NULL,
  `taken_at_local` datetime DEFAULT NULL,
  `taken_src` varbinary(8) DEFAULT NULL,
  `photo_uid` varbinary(42) DEFAULT NULL,
  `photo_type` varbinary(8) DEFAULT 'image',
  `type_src` varbinary(8) DEFAULT NULL,
  `photo_title` varchar(200) DEFAULT NULL,
  `title_src` varbinary(8) DEFAULT NULL,
  `photo_description` varchar(4096) DEFAULT NULL,
  `description_src` varbinary(8) DEFAULT NULL,
  `photo_path` varbinary(1024) DEFAULT NULL,
  `photo_name` varbinary(255) DEFAULT NULL,
  `original_name` varbinary(755) DEFAULT NULL,
  `photo_stack` tinyint(4) DEFAULT NULL,
  `photo_favorite` tinyint(1) DEFAULT NULL,
  `photo_private` tinyint(1) DEFAULT NULL,
  `photo_scan` tinyint(1) DEFAULT NULL,
  `photo_panorama` tinyint(1) DEFAULT NULL,
  `time_zone` varbinary(64) DEFAULT NULL,
  `place_id` varbinary(42) DEFAULT 'zz',
  `place_src` varbinary(8) DEFAULT NULL,
  `cell_id` varbinary(42) DEFAULT 'zz',
  `cell_accuracy` int(11) DEFAULT NULL,
  `photo_altitude` int(11) DEFAULT NULL,
  `photo_lat` float DEFAULT NULL,
  `photo_lng` float DEFAULT NULL,
  `photo_country` varbinary(2) DEFAULT 'zz',
  `photo_year` int(11) DEFAULT NULL,
  `photo_month` int(11) DEFAULT NULL,
  `photo_day` int(11) DEFAULT NULL,
  `photo_iso` int(11) DEFAULT NULL,
  `photo_exposure` varbinary(64) DEFAULT NULL,
  `photo_f_number` float DEFAULT NULL,
  `photo_focal_length` int(11) DEFAULT NULL,
  `photo_quality` smallint(6) DEFAULT NULL,
  `photo_faces` int(11) DEFAULT NULL,
  `photo_resolution` smallint(6) DEFAULT NULL,
  `photo_duration` bigint(20) DEFAULT NULL,
  `photo_color` smallint(6) DEFAULT -1,
  `camera_id` int(10) unsigned DEFAULT 1,
  `camera_serial` varbinary(160) DEFAULT NULL,
  `camera_src` varbinary(8) DEFAULT NULL,
  `lens_id` int(10) unsigned DEFAULT 1,
  `created_by` varbinary(42) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `edited_at` datetime DEFAULT NULL,
  `published_at` datetime DEFAULT NULL,
  `checked_at` datetime DEFAULT NULL,
  `estimated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uix_photos_photo_uid` (`photo_uid`),
  KEY `idx_photos_photo_lng` (`photo_lng`),
  KEY `idx_photos_camera_lens` (`camera_id`,`lens_id`),
  KEY `idx_photos_taken_uid` (`taken_at`,`photo_uid`),
  KEY `idx_photos_ymd` (`photo_day`),
  KEY `idx_photos_checked_at` (`checked_at`),
  KEY `idx_photos_uuid` (`uuid`),
  KEY `idx_photos_path_name` (`photo_path`,`photo_name`),
  KEY `idx_photos_place_id` (`place_id`),
  KEY `idx_photos_cell_id` (`cell_id`),
  KEY `idx_photos_created_by` (`created_by`),
  KEY `idx_photos_deleted_at` (`deleted_at`),
  KEY `idx_photos_photo_lat` (`photo_lat`),
  KEY `idx_photos_country_year_month` (`photo_country`,`photo_year`,`photo_month`),
  KEY `idx_photos_published_at` (`published_at`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `photos_albums` (
  `photo_uid` varbinary(42) NOT NULL,
  `album_uid` varbinary(42) NOT NULL,
  `order` int(11) DEFAULT NULL,
  `hidden` tinyint(1) DEFAULT NULL,
  `missing` tinyint(1) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`photo_uid`,`album_uid`),
  KEY `idx_photos_albums_album_uid` (`album_uid`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `photos_keywords` (
  `photo_id` int(10) unsigned NOT NULL,
  `keyword_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`photo_id`,`keyword_id`),
  KEY `idx_photos_keywords_keyword_id` (`keyword_id`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `photos_labels` (
  `photo_id` int(10) unsigned NOT NULL,
  `label_id` int(10) unsigned NOT NULL,
  `label_src` varbinary(8) DEFAULT NULL,
  `uncertainty` smallint(6) DEFAULT NULL,
  PRIMARY KEY (`photo_id`,`label_id`),
  KEY `idx_photos_labels_label_id` (`label_id`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `photos_users` (
  `uid` varbinary(42) NOT NULL,
  `user_uid` varbinary(42) NOT NULL,
  `team_uid` varbinary(42) DEFAULT NULL,
  `perm` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`uid`,`user_uid`),
  KEY `idx_photos_users_user_uid` (`user_uid`),
  KEY `idx_photos_users_team_uid` (`team_uid`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `places` (
  `id` varbinary(42) NOT NULL,
  `place_label` varchar(400) DEFAULT NULL,
  `place_district` varchar(100) DEFAULT NULL,
  `place_city` varchar(100) DEFAULT NULL,
  `place_state` varchar(100) DEFAULT NULL,
  `place_country` varbinary(2) DEFAULT NULL,
  `place_keywords` varchar(300) DEFAULT NULL,
  `place_favorite` tinyint(1) DEFAULT NULL,
  `photo_count` int(11) DEFAULT 1,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_places_place_state` (`place_state`),
  KEY `idx_places_place_district` (`place_district`),
  KEY `idx_places_place_city` (`place_city`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `reactions` (
  `uid` varbinary(42) NOT NULL,
  `user_uid` varbinary(42) NOT NULL,
  `reaction` varbinary(64) NOT NULL,
  `reacted` int(11) DEFAULT NULL,
  `reacted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`uid`,`user_uid`,`reaction`),
  KEY `idx_reactions_reacted_at` (`reacted_at`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `services` (
  `id` int(10) unsigned NOT NULL,
  `acc_name` varchar(160) DEFAULT NULL,
  `acc_owner` varchar(160) DEFAULT NULL,
  `acc_url` varchar(255) DEFAULT NULL,
  `acc_type` varbinary(255) DEFAULT NULL,
  `acc_key` varbinary(255) DEFAULT NULL,
  `acc_user` varbinary(255) DEFAULT NULL,
  `acc_pass` varbinary(255) DEFAULT NULL,
  `acc_timeout` varbinary(16) DEFAULT NULL,
  `acc_error` varbinary(512) DEFAULT NULL,
  `acc_errors` int(11) DEFAULT NULL,
  `acc_share` tinyint(1) DEFAULT NULL,
  `acc_sync` tinyint(1) DEFAULT NULL,
  `retry_limit` int(11) DEFAULT NULL,
  `share_path` varbinary(1024) DEFAULT NULL,
  `share_size` varbinary(16) DEFAULT NULL,
  `share_expires` int(11) DEFAULT NULL,
  `sync_path` varbinary(1024) DEFAULT NULL,
  `sync_status` varbinary(16) DEFAULT NULL,
  `sync_interval` int(11) DEFAULT NULL,
  `sync_date` datetime DEFAULT NULL,
  `sync_upload` tinyint(1) DEFAULT NULL,
  `sync_download` tinyint(1) DEFAULT NULL,
  `sync_filenames` tinyint(1) DEFAULT NULL,
  `sync_raw` tinyint(1) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_services_deleted_at` (`deleted_at`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `subjects` (
  `subj_uid` varbinary(42) NOT NULL,
  `subj_type` varbinary(8) DEFAULT '',
  `subj_src` varbinary(8) DEFAULT '',
  `subj_slug` varbinary(160) DEFAULT '',
  `subj_name` varchar(160) DEFAULT '',
  `subj_alias` varchar(160) DEFAULT '',
  `subj_about` varchar(512) DEFAULT NULL,
  `subj_bio` varchar(2048) DEFAULT NULL,
  `subj_notes` varchar(1024) DEFAULT NULL,
  `subj_favorite` tinyint(1) DEFAULT 0,
  `subj_hidden` tinyint(1) DEFAULT 0,
  `subj_private` tinyint(1) DEFAULT 0,
  `subj_excluded` tinyint(1) DEFAULT 0,
  `file_count` int(11) DEFAULT 0,
  `photo_count` int(11) DEFAULT 0,
  `thumb` varbinary(128) DEFAULT '',
  `thumb_src` varbinary(8) DEFAULT '',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`subj_uid`),
  UNIQUE KEY `uix_subjects_subj_name` (`subj_name`),
  KEY `idx_subjects_subj_slug` (`subj_slug`),
  KEY `idx_subjects_thumb` (`thumb`),
  KEY `idx_subjects_deleted_at` (`deleted_at`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `versions` (
  `id` int(10) unsigned NOT NULL,
  `version` varchar(255) DEFAULT NULL,
  `edition` varchar(255) DEFAULT NULL,
  `error` varchar(255) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `migrated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_version_edition` (`version`,`edition`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

