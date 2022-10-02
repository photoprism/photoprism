CREATE DATABASE IF NOT EXISTS `migrate`;
CREATE USER IF NOT EXISTS 'migrate'@'%' IDENTIFIED BY 'migrate';
GRANT ALL PRIVILEGES ON `migrate`.* TO 'migrate'@'%';

FLUSH PRIVILEGES;

-- ----------------------------------------------------------------------------------------
-- init "migrate" db
-- ----------------------------------------------------------------------------------------

USE migrate;

-- MariaDB dump 10.19  Distrib 10.6.8-MariaDB, for debian-linux-gnu (x86_64)
--
-- Host: mariadb    Database: photoprism
-- ------------------------------------------------------
-- Server version	10.8.3-MariaDB-1:10.8.3+maria~jammy

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `accounts`
--

DROP TABLE IF EXISTS `accounts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `accounts` (
                            `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
                            `acc_name` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                            `acc_owner` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                            `acc_url` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
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
                            `share_path` varbinary(500) DEFAULT NULL,
                            `share_size` varbinary(16) DEFAULT NULL,
                            `share_expires` int(11) DEFAULT NULL,
                            `sync_path` varbinary(500) DEFAULT NULL,
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
                            KEY `idx_accounts_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `accounts`
--

LOCK TABLES `accounts` WRITE;
/*!40000 ALTER TABLE `accounts` DISABLE KEYS */;
/*!40000 ALTER TABLE `accounts` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `addresses`
--

DROP TABLE IF EXISTS `addresses`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `addresses` (
                             `id` int(11) NOT NULL AUTO_INCREMENT,
                             `cell_id` varbinary(64) DEFAULT 'zz',
                             `address_src` varbinary(8) DEFAULT NULL,
                             `address_lat` float DEFAULT NULL,
                             `address_lng` float DEFAULT NULL,
                             `address_line1` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                             `address_line2` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                             `address_zip` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                             `address_city` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                             `address_state` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                             `address_country` varbinary(2) DEFAULT 'zz',
                             `address_notes` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                             `created_at` datetime DEFAULT NULL,
                             `updated_at` datetime DEFAULT NULL,
                             `deleted_at` datetime DEFAULT NULL,
                             PRIMARY KEY (`id`),
                             KEY `idx_addresses_cell_id` (`cell_id`),
                             KEY `idx_addresses_address_lat` (`address_lat`),
                             KEY `idx_addresses_address_lng` (`address_lng`),
                             KEY `idx_addresses_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `addresses`
--

LOCK TABLES `addresses` WRITE;
/*!40000 ALTER TABLE `addresses` DISABLE KEYS */;
INSERT INTO `addresses` VALUES (1,'zz','default',0,0,'','','','','','zz','','2022-06-15 18:32:06','2022-06-15 18:32:06',NULL);
/*!40000 ALTER TABLE `addresses` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `albums`
--

DROP TABLE IF EXISTS `albums`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `albums` (
                          `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
                          `album_uid` varbinary(64) DEFAULT NULL,
                          `parent_uid` varbinary(64) DEFAULT '',
                          `album_slug` varbinary(160) DEFAULT NULL,
                          `album_path` varbinary(500) DEFAULT NULL,
                          `album_type` varbinary(8) DEFAULT 'album',
                          `album_title` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `album_location` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `album_category` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `album_caption` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `album_description` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `album_notes` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `album_filter` varbinary(2048) DEFAULT '',
                          `album_order` varbinary(32) DEFAULT NULL,
                          `album_template` varbinary(255) DEFAULT NULL,
                          `album_state` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `album_country` varbinary(2) DEFAULT 'zz',
                          `album_year` int(11) DEFAULT NULL,
                          `album_month` int(11) DEFAULT NULL,
                          `album_day` int(11) DEFAULT NULL,
                          `album_favorite` tinyint(1) DEFAULT NULL,
                          `album_private` tinyint(1) DEFAULT NULL,
                          `thumb` varbinary(128) DEFAULT '',
                          `thumb_src` varbinary(8) DEFAULT '',
                          `created_at` datetime DEFAULT NULL,
                          `updated_at` datetime DEFAULT NULL,
                          `deleted_at` datetime DEFAULT NULL,
                          PRIMARY KEY (`id`),
                          UNIQUE KEY `uix_albums_album_uid` (`album_uid`),
                          KEY `idx_albums_album_path` (`album_path`),
                          KEY `idx_albums_album_title` (`album_title`),
                          KEY `idx_albums_album_category` (`album_category`),
                          KEY `idx_albums_ymd` (`album_day`),
                          KEY `idx_albums_deleted_at` (`deleted_at`),
                          KEY `idx_albums_album_slug` (`album_slug`),
                          KEY `idx_albums_album_state` (`album_state`),
                          KEY `idx_albums_country_year_month` (`album_country`,`album_year`,`album_month`),
                          KEY `idx_albums_thumb` (`thumb`),
                          KEY `idx_albums_album_filter` (`album_filter`(512))
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `albums`
--

LOCK TABLES `albums` WRITE;
/*!40000 ALTER TABLE `albums` DISABLE KEYS */;
INSERT INTO `albums` VALUES (1,'ardj7i03be6hqp4x','','bar','bar','folder','Bar','','','','','','path:bar public:true','added','','','zz',2022,6,0,0,0,'e74a3cab4ddfa28d5872d1ecb358e7165f9723fb','','2022-06-15 18:32:24','2022-06-15 18:32:24',NULL),(2,'ardj7i02xnj6yn23','','foo','foo','folder','Foo','','','','','','path:foo public:true','added','','','zz',2022,6,0,0,0,'57e4a82eb42e553519d2bbe6c707d37f514cadb4','','2022-06-15 18:32:24','2022-06-15 18:32:24',NULL);
/*!40000 ALTER TABLE `albums` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `cameras`
--

DROP TABLE IF EXISTS `cameras`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cameras` (
                           `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
                           `camera_slug` varbinary(160) DEFAULT NULL,
                           `camera_name` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `camera_make` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `camera_model` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `camera_type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `camera_description` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `camera_notes` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `created_at` datetime DEFAULT NULL,
                           `updated_at` datetime DEFAULT NULL,
                           `deleted_at` datetime DEFAULT NULL,
                           PRIMARY KEY (`id`),
                           UNIQUE KEY `uix_cameras_camera_slug` (`camera_slug`),
                           KEY `idx_cameras_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `cameras`
--

LOCK TABLES `cameras` WRITE;
/*!40000 ALTER TABLE `cameras` DISABLE KEYS */;
INSERT INTO `cameras` VALUES (1,'zz','Unknown','','Unknown','','','','2022-06-15 18:32:06','2022-06-15 18:32:06',NULL),(2,'canon-eos-6d','Canon EOS 6D','Canon','EOS 6D','','','','2022-06-15 18:32:24','2022-06-15 18:32:24',NULL),(3,'canon-eos-5d','Canon EOS 5D','Canon','EOS 5D','','','','2022-06-15 18:32:32','2022-06-15 18:32:32',NULL),(4,'olympus-c2500l','Olympus C2500L','Olympus','C2500L','','','','2022-06-15 18:32:32','2022-06-15 18:32:32',NULL);
/*!40000 ALTER TABLE `cameras` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `categories`
--

DROP TABLE IF EXISTS `categories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `categories` (
                              `label_id` int(10) unsigned NOT NULL,
                              `category_id` int(10) unsigned NOT NULL,
                              PRIMARY KEY (`label_id`,`category_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `categories`
--

LOCK TABLES `categories` WRITE;
/*!40000 ALTER TABLE `categories` DISABLE KEYS */;
INSERT INTO `categories` VALUES (1,2),(3,4),(3,5);
/*!40000 ALTER TABLE `categories` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `cells`
--

DROP TABLE IF EXISTS `cells`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cells` (
                         `id` varbinary(64) NOT NULL,
                         `cell_name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `cell_street` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `cell_postcode` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `cell_category` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `place_id` varbinary(64) DEFAULT 'zz',
                         `created_at` datetime DEFAULT NULL,
                         `updated_at` datetime DEFAULT NULL,
                         PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `cells`
--

LOCK TABLES `cells` WRITE;
/*!40000 ALTER TABLE `cells` DISABLE KEYS */;
INSERT INTO `cells` VALUES ('s2:47a85a6482d4','Zentralasiatisches Gebirge Altai','Willdenowstraße','12203','botanical garden','de:z0vP8a5RZU2e','2022-06-15 18:32:33','2022-06-15 18:32:33'),('s2:89c25903bac4','Turtle Bay','3rd Avenue','10017','','us:a4cAIZhm2D6f','2022-06-15 18:32:33','2022-06-15 18:32:33'),('zz','','','','','zz','2022-06-15 18:32:06','2022-06-15 18:32:06');
/*!40000 ALTER TABLE `cells` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `countries`
--

DROP TABLE IF EXISTS `countries`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `countries` (
                             `id` varbinary(2) NOT NULL,
                             `country_slug` varbinary(160) DEFAULT NULL,
                             `country_name` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                             `country_description` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                             `country_notes` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                             `country_photo_id` int(10) unsigned DEFAULT NULL,
                             PRIMARY KEY (`id`),
                             UNIQUE KEY `uix_countries_country_slug` (`country_slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `countries`
--

LOCK TABLES `countries` WRITE;
/*!40000 ALTER TABLE `countries` DISABLE KEYS */;
INSERT INTO `countries` VALUES ('de','germany','Germany','','',0),('us','usa','USA','','',0),('zz','zz','Unknown','','',0);
/*!40000 ALTER TABLE `countries` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `details`
--

DROP TABLE IF EXISTS `details`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `details` (
                           `photo_id` int(10) unsigned NOT NULL,
                           `keywords` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `keywords_src` varbinary(8) DEFAULT NULL,
                           `notes` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `notes_src` varbinary(8) DEFAULT NULL,
                           `subject` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `subject_src` varbinary(8) DEFAULT NULL,
                           `artist` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `artist_src` varbinary(8) DEFAULT NULL,
                           `copyright` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `copyright_src` varbinary(8) DEFAULT NULL,
                           `license` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `license_src` varbinary(8) DEFAULT NULL,
                           `software` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `software_src` varbinary(8) DEFAULT NULL,
                           `created_at` datetime DEFAULT NULL,
                           `updated_at` datetime DEFAULT NULL,
                           PRIMARY KEY (`photo_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `details`
--

LOCK TABLES `details` WRITE;
/*!40000 ALTER TABLE `details` DISABLE KEYS */;
INSERT INTO `details` VALUES (1,'altai, berlin, blue, botanical, dahlem, demo, garden, gebirge, germany, lichterfelde, willdenowstraße, zentralasiatisches','','','','','','','','','','','','Adobe Photoshop CS6 (Macintosh)','meta','2022-06-15 18:32:24','2022-06-15 18:32:33'),(2,'bakery, bar, demo, green, reunion','meta','','','Reunion','meta','','','','','','','Adobe Photoshop CS6 (Macintosh)','meta','2022-06-15 18:32:32','2022-06-15 18:32:32'),(3,'avenue, bay, cab, demo, east, foo, manhattan, midtown, new-york, red, turtle, usa','','','','','','','','','','','','09-0120','meta','2022-06-15 18:32:33','2022-06-15 18:32:33');
/*!40000 ALTER TABLE `details` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `duplicates`
--

DROP TABLE IF EXISTS `duplicates`;
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `duplicates`
--

LOCK TABLES `duplicates` WRITE;
/*!40000 ALTER TABLE `duplicates` DISABLE KEYS */;
/*!40000 ALTER TABLE `duplicates` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `errors`
--

DROP TABLE IF EXISTS `errors`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `errors` (
                          `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
                          `error_time` datetime DEFAULT NULL,
                          `error_level` varbinary(32) DEFAULT NULL,
                          `error_message` varbinary(2048) DEFAULT NULL,
                          PRIMARY KEY (`id`),
                          KEY `idx_errors_error_time` (`error_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `errors`
--

LOCK TABLES `errors` WRITE;
/*!40000 ALTER TABLE `errors` DISABLE KEYS */;
/*!40000 ALTER TABLE `errors` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `faces`
--

DROP TABLE IF EXISTS `faces`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `faces` (
                         `id` varbinary(64) NOT NULL,
                         `face_src` varbinary(8) DEFAULT NULL,
                         `face_kind` int(11) DEFAULT NULL,
                         `face_hidden` tinyint(1) DEFAULT NULL,
                         `subj_uid` varbinary(64) DEFAULT '',
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `faces`
--

LOCK TABLES `faces` WRITE;
/*!40000 ALTER TABLE `faces` DISABLE KEYS */;
/*!40000 ALTER TABLE `faces` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `files`
--

DROP TABLE IF EXISTS `files`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `files` (
                         `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
                         `photo_id` int(10) unsigned DEFAULT NULL,
                         `photo_uid` varbinary(42) DEFAULT NULL,
                         `photo_taken_at` datetime DEFAULT NULL,
                         `time_index` varbinary(48) DEFAULT NULL,
                         `media_id` varbinary(32) DEFAULT NULL,
                         `media_utc` bigint(20) DEFAULT NULL,
                         `instance_id` varbinary(42) DEFAULT NULL,
                         `file_uid` varbinary(42) DEFAULT NULL,
                         `file_name` varbinary(755) DEFAULT NULL,
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
                         `file_projection` varbinary(64) DEFAULT NULL,
                         `file_aspect_ratio` float DEFAULT NULL,
                         `file_hdr` tinyint(1) DEFAULT NULL,
                         `file_watermark` tinyint(1) DEFAULT NULL,
                         `file_color_profile` varbinary(64) DEFAULT NULL,
                         `file_main_color` varbinary(16) DEFAULT NULL,
                         `file_colors` varbinary(9) DEFAULT NULL,
                         `file_luminance` varbinary(9) DEFAULT NULL,
                         `file_diff` int(11) DEFAULT -1,
                         `file_chroma` smallint(6) DEFAULT -1,
                         `file_software` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `file_error` varbinary(512) DEFAULT NULL,
                         `mod_time` bigint(20) DEFAULT NULL,
                         `created_at` datetime DEFAULT NULL,
                         `created_in` bigint(20) DEFAULT NULL,
                         `updated_at` datetime DEFAULT NULL,
                         `updated_in` bigint(20) DEFAULT NULL,
                         `deleted_at` datetime DEFAULT NULL,
                         PRIMARY KEY (`id`),
                         UNIQUE KEY `uix_files_file_uid` (`file_uid`),
                         UNIQUE KEY `idx_files_name_root` (`file_name`,`file_root`),
                         UNIQUE KEY `idx_files_search_media` (`media_id`),
                         UNIQUE KEY `idx_files_search_timeline` (`time_index`),
                         KEY `idx_files_photo_id` (`photo_id`,`file_primary`),
                         KEY `idx_files_photo_uid` (`photo_uid`),
                         KEY `idx_files_photo_taken_at` (`photo_taken_at`),
                         KEY `idx_files_media_utc` (`media_utc`),
                         KEY `idx_files_instance_id` (`instance_id`),
                         KEY `idx_files_file_hash` (`file_hash`),
                         KEY `idx_files_file_main_color` (`file_main_color`),
                         KEY `idx_files_deleted_at` (`deleted_at`),
                         KEY `idx_files_missing_root` (`file_missing`,`file_root`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `files`
--

LOCK TABLES `files` WRITE;
/*!40000 ALTER TABLE `files` DISABLE KEYS */;
INSERT INTO `files` VALUES (1,1,'pqowep21b8kn01jf','2013-06-05 18:35:44','79869394816456-9999999999-1-frdj7i0373r2qmt4','9999999999-1-frdj7i0373r2qmt4',1370450144000,'','frdj7i0373r2qmt4','20130605-183544-Botanical-Garden-Dahlem-2013.raw','/','','d9cacc1087b08b1c66fd4dfd96b383f00bc02e2e',22274437,'cr2','raw','raw','image/x-canon-cr2',0,0,0,0,0,0,0,0,5472,3648,1,'',1.5,0,0,'','','','',-1,-1,'','',1611248724,'2022-06-15 18:32:24',77576536,'2022-06-15 18:32:33',0,NULL),(2,2,'pqbcend17jiprb08','2010-03-12 17:10:07','79899687828993-9999999998-0-frdj7i81jysflc8r','9999999998-0-frdj7i81jysflc8r',1268413807000,'','frdj7i81jysflc8r','bar/20100312-171007-Bakery-2010.jpg','/','','e74a3cab4ddfa28d5872d1ecb358e7165f9723fb',756516,'jpeg','jpg','image','image/jpeg',1,0,0,0,0,0,0,0,2000,1333,1,'',1.5,0,0,'Camera RGB Profile','green','9991A0224','B873B174F',1023,22,'Adobe Photoshop CS6 (Macintosh)','',1595373641,'2022-06-15 18:32:32',8886824471,'2022-06-15 18:32:32',0,NULL),(3,2,'pqbcend17jiprb08','2010-03-12 17:10:07','79899687828993-9999999998-2-frdj7i8vofrb0i7e','9999999998-2-frdj7i8vofrb0i7e',1268413807000,'','frdj7i8vofrb0i7e','bar/20100312-171007-Bakery-2010.yml','/','','206f19fe31bd1e7e98fc41bb8a761c6d122e3cc1',655,'','yml','sidecar','',0,1,0,0,0,0,0,0,0,0,1,'',0,0,0,'','','','',-1,-1,'','',1655317901,'2022-06-15 18:32:32',14799302,'2022-06-15 18:32:32',0,NULL),(4,1,'pqowep21b8kn01jf','2013-06-05 18:35:44','79869394816456-9999999999-0-frdj7i939lkj4iwl','9999999999-0-frdj7i939lkj4iwl',1370450144000,'','frdj7i939lkj4iwl','20130605-183544-Botanical-Garden-Dahlem-2013.jpg','/','','426eafe429d64abbb39fe6aaa487b6793d946075',824025,'jpeg','jpg','image','image/jpeg',1,0,0,0,0,0,0,0,2048,1365,1,'',1.5,0,0,'Adobe RGB (1998)','blue','666AA6999','677BA7633',949,49,'Adobe Photoshop CS6 (Macintosh)','',1611248717,'2022-06-15 18:32:33',9051753722,'2022-06-15 18:32:33',0,NULL),(5,1,'pqowep21b8kn01jf','2013-06-05 18:35:44','79869394816456-9999999999-2-frdj7i92rk5hc8hb','9999999999-2-frdj7i92rk5hc8hb',1370457344000,'','frdj7i92rk5hc8hb','20130605-183544-Botanical-Garden-Dahlem-2013.yml','/','','f0a6486c475de593109eb3b1478806af43baa246',560,'','yml','sidecar','',0,1,0,0,0,0,0,0,0,0,1,'',0,0,0,'','','','',-1,-1,'','',1655317901,'2022-06-15 18:32:33',11178661,'2022-06-15 18:32:33',0,NULL),(6,3,'pqbcen521n7cdwa7','2002-03-12 19:01:54','79979687809846-9999999997-0-frdj7i92kt369bm4','9999999997-0-frdj7i92kt369bm4',1015959714000,'','frdj7i92kt369bm4','foo/20020312-190154-Cab-2002.jpg','/','','57e4a82eb42e553519d2bbe6c707d37f514cadb4',534188,'jpeg','jpg','image','image/jpeg',1,0,0,0,0,0,0,0,1712,1368,1,'',1.25,0,0,'Camera RGB Profile','red','01EE1EB33','2345C4BA9',1023,42,'09-0120','',1595373641,'2022-06-15 18:32:33',9179558186,'2022-06-15 18:32:33',0,NULL),(7,3,'pqbcen521n7cdwa7','2002-03-12 19:01:54','79979687809846-9999999997-2-frdj7i93ro6638ts','9999999997-2-frdj7i93ro6638ts',1015959714000,'','frdj7i93ro6638ts','foo/20020312-190154-Cab-2002.yml','/','','4cc8adc882f3aa0219ab31d0d4f4aca88593b96a',648,'','yml','sidecar','',0,1,0,0,0,0,0,0,0,0,1,'',0,0,0,'','','','',-1,-1,'','',1655317901,'2022-06-15 18:32:33',10984077,'2022-06-15 18:32:33',0,NULL);
/*!40000 ALTER TABLE `files` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `files_share`
--

DROP TABLE IF EXISTS `files_share`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `files_share` (
                               `file_id` int(10) unsigned NOT NULL,
                               `account_id` int(10) unsigned NOT NULL,
                               `remote_name` varbinary(255) NOT NULL,
                               `status` varbinary(16) DEFAULT NULL,
                               `error` varbinary(512) DEFAULT NULL,
                               `errors` int(11) DEFAULT NULL,
                               `created_at` datetime DEFAULT NULL,
                               `updated_at` datetime DEFAULT NULL,
                               PRIMARY KEY (`file_id`,`account_id`,`remote_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `files_share`
--

LOCK TABLES `files_share` WRITE;
/*!40000 ALTER TABLE `files_share` DISABLE KEYS */;
/*!40000 ALTER TABLE `files_share` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `files_sync`
--

DROP TABLE IF EXISTS `files_sync`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `files_sync` (
                              `remote_name` varbinary(255) NOT NULL,
                              `account_id` int(10) unsigned NOT NULL,
                              `file_id` int(10) unsigned DEFAULT NULL,
                              `remote_date` datetime DEFAULT NULL,
                              `remote_size` bigint(20) DEFAULT NULL,
                              `status` varbinary(16) DEFAULT NULL,
                              `error` varbinary(512) DEFAULT NULL,
                              `errors` int(11) DEFAULT NULL,
                              `created_at` datetime DEFAULT NULL,
                              `updated_at` datetime DEFAULT NULL,
                              PRIMARY KEY (`remote_name`,`account_id`),
                              KEY `idx_files_sync_file_id` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `files_sync`
--

LOCK TABLES `files_sync` WRITE;
/*!40000 ALTER TABLE `files_sync` DISABLE KEYS */;
/*!40000 ALTER TABLE `files_sync` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `folders`
--

DROP TABLE IF EXISTS `folders`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `folders` (
                           `path` varbinary(500) DEFAULT NULL,
                           `root` varbinary(16) DEFAULT '',
                           `folder_uid` varbinary(42) NOT NULL,
                           `folder_type` varbinary(16) DEFAULT NULL,
                           `folder_title` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `folder_category` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `folder_description` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
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
                           `deleted_at` datetime DEFAULT NULL,
                           PRIMARY KEY (`folder_uid`),
                           UNIQUE KEY `idx_folders_path_root` (`path`,`root`),
                           KEY `idx_folders_folder_category` (`folder_category`),
                           KEY `idx_folders_country_year_month` (`folder_country`,`folder_year`,`folder_month`),
                           KEY `idx_folders_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `folders`
--

LOCK TABLES `folders` WRITE;
/*!40000 ALTER TABLE `folders` DISABLE KEYS */;
INSERT INTO `folders` VALUES ('bar','/','drdj7i01kv0hpbs1','','Bar','','','name','zz',2022,6,0,0,0,0,0,'2022-06-15 18:32:24','2022-06-15 18:32:24','2022-06-15 18:31:41',NULL),('foo','/','drdj7i0ds27xkpyf','','Foo','','','name','zz',2022,6,0,0,0,0,0,'2022-06-15 18:32:24','2022-06-15 18:32:24','2022-06-15 18:31:41',NULL),('','/','drdj7i0t7ixeafsr','','Originals','','','name','zz',2022,6,0,0,0,0,0,'2022-06-15 18:32:24','2022-06-15 18:32:24','2022-06-15 18:31:41',NULL);
/*!40000 ALTER TABLE `folders` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `keywords`
--

DROP TABLE IF EXISTS `keywords`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `keywords` (
                            `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
                            `keyword` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                            `skip` tinyint(1) DEFAULT NULL,
                            PRIMARY KEY (`id`),
                            KEY `idx_keywords_keyword` (`keyword`)
) ENGINE=InnoDB AUTO_INCREMENT=60 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `keywords`
--

LOCK TABLES `keywords` WRITE;
/*!40000 ALTER TABLE `keywords` DISABLE KEYS */;
INSERT INTO `keywords` VALUES (1,'alteration',0),(2,'available',0),(3,'bakery',0),(4,'bar',0),(5,'believable',0),(6,'demo',0),(7,'form',0),(8,'green',0),(9,'humour',0),(10,'injected',0),(11,'ipsum',0),(12,'look',0),(13,'lorem',0),(14,'majority',0),(15,'many',0),(16,'passages',0),(17,'randomised',0),(18,'reunion',0),(19,'suffered',0),(20,'variations',0),(21,'words',0),(22,'avenue',0),(23,'bay',0),(24,'belief',0),(25,'cab',0),(26,'classical',0),(27,'contrary',0),(28,'altai',0),(29,'east',0),(30,'berlin',0),(31,'foo',0),(32,'blue',0),(33,'botanical',0),(34,'latin',0),(35,'dahlem',0),(36,'literature',0),(37,'garden',0),(38,'manhattan',0),(39,'gebirge',0),(40,'midtown',0),(41,'germany',0),(42,'new',0),(43,'lichterfelde',0),(44,'new-york',0),(45,'short',0),(46,'old',0),(47,'willdenowstraße',0),(48,'piece',0),(49,'zentralasiatisches',0),(50,'popular',0),(51,'random',0),(52,'red',0),(53,'roots',0),(54,'simply',0),(55,'text',0),(56,'turtle',0),(57,'usa',0),(58,'years',0),(59,'york',0);
/*!40000 ALTER TABLE `keywords` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `labels`
--

DROP TABLE IF EXISTS `labels`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `labels` (
                          `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
                          `label_uid` varbinary(42) DEFAULT NULL,
                          `label_slug` varbinary(160) DEFAULT NULL,
                          `custom_slug` varbinary(160) DEFAULT NULL,
                          `label_name` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `label_priority` int(11) DEFAULT NULL,
                          `label_favorite` tinyint(1) DEFAULT NULL,
                          `label_description` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `label_notes` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `photo_count` int(11) DEFAULT 1,
                          `thumb` varbinary(128) DEFAULT '',
                          `thumb_src` varbinary(8) DEFAULT '',
                          `created_at` datetime DEFAULT NULL,
                          `updated_at` datetime DEFAULT NULL,
                          `deleted_at` datetime DEFAULT NULL,
                          PRIMARY KEY (`id`),
                          UNIQUE KEY `uix_labels_label_uid` (`label_uid`),
                          UNIQUE KEY `uix_labels_label_slug` (`label_slug`),
                          KEY `idx_labels_custom_slug` (`custom_slug`),
                          KEY `idx_labels_thumb` (`thumb`),
                          KEY `idx_labels_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `labels`
--

LOCK TABLES `labels` WRITE;
/*!40000 ALTER TABLE `labels` DISABLE KEYS */;
INSERT INTO `labels` VALUES (1,'lrdj7i83abtybuaa','bakery','bakery','Bakery',0,0,'','',1,'e74a3cab4ddfa28d5872d1ecb358e7165f9723fb','','2022-06-15 18:32:32','2022-06-15 18:32:32',NULL),(2,'lrdj7i8295f0tug1','shop','shop','Shop',-3,0,'','',1,'e74a3cab4ddfa28d5872d1ecb358e7165f9723fb','','2022-06-15 18:32:32','2022-06-15 18:32:32',NULL),(3,'lrdj7i91mhs84jpy','cab','cab','Cab',0,0,'','',1,'57e4a82eb42e553519d2bbe6c707d37f514cadb4','','2022-06-15 18:32:33','2022-06-15 18:32:33',NULL),(4,'lrdj7i9cxj8ptza2','car','car','Car',-3,0,'','',1,'57e4a82eb42e553519d2bbe6c707d37f514cadb4','','2022-06-15 18:32:33','2022-06-15 18:32:33',NULL),(5,'lrdj7i91lnwrs6qm','vehicle','vehicle','Vehicle',-3,0,'','',1,'57e4a82eb42e553519d2bbe6c707d37f514cadb4','','2022-06-15 18:32:33','2022-06-15 18:32:33',NULL),(6,'lrdj7i9qazsdkln6','botanical-garden','botanical-garden','Botanical Garden',-1,0,'','',1,'426eafe429d64abbb39fe6aaa487b6793d946075','','2022-06-15 18:32:33','2022-06-15 18:32:33',NULL);
/*!40000 ALTER TABLE `labels` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `lenses`
--

DROP TABLE IF EXISTS `lenses`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `lenses` (
                          `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
                          `lens_slug` varbinary(160) DEFAULT NULL,
                          `lens_name` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `lens_make` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `lens_model` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `lens_type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `lens_description` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `lens_notes` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `created_at` datetime DEFAULT NULL,
                          `updated_at` datetime DEFAULT NULL,
                          `deleted_at` datetime DEFAULT NULL,
                          PRIMARY KEY (`id`),
                          UNIQUE KEY `uix_lenses_lens_slug` (`lens_slug`),
                          KEY `idx_lenses_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `lenses`
--

LOCK TABLES `lenses` WRITE;
/*!40000 ALTER TABLE `lenses` DISABLE KEYS */;
INSERT INTO `lenses` VALUES (1,'zz','Unknown','','Unknown','','','','2022-06-15 18:32:06','2022-06-15 18:32:06',NULL),(2,'ef16-35mm-f-2-8l-ii-usm','EF16-35mm f/2.8L II USM','','EF16-35mm f/2.8L II USM','','','','2022-06-15 18:32:24','2022-06-15 18:32:24',NULL);
/*!40000 ALTER TABLE `lenses` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `links`
--

DROP TABLE IF EXISTS `links`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `links` (
                         `link_uid` varbinary(42) NOT NULL,
                         `shared_uid` varbinary(42) DEFAULT NULL,
                         `share_slug` varbinary(160) DEFAULT NULL,
                         `link_token` varbinary(160) DEFAULT NULL,
                         `link_expires` int(11) DEFAULT NULL,
                         `link_views` int(10) unsigned DEFAULT NULL,
                         `max_views` int(10) unsigned DEFAULT NULL,
                         `has_password` tinyint(1) DEFAULT NULL,
                         `can_comment` tinyint(1) DEFAULT NULL,
                         `can_edit` tinyint(1) DEFAULT NULL,
                         `created_at` datetime DEFAULT NULL,
                         `modified_at` datetime DEFAULT NULL,
                         PRIMARY KEY (`link_uid`),
                         UNIQUE KEY `idx_links_uid_token` (`shared_uid`,`link_token`),
                         KEY `idx_links_share_slug` (`share_slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `links`
--

LOCK TABLES `links` WRITE;
/*!40000 ALTER TABLE `links` DISABLE KEYS */;
/*!40000 ALTER TABLE `links` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `markers`
--

DROP TABLE IF EXISTS `markers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `markers` (
                           `marker_uid` varbinary(42) NOT NULL,
                           `file_uid` varbinary(42) DEFAULT '',
                           `marker_type` varbinary(8) DEFAULT '',
                           `marker_src` varbinary(8) DEFAULT '',
                           `marker_name` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                           `marker_review` tinyint(1) DEFAULT NULL,
                           `marker_invalid` tinyint(1) DEFAULT NULL,
                           `subj_uid` varbinary(42) DEFAULT NULL,
                           `subj_src` varbinary(8) DEFAULT '',
                           `face_id` varbinary(42) DEFAULT NULL,
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `markers`
--

LOCK TABLES `markers` WRITE;
/*!40000 ALTER TABLE `markers` DISABLE KEYS */;
/*!40000 ALTER TABLE `markers` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `migrations`
--

DROP TABLE IF EXISTS `migrations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `migrations` (
                              `id` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL,
                              `dialect` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                              `error` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                              `source` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                              `started_at` datetime DEFAULT NULL,
                              `finished_at` datetime DEFAULT NULL,
                              PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `passwords`
--

DROP TABLE IF EXISTS `passwords`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `passwords` (
                             `uid` varbinary(255) NOT NULL,
                             `hash` varbinary(255) DEFAULT NULL,
                             `created_at` datetime DEFAULT NULL,
                             `updated_at` datetime DEFAULT NULL,
                             PRIMARY KEY (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `passwords`
--

LOCK TABLES `passwords` WRITE;
/*!40000 ALTER TABLE `passwords` DISABLE KEYS */;
INSERT INTO `passwords` VALUES ('urdj7hi2ng0eivkl','$2a$14$ynzpBwbDX426TMSjmlID6eQrn7ypuAct5od4.70lyiOspkasb8iyy','2022-06-15 18:32:07','2022-06-15 18:32:07');
/*!40000 ALTER TABLE `passwords` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `photos`
--

DROP TABLE IF EXISTS `photos`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `photos` (
                          `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
                          `uuid` varbinary(42) DEFAULT NULL,
                          `taken_at` datetime DEFAULT NULL,
                          `taken_at_local` datetime DEFAULT NULL,
                          `taken_src` varbinary(8) DEFAULT NULL,
                          `photo_uid` varbinary(42) DEFAULT NULL,
                          `photo_type` varbinary(8) DEFAULT 'image',
                          `type_src` varbinary(8) DEFAULT NULL,
                          `photo_title` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `title_src` varbinary(8) DEFAULT NULL,
                          `photo_description` TEXT COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `description_src` varbinary(8) DEFAULT NULL,
                          `photo_path` varbinary(500) DEFAULT NULL,
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
                          `photo_color` smallint(6) DEFAULT -1,
                          `camera_id` int(10) unsigned DEFAULT 1,
                          `camera_serial` varbinary(160) DEFAULT NULL,
                          `camera_src` varbinary(8) DEFAULT NULL,
                          `lens_id` int(10) unsigned DEFAULT 1,
                          `created_at` datetime DEFAULT NULL,
                          `updated_at` datetime DEFAULT NULL,
                          `edited_at` datetime DEFAULT NULL,
                          `checked_at` datetime DEFAULT NULL,
                          `estimated_at` datetime DEFAULT NULL,
                          `deleted_at` datetime DEFAULT NULL,
                          PRIMARY KEY (`id`),
                          UNIQUE KEY `uix_photos_photo_uid` (`photo_uid`),
                          KEY `idx_photos_uuid` (`uuid`),
                          KEY `idx_photos_taken_uid` (`taken_at`,`photo_uid`),
                          KEY `idx_photos_place_id` (`place_id`),
                          KEY `idx_photos_photo_lat` (`photo_lat`),
                          KEY `idx_photos_checked_at` (`checked_at`),
                          KEY `idx_photos_deleted_at` (`deleted_at`),
                          KEY `idx_photos_path_name` (`photo_path`,`photo_name`),
                          KEY `idx_photos_cell_id` (`cell_id`),
                          KEY `idx_photos_photo_lng` (`photo_lng`),
                          KEY `idx_photos_country_year_month` (`photo_country`,`photo_year`,`photo_month`),
                          KEY `idx_photos_ymd` (`photo_day`),
                          KEY `idx_photos_camera_lens` (`camera_id`,`lens_id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `photos`
--

LOCK TABLES `photos` WRITE;
/*!40000 ALTER TABLE `photos` DISABLE KEYS */;
INSERT INTO `photos` VALUES (1,'','2013-06-05 16:35:44','2013-06-05 18:35:44','meta','pqowep21b8kn01jf','raw','','Botanical Garden / Berlin / 2013','','Short Description','xmp','','20130605-183544-Botanical-Garden-Dahlem-2013','',0,0,0,0,0,'Europe/Berlin','de:z0vP8a5RZU2e','meta','s2:47a85a6482d4',0,47,52.4548,13.3045,'de',2013,6,5,100,'1/250',7.1,27,4,0,20,6,2,'033024001432','meta',2,'2021-02-21 21:36:38','2022-06-15 18:32:33',NULL,NULL,NULL,NULL),(2,'3257f85974879dc245cfd51d2df0dae4','2010-03-12 17:10:07','2010-03-12 17:10:07','meta','pqbcend17jiprb08','image','','Bakery / 2010','','There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour, or randomised words which don\'t look even slightly believable.','manual','bar','20100312-171007-Bakery-2010','',0,0,0,0,0,'','zz','','zz',0,30,0,0,'zz',2010,3,12,500,'1/60',4,35,3,0,3,9,3,'','meta',1,'2020-06-03 09:06:49','2022-06-15 18:32:32',NULL,NULL,NULL,NULL),(3,'','2002-03-13 00:01:54','2002-03-12 19:01:54','meta','pqbcen521n7cdwa7','image','','Cab / New York / 2002','','Contrary to popular belief, Lorem Ipsum is not simply random text. It has roots in a piece of classical Latin literature from 45 BC, making it over 2000 years old.','meta','foo','20020312-190154-Cab-2002','',0,0,0,0,0,'America/New_York','us:a4cAIZhm2D6f','','s2:89c25903bac4',0,0,40.7508,-73.975,'us',2002,3,12,100,'1/200',3.9,28,3,0,2,14,4,'2071381','meta',1,'2020-06-03 09:06:41','2022-06-15 18:32:33','2020-06-04 12:22:36',NULL,NULL,NULL);
/*!40000 ALTER TABLE `photos` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `photos_albums`
--

DROP TABLE IF EXISTS `photos_albums`;
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `photos_albums`
--

LOCK TABLES `photos_albums` WRITE;
/*!40000 ALTER TABLE `photos_albums` DISABLE KEYS */;
/*!40000 ALTER TABLE `photos_albums` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `photos_keywords`
--

DROP TABLE IF EXISTS `photos_keywords`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `photos_keywords` (
                                   `photo_id` int(10) unsigned NOT NULL,
                                   `keyword_id` int(10) unsigned NOT NULL,
                                   PRIMARY KEY (`photo_id`,`keyword_id`),
                                   KEY `idx_photos_keywords_keyword_id` (`keyword_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `photos_keywords`
--

LOCK TABLES `photos_keywords` WRITE;
/*!40000 ALTER TABLE `photos_keywords` DISABLE KEYS */;
INSERT INTO `photos_keywords` VALUES (1,6),(1,28),(1,30),(1,32),(1,33),(1,35),(1,37),(1,39),(1,41),(1,43),(1,45),(1,47),(1,49),(2,1),(2,2),(2,3),(2,4),(2,5),(2,6),(2,7),(2,8),(2,9),(2,10),(2,11),(2,12),(2,13),(2,14),(2,15),(2,16),(2,17),(2,18),(2,19),(2,20),(2,21),(3,6),(3,11),(3,13),(3,22),(3,23),(3,24),(3,25),(3,26),(3,27),(3,29),(3,31),(3,34),(3,36),(3,38),(3,40),(3,42),(3,44),(3,46),(3,48),(3,50),(3,51),(3,52),(3,53),(3,54),(3,55),(3,56),(3,57),(3,58),(3,59);
/*!40000 ALTER TABLE `photos_keywords` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `photos_labels`
--

DROP TABLE IF EXISTS `photos_labels`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `photos_labels` (
                                 `photo_id` int(10) unsigned NOT NULL,
                                 `label_id` int(10) unsigned NOT NULL,
                                 `label_src` varbinary(8) DEFAULT NULL,
                                 `uncertainty` smallint(6) DEFAULT NULL,
                                 PRIMARY KEY (`photo_id`,`label_id`),
                                 KEY `idx_photos_labels_label_id` (`label_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `photos_labels`
--

LOCK TABLES `photos_labels` WRITE;
/*!40000 ALTER TABLE `photos_labels` DISABLE KEYS */;
INSERT INTO `photos_labels` VALUES (1,6,'location',0),(2,1,'image',22),(3,3,'image',14);
/*!40000 ALTER TABLE `photos_labels` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `places`
--

DROP TABLE IF EXISTS `places`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `places` (
                          `id` varbinary(42) NOT NULL,
                          `place_label` varchar(400) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `place_district` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `place_city` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `place_state` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `place_country` varbinary(2) DEFAULT NULL,
                          `place_keywords` varchar(300) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                          `place_favorite` tinyint(1) DEFAULT NULL,
                          `photo_count` int(11) DEFAULT 1,
                          `created_at` datetime DEFAULT NULL,
                          `updated_at` datetime DEFAULT NULL,
                          PRIMARY KEY (`id`),
                          KEY `idx_places_place_state` (`place_state`),
                          KEY `idx_places_place_district` (`place_district`),
                          KEY `idx_places_place_city` (`place_city`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `places`
--

LOCK TABLES `places` WRITE;
/*!40000 ALTER TABLE `places` DISABLE KEYS */;
INSERT INTO `places` VALUES ('de:z0vP8a5RZU2e','Lichterfelde, Berlin, Germany','Lichterfelde','Berlin','Berlin','de','',0,1,'2022-06-15 18:32:33','2022-06-15 18:32:33'),('us:a4cAIZhm2D6f','Manhattan, New York, USA','Manhattan','New York','New York','us','',0,1,'2022-06-15 18:32:33','2022-06-15 18:32:33'),('zz','Unknown','Unknown','Unknown','Unknown','zz','',0,1,'2022-06-15 18:32:06','2022-06-15 18:32:06');
/*!40000 ALTER TABLE `places` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `subjects`
--

DROP TABLE IF EXISTS `subjects`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `subjects` (
                            `subj_uid` varbinary(42) NOT NULL,
                            `subj_type` varbinary(8) DEFAULT '',
                            `subj_src` varbinary(8) DEFAULT '',
                            `subj_slug` varbinary(160) DEFAULT '',
                            `subj_name` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT '',
                            `subj_alias` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT '',
                            `subj_bio` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                            `subj_notes` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                            `subj_favorite` tinyint(1) DEFAULT 0,
                            `subj_hidden` tinyint(1) DEFAULT 0,
                            `subj_private` tinyint(1) DEFAULT 0,
                            `subj_excluded` tinyint(1) DEFAULT 0,
                            `file_count` int(11) DEFAULT 0,
                            `photo_count` int(11) DEFAULT 0,
                            `thumb` varbinary(128) DEFAULT '',
                            `thumb_src` varbinary(8) DEFAULT '',
                            `metadata_json` mediumblob DEFAULT NULL,
                            `created_at` datetime DEFAULT NULL,
                            `updated_at` datetime DEFAULT NULL,
                            `deleted_at` datetime DEFAULT NULL,
                            PRIMARY KEY (`subj_uid`),
                            UNIQUE KEY `uix_subjects_subj_name` (`subj_name`),
                            KEY `idx_subjects_deleted_at` (`deleted_at`),
                            KEY `idx_subjects_subj_slug` (`subj_slug`),
                            KEY `idx_subjects_thumb` (`thumb`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `subjects`
--

LOCK TABLES `subjects` WRITE;
/*!40000 ALTER TABLE `subjects` DISABLE KEYS */;
/*!40000 ALTER TABLE `subjects` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `users` (
                         `id` int(11) NOT NULL AUTO_INCREMENT,
                         `address_id` int(11) DEFAULT 1,
                         `user_uid` varbinary(42) DEFAULT NULL,
                         `mother_uid` varbinary(42) DEFAULT NULL,
                         `father_uid` varbinary(42) DEFAULT NULL,
                         `global_uid` varbinary(42) DEFAULT NULL,
                         `full_name` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `nick_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `maiden_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `artist_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `user_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `user_status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `user_disabled` tinyint(1) DEFAULT NULL,
                         `user_settings` longtext COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `primary_email` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `email_confirmed` tinyint(1) DEFAULT NULL,
                         `backup_email` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `person_url` varbinary(255) DEFAULT NULL,
                         `person_phone` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `person_status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `person_avatar` varbinary(255) DEFAULT NULL,
                         `person_location` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `person_bio` text COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `person_accounts` longtext COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `business_url` varbinary(255) DEFAULT NULL,
                         `business_phone` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `business_email` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `company_name` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `department_name` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `job_title` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
                         `birth_year` int(11) DEFAULT NULL,
                         `birth_month` int(11) DEFAULT NULL,
                         `birth_day` int(11) DEFAULT NULL,
                         `terms_accepted` tinyint(1) DEFAULT NULL,
                         `is_artist` tinyint(1) DEFAULT NULL,
                         `is_subject` tinyint(1) DEFAULT NULL,
                         `role_admin` tinyint(1) DEFAULT NULL,
                         `role_guest` tinyint(1) DEFAULT NULL,
                         `role_child` tinyint(1) DEFAULT NULL,
                         `role_family` tinyint(1) DEFAULT NULL,
                         `role_friend` tinyint(1) DEFAULT NULL,
                         `webdav` tinyint(1) DEFAULT NULL,
                         `storage_path` varbinary(500) DEFAULT NULL,
                         `can_invite` tinyint(1) DEFAULT NULL,
                         `invite_token` varbinary(32) DEFAULT NULL,
                         `invited_by` varbinary(32) DEFAULT NULL,
                         `confirm_token` varbinary(64) DEFAULT NULL,
                         `reset_token` varbinary(64) DEFAULT NULL,
                         `api_token` varbinary(128) DEFAULT NULL,
                         `api_secret` varbinary(128) DEFAULT NULL,
                         `login_attempts` int(11) DEFAULT NULL,
                         `login_at` datetime DEFAULT NULL,
                         `created_at` datetime DEFAULT NULL,
                         `updated_at` datetime DEFAULT NULL,
                         `deleted_at` datetime DEFAULT NULL,
                         PRIMARY KEY (`id`),
                         UNIQUE KEY `uix_users_user_uid` (`user_uid`),
                         KEY `idx_users_global_uid` (`global_uid`),
                         KEY `idx_users_primary_email` (`primary_email`),
                         KEY `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (-2,1,'u000000000000002','','','','Guest','','','','','',1,'','',0,'','','','','','','','','','','','','','',0,0,0,0,0,0,0,1,0,0,0,0,'',0,'','','','','','',0,NULL,'2022-06-15 18:32:06','2022-06-15 18:32:06',NULL),(-1,1,'u000000000000001','','','','Anonymous','','','','','',1,'','',0,'','','','','','','','','','','','','','',0,0,0,0,0,0,0,0,0,0,0,0,'',0,'','','','','','',0,NULL,'2022-06-15 18:32:06','2022-06-15 18:32:06',NULL),(1,1,'urdj7hi2ng0eivkl','','','','Admin','','','','admin','',0,'','',0,'','','','','','','','','','','','','','',0,0,0,0,0,0,1,0,0,0,0,0,'',0,'','','','','','',0,NULL,'2022-06-15 18:32:06','2022-06-15 18:32:06',NULL);
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2022-06-15 18:32:44
