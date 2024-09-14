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

CREATE DATABASE IF NOT EXISTS keycloak;
CREATE USER IF NOT EXISTS keycloak@'%' IDENTIFIED BY 'keycloak';
GRANT ALL PRIVILEGES ON keycloak.* TO keycloak@'%';

CREATE DATABASE IF NOT EXISTS `local`;
CREATE USER IF NOT EXISTS 'local'@'%' IDENTIFIED BY 'local';
GRANT ALL PRIVILEGES ON `local`.* TO 'local'@'%';

CREATE DATABASE IF NOT EXISTS latest;
CREATE USER IF NOT EXISTS latest@'%' IDENTIFIED BY 'latest';
GRANT ALL PRIVILEGES ON latest.* TO latest@'%';

CREATE DATABASE IF NOT EXISTS preview;
CREATE USER IF NOT EXISTS preview@'%' IDENTIFIED BY 'preview';
GRANT ALL PRIVILEGES ON preview.* TO preview@preview;

CREATE DATABASE IF NOT EXISTS testdb;
CREATE USER IF NOT EXISTS testdb@'%' IDENTIFIED BY 'testdb';
GRANT ALL PRIVILEGES ON testdb.* TO testdb@'%';

CREATE DATABASE IF NOT EXISTS `migrate`;
CREATE USER IF NOT EXISTS 'migrate'@'%' IDENTIFIED BY 'migrate';
GRANT ALL PRIVILEGES ON `migrate`.* TO 'migrate'@'%';

CREATE DATABASE IF NOT EXISTS acceptance;
CREATE USER IF NOT EXISTS acceptance@'%' IDENTIFIED BY 'acceptance';
GRANT ALL PRIVILEGES ON acceptance.* TO acceptance@'%';

CREATE DATABASE IF NOT EXISTS photoprism_01;
CREATE USER IF NOT EXISTS photoprism_01@'%' IDENTIFIED BY 'photoprism_01';
GRANT ALL PRIVILEGES ON photoprism_01.* TO photoprism_01@'%';
CREATE DATABASE IF NOT EXISTS photoprism_02;
CREATE USER IF NOT EXISTS photoprism_02@'%' IDENTIFIED BY 'photoprism_02';
GRANT ALL PRIVILEGES ON photoprism_02.* TO photoprism_02@'%';
CREATE DATABASE IF NOT EXISTS photoprism_03;
CREATE USER IF NOT EXISTS photoprism_03@'%' IDENTIFIED BY 'photoprism_03';
GRANT ALL PRIVILEGES ON photoprism_03.* TO photoprism_03@'%';
CREATE DATABASE IF NOT EXISTS photoprism_04;
CREATE USER IF NOT EXISTS photoprism_04@'%' IDENTIFIED BY 'photoprism_04';
GRANT ALL PRIVILEGES ON photoprism_04.* TO photoprism_04@'%';
CREATE DATABASE IF NOT EXISTS photoprism_05;
CREATE USER IF NOT EXISTS photoprism_05@'%' IDENTIFIED BY 'photoprism_05';
GRANT ALL PRIVILEGES ON photoprism_05.* TO photoprism_05@'%';

FLUSH PRIVILEGES;

-- ----------------------------------------------------------------------------------------
-- init "keycloak" db
-- ----------------------------------------------------------------------------------------

USE keycloak;

--
-- Table structure for table `ADMIN_EVENT_ENTITY`
--

DROP TABLE IF EXISTS `ADMIN_EVENT_ENTITY`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ADMIN_EVENT_ENTITY` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ADMIN_EVENT_TIME` bigint(20) DEFAULT NULL,
  `REALM_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `OPERATION_TYPE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `AUTH_REALM_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `AUTH_CLIENT_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `AUTH_USER_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `IP_ADDRESS` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `RESOURCE_PATH` text COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REPRESENTATION` text COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ERROR` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `RESOURCE_TYPE` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ADMIN_EVENT_ENTITY`
--

LOCK TABLES `ADMIN_EVENT_ENTITY` WRITE;
/*!40000 ALTER TABLE `ADMIN_EVENT_ENTITY` DISABLE KEYS */;
/*!40000 ALTER TABLE `ADMIN_EVENT_ENTITY` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ASSOCIATED_POLICY`
--

DROP TABLE IF EXISTS `ASSOCIATED_POLICY`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ASSOCIATED_POLICY` (
  `POLICY_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ASSOCIATED_POLICY_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`POLICY_ID`,`ASSOCIATED_POLICY_ID`),
  KEY `IDX_ASSOC_POL_ASSOC_POL_ID` (`ASSOCIATED_POLICY_ID`),
  CONSTRAINT `FK_FRSR5S213XCX4WNKOG82SSRFY` FOREIGN KEY (`ASSOCIATED_POLICY_ID`) REFERENCES `RESOURCE_SERVER_POLICY` (`ID`),
  CONSTRAINT `FK_FRSRPAS14XCX4WNKOG82SSRFY` FOREIGN KEY (`POLICY_ID`) REFERENCES `RESOURCE_SERVER_POLICY` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ASSOCIATED_POLICY`
--

LOCK TABLES `ASSOCIATED_POLICY` WRITE;
/*!40000 ALTER TABLE `ASSOCIATED_POLICY` DISABLE KEYS */;
/*!40000 ALTER TABLE `ASSOCIATED_POLICY` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `AUTHENTICATION_EXECUTION`
--

DROP TABLE IF EXISTS `AUTHENTICATION_EXECUTION`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `AUTHENTICATION_EXECUTION` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ALIAS` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `AUTHENTICATOR` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `FLOW_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REQUIREMENT` int(11) DEFAULT NULL,
  `PRIORITY` int(11) DEFAULT NULL,
  `AUTHENTICATOR_FLOW` bit(1) NOT NULL DEFAULT b'0',
  `AUTH_FLOW_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `AUTH_CONFIG` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_AUTH_EXEC_REALM_FLOW` (`REALM_ID`,`FLOW_ID`),
  KEY `IDX_AUTH_EXEC_FLOW` (`FLOW_ID`),
  CONSTRAINT `FK_AUTH_EXEC_FLOW` FOREIGN KEY (`FLOW_ID`) REFERENCES `AUTHENTICATION_FLOW` (`ID`),
  CONSTRAINT `FK_AUTH_EXEC_REALM` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `AUTHENTICATION_EXECUTION`
--

LOCK TABLES `AUTHENTICATION_EXECUTION` WRITE;
/*!40000 ALTER TABLE `AUTHENTICATION_EXECUTION` DISABLE KEYS */;
INSERT INTO `AUTHENTICATION_EXECUTION` VALUES ('10d29d2a-e6ef-412f-9d8d-a73245d94886',NULL,'direct-grant-validate-otp','master','124c5389-5d37-4d77-a226-f39350c568e9',0,20,'\0',NULL,NULL),('10ff89c6-f5b0-4076-8763-a2d9f5467c1a',NULL,'client-jwt','master','d180151b-30b4-4438-8ec3-9033d2e60c38',2,20,'\0',NULL,NULL),('15633fb6-e4d4-48a1-817b-dd22e177de5c',NULL,'conditional-user-configured','master','b9691c0e-fca4-43f2-bca1-40558f346594',0,10,'\0',NULL,NULL),('1723bba0-6f63-4c54-baff-8e96e803ea4e',NULL,'identity-provider-redirector','master','5adc5270-4510-48ee-898c-cbae8f28a3cd',2,25,'\0',NULL,NULL),('235d8c25-fd4c-4631-baaf-4c46259b3ed5',NULL,'auth-spnego','master','5adc5270-4510-48ee-898c-cbae8f28a3cd',3,20,'\0',NULL,NULL),('236ec61c-6e8d-48d3-8453-03d092f38445',NULL,'idp-email-verification','master','3e2f3567-7f57-4dfd-95b1-90a6fff2e79c',2,10,'\0',NULL,NULL),('27880ce3-5679-4fbe-a0ee-680ad9da4f5a',NULL,'client-x509','master','d180151b-30b4-4438-8ec3-9033d2e60c38',2,40,'\0',NULL,NULL),('295cf641-f50a-445c-a3cf-a43afcd5c6d6',NULL,'idp-review-profile','master','3dc0e0a6-fec2-47b9-b48d-e31bf31a164d',0,10,'\0',NULL,'98d35458-7329-4099-9c58-c12984adf496'),('2f647074-0625-441d-a125-ebbb93a73af6',NULL,'registration-password-action','master','8a75833c-68ef-4248-9602-5781e764bb55',0,50,'\0',NULL,NULL),('3034a537-c4af-44eb-8f29-6231d566f6b5',NULL,NULL,'master','5157a4d6-247e-447e-bb54-12e9136c3dc4',2,20,'','a6df1ac7-e51e-416c-afc2-3fb82e5594c3',NULL),('38807fba-6bb3-4563-9165-5de5e02ddf8e',NULL,'conditional-user-configured','master','124c5389-5d37-4d77-a226-f39350c568e9',0,10,'\0',NULL,NULL),('3ad4351a-b76e-4d6f-b1ac-2a6e4e28f9e3',NULL,NULL,'master','6137a5dc-223b-4817-aed0-4b33b126164d',0,20,'','1bd473b0-79f0-46a5-a47c-fc9fad029026',NULL),('3d96b58d-1156-48cc-9727-a038a34f3067',NULL,'registration-profile-action','master','8a75833c-68ef-4248-9602-5781e764bb55',0,40,'\0',NULL,NULL),('4283107d-e756-40f0-ba37-df7e311113f7',NULL,'auth-otp-form','master','b9691c0e-fca4-43f2-bca1-40558f346594',0,20,'\0',NULL,NULL),('43bcf0f9-6d2d-411b-8f19-28fa076b8bce',NULL,NULL,'master','5adc5270-4510-48ee-898c-cbae8f28a3cd',2,30,'','7d35ce2d-9375-42b1-ba30-249cdbbb1944',NULL),('539c9f31-df42-4dcb-a128-9359bbb553e9',NULL,'idp-create-user-if-unique','master','5157a4d6-247e-447e-bb54-12e9136c3dc4',2,10,'\0',NULL,'d3806a77-7749-48bb-bbd8-0981d7bad74f'),('5f94d282-18e2-422a-86cc-3a79b79687c4',NULL,'reset-otp','master','9f0eb3c3-59a5-43ae-aa6c-fc70b158ea7a',0,20,'\0',NULL,NULL),('667e16c2-69d5-4f29-b37b-9502736ff929',NULL,'auth-cookie','master','5adc5270-4510-48ee-898c-cbae8f28a3cd',2,10,'\0',NULL,NULL),('67410c98-7dae-4c92-9c94-14614aaefe91',NULL,NULL,'master','9e586e10-7d4f-4ad6-977a-fd05dffd6ee6',1,40,'','9f0eb3c3-59a5-43ae-aa6c-fc70b158ea7a',NULL),('681e8bca-e83a-410f-8ede-bedbd406bf5f',NULL,NULL,'master','3e2f3567-7f57-4dfd-95b1-90a6fff2e79c',2,20,'','12af5db4-211e-40f6-9805-89e2b48a0b50',NULL),('6a3101e1-a32f-4545-872c-59160b678493',NULL,'registration-user-creation','master','8a75833c-68ef-4248-9602-5781e764bb55',0,20,'\0',NULL,NULL),('6b937d13-a35f-4ff2-bfe5-b7efe66def55',NULL,'reset-credential-email','master','9e586e10-7d4f-4ad6-977a-fd05dffd6ee6',0,20,'\0',NULL,NULL),('7be2fcbe-d927-4534-a87c-e83ca5e2d015',NULL,NULL,'master','7083d693-5892-42b4-a592-f63c604dd8dc',1,30,'','124c5389-5d37-4d77-a226-f39350c568e9',NULL),('7d145ace-a68b-4b81-ae70-5c49cec28d7c',NULL,NULL,'master','3dc0e0a6-fec2-47b9-b48d-e31bf31a164d',0,20,'','5157a4d6-247e-447e-bb54-12e9136c3dc4',NULL),('8152d141-281c-4bf2-a62c-814a89209e7d',NULL,'direct-grant-validate-username','master','7083d693-5892-42b4-a592-f63c604dd8dc',0,10,'\0',NULL,NULL),('87b696c3-9d25-4ba3-83a3-39b849c0629f',NULL,'basic-auth','master','1bd473b0-79f0-46a5-a47c-fc9fad029026',0,10,'\0',NULL,NULL),('87ea521b-38be-490a-9c91-ba9d97d39583',NULL,'auth-spnego','master','1bd473b0-79f0-46a5-a47c-fc9fad029026',3,30,'\0',NULL,NULL),('898c48ac-05f7-46c1-9fd1-a4a5618ddebb',NULL,'registration-page-form','master','a1a80a32-2bfd-4d72-9702-b41896868e69',0,10,'','8a75833c-68ef-4248-9602-5781e764bb55',NULL),('8d478b03-61d0-4c84-bd80-b196e083455d',NULL,'auth-otp-form','master','2a4ea6d6-302a-44dc-8b3d-e57dad183c9f',0,20,'\0',NULL,NULL),('96954bdc-28b2-4234-b6e8-98b3d97c9ab8',NULL,'client-secret','master','d180151b-30b4-4438-8ec3-9033d2e60c38',2,10,'\0',NULL,NULL),('a0a1ccbd-2245-4ba6-b4fd-ce08c3353535',NULL,'http-basic-authenticator','master','5fdff352-e4e5-4dd0-8eb9-a4c8462cb539',0,10,'\0',NULL,NULL),('a1509e28-8554-4ce1-85b5-bb44d0c0a4fe',NULL,'idp-confirm-link','master','a6df1ac7-e51e-416c-afc2-3fb82e5594c3',0,10,'\0',NULL,NULL),('a15b00a9-e28f-41d4-89bf-32bac833e6d5',NULL,'no-cookie-redirect','master','6137a5dc-223b-4817-aed0-4b33b126164d',0,10,'\0',NULL,NULL),('a168dedd-ade1-437d-93f1-a56ea68e27ee',NULL,NULL,'master','a6df1ac7-e51e-416c-afc2-3fb82e5594c3',0,20,'','3e2f3567-7f57-4dfd-95b1-90a6fff2e79c',NULL),('a4c4e437-9d6e-4639-b81d-55be2b3f0b0e',NULL,'idp-username-password-form','master','12af5db4-211e-40f6-9805-89e2b48a0b50',0,10,'\0',NULL,NULL),('adfc31b0-9a5a-4137-aff6-71c7800c2aed',NULL,'direct-grant-validate-password','master','7083d693-5892-42b4-a592-f63c604dd8dc',0,20,'\0',NULL,NULL),('b62a0854-bf37-4886-9e9f-0a4b14cdc1b1',NULL,NULL,'master','12af5db4-211e-40f6-9805-89e2b48a0b50',1,20,'','2a4ea6d6-302a-44dc-8b3d-e57dad183c9f',NULL),('bdc22f68-2bb6-422b-ac66-d4fbb52592fb',NULL,'auth-username-password-form','master','7d35ce2d-9375-42b1-ba30-249cdbbb1944',0,10,'\0',NULL,NULL),('c3c7c44b-e5bf-4929-873e-7a8f17659089',NULL,NULL,'master','7d35ce2d-9375-42b1-ba30-249cdbbb1944',1,20,'','b9691c0e-fca4-43f2-bca1-40558f346594',NULL),('c86853bf-6edb-4e93-9499-c779186f4c63',NULL,'reset-credentials-choose-user','master','9e586e10-7d4f-4ad6-977a-fd05dffd6ee6',0,10,'\0',NULL,NULL),('d6afca0a-e5f2-4532-b99e-4e73b9eb0545',NULL,'basic-auth-otp','master','1bd473b0-79f0-46a5-a47c-fc9fad029026',3,20,'\0',NULL,NULL),('d9a09526-9cb4-4fea-8c2c-d6dde18cfcbe',NULL,'client-secret-jwt','master','d180151b-30b4-4438-8ec3-9033d2e60c38',2,30,'\0',NULL,NULL),('da6090f5-1109-4f24-a199-dec7dcdb202e',NULL,'conditional-user-configured','master','2a4ea6d6-302a-44dc-8b3d-e57dad183c9f',0,10,'\0',NULL,NULL),('df617d17-ac56-46cc-bb90-b70211881606',NULL,'reset-password','master','9e586e10-7d4f-4ad6-977a-fd05dffd6ee6',0,30,'\0',NULL,NULL),('e50e9d13-59af-4a2b-9cf6-10b64d6583de',NULL,'conditional-user-configured','master','9f0eb3c3-59a5-43ae-aa6c-fc70b158ea7a',0,10,'\0',NULL,NULL),('f18aa542-0611-4ec3-8045-3c2084d97c35',NULL,'docker-http-basic-authenticator','master','df0e3e20-eeea-4996-ba53-8ec117fcbec1',0,10,'\0',NULL,NULL),('f486afcd-90a1-469a-8355-6e10b4abe9c1',NULL,'registration-recaptcha-action','master','8a75833c-68ef-4248-9602-5781e764bb55',3,60,'\0',NULL,NULL);
/*!40000 ALTER TABLE `AUTHENTICATION_EXECUTION` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `AUTHENTICATION_FLOW`
--

DROP TABLE IF EXISTS `AUTHENTICATION_FLOW`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `AUTHENTICATION_FLOW` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ALIAS` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb3 DEFAULT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `PROVIDER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'basic-flow',
  `TOP_LEVEL` bit(1) NOT NULL DEFAULT b'0',
  `BUILT_IN` bit(1) NOT NULL DEFAULT b'0',
  PRIMARY KEY (`ID`),
  KEY `IDX_AUTH_FLOW_REALM` (`REALM_ID`),
  CONSTRAINT `FK_AUTH_FLOW_REALM` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `AUTHENTICATION_FLOW`
--

LOCK TABLES `AUTHENTICATION_FLOW` WRITE;
/*!40000 ALTER TABLE `AUTHENTICATION_FLOW` DISABLE KEYS */;
INSERT INTO `AUTHENTICATION_FLOW` VALUES ('124c5389-5d37-4d77-a226-f39350c568e9','Direct Grant - Conditional OTP','Flow to determine if the OTP is required for the authentication','master','basic-flow','\0',''),('12af5db4-211e-40f6-9805-89e2b48a0b50','Verify Existing Account by Re-authentication','Reauthentication of existing account','master','basic-flow','\0',''),('1bd473b0-79f0-46a5-a47c-fc9fad029026','Authentication Options','Authentication options.','master','basic-flow','\0',''),('2a4ea6d6-302a-44dc-8b3d-e57dad183c9f','First broker login - Conditional OTP','Flow to determine if the OTP is required for the authentication','master','basic-flow','\0',''),('3dc0e0a6-fec2-47b9-b48d-e31bf31a164d','first broker login','Actions taken after first broker login with identity provider account, which is not yet linked to any Keycloak account','master','basic-flow','',''),('3e2f3567-7f57-4dfd-95b1-90a6fff2e79c','Account verification options','Method with which to verity the existing account','master','basic-flow','\0',''),('5157a4d6-247e-447e-bb54-12e9136c3dc4','User creation or linking','Flow for the existing/non-existing user alternatives','master','basic-flow','\0',''),('5adc5270-4510-48ee-898c-cbae8f28a3cd','browser','browser based authentication','master','basic-flow','',''),('5fdff352-e4e5-4dd0-8eb9-a4c8462cb539','saml ecp','SAML ECP Profile Authentication Flow','master','basic-flow','',''),('6137a5dc-223b-4817-aed0-4b33b126164d','http challenge','An authentication flow based on challenge-response HTTP Authentication Schemes','master','basic-flow','',''),('7083d693-5892-42b4-a592-f63c604dd8dc','direct grant','OpenID Connect Resource Owner Grant','master','basic-flow','',''),('7d35ce2d-9375-42b1-ba30-249cdbbb1944','forms','Username, password, otp and other auth forms.','master','basic-flow','\0',''),('8a75833c-68ef-4248-9602-5781e764bb55','registration form','registration form','master','form-flow','\0',''),('9e586e10-7d4f-4ad6-977a-fd05dffd6ee6','reset credentials','Reset credentials for a user if they forgot their password or something','master','basic-flow','',''),('9f0eb3c3-59a5-43ae-aa6c-fc70b158ea7a','Reset - Conditional OTP','Flow to determine if the OTP should be reset or not. Set to REQUIRED to force.','master','basic-flow','\0',''),('a1a80a32-2bfd-4d72-9702-b41896868e69','registration','registration flow','master','basic-flow','',''),('a6df1ac7-e51e-416c-afc2-3fb82e5594c3','Handle Existing Account','Handle what to do if there is existing account with same email/username like authenticated identity provider','master','basic-flow','\0',''),('b9691c0e-fca4-43f2-bca1-40558f346594','Browser - Conditional OTP','Flow to determine if the OTP is required for the authentication','master','basic-flow','\0',''),('d180151b-30b4-4438-8ec3-9033d2e60c38','clients','Base authentication for clients','master','client-flow','',''),('df0e3e20-eeea-4996-ba53-8ec117fcbec1','docker auth','Used by Docker clients to authenticate against the IDP','master','basic-flow','','');
/*!40000 ALTER TABLE `AUTHENTICATION_FLOW` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `AUTHENTICATOR_CONFIG`
--

DROP TABLE IF EXISTS `AUTHENTICATOR_CONFIG`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `AUTHENTICATOR_CONFIG` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ALIAS` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_AUTH_CONFIG_REALM` (`REALM_ID`),
  CONSTRAINT `FK_AUTH_REALM` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `AUTHENTICATOR_CONFIG`
--

LOCK TABLES `AUTHENTICATOR_CONFIG` WRITE;
/*!40000 ALTER TABLE `AUTHENTICATOR_CONFIG` DISABLE KEYS */;
INSERT INTO `AUTHENTICATOR_CONFIG` VALUES ('98d35458-7329-4099-9c58-c12984adf496','review profile config','master'),('d3806a77-7749-48bb-bbd8-0981d7bad74f','create unique user config','master');
/*!40000 ALTER TABLE `AUTHENTICATOR_CONFIG` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `AUTHENTICATOR_CONFIG_ENTRY`
--

DROP TABLE IF EXISTS `AUTHENTICATOR_CONFIG_ENTRY`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `AUTHENTICATOR_CONFIG_ENTRY` (
  `AUTHENTICATOR_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` longtext COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`AUTHENTICATOR_ID`,`NAME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `AUTHENTICATOR_CONFIG_ENTRY`
--

LOCK TABLES `AUTHENTICATOR_CONFIG_ENTRY` WRITE;
/*!40000 ALTER TABLE `AUTHENTICATOR_CONFIG_ENTRY` DISABLE KEYS */;
INSERT INTO `AUTHENTICATOR_CONFIG_ENTRY` VALUES ('98d35458-7329-4099-9c58-c12984adf496','missing','update.profile.on.first.login'),('d3806a77-7749-48bb-bbd8-0981d7bad74f','false','require.password.update.after.registration');
/*!40000 ALTER TABLE `AUTHENTICATOR_CONFIG_ENTRY` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `BROKER_LINK`
--

DROP TABLE IF EXISTS `BROKER_LINK`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `BROKER_LINK` (
  `IDENTITY_PROVIDER` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `STORAGE_PROVIDER_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `BROKER_USER_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `BROKER_USERNAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `TOKEN` text COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `USER_ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`IDENTITY_PROVIDER`,`USER_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `BROKER_LINK`
--

LOCK TABLES `BROKER_LINK` WRITE;
/*!40000 ALTER TABLE `BROKER_LINK` DISABLE KEYS */;
/*!40000 ALTER TABLE `BROKER_LINK` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT`
--

DROP TABLE IF EXISTS `CLIENT`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ENABLED` bit(1) NOT NULL DEFAULT b'0',
  `FULL_SCOPE_ALLOWED` bit(1) NOT NULL DEFAULT b'0',
  `CLIENT_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NOT_BEFORE` int(11) DEFAULT NULL,
  `PUBLIC_CLIENT` bit(1) NOT NULL DEFAULT b'0',
  `SECRET` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `BASE_URL` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `BEARER_ONLY` bit(1) NOT NULL DEFAULT b'0',
  `MANAGEMENT_URL` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `SURROGATE_AUTH_REQUIRED` bit(1) NOT NULL DEFAULT b'0',
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `PROTOCOL` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NODE_REREG_TIMEOUT` int(11) DEFAULT 0,
  `FRONTCHANNEL_LOGOUT` bit(1) NOT NULL DEFAULT b'0',
  `CONSENT_REQUIRED` bit(1) NOT NULL DEFAULT b'0',
  `NAME` varchar(255) CHARACTER SET utf8mb3 DEFAULT NULL,
  `SERVICE_ACCOUNTS_ENABLED` bit(1) NOT NULL DEFAULT b'0',
  `CLIENT_AUTHENTICATOR_TYPE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ROOT_URL` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb3 DEFAULT NULL,
  `REGISTRATION_TOKEN` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `STANDARD_FLOW_ENABLED` bit(1) NOT NULL DEFAULT b'1',
  `IMPLICIT_FLOW_ENABLED` bit(1) NOT NULL DEFAULT b'0',
  `DIRECT_ACCESS_GRANTS_ENABLED` bit(1) NOT NULL DEFAULT b'0',
  `ALWAYS_DISPLAY_IN_CONSOLE` bit(1) NOT NULL DEFAULT b'0',
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_B71CJLBENV945RB6GCON438AT` (`REALM_ID`,`CLIENT_ID`),
  KEY `IDX_CLIENT_ID` (`CLIENT_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT`
--

LOCK TABLES `CLIENT` WRITE;
/*!40000 ALTER TABLE `CLIENT` DISABLE KEYS */;
INSERT INTO `CLIENT` VALUES ('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','','\0','account',0,'',NULL,'/realms/master/account/','\0',NULL,'\0','master','openid-connect',0,'\0','\0','${client_account}','\0','client-secret','${authBaseUrl}',NULL,NULL,'','\0','\0','\0'),('54905dd0-4ade-494e-9c35-ab2d445a99f5','','\0','account-console',0,'',NULL,'/realms/master/account/','\0',NULL,'\0','master','openid-connect',0,'\0','\0','${client_account-console}','\0','client-secret','${authBaseUrl}',NULL,NULL,'','\0','\0','\0'),('5a059221-51fd-434f-84a6-40fa51cda5ce','','','photoprism-develop',0,'\0','9d8351a0-ca01-4556-9c37-85eb634869b9',NULL,'\0','https://app.localssl.dev/','\0','master','openid-connect',-1,'\0','\0','PhotoPrism','\0','client-secret','https://app.localssl.dev/',NULL,NULL,'','\0','','\0'),('5b62e4f6-f646-4e0b-aa07-83a17a324137','','\0','broker',0,'\0',NULL,NULL,'',NULL,'\0','master','openid-connect',0,'\0','\0','${client_broker}','\0','client-secret',NULL,NULL,NULL,'','\0','\0','\0'),('8a6bade2-ad19-45f1-9923-b357684d765c','','\0','admin-cli',0,'',NULL,NULL,'\0',NULL,'\0','master','openid-connect',0,'\0','\0','${client_admin-cli}','\0','client-secret',NULL,NULL,NULL,'\0','\0','','\0'),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','','\0','security-admin-console',0,'',NULL,'/admin/master/console/','\0',NULL,'\0','master','openid-connect',0,'\0','\0','${client_security-admin-console}','\0','client-secret','${authAdminUrl}',NULL,NULL,'','\0','\0','\0'),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','','\0','master-realm',0,'\0',NULL,NULL,'',NULL,'\0','master',NULL,0,'\0','\0','master Realm','\0','client-secret',NULL,NULL,NULL,'','\0','\0','\0');
/*!40000 ALTER TABLE `CLIENT` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT_ATTRIBUTES`
--

DROP TABLE IF EXISTS `CLIENT_ATTRIBUTES`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT_ATTRIBUTES` (
  `CLIENT_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` text COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`CLIENT_ID`,`NAME`),
  KEY `IDX_CLIENT_ATT_BY_NAME_VALUE` (`NAME`,`VALUE`(255)),
  CONSTRAINT `FK3C47C64BEACCA966` FOREIGN KEY (`CLIENT_ID`) REFERENCES `CLIENT` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT_ATTRIBUTES`
--

LOCK TABLES `CLIENT_ATTRIBUTES` WRITE;
/*!40000 ALTER TABLE `CLIENT_ATTRIBUTES` DISABLE KEYS */;
INSERT INTO `CLIENT_ATTRIBUTES` VALUES ('54905dd0-4ade-494e-9c35-ab2d445a99f5','S256','pkce.code.challenge.method'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','backchannel.logout.revoke.offline.tokens'),('5a059221-51fd-434f-84a6-40fa51cda5ce','true','backchannel.logout.session.required'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','client_credentials.use_refresh_token'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','display.on.consent.screen'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','exclude.session.state.from.auth.response'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','id.token.as.detached.signature'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','oauth2.device.authorization.grant.enabled'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','oidc.ciba.grant.enabled'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','require.pushed.authorization.requests'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','saml_force_name_id_format'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','saml.artifact.binding'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','saml.assertion.signature'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','saml.authnstatement'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','saml.client.signature'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','saml.encrypt'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','saml.force.post.binding'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','saml.multivalued.roles'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','saml.onetimeuse.condition'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','saml.server.signature'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','saml.server.signature.keyinfo.ext'),('5a059221-51fd-434f-84a6-40fa51cda5ce','false','tls.client.certificate.bound.access.tokens'),('5a059221-51fd-434f-84a6-40fa51cda5ce','true','use.refresh.tokens'),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','S256','pkce.code.challenge.method');
/*!40000 ALTER TABLE `CLIENT_ATTRIBUTES` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT_AUTH_FLOW_BINDINGS`
--

DROP TABLE IF EXISTS `CLIENT_AUTH_FLOW_BINDINGS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT_AUTH_FLOW_BINDINGS` (
  `CLIENT_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `FLOW_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `BINDING_NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`CLIENT_ID`,`BINDING_NAME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT_AUTH_FLOW_BINDINGS`
--

LOCK TABLES `CLIENT_AUTH_FLOW_BINDINGS` WRITE;
/*!40000 ALTER TABLE `CLIENT_AUTH_FLOW_BINDINGS` DISABLE KEYS */;
/*!40000 ALTER TABLE `CLIENT_AUTH_FLOW_BINDINGS` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT_INITIAL_ACCESS`
--

DROP TABLE IF EXISTS `CLIENT_INITIAL_ACCESS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT_INITIAL_ACCESS` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `TIMESTAMP` int(11) DEFAULT NULL,
  `EXPIRATION` int(11) DEFAULT NULL,
  `COUNT` int(11) DEFAULT NULL,
  `REMAINING_COUNT` int(11) DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_CLIENT_INIT_ACC_REALM` (`REALM_ID`),
  CONSTRAINT `FK_CLIENT_INIT_ACC_REALM` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT_INITIAL_ACCESS`
--

LOCK TABLES `CLIENT_INITIAL_ACCESS` WRITE;
/*!40000 ALTER TABLE `CLIENT_INITIAL_ACCESS` DISABLE KEYS */;
/*!40000 ALTER TABLE `CLIENT_INITIAL_ACCESS` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT_NODE_REGISTRATIONS`
--

DROP TABLE IF EXISTS `CLIENT_NODE_REGISTRATIONS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT_NODE_REGISTRATIONS` (
  `CLIENT_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` int(11) DEFAULT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`CLIENT_ID`,`NAME`),
  CONSTRAINT `FK4129723BA992F594` FOREIGN KEY (`CLIENT_ID`) REFERENCES `CLIENT` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT_NODE_REGISTRATIONS`
--

LOCK TABLES `CLIENT_NODE_REGISTRATIONS` WRITE;
/*!40000 ALTER TABLE `CLIENT_NODE_REGISTRATIONS` DISABLE KEYS */;
/*!40000 ALTER TABLE `CLIENT_NODE_REGISTRATIONS` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT_SCOPE`
--

DROP TABLE IF EXISTS `CLIENT_SCOPE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT_SCOPE` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb3 DEFAULT NULL,
  `PROTOCOL` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_CLI_SCOPE` (`REALM_ID`,`NAME`),
  KEY `IDX_REALM_CLSCOPE` (`REALM_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT_SCOPE`
--

LOCK TABLES `CLIENT_SCOPE` WRITE;
/*!40000 ALTER TABLE `CLIENT_SCOPE` DISABLE KEYS */;
INSERT INTO `CLIENT_SCOPE` VALUES ('13052fde-d239-4154-b80b-0f406ed76ded','phone','master','OpenID Connect built-in scope: phone','openid-connect'),('395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','offline_access','master','OpenID Connect built-in scope: offline_access','openid-connect'),('a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb','profile','master','OpenID Connect built-in scope: profile','openid-connect'),('abde17dd-48e0-4d26-a2b7-e75c04b1ac7f','microprofile-jwt','master','Microprofile - JWT built-in scope','openid-connect'),('c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','address','master','OpenID Connect built-in scope: address','openid-connect'),('cb25d275-eff3-4655-b032-e163a0a23c0f','email','master','OpenID Connect built-in scope: email','openid-connect'),('e4f019f4-8a8a-4682-bf50-8e883c89cd03','web-origins','master','OpenID Connect scope for add allowed web origins to the access token','openid-connect'),('f0e07760-3d3d-45d5-b651-403f8b19de35','roles','master','OpenID Connect scope for add user roles to the access token','openid-connect'),('f55ceb89-6d3c-4bcb-882e-44c498d8b305','role_list','master','SAML role list','saml');
/*!40000 ALTER TABLE `CLIENT_SCOPE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT_SCOPE_ATTRIBUTES`
--

DROP TABLE IF EXISTS `CLIENT_SCOPE_ATTRIBUTES`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT_SCOPE_ATTRIBUTES` (
  `SCOPE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` text COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`SCOPE_ID`,`NAME`),
  KEY `IDX_CLSCOPE_ATTRS` (`SCOPE_ID`),
  CONSTRAINT `FK_CL_SCOPE_ATTR_SCOPE` FOREIGN KEY (`SCOPE_ID`) REFERENCES `CLIENT_SCOPE` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT_SCOPE_ATTRIBUTES`
--

LOCK TABLES `CLIENT_SCOPE_ATTRIBUTES` WRITE;
/*!40000 ALTER TABLE `CLIENT_SCOPE_ATTRIBUTES` DISABLE KEYS */;
INSERT INTO `CLIENT_SCOPE_ATTRIBUTES` VALUES ('13052fde-d239-4154-b80b-0f406ed76ded','${phoneScopeConsentText}','consent.screen.text'),('13052fde-d239-4154-b80b-0f406ed76ded','true','display.on.consent.screen'),('13052fde-d239-4154-b80b-0f406ed76ded','true','include.in.token.scope'),('395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','${offlineAccessScopeConsentText}','consent.screen.text'),('395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','true','display.on.consent.screen'),('a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb','${profileScopeConsentText}','consent.screen.text'),('a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb','true','display.on.consent.screen'),('a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb','true','include.in.token.scope'),('abde17dd-48e0-4d26-a2b7-e75c04b1ac7f','false','display.on.consent.screen'),('abde17dd-48e0-4d26-a2b7-e75c04b1ac7f','true','include.in.token.scope'),('c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','${addressScopeConsentText}','consent.screen.text'),('c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','true','display.on.consent.screen'),('c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','true','include.in.token.scope'),('cb25d275-eff3-4655-b032-e163a0a23c0f','${emailScopeConsentText}','consent.screen.text'),('cb25d275-eff3-4655-b032-e163a0a23c0f','true','display.on.consent.screen'),('cb25d275-eff3-4655-b032-e163a0a23c0f','true','include.in.token.scope'),('e4f019f4-8a8a-4682-bf50-8e883c89cd03','','consent.screen.text'),('e4f019f4-8a8a-4682-bf50-8e883c89cd03','false','display.on.consent.screen'),('e4f019f4-8a8a-4682-bf50-8e883c89cd03','false','include.in.token.scope'),('f0e07760-3d3d-45d5-b651-403f8b19de35','${rolesScopeConsentText}','consent.screen.text'),('f0e07760-3d3d-45d5-b651-403f8b19de35','true','display.on.consent.screen'),('f0e07760-3d3d-45d5-b651-403f8b19de35','false','include.in.token.scope'),('f55ceb89-6d3c-4bcb-882e-44c498d8b305','${samlRoleListScopeConsentText}','consent.screen.text'),('f55ceb89-6d3c-4bcb-882e-44c498d8b305','true','display.on.consent.screen');
/*!40000 ALTER TABLE `CLIENT_SCOPE_ATTRIBUTES` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT_SCOPE_CLIENT`
--

DROP TABLE IF EXISTS `CLIENT_SCOPE_CLIENT`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT_SCOPE_CLIENT` (
  `CLIENT_ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `SCOPE_ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `DEFAULT_SCOPE` bit(1) NOT NULL DEFAULT b'0',
  PRIMARY KEY (`CLIENT_ID`,`SCOPE_ID`),
  KEY `IDX_CLSCOPE_CL` (`CLIENT_ID`),
  KEY `IDX_CL_CLSCOPE` (`SCOPE_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT_SCOPE_CLIENT`
--

LOCK TABLES `CLIENT_SCOPE_CLIENT` WRITE;
/*!40000 ALTER TABLE `CLIENT_SCOPE_CLIENT` DISABLE KEYS */;
INSERT INTO `CLIENT_SCOPE_CLIENT` VALUES ('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','13052fde-d239-4154-b80b-0f406ed76ded','\0'),('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','\0'),('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',''),('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f','\0'),('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','\0'),('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','cb25d275-eff3-4655-b032-e163a0a23c0f',''),('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','e4f019f4-8a8a-4682-bf50-8e883c89cd03',''),('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','f0e07760-3d3d-45d5-b651-403f8b19de35',''),('54905dd0-4ade-494e-9c35-ab2d445a99f5','13052fde-d239-4154-b80b-0f406ed76ded','\0'),('54905dd0-4ade-494e-9c35-ab2d445a99f5','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','\0'),('54905dd0-4ade-494e-9c35-ab2d445a99f5','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',''),('54905dd0-4ade-494e-9c35-ab2d445a99f5','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f','\0'),('54905dd0-4ade-494e-9c35-ab2d445a99f5','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','\0'),('54905dd0-4ade-494e-9c35-ab2d445a99f5','cb25d275-eff3-4655-b032-e163a0a23c0f',''),('54905dd0-4ade-494e-9c35-ab2d445a99f5','e4f019f4-8a8a-4682-bf50-8e883c89cd03',''),('54905dd0-4ade-494e-9c35-ab2d445a99f5','f0e07760-3d3d-45d5-b651-403f8b19de35',''),('5a059221-51fd-434f-84a6-40fa51cda5ce','13052fde-d239-4154-b80b-0f406ed76ded','\0'),('5a059221-51fd-434f-84a6-40fa51cda5ce','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','\0'),('5a059221-51fd-434f-84a6-40fa51cda5ce','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',''),('5a059221-51fd-434f-84a6-40fa51cda5ce','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f','\0'),('5a059221-51fd-434f-84a6-40fa51cda5ce','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','\0'),('5a059221-51fd-434f-84a6-40fa51cda5ce','cb25d275-eff3-4655-b032-e163a0a23c0f',''),('5a059221-51fd-434f-84a6-40fa51cda5ce','e4f019f4-8a8a-4682-bf50-8e883c89cd03',''),('5a059221-51fd-434f-84a6-40fa51cda5ce','f0e07760-3d3d-45d5-b651-403f8b19de35',''),('5b62e4f6-f646-4e0b-aa07-83a17a324137','13052fde-d239-4154-b80b-0f406ed76ded','\0'),('5b62e4f6-f646-4e0b-aa07-83a17a324137','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','\0'),('5b62e4f6-f646-4e0b-aa07-83a17a324137','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',''),('5b62e4f6-f646-4e0b-aa07-83a17a324137','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f','\0'),('5b62e4f6-f646-4e0b-aa07-83a17a324137','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','\0'),('5b62e4f6-f646-4e0b-aa07-83a17a324137','cb25d275-eff3-4655-b032-e163a0a23c0f',''),('5b62e4f6-f646-4e0b-aa07-83a17a324137','e4f019f4-8a8a-4682-bf50-8e883c89cd03',''),('5b62e4f6-f646-4e0b-aa07-83a17a324137','f0e07760-3d3d-45d5-b651-403f8b19de35',''),('8a6bade2-ad19-45f1-9923-b357684d765c','13052fde-d239-4154-b80b-0f406ed76ded','\0'),('8a6bade2-ad19-45f1-9923-b357684d765c','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','\0'),('8a6bade2-ad19-45f1-9923-b357684d765c','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',''),('8a6bade2-ad19-45f1-9923-b357684d765c','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f','\0'),('8a6bade2-ad19-45f1-9923-b357684d765c','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','\0'),('8a6bade2-ad19-45f1-9923-b357684d765c','cb25d275-eff3-4655-b032-e163a0a23c0f',''),('8a6bade2-ad19-45f1-9923-b357684d765c','e4f019f4-8a8a-4682-bf50-8e883c89cd03',''),('8a6bade2-ad19-45f1-9923-b357684d765c','f0e07760-3d3d-45d5-b651-403f8b19de35',''),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','13052fde-d239-4154-b80b-0f406ed76ded','\0'),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','\0'),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',''),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f','\0'),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','\0'),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','cb25d275-eff3-4655-b032-e163a0a23c0f',''),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','e4f019f4-8a8a-4682-bf50-8e883c89cd03',''),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','f0e07760-3d3d-45d5-b651-403f8b19de35',''),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','13052fde-d239-4154-b80b-0f406ed76ded','\0'),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','\0'),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',''),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f','\0'),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','\0'),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','cb25d275-eff3-4655-b032-e163a0a23c0f',''),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','e4f019f4-8a8a-4682-bf50-8e883c89cd03',''),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','f0e07760-3d3d-45d5-b651-403f8b19de35','');
/*!40000 ALTER TABLE `CLIENT_SCOPE_CLIENT` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT_SCOPE_ROLE_MAPPING`
--

DROP TABLE IF EXISTS `CLIENT_SCOPE_ROLE_MAPPING`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT_SCOPE_ROLE_MAPPING` (
  `SCOPE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ROLE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`SCOPE_ID`,`ROLE_ID`),
  KEY `IDX_CLSCOPE_ROLE` (`SCOPE_ID`),
  KEY `IDX_ROLE_CLSCOPE` (`ROLE_ID`),
  CONSTRAINT `FK_CL_SCOPE_RM_SCOPE` FOREIGN KEY (`SCOPE_ID`) REFERENCES `CLIENT_SCOPE` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT_SCOPE_ROLE_MAPPING`
--

LOCK TABLES `CLIENT_SCOPE_ROLE_MAPPING` WRITE;
/*!40000 ALTER TABLE `CLIENT_SCOPE_ROLE_MAPPING` DISABLE KEYS */;
INSERT INTO `CLIENT_SCOPE_ROLE_MAPPING` VALUES ('395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','e06c7506-138d-4968-9186-cd958b29e577');
/*!40000 ALTER TABLE `CLIENT_SCOPE_ROLE_MAPPING` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT_SESSION`
--

DROP TABLE IF EXISTS `CLIENT_SESSION`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT_SESSION` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `CLIENT_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REDIRECT_URI` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `STATE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `TIMESTAMP` int(11) DEFAULT NULL,
  `SESSION_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `AUTH_METHOD` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REALM_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `AUTH_USER_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CURRENT_ACTION` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_CLIENT_SESSION_SESSION` (`SESSION_ID`),
  CONSTRAINT `FK_B4AO2VCVAT6UKAU74WBWTFQO1` FOREIGN KEY (`SESSION_ID`) REFERENCES `USER_SESSION` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT_SESSION`
--

LOCK TABLES `CLIENT_SESSION` WRITE;
/*!40000 ALTER TABLE `CLIENT_SESSION` DISABLE KEYS */;
/*!40000 ALTER TABLE `CLIENT_SESSION` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT_SESSION_AUTH_STATUS`
--

DROP TABLE IF EXISTS `CLIENT_SESSION_AUTH_STATUS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT_SESSION_AUTH_STATUS` (
  `AUTHENTICATOR` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `STATUS` int(11) DEFAULT NULL,
  `CLIENT_SESSION` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`CLIENT_SESSION`,`AUTHENTICATOR`),
  CONSTRAINT `AUTH_STATUS_CONSTRAINT` FOREIGN KEY (`CLIENT_SESSION`) REFERENCES `CLIENT_SESSION` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT_SESSION_AUTH_STATUS`
--

LOCK TABLES `CLIENT_SESSION_AUTH_STATUS` WRITE;
/*!40000 ALTER TABLE `CLIENT_SESSION_AUTH_STATUS` DISABLE KEYS */;
/*!40000 ALTER TABLE `CLIENT_SESSION_AUTH_STATUS` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT_SESSION_NOTE`
--

DROP TABLE IF EXISTS `CLIENT_SESSION_NOTE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT_SESSION_NOTE` (
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CLIENT_SESSION` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`CLIENT_SESSION`,`NAME`),
  CONSTRAINT `FK5EDFB00FF51C2736` FOREIGN KEY (`CLIENT_SESSION`) REFERENCES `CLIENT_SESSION` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT_SESSION_NOTE`
--

LOCK TABLES `CLIENT_SESSION_NOTE` WRITE;
/*!40000 ALTER TABLE `CLIENT_SESSION_NOTE` DISABLE KEYS */;
/*!40000 ALTER TABLE `CLIENT_SESSION_NOTE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT_SESSION_PROT_MAPPER`
--

DROP TABLE IF EXISTS `CLIENT_SESSION_PROT_MAPPER`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT_SESSION_PROT_MAPPER` (
  `PROTOCOL_MAPPER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `CLIENT_SESSION` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`CLIENT_SESSION`,`PROTOCOL_MAPPER_ID`),
  CONSTRAINT `FK_33A8SGQW18I532811V7O2DK89` FOREIGN KEY (`CLIENT_SESSION`) REFERENCES `CLIENT_SESSION` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT_SESSION_PROT_MAPPER`
--

LOCK TABLES `CLIENT_SESSION_PROT_MAPPER` WRITE;
/*!40000 ALTER TABLE `CLIENT_SESSION_PROT_MAPPER` DISABLE KEYS */;
/*!40000 ALTER TABLE `CLIENT_SESSION_PROT_MAPPER` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT_SESSION_ROLE`
--

DROP TABLE IF EXISTS `CLIENT_SESSION_ROLE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT_SESSION_ROLE` (
  `ROLE_ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `CLIENT_SESSION` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`CLIENT_SESSION`,`ROLE_ID`),
  CONSTRAINT `FK_11B7SGQW18I532811V7O2DV76` FOREIGN KEY (`CLIENT_SESSION`) REFERENCES `CLIENT_SESSION` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT_SESSION_ROLE`
--

LOCK TABLES `CLIENT_SESSION_ROLE` WRITE;
/*!40000 ALTER TABLE `CLIENT_SESSION_ROLE` DISABLE KEYS */;
/*!40000 ALTER TABLE `CLIENT_SESSION_ROLE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CLIENT_USER_SESSION_NOTE`
--

DROP TABLE IF EXISTS `CLIENT_USER_SESSION_NOTE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CLIENT_USER_SESSION_NOTE` (
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` text COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CLIENT_SESSION` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`CLIENT_SESSION`,`NAME`),
  CONSTRAINT `FK_CL_USR_SES_NOTE` FOREIGN KEY (`CLIENT_SESSION`) REFERENCES `CLIENT_SESSION` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CLIENT_USER_SESSION_NOTE`
--

LOCK TABLES `CLIENT_USER_SESSION_NOTE` WRITE;
/*!40000 ALTER TABLE `CLIENT_USER_SESSION_NOTE` DISABLE KEYS */;
/*!40000 ALTER TABLE `CLIENT_USER_SESSION_NOTE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `COMPONENT`
--

DROP TABLE IF EXISTS `COMPONENT`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `COMPONENT` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `PARENT_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `PROVIDER_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `PROVIDER_TYPE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `SUB_TYPE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_COMPONENT_REALM` (`REALM_ID`),
  KEY `IDX_COMPONENT_PROVIDER_TYPE` (`PROVIDER_TYPE`),
  CONSTRAINT `FK_COMPONENT_REALM` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `COMPONENT`
--

LOCK TABLES `COMPONENT` WRITE;
/*!40000 ALTER TABLE `COMPONENT` DISABLE KEYS */;
INSERT INTO `COMPONENT` VALUES ('08b208d4-dd60-4eea-9ca1-9213f05b508c','Consent Required','master','consent-required','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','anonymous'),('12fe216f-5e34-43fb-bc92-3c11a8a6abe1','Full Scope Disabled','master','scope','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','anonymous'),('3b786f80-369b-4cc3-a54f-e7efa9dfca00','rsa-generated','master','rsa-generated','org.keycloak.keys.KeyProvider','master',NULL),('41e477a9-2312-4054-9bb2-48c803f200a5','hmac-generated','master','hmac-generated','org.keycloak.keys.KeyProvider','master',NULL),('4bfbde17-6bcf-4542-887f-8adb3ad83893','Max Clients Limit','master','max-clients','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','anonymous'),('522398f4-90ab-4f2b-bdbf-9b8065f6533e','aes-generated','master','aes-generated','org.keycloak.keys.KeyProvider','master',NULL),('59a894bc-218d-4acf-9a44-6d79def59870','Allowed Client Scopes','master','allowed-client-templates','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','anonymous'),('690008b2-bc5a-42f4-82f3-2ff750fe6e9a','Allowed Protocol Mapper Types','master','allowed-protocol-mappers','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','authenticated'),('984d324a-40e0-448a-bef3-11bf65cb9723','rsa-enc-generated','master','rsa-generated','org.keycloak.keys.KeyProvider','master',NULL),('99846733-ac9b-4ec5-8d38-df7b359d016f','Trusted Hosts','master','trusted-hosts','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','anonymous'),('a8274b71-1b85-4cef-bcfc-ae84d37ab194','Allowed Protocol Mapper Types','master','allowed-protocol-mappers','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','anonymous'),('f6858356-dd21-4e56-892e-35b6f776978c','Allowed Client Scopes','master','allowed-client-templates','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','authenticated');
/*!40000 ALTER TABLE `COMPONENT` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `COMPONENT_CONFIG`
--

DROP TABLE IF EXISTS `COMPONENT_CONFIG`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `COMPONENT_CONFIG` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `COMPONENT_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(4000) CHARACTER SET utf8mb3 DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_COMPO_CONFIG_COMPO` (`COMPONENT_ID`),
  CONSTRAINT `FK_COMPONENT_CONFIG` FOREIGN KEY (`COMPONENT_ID`) REFERENCES `COMPONENT` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `COMPONENT_CONFIG`
--

LOCK TABLES `COMPONENT_CONFIG` WRITE;
/*!40000 ALTER TABLE `COMPONENT_CONFIG` DISABLE KEYS */;
INSERT INTO `COMPONENT_CONFIG` VALUES ('025dc769-74da-46ea-992e-651f8ed8b94b','984d324a-40e0-448a-bef3-11bf65cb9723','priority','100'),('050174df-d8b7-43ed-b790-0cf9ee4220d0','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','oidc-address-mapper'),('108d3e53-4f6d-4997-9e28-e6907c2220ac','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','saml-user-property-mapper'),('1245310b-9eb4-4fe6-bad2-ff739d504d58','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','saml-user-attribute-mapper'),('181e98a6-98bb-4d41-b357-7291d4d045a5','4bfbde17-6bcf-4542-887f-8adb3ad83893','max-clients','200'),('1c430fc1-0647-43ef-8db0-e0f5c4bcf771','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','saml-role-list-mapper'),('1e83b262-7ef9-458e-b231-443fd9030279','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','oidc-full-name-mapper'),('2192a5f2-2fd7-4eec-9d32-d6b2ca41c992','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','oidc-full-name-mapper'),('26841fbc-8552-454a-9534-db66a7d2f641','59a894bc-218d-4acf-9a44-6d79def59870','allow-default-scopes','true'),('2bca5418-276f-4310-8ec8-88fd49d92fdf','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','saml-role-list-mapper'),('351d0ecc-97e7-4dda-9aee-fd74bfa4914c','984d324a-40e0-448a-bef3-11bf65cb9723','privateKey','MIIEowIBAAKCAQEAh8ppU5X7uFJ0hBJC78wR6wN6s1qVxL9u1JcwwOocRPoG8Ua1Rsv9g0P9IXJMDEBdqzWY7voMD7wngJchN301tjimgYEfHx10KkX6oWkXuk30oaSTqXwjb19ps8hRvwQWJC6P1Y/SylgwaHved14p4hpLObMOoiFc2PrCH/TXP81T2BQqXSIvfJytdnO9E90ASDX0Hhv5wIr12Kuz4arh5/b78OWWJ9oWNfIkzgQoktbDdMysofij7zQgD8mYxcfQr4DvHetk+dn6x4GJWqhA1CTgHN0aF/sltm7G9bImmsMuVf9obZg77oHcFYcXhWBXbkG3baKtKYfGq6JoYA6CKwIDAQABAoIBAAQ6mIci36EA6GIIk48WQuSXyiV1x75F2/TA9KK9Z736L2cqNZEL30xMPMDi511mT8R6OdYPcXq3+F731e/9dUPEheL4m3iDmU+LuF94f2Ws8dZq4rJfjFb2mLshnPIe9XWRAae7/+uPTYqjeO0swI8rFHaqjeUctuCHBq6qGF4DQmEqAuZK0TMiZxPQlACR9G8WAmnAV66weKlhCshOkVKcSDFoH0NMMJ/KHmVzFo3drwJkKZNXbelv8EsuUnL/Pcj7WZfPcbCbR0M8TRDL7QUWX3CYKjQHgCnj6miA/7ydsUoa7BnJEzmF885ncrqfxfPAVSHuHqG9eFh7NMEG8JECgYEAu9o+TTqoM4ju6R+iww4PW4Ny7/IfyAu8qEcR22XewnyDz+qTOTePlPfEIx6al5Zymsz41pzvliPwQ0tgr49kuuM2GRVV9RzWq2JULCjrxj4tPJbeKd6m5JqlTRru+FiuLma/LXHAf6KU4cfbQQpStYRlP4XROk3F5voX+lDG00kCgYEAuQ02z+00b1o0MzDayHWw64FZ2C0lB47oOQRnx4HGJguL5/S44TaZiOWIqxFv1lqO8hWudmE5kCGUK6WaQDhOqueebMdbLODHn1qyJ7/ZGYCYX0tYRN0tGKvlCGoTjOEyfMLysIwQTXfNtGUYMQf43U0KI5WhduPFoDuRjMKuddMCgYEAqoluld3yZRajDbBSqpFRD9s9tOcyQwGku4AJjgvlNtqjL1XdYcw25R4pSVi3L3a9hBsgrHS8bKkjrXP4ymh7Ic6zhgIAjw0nNV+G2rArm0VG/AJandgr2s0p093npD2dozJTzIXAJB8M2gv92AXvICqZYBmz4CJKz22r5ur+FUECgYAH9yal2psINATNM0wnltFPwdihMohGhANA+QySjOZ/mr2h9WnD3/rJ5r90RaLfwjQm/YHt/I9iwd9D5bP3EbVpK+Eo44fsLZzKIjhK97obm+pzJ6YcCL05M6T/MLm4tbTbo/SYXt8QxphnLHbXHXW76OYH1BgIKxPFquq/+V1TGwKBgA5QZHYLMTtLDakpAZI4rMoSXawyvDEh1vuc+s9jwT1djyQVQbLdlyPW2vO0RYRVVXBi54azrFW+xLL4FavEqbT6sdrrzUl+8TLL7bJ1btIlReQfuuzv1Na/AWc1qJIkt/aWgzYeOiy4wZosnZvckU+vHZe5b74Vn0K0AdMH+tnM'),('37670004-a54c-49ad-a95a-3a93d22301d8','522398f4-90ab-4f2b-bdbf-9b8065f6533e','secret','QYrPTNemXvuh_s8Up4yu_A'),('3d2f9154-e187-40b8-86ce-15b77737c7cb','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','oidc-usermodel-property-mapper'),('3dd3b966-67b9-4c36-879b-664038038329','f6858356-dd21-4e56-892e-35b6f776978c','allow-default-scopes','true'),('4add7b8b-2a3b-4640-a632-7193c366fd57','41e477a9-2312-4054-9bb2-48c803f200a5','algorithm','HS256'),('4d551be7-5fdf-4a30-9f35-97ab28a654e6','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','oidc-usermodel-attribute-mapper'),('4e700315-b68f-45b0-a719-bd62aadcd927','99846733-ac9b-4ec5-8d38-df7b359d016f','host-sending-registration-request-must-match','true'),('540efe8b-23ad-423e-9d12-3493e593664a','984d324a-40e0-448a-bef3-11bf65cb9723','keyUse','enc'),('642ab800-084f-4e68-b8d2-c6d37a24557d','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','oidc-usermodel-attribute-mapper'),('6c387d3e-ce0b-4570-99c3-96d47cce9fef','41e477a9-2312-4054-9bb2-48c803f200a5','secret','wE0La3ld_-jLiWfKAICSC5QOrife31YSvkmWrpm_Fwom4ksV40GWKiuTr8QA_SjPvauBFBfJ2l9JFaFLgWcomw'),('903b4806-0a50-4244-b963-fe2e3838b416','522398f4-90ab-4f2b-bdbf-9b8065f6533e','kid','f59924e8-f425-4c59-8edb-a5d0155619d6'),('9ab02801-5e80-4c01-b711-24633d916450','41e477a9-2312-4054-9bb2-48c803f200a5','priority','100'),('9f41d424-3f2b-4ab0-949b-0e6de10ed573','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','oidc-sha256-pairwise-sub-mapper'),('a25bd806-bbbc-44f7-a43b-da98672c3040','984d324a-40e0-448a-bef3-11bf65cb9723','certificate','MIICmzCCAYMCBgF83P1ndTANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDDAZtYXN0ZXIwHhcNMjExMTAxMTkzMTA3WhcNMzExMTAxMTkzMjQ3WjARMQ8wDQYDVQQDDAZtYXN0ZXIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCHymlTlfu4UnSEEkLvzBHrA3qzWpXEv27UlzDA6hxE+gbxRrVGy/2DQ/0hckwMQF2rNZju+gwPvCeAlyE3fTW2OKaBgR8fHXQqRfqhaRe6TfShpJOpfCNvX2mzyFG/BBYkLo/Vj9LKWDBoe953XiniGks5sw6iIVzY+sIf9Nc/zVPYFCpdIi98nK12c70T3QBINfQeG/nAivXYq7PhquHn9vvw5ZYn2hY18iTOBCiS1sN0zKyh+KPvNCAPyZjFx9CvgO8d62T52frHgYlaqEDUJOAc3RoX+yW2bsb1siaawy5V/2htmDvugdwVhxeFYFduQbdtoq0ph8aromhgDoIrAgMBAAEwDQYJKoZIhvcNAQELBQADggEBABwB04rw1lSu7fpEc+/3zbQXDQSlFjn/UtwTwEitfwiKhRXC8g125wxg0CTzc642RLDIvtghifa+A4P2x/YWuyxwKq7xQG+EroHZ8Lc1gWePXqFVwoT6++146B2tvxG69o2G8xKdxGWafLXd1CFGe3FokRRMXYWTgXJWMuo/EE+3AY61ZPcK8BI0mSASjIz5J2wA0BCEbehxaJ7x8QWaGupqfvLetkwEyPT9s7GTvGKCS9tqSJQRuPveyapybuKWZ80xvrrH3vSDJGhfiplf5G4/X1Ir8FSjwpZ7kMoqz2EfWABldZXRHzdcGfy9w5OnTb/NAk6ULOMNWQCBcILetKU='),('a57bebf1-fc77-4661-9a8f-227d92cd0099','3b786f80-369b-4cc3-a54f-e7efa9dfca00','priority','100'),('a5ed2232-35fc-476a-85b0-dabc6dd4e2fc','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','oidc-usermodel-property-mapper'),('b4dbd7d0-958c-4675-a8ab-b5f7118e4a40','3b786f80-369b-4cc3-a54f-e7efa9dfca00','certificate','MIICmzCCAYMCBgF83P1lIjANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDDAZtYXN0ZXIwHhcNMjExMTAxMTkzMTA2WhcNMzExMTAxMTkzMjQ2WjARMQ8wDQYDVQQDDAZtYXN0ZXIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDJ5qTFqw8c0tPqW5foRjxXei+WKfuudGMY13s2+QpO7qCBBJZTuo9nkj1Ij1PG19Aq6IC/fVNwaf1xk8u4m8pFj1hQKWE64t9glMu3Ew1RbRyrV1RDuu3uPxSV3gUmKbQda11rpgHlSfigouJERCMU79usgZ5URsnyxWQgUxN9iif8Psu0I+DQcc7K8JZrQw+uQkCxWrGuv6O6pnwS7m/pbm3XoV6FLM7mqUxnL+81UVCJRdpCtgtQxgw5A/8UxVBHocEkkH23IFP4R2UMak5quWtgXjMIUmAiOhE0LuHjUi+NWincWJ0/DENuvRc2Ukr//U7h4FeWEhxu2EaFcwQvAgMBAAEwDQYJKoZIhvcNAQELBQADggEBALN4mBahtUhFRFSvIDqF1vhMkri60BIIu6UHy6CivBrs0pQOAOIo/G7STJrtGyDhybAMT4J3FzRXYASIDHGDLSmGqz9gmxVfMt08EEAyM+9Ep7d1FLTiC1iLHqP5BLe9WVVAIro3/w2yqExqj8XJhUa/xpQA42lIu3didUe/13SZ8oiAgLsO55tV/Uz2uGsIfQ8FeImCEkFQNeJQZspc5RT32ARBgGP8RlpgystIbiXCUpq/l90EoNLUJZJZ3iVlCl7SwPY+eqfY86IZl40xM3RmX2U9VVEh9KWhurM04M5up2MiAoleBwjlmY3HCC08aU3lQtzK1guRfDYWy/bXQFA='),('c769fd0f-bb79-4466-ae9c-7b3c403029bd','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','saml-user-property-mapper'),('d477301c-86b9-4fdb-ba3f-7874d7f3fa45','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','saml-user-attribute-mapper'),('e033eb96-3a12-4f88-9402-31aa00f6113a','3b786f80-369b-4cc3-a54f-e7efa9dfca00','keyUse','sig'),('e24bbbfa-620b-44cb-b070-e0c2383555f2','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','oidc-sha256-pairwise-sub-mapper'),('e3eb40d3-45f7-4edf-9347-d41334635763','41e477a9-2312-4054-9bb2-48c803f200a5','kid','7ca5813e-1813-49c6-a1f8-d1ecf58f56e9'),('eef9b1d5-555e-4b75-9e37-17e4d4b91403','522398f4-90ab-4f2b-bdbf-9b8065f6533e','priority','100'),('f5a7c7b4-ce7e-458d-839f-78e660238f6b','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','oidc-address-mapper'),('f5b7f63b-8d3a-4575-a3e6-4346cda1a549','3b786f80-369b-4cc3-a54f-e7efa9dfca00','privateKey','MIIEpQIBAAKCAQEAyeakxasPHNLT6luX6EY8V3ovlin7rnRjGNd7NvkKTu6ggQSWU7qPZ5I9SI9TxtfQKuiAv31TcGn9cZPLuJvKRY9YUClhOuLfYJTLtxMNUW0cq1dUQ7rt7j8Uld4FJim0HWtda6YB5Un4oKLiREQjFO/brIGeVEbJ8sVkIFMTfYon/D7LtCPg0HHOyvCWa0MPrkJAsVqxrr+juqZ8Eu5v6W5t16FehSzO5qlMZy/vNVFQiUXaQrYLUMYMOQP/FMVQR6HBJJB9tyBT+EdlDGpOarlrYF4zCFJgIjoRNC7h41IvjVop3FidPwxDbr0XNlJK//1O4eBXlhIcbthGhXMELwIDAQABAoIBAQCea3I4g5tNE4QiLJJKN+oa/Y2fNvv7i+lB0boljU1wV77q3Q2TTxw8uTuK1qN2r1nwgRScrBqvZwrtdnlwNhWFdQ9nfsCC8wcxAi/CS5m0nXfUXaaJqoAM48QkP9wscKaaOudHky+DmQIUERqXVBtuzzG/7sir+gt1iTqiPm1Zn4v/U7l1eRhIPx/qJ8+4iTw6crRLooUzSHqCirZWs6pRbzAoxd3OBdxr+wXqmzV1+F6HguonhawFnqilaHEf0Lj+N0Lc1vOdmzNSlJ3azV/kRvMBqsDRKPcNIfCnjeNA+gHJd4u++OzX3IDRsJtlyj51+TmYkmMnLOfsWTRHmtfxAoGBAOrZtHJ3ronteZG55DyHN3idDLW8pi783oKSB7HYNJnIXoAupdjsSyLcN3A8EhTRbvAiA0XjfeqbpT8Xg3OLOxEZi1PgndguizutkPOIDeL0Z+p0H4CHdY3p2VWSUTquojlwaOgf7iUvPWdlDD/DM1glfvH05jThg59zhlt0qULHAoGBANwVUH3cVEtnp50oG/LlzECM9eLmsSu/4APyIGLXtk/LMtu2pJPhatKMgtOfXBTrWud0q2E6PwPaDIX15EITcK8gtGIpzDOoHvb6YLwLmxaeUy4M+oHBkR1m/CMb77r5sq5DfOR9UoTr+3tVRPxlk6ES1vqrbsJNZ9N70xzFIctZAoGBAJrsrNYKT7CbYOQaLg8j4BsH91d4MGS02ZBnBv5yMxjzjiufGjcEgfhoL4YxingDRNzSgzg6f1kh/hulxkiVo4x/PmNBvL7czWq77/BHY2nBcz++BP4D3i+VAZMqp70/cLLVjc77KV2MUUSA61iwy5EtgxXYSXi+/9ZTHmH8jqAHAoGBALaGkuAfaGW1TOThC/TyMujiP1d0fkHLe22qVMPFJXWeD8r6+hmPXTnLwQDj7MmIvDazoyMa3IJESBid6zYFy3HjDNdQ1QOOjkfFNY8fjPtASboql2Qf9ktNSxWPKM6IInG2lREnAtYspMAP4wv07nArINJ6dXx+F/rkeh0lPTbZAoGAN11eATq+0KGYoF7lEV72d74roNrg+uKffn8GsP6KlmV8nqubg/QOug7c/tuEa2hqpcA5dYTFNjKqaSluKOfje+U9RTM9/wICmQYn/VBBRzG6l0oO3X5TTefy4Qtd292N1R/px/0wmF6RhlDqIW5H7NNIpki9uwG78IA5jvWknSI='),('f7f2cc71-2362-4e27-94a3-2624238eeb28','99846733-ac9b-4ec5-8d38-df7b359d016f','client-uris-must-match','true');
/*!40000 ALTER TABLE `COMPONENT_CONFIG` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `COMPOSITE_ROLE`
--

DROP TABLE IF EXISTS `COMPOSITE_ROLE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `COMPOSITE_ROLE` (
  `COMPOSITE` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `CHILD_ROLE` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`COMPOSITE`,`CHILD_ROLE`),
  KEY `IDX_COMPOSITE` (`COMPOSITE`),
  KEY `IDX_COMPOSITE_CHILD` (`CHILD_ROLE`),
  CONSTRAINT `FK_A63WVEKFTU8JO1PNJ81E7MCE2` FOREIGN KEY (`COMPOSITE`) REFERENCES `KEYCLOAK_ROLE` (`ID`),
  CONSTRAINT `FK_GR7THLLB9LU8Q4VQA4524JJY8` FOREIGN KEY (`CHILD_ROLE`) REFERENCES `KEYCLOAK_ROLE` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `COMPOSITE_ROLE`
--

LOCK TABLES `COMPOSITE_ROLE` WRITE;
/*!40000 ALTER TABLE `COMPOSITE_ROLE` DISABLE KEYS */;
INSERT INTO `COMPOSITE_ROLE` VALUES ('1aa723ff-209c-4637-b3c8-8159d72e9b09','2a2bb3c8-e26f-4adb-9a9e-bbfb1685745e'),('4d227aaf-b4fa-4a86-9535-30210f612f2e','1aa723ff-209c-4637-b3c8-8159d72e9b09'),('4d227aaf-b4fa-4a86-9535-30210f612f2e','1d4a0b4b-1f99-4a6e-a9e1-463eec8147fd'),('4d227aaf-b4fa-4a86-9535-30210f612f2e','b3794300-e7e9-4f3d-af84-7d032a01df6b'),('4d227aaf-b4fa-4a86-9535-30210f612f2e','e06c7506-138d-4968-9186-cd958b29e577'),('5827ab16-b5bc-4738-b05e-89406e065439','0b1ad207-9fa7-4ed3-84ba-f800607e7d09'),('5827ab16-b5bc-4738-b05e-89406e065439','16801c90-306a-4dd9-8f04-6a79dc56ce7a'),('5827ab16-b5bc-4738-b05e-89406e065439','1b33555e-7a77-48cd-8c2a-5b7fae1318df'),('5827ab16-b5bc-4738-b05e-89406e065439','3b8b7ca4-89c8-4038-a996-c7ce380efa2c'),('5827ab16-b5bc-4738-b05e-89406e065439','70e66b49-b385-466c-95a8-7bff51eee65f'),('5827ab16-b5bc-4738-b05e-89406e065439','766d6f0c-d33b-480c-b193-98f7dd41b5e6'),('5827ab16-b5bc-4738-b05e-89406e065439','7b199ec2-c91d-4300-a565-f21a8c16b647'),('5827ab16-b5bc-4738-b05e-89406e065439','8340533b-d129-4b11-ae94-708efcb2a14a'),('5827ab16-b5bc-4738-b05e-89406e065439','839188a4-115b-4eab-9fa5-9573ee95fbf0'),('5827ab16-b5bc-4738-b05e-89406e065439','9066e812-fa7d-4d41-8b29-7c350d08bedf'),('5827ab16-b5bc-4738-b05e-89406e065439','982e7b3f-64e8-4d41-9323-7d308fe31b8f'),('5827ab16-b5bc-4738-b05e-89406e065439','a433b041-695e-4da9-a275-1b6abf9184c9'),('5827ab16-b5bc-4738-b05e-89406e065439','b52948b8-e6cc-45df-aeeb-bfa8210e9378'),('5827ab16-b5bc-4738-b05e-89406e065439','c9a3a64a-f4f7-46a4-aaf0-b1af52473ce8'),('5827ab16-b5bc-4738-b05e-89406e065439','ce72ef0f-db0f-4eb7-bf60-642c203cc1e9'),('5827ab16-b5bc-4738-b05e-89406e065439','ceab846f-983c-4c02-a89e-bfac58410427'),('5827ab16-b5bc-4738-b05e-89406e065439','d3540296-65b3-4fa5-a007-fe0feb99d6d1'),('5827ab16-b5bc-4738-b05e-89406e065439','d726a590-eea5-4040-85aa-b44f10f3bf58'),('5827ab16-b5bc-4738-b05e-89406e065439','e498256c-c0b0-4bfa-a7c3-f10f05de507e'),('839188a4-115b-4eab-9fa5-9573ee95fbf0','70e66b49-b385-466c-95a8-7bff51eee65f'),('839188a4-115b-4eab-9fa5-9573ee95fbf0','a433b041-695e-4da9-a275-1b6abf9184c9'),('ce72ef0f-db0f-4eb7-bf60-642c203cc1e9','8340533b-d129-4b11-ae94-708efcb2a14a'),('f660cd73-8ffe-4077-b73a-b26c6a24f149','8462f2ed-3177-493c-88af-ee6dd888b246');
/*!40000 ALTER TABLE `COMPOSITE_ROLE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `CREDENTIAL`
--

DROP TABLE IF EXISTS `CREDENTIAL`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CREDENTIAL` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `SALT` tinyblob DEFAULT NULL,
  `TYPE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `USER_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CREATED_DATE` bigint(20) DEFAULT NULL,
  `USER_LABEL` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `SECRET_DATA` longtext COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CREDENTIAL_DATA` longtext COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `PRIORITY` int(11) DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_USER_CREDENTIAL` (`USER_ID`),
  CONSTRAINT `FK_PFYR0GLASQYL0DEI3KL69R6V0` FOREIGN KEY (`USER_ID`) REFERENCES `USER_ENTITY` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CREDENTIAL`
--

LOCK TABLES `CREDENTIAL` WRITE;
/*!40000 ALTER TABLE `CREDENTIAL` DISABLE KEYS */;
INSERT INTO `CREDENTIAL` VALUES ('9e4eb727-5eef-4686-b4c7-4d84282fade1',NULL,'password','563bb06b-d712-48c1-9381-cd6473e18590',1635795167967,NULL,'{\"value\":\"PEMYsigNJ+5xMOOdQkhjMh/7x9e2qKC+Mv9usfICUOwXv79Fn9Dar3fee5FJCw86tpQLP+hz2Of1m+pAksYjdg==\",\"salt\":\"IbVs55tBY3rj4s9WcG5QMg==\",\"additionalParameters\":{}}','{\"hashIterations\":27500,\"algorithm\":\"pbkdf2-sha256\",\"additionalParameters\":{}}',10),('df81f1dd-0bd8-4664-aa2b-83304f8f54c2',NULL,'password','744b396e-3cf9-4e9f-9493-61b7b188fb10',1635845862853,NULL,'{\"value\":\"SsI8poDKv1KUYXywcU8od170skwnBD3ibXh58RjvF5zzkw/Kndk7108OjomGfORghJC5Y7rnW9HI2aPhO6DdJg==\",\"salt\":\"+QSBSxjvcC7J2FwVtbH5WQ==\",\"additionalParameters\":{}}','{\"hashIterations\":27500,\"algorithm\":\"pbkdf2-sha256\",\"additionalParameters\":{}}',10);
/*!40000 ALTER TABLE `CREDENTIAL` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `DATABASECHANGELOG`
--

DROP TABLE IF EXISTS `DATABASECHANGELOG`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `DATABASECHANGELOG` (
  `ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `AUTHOR` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `FILENAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `DATEEXECUTED` datetime NOT NULL,
  `ORDEREXECUTED` int(11) NOT NULL,
  `EXECTYPE` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `MD5SUM` varchar(35) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `DESCRIPTION` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `COMMENTS` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `TAG` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `LIQUIBASE` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CONTEXTS` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `LABELS` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `DEPLOYMENT_ID` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `DATABASECHANGELOG`
--

LOCK TABLES `DATABASECHANGELOG` WRITE;
/*!40000 ALTER TABLE `DATABASECHANGELOG` DISABLE KEYS */;
INSERT INTO `DATABASECHANGELOG` VALUES ('1.0.0.Final-KEYCLOAK-5461','sthorger@redhat.com','META-INF/jpa-changelog-1.0.0.Final.xml','2021-11-01 19:32:34',1,'EXECUTED','7:4e70412f24a3f382c82183742ec79317','createTable tableName=APPLICATION_DEFAULT_ROLES; createTable tableName=CLIENT; createTable tableName=CLIENT_SESSION; createTable tableName=CLIENT_SESSION_ROLE; createTable tableName=COMPOSITE_ROLE; createTable tableName=CREDENTIAL; createTable tab...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.0.0.Final-KEYCLOAK-5461','sthorger@redhat.com','META-INF/db2-jpa-changelog-1.0.0.Final.xml','2021-11-01 19:32:34',2,'MARK_RAN','7:cb16724583e9675711801c6875114f28','createTable tableName=APPLICATION_DEFAULT_ROLES; createTable tableName=CLIENT; createTable tableName=CLIENT_SESSION; createTable tableName=CLIENT_SESSION_ROLE; createTable tableName=COMPOSITE_ROLE; createTable tableName=CREDENTIAL; createTable tab...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.1.0.Beta1','sthorger@redhat.com','META-INF/jpa-changelog-1.1.0.Beta1.xml','2021-11-01 19:32:34',3,'EXECUTED','7:0310eb8ba07cec616460794d42ade0fa','delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION; createTable tableName=CLIENT_ATTRIBUTES; createTable tableName=CLIENT_SESSION_NOTE; createTable tableName=APP_NODE_REGISTRATIONS; addColumn table...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.1.0.Final','sthorger@redhat.com','META-INF/jpa-changelog-1.1.0.Final.xml','2021-11-01 19:32:34',4,'EXECUTED','7:5d25857e708c3233ef4439df1f93f012','renameColumn newColumnName=EVENT_TIME, oldColumnName=TIME, tableName=EVENT_ENTITY','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.2.0.Beta1','psilva@redhat.com','META-INF/jpa-changelog-1.2.0.Beta1.xml','2021-11-01 19:32:34',5,'EXECUTED','7:c7a54a1041d58eb3817a4a883b4d4e84','delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION; createTable tableName=PROTOCOL_MAPPER; createTable tableName=PROTOCOL_MAPPER_CONFIG; createTable tableName=...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.2.0.Beta1','psilva@redhat.com','META-INF/db2-jpa-changelog-1.2.0.Beta1.xml','2021-11-01 19:32:34',6,'MARK_RAN','7:2e01012df20974c1c2a605ef8afe25b7','delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION; createTable tableName=PROTOCOL_MAPPER; createTable tableName=PROTOCOL_MAPPER_CONFIG; createTable tableName=...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.2.0.RC1','bburke@redhat.com','META-INF/jpa-changelog-1.2.0.CR1.xml','2021-11-01 19:32:35',7,'EXECUTED','7:0f08df48468428e0f30ee59a8ec01a41','delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete tableName=USER_SESSION; createTable tableName=MIGRATION_MODEL; createTable tableName=IDENTITY_P...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.2.0.RC1','bburke@redhat.com','META-INF/db2-jpa-changelog-1.2.0.CR1.xml','2021-11-01 19:32:35',8,'MARK_RAN','7:a77ea2ad226b345e7d689d366f185c8c','delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete tableName=USER_SESSION; createTable tableName=MIGRATION_MODEL; createTable tableName=IDENTITY_P...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.2.0.Final','keycloak','META-INF/jpa-changelog-1.2.0.Final.xml','2021-11-01 19:32:35',9,'EXECUTED','7:a3377a2059aefbf3b90ebb4c4cc8e2ab','update tableName=CLIENT; update tableName=CLIENT; update tableName=CLIENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.3.0','bburke@redhat.com','META-INF/jpa-changelog-1.3.0.xml','2021-11-01 19:32:35',10,'EXECUTED','7:04c1dbedc2aa3e9756d1a1668e003451','delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_PROT_MAPPER; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete tableName=USER_SESSION; createTable tableName=ADMI...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.4.0','bburke@redhat.com','META-INF/jpa-changelog-1.4.0.xml','2021-11-01 19:32:35',11,'EXECUTED','7:36ef39ed560ad07062d956db861042ba','delete tableName=CLIENT_SESSION_AUTH_STATUS; delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_PROT_MAPPER; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete table...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.4.0','bburke@redhat.com','META-INF/db2-jpa-changelog-1.4.0.xml','2021-11-01 19:32:35',12,'MARK_RAN','7:d909180b2530479a716d3f9c9eaea3d7','delete tableName=CLIENT_SESSION_AUTH_STATUS; delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_PROT_MAPPER; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete table...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.5.0','bburke@redhat.com','META-INF/jpa-changelog-1.5.0.xml','2021-11-01 19:32:35',13,'EXECUTED','7:cf12b04b79bea5152f165eb41f3955f6','delete tableName=CLIENT_SESSION_AUTH_STATUS; delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_PROT_MAPPER; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete table...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.6.1_from15','mposolda@redhat.com','META-INF/jpa-changelog-1.6.1.xml','2021-11-01 19:32:35',14,'EXECUTED','7:7e32c8f05c755e8675764e7d5f514509','addColumn tableName=REALM; addColumn tableName=KEYCLOAK_ROLE; addColumn tableName=CLIENT; createTable tableName=OFFLINE_USER_SESSION; createTable tableName=OFFLINE_CLIENT_SESSION; addPrimaryKey constraintName=CONSTRAINT_OFFL_US_SES_PK2, tableName=...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.6.1_from16-pre','mposolda@redhat.com','META-INF/jpa-changelog-1.6.1.xml','2021-11-01 19:32:35',15,'MARK_RAN','7:980ba23cc0ec39cab731ce903dd01291','delete tableName=OFFLINE_CLIENT_SESSION; delete tableName=OFFLINE_USER_SESSION','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.6.1_from16','mposolda@redhat.com','META-INF/jpa-changelog-1.6.1.xml','2021-11-01 19:32:35',16,'MARK_RAN','7:2fa220758991285312eb84f3b4ff5336','dropPrimaryKey constraintName=CONSTRAINT_OFFLINE_US_SES_PK, tableName=OFFLINE_USER_SESSION; dropPrimaryKey constraintName=CONSTRAINT_OFFLINE_CL_SES_PK, tableName=OFFLINE_CLIENT_SESSION; addColumn tableName=OFFLINE_USER_SESSION; update tableName=OF...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.6.1','mposolda@redhat.com','META-INF/jpa-changelog-1.6.1.xml','2021-11-01 19:32:35',17,'EXECUTED','7:d41d8cd98f00b204e9800998ecf8427e','empty','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.7.0','bburke@redhat.com','META-INF/jpa-changelog-1.7.0.xml','2021-11-01 19:32:35',18,'EXECUTED','7:91ace540896df890cc00a0490ee52bbc','createTable tableName=KEYCLOAK_GROUP; createTable tableName=GROUP_ROLE_MAPPING; createTable tableName=GROUP_ATTRIBUTE; createTable tableName=USER_GROUP_MEMBERSHIP; createTable tableName=REALM_DEFAULT_GROUPS; addColumn tableName=IDENTITY_PROVIDER; ...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.8.0','mposolda@redhat.com','META-INF/jpa-changelog-1.8.0.xml','2021-11-01 19:32:36',19,'EXECUTED','7:c31d1646dfa2618a9335c00e07f89f24','addColumn tableName=IDENTITY_PROVIDER; createTable tableName=CLIENT_TEMPLATE; createTable tableName=CLIENT_TEMPLATE_ATTRIBUTES; createTable tableName=TEMPLATE_SCOPE_MAPPING; dropNotNullConstraint columnName=CLIENT_ID, tableName=PROTOCOL_MAPPER; ad...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.8.0-2','keycloak','META-INF/jpa-changelog-1.8.0.xml','2021-11-01 19:32:36',20,'EXECUTED','7:df8bc21027a4f7cbbb01f6344e89ce07','dropDefaultValue columnName=ALGORITHM, tableName=CREDENTIAL; update tableName=CREDENTIAL','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.8.0','mposolda@redhat.com','META-INF/db2-jpa-changelog-1.8.0.xml','2021-11-01 19:32:36',21,'MARK_RAN','7:f987971fe6b37d963bc95fee2b27f8df','addColumn tableName=IDENTITY_PROVIDER; createTable tableName=CLIENT_TEMPLATE; createTable tableName=CLIENT_TEMPLATE_ATTRIBUTES; createTable tableName=TEMPLATE_SCOPE_MAPPING; dropNotNullConstraint columnName=CLIENT_ID, tableName=PROTOCOL_MAPPER; ad...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.8.0-2','keycloak','META-INF/db2-jpa-changelog-1.8.0.xml','2021-11-01 19:32:36',22,'MARK_RAN','7:df8bc21027a4f7cbbb01f6344e89ce07','dropDefaultValue columnName=ALGORITHM, tableName=CREDENTIAL; update tableName=CREDENTIAL','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.9.0','mposolda@redhat.com','META-INF/jpa-changelog-1.9.0.xml','2021-11-01 19:32:36',23,'EXECUTED','7:ed2dc7f799d19ac452cbcda56c929e47','update tableName=REALM; update tableName=REALM; update tableName=REALM; update tableName=REALM; update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=REALM; update tableName=REALM; customChange; dr...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.9.1','keycloak','META-INF/jpa-changelog-1.9.1.xml','2021-11-01 19:32:36',24,'EXECUTED','7:80b5db88a5dda36ece5f235be8757615','modifyDataType columnName=PRIVATE_KEY, tableName=REALM; modifyDataType columnName=PUBLIC_KEY, tableName=REALM; modifyDataType columnName=CERTIFICATE, tableName=REALM','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.9.1','keycloak','META-INF/db2-jpa-changelog-1.9.1.xml','2021-11-01 19:32:36',25,'MARK_RAN','7:1437310ed1305a9b93f8848f301726ce','modifyDataType columnName=PRIVATE_KEY, tableName=REALM; modifyDataType columnName=CERTIFICATE, tableName=REALM','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.9.2','keycloak','META-INF/jpa-changelog-1.9.2.xml','2021-11-01 19:32:36',26,'EXECUTED','7:b82ffb34850fa0836be16deefc6a87c4','createIndex indexName=IDX_USER_EMAIL, tableName=USER_ENTITY; createIndex indexName=IDX_USER_ROLE_MAPPING, tableName=USER_ROLE_MAPPING; createIndex indexName=IDX_USER_GROUP_MAPPING, tableName=USER_GROUP_MEMBERSHIP; createIndex indexName=IDX_USER_CO...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-2.0.0','psilva@redhat.com','META-INF/jpa-changelog-authz-2.0.0.xml','2021-11-01 19:32:36',27,'EXECUTED','7:9cc98082921330d8d9266decdd4bd658','createTable tableName=RESOURCE_SERVER; addPrimaryKey constraintName=CONSTRAINT_FARS, tableName=RESOURCE_SERVER; addUniqueConstraint constraintName=UK_AU8TT6T700S9V50BU18WS5HA6, tableName=RESOURCE_SERVER; createTable tableName=RESOURCE_SERVER_RESOU...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-2.5.1','psilva@redhat.com','META-INF/jpa-changelog-authz-2.5.1.xml','2021-11-01 19:32:36',28,'EXECUTED','7:03d64aeed9cb52b969bd30a7ac0db57e','update tableName=RESOURCE_SERVER_POLICY','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.1.0-KEYCLOAK-5461','bburke@redhat.com','META-INF/jpa-changelog-2.1.0.xml','2021-11-01 19:32:36',29,'EXECUTED','7:f1f9fd8710399d725b780f463c6b21cd','createTable tableName=BROKER_LINK; createTable tableName=FED_USER_ATTRIBUTE; createTable tableName=FED_USER_CONSENT; createTable tableName=FED_USER_CONSENT_ROLE; createTable tableName=FED_USER_CONSENT_PROT_MAPPER; createTable tableName=FED_USER_CR...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.2.0','bburke@redhat.com','META-INF/jpa-changelog-2.2.0.xml','2021-11-01 19:32:36',30,'EXECUTED','7:53188c3eb1107546e6f765835705b6c1','addColumn tableName=ADMIN_EVENT_ENTITY; createTable tableName=CREDENTIAL_ATTRIBUTE; createTable tableName=FED_CREDENTIAL_ATTRIBUTE; modifyDataType columnName=VALUE, tableName=CREDENTIAL; addForeignKeyConstraint baseTableName=FED_CREDENTIAL_ATTRIBU...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.3.0','bburke@redhat.com','META-INF/jpa-changelog-2.3.0.xml','2021-11-01 19:32:37',31,'EXECUTED','7:d6e6f3bc57a0c5586737d1351725d4d4','createTable tableName=FEDERATED_USER; addPrimaryKey constraintName=CONSTR_FEDERATED_USER, tableName=FEDERATED_USER; dropDefaultValue columnName=TOTP, tableName=USER_ENTITY; dropColumn columnName=TOTP, tableName=USER_ENTITY; addColumn tableName=IDE...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.4.0','bburke@redhat.com','META-INF/jpa-changelog-2.4.0.xml','2021-11-01 19:32:37',32,'EXECUTED','7:454d604fbd755d9df3fd9c6329043aa5','customChange','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.5.0','bburke@redhat.com','META-INF/jpa-changelog-2.5.0.xml','2021-11-01 19:32:37',33,'EXECUTED','7:57e98a3077e29caf562f7dbf80c72600','customChange; modifyDataType columnName=USER_ID, tableName=OFFLINE_USER_SESSION','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.5.0-unicode-oracle','hmlnarik@redhat.com','META-INF/jpa-changelog-2.5.0.xml','2021-11-01 19:32:37',34,'MARK_RAN','7:e4c7e8f2256210aee71ddc42f538b57a','modifyDataType columnName=DESCRIPTION, tableName=AUTHENTICATION_FLOW; modifyDataType columnName=DESCRIPTION, tableName=CLIENT_TEMPLATE; modifyDataType columnName=DESCRIPTION, tableName=RESOURCE_SERVER_POLICY; modifyDataType columnName=DESCRIPTION,...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.5.0-unicode-other-dbs','hmlnarik@redhat.com','META-INF/jpa-changelog-2.5.0.xml','2021-11-01 19:32:37',35,'EXECUTED','7:09a43c97e49bc626460480aa1379b522','modifyDataType columnName=DESCRIPTION, tableName=AUTHENTICATION_FLOW; modifyDataType columnName=DESCRIPTION, tableName=CLIENT_TEMPLATE; modifyDataType columnName=DESCRIPTION, tableName=RESOURCE_SERVER_POLICY; modifyDataType columnName=DESCRIPTION,...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.5.0-duplicate-email-support','slawomir@dabek.name','META-INF/jpa-changelog-2.5.0.xml','2021-11-01 19:32:37',36,'EXECUTED','7:26bfc7c74fefa9126f2ce702fb775553','addColumn tableName=REALM','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.5.0-unique-group-names','hmlnarik@redhat.com','META-INF/jpa-changelog-2.5.0.xml','2021-11-01 19:32:37',37,'EXECUTED','7:a161e2ae671a9020fff61e996a207377','addUniqueConstraint constraintName=SIBLING_NAMES, tableName=KEYCLOAK_GROUP','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.5.1','bburke@redhat.com','META-INF/jpa-changelog-2.5.1.xml','2021-11-01 19:32:37',38,'EXECUTED','7:37fc1781855ac5388c494f1442b3f717','addColumn tableName=FED_USER_CONSENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.0.0','bburke@redhat.com','META-INF/jpa-changelog-3.0.0.xml','2021-11-01 19:32:37',39,'EXECUTED','7:13a27db0dae6049541136adad7261d27','addColumn tableName=IDENTITY_PROVIDER','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.2.0-fix','keycloak','META-INF/jpa-changelog-3.2.0.xml','2021-11-01 19:32:37',40,'MARK_RAN','7:550300617e3b59e8af3a6294df8248a3','addNotNullConstraint columnName=REALM_ID, tableName=CLIENT_INITIAL_ACCESS','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.2.0-fix-with-keycloak-5416','keycloak','META-INF/jpa-changelog-3.2.0.xml','2021-11-01 19:32:37',41,'MARK_RAN','7:e3a9482b8931481dc2772a5c07c44f17','dropIndex indexName=IDX_CLIENT_INIT_ACC_REALM, tableName=CLIENT_INITIAL_ACCESS; addNotNullConstraint columnName=REALM_ID, tableName=CLIENT_INITIAL_ACCESS; createIndex indexName=IDX_CLIENT_INIT_ACC_REALM, tableName=CLIENT_INITIAL_ACCESS','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.2.0-fix-offline-sessions','hmlnarik','META-INF/jpa-changelog-3.2.0.xml','2021-11-01 19:32:37',42,'EXECUTED','7:72b07d85a2677cb257edb02b408f332d','customChange','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.2.0-fixed','keycloak','META-INF/jpa-changelog-3.2.0.xml','2021-11-01 19:32:37',43,'EXECUTED','7:a72a7858967bd414835d19e04d880312','addColumn tableName=REALM; dropPrimaryKey constraintName=CONSTRAINT_OFFL_CL_SES_PK2, tableName=OFFLINE_CLIENT_SESSION; dropColumn columnName=CLIENT_SESSION_ID, tableName=OFFLINE_CLIENT_SESSION; addPrimaryKey constraintName=CONSTRAINT_OFFL_CL_SES_P...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.3.0','keycloak','META-INF/jpa-changelog-3.3.0.xml','2021-11-01 19:32:37',44,'EXECUTED','7:94edff7cf9ce179e7e85f0cd78a3cf2c','addColumn tableName=USER_ENTITY','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-3.4.0.CR1-resource-server-pk-change-part1','glavoie@gmail.com','META-INF/jpa-changelog-authz-3.4.0.CR1.xml','2021-11-01 19:32:37',45,'EXECUTED','7:6a48ce645a3525488a90fbf76adf3bb3','addColumn tableName=RESOURCE_SERVER_POLICY; addColumn tableName=RESOURCE_SERVER_RESOURCE; addColumn tableName=RESOURCE_SERVER_SCOPE','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-3.4.0.CR1-resource-server-pk-change-part2-KEYCLOAK-6095','hmlnarik@redhat.com','META-INF/jpa-changelog-authz-3.4.0.CR1.xml','2021-11-01 19:32:37',46,'EXECUTED','7:e64b5dcea7db06077c6e57d3b9e5ca14','customChange','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-3.4.0.CR1-resource-server-pk-change-part3-fixed','glavoie@gmail.com','META-INF/jpa-changelog-authz-3.4.0.CR1.xml','2021-11-01 19:32:37',47,'MARK_RAN','7:fd8cf02498f8b1e72496a20afc75178c','dropIndex indexName=IDX_RES_SERV_POL_RES_SERV, tableName=RESOURCE_SERVER_POLICY; dropIndex indexName=IDX_RES_SRV_RES_RES_SRV, tableName=RESOURCE_SERVER_RESOURCE; dropIndex indexName=IDX_RES_SRV_SCOPE_RES_SRV, tableName=RESOURCE_SERVER_SCOPE','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-3.4.0.CR1-resource-server-pk-change-part3-fixed-nodropindex','glavoie@gmail.com','META-INF/jpa-changelog-authz-3.4.0.CR1.xml','2021-11-01 19:32:38',48,'EXECUTED','7:542794f25aa2b1fbabb7e577d6646319','addNotNullConstraint columnName=RESOURCE_SERVER_CLIENT_ID, tableName=RESOURCE_SERVER_POLICY; addNotNullConstraint columnName=RESOURCE_SERVER_CLIENT_ID, tableName=RESOURCE_SERVER_RESOURCE; addNotNullConstraint columnName=RESOURCE_SERVER_CLIENT_ID, ...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authn-3.4.0.CR1-refresh-token-max-reuse','glavoie@gmail.com','META-INF/jpa-changelog-authz-3.4.0.CR1.xml','2021-11-01 19:32:38',49,'EXECUTED','7:edad604c882df12f74941dac3cc6d650','addColumn tableName=REALM','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.4.0','keycloak','META-INF/jpa-changelog-3.4.0.xml','2021-11-01 19:32:38',50,'EXECUTED','7:0f88b78b7b46480eb92690cbf5e44900','addPrimaryKey constraintName=CONSTRAINT_REALM_DEFAULT_ROLES, tableName=REALM_DEFAULT_ROLES; addPrimaryKey constraintName=CONSTRAINT_COMPOSITE_ROLE, tableName=COMPOSITE_ROLE; addPrimaryKey constraintName=CONSTR_REALM_DEFAULT_GROUPS, tableName=REALM...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.4.0-KEYCLOAK-5230','hmlnarik@redhat.com','META-INF/jpa-changelog-3.4.0.xml','2021-11-01 19:32:38',51,'EXECUTED','7:d560e43982611d936457c327f872dd59','createIndex indexName=IDX_FU_ATTRIBUTE, tableName=FED_USER_ATTRIBUTE; createIndex indexName=IDX_FU_CONSENT, tableName=FED_USER_CONSENT; createIndex indexName=IDX_FU_CONSENT_RU, tableName=FED_USER_CONSENT; createIndex indexName=IDX_FU_CREDENTIAL, t...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.4.1','psilva@redhat.com','META-INF/jpa-changelog-3.4.1.xml','2021-11-01 19:32:38',52,'EXECUTED','7:c155566c42b4d14ef07059ec3b3bbd8e','modifyDataType columnName=VALUE, tableName=CLIENT_ATTRIBUTES','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.4.2','keycloak','META-INF/jpa-changelog-3.4.2.xml','2021-11-01 19:32:38',53,'EXECUTED','7:b40376581f12d70f3c89ba8ddf5b7dea','update tableName=REALM','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.4.2-KEYCLOAK-5172','mkanis@redhat.com','META-INF/jpa-changelog-3.4.2.xml','2021-11-01 19:32:38',54,'EXECUTED','7:a1132cc395f7b95b3646146c2e38f168','update tableName=CLIENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.0.0-KEYCLOAK-6335','bburke@redhat.com','META-INF/jpa-changelog-4.0.0.xml','2021-11-01 19:32:38',55,'EXECUTED','7:d8dc5d89c789105cfa7ca0e82cba60af','createTable tableName=CLIENT_AUTH_FLOW_BINDINGS; addPrimaryKey constraintName=C_CLI_FLOW_BIND, tableName=CLIENT_AUTH_FLOW_BINDINGS','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.0.0-CLEANUP-UNUSED-TABLE','bburke@redhat.com','META-INF/jpa-changelog-4.0.0.xml','2021-11-01 19:32:38',56,'EXECUTED','7:7822e0165097182e8f653c35517656a3','dropTable tableName=CLIENT_IDENTITY_PROV_MAPPING','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.0.0-KEYCLOAK-6228','bburke@redhat.com','META-INF/jpa-changelog-4.0.0.xml','2021-11-01 19:32:38',57,'EXECUTED','7:c6538c29b9c9a08f9e9ea2de5c2b6375','dropUniqueConstraint constraintName=UK_JKUWUVD56ONTGSUHOGM8UEWRT, tableName=USER_CONSENT; dropNotNullConstraint columnName=CLIENT_ID, tableName=USER_CONSENT; addColumn tableName=USER_CONSENT; addUniqueConstraint constraintName=UK_JKUWUVD56ONTGSUHO...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.0.0-KEYCLOAK-5579-fixed','mposolda@redhat.com','META-INF/jpa-changelog-4.0.0.xml','2021-11-01 19:32:39',58,'EXECUTED','7:6d4893e36de22369cf73bcb051ded875','dropForeignKeyConstraint baseTableName=CLIENT_TEMPLATE_ATTRIBUTES, constraintName=FK_CL_TEMPL_ATTR_TEMPL; renameTable newTableName=CLIENT_SCOPE_ATTRIBUTES, oldTableName=CLIENT_TEMPLATE_ATTRIBUTES; renameColumn newColumnName=SCOPE_ID, oldColumnName...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-4.0.0.CR1','psilva@redhat.com','META-INF/jpa-changelog-authz-4.0.0.CR1.xml','2021-11-01 19:32:39',59,'EXECUTED','7:57960fc0b0f0dd0563ea6f8b2e4a1707','createTable tableName=RESOURCE_SERVER_PERM_TICKET; addPrimaryKey constraintName=CONSTRAINT_FAPMT, tableName=RESOURCE_SERVER_PERM_TICKET; addForeignKeyConstraint baseTableName=RESOURCE_SERVER_PERM_TICKET, constraintName=FK_FRSRHO213XCX4WNKOG82SSPMT...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-4.0.0.Beta3','psilva@redhat.com','META-INF/jpa-changelog-authz-4.0.0.Beta3.xml','2021-11-01 19:32:39',60,'EXECUTED','7:2b4b8bff39944c7097977cc18dbceb3b','addColumn tableName=RESOURCE_SERVER_POLICY; addColumn tableName=RESOURCE_SERVER_PERM_TICKET; addForeignKeyConstraint baseTableName=RESOURCE_SERVER_PERM_TICKET, constraintName=FK_FRSRPO2128CX4WNKOG82SSRFY, referencedTableName=RESOURCE_SERVER_POLICY','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-4.2.0.Final','mhajas@redhat.com','META-INF/jpa-changelog-authz-4.2.0.Final.xml','2021-11-01 19:32:39',61,'EXECUTED','7:2aa42a964c59cd5b8ca9822340ba33a8','createTable tableName=RESOURCE_URIS; addForeignKeyConstraint baseTableName=RESOURCE_URIS, constraintName=FK_RESOURCE_SERVER_URIS, referencedTableName=RESOURCE_SERVER_RESOURCE; customChange; dropColumn columnName=URI, tableName=RESOURCE_SERVER_RESO...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-4.2.0.Final-KEYCLOAK-9944','hmlnarik@redhat.com','META-INF/jpa-changelog-authz-4.2.0.Final.xml','2021-11-01 19:32:39',62,'EXECUTED','7:9ac9e58545479929ba23f4a3087a0346','addPrimaryKey constraintName=CONSTRAINT_RESOUR_URIS_PK, tableName=RESOURCE_URIS','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.2.0-KEYCLOAK-6313','wadahiro@gmail.com','META-INF/jpa-changelog-4.2.0.xml','2021-11-01 19:32:39',63,'EXECUTED','7:14d407c35bc4fe1976867756bcea0c36','addColumn tableName=REQUIRED_ACTION_PROVIDER','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.3.0-KEYCLOAK-7984','wadahiro@gmail.com','META-INF/jpa-changelog-4.3.0.xml','2021-11-01 19:32:39',64,'EXECUTED','7:241a8030c748c8548e346adee548fa93','update tableName=REQUIRED_ACTION_PROVIDER','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.6.0-KEYCLOAK-7950','psilva@redhat.com','META-INF/jpa-changelog-4.6.0.xml','2021-11-01 19:32:39',65,'EXECUTED','7:7d3182f65a34fcc61e8d23def037dc3f','update tableName=RESOURCE_SERVER_RESOURCE','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.6.0-KEYCLOAK-8377','keycloak','META-INF/jpa-changelog-4.6.0.xml','2021-11-01 19:32:39',66,'EXECUTED','7:b30039e00a0b9715d430d1b0636728fa','createTable tableName=ROLE_ATTRIBUTE; addPrimaryKey constraintName=CONSTRAINT_ROLE_ATTRIBUTE_PK, tableName=ROLE_ATTRIBUTE; addForeignKeyConstraint baseTableName=ROLE_ATTRIBUTE, constraintName=FK_ROLE_ATTRIBUTE_ID, referencedTableName=KEYCLOAK_ROLE...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.6.0-KEYCLOAK-8555','gideonray@gmail.com','META-INF/jpa-changelog-4.6.0.xml','2021-11-01 19:32:39',67,'EXECUTED','7:3797315ca61d531780f8e6f82f258159','createIndex indexName=IDX_COMPONENT_PROVIDER_TYPE, tableName=COMPONENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.7.0-KEYCLOAK-1267','sguilhen@redhat.com','META-INF/jpa-changelog-4.7.0.xml','2021-11-01 19:32:39',68,'EXECUTED','7:c7aa4c8d9573500c2d347c1941ff0301','addColumn tableName=REALM','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.7.0-KEYCLOAK-7275','keycloak','META-INF/jpa-changelog-4.7.0.xml','2021-11-01 19:32:39',69,'EXECUTED','7:b207faee394fc074a442ecd42185a5dd','renameColumn newColumnName=CREATED_ON, oldColumnName=LAST_SESSION_REFRESH, tableName=OFFLINE_USER_SESSION; addNotNullConstraint columnName=CREATED_ON, tableName=OFFLINE_USER_SESSION; addColumn tableName=OFFLINE_USER_SESSION; customChange; createIn...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.8.0-KEYCLOAK-8835','sguilhen@redhat.com','META-INF/jpa-changelog-4.8.0.xml','2021-11-01 19:32:39',70,'EXECUTED','7:ab9a9762faaba4ddfa35514b212c4922','addNotNullConstraint columnName=SSO_MAX_LIFESPAN_REMEMBER_ME, tableName=REALM; addNotNullConstraint columnName=SSO_IDLE_TIMEOUT_REMEMBER_ME, tableName=REALM','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-7.0.0-KEYCLOAK-10443','psilva@redhat.com','META-INF/jpa-changelog-authz-7.0.0.xml','2021-11-01 19:32:39',71,'EXECUTED','7:b9710f74515a6ccb51b72dc0d19df8c4','addColumn tableName=RESOURCE_SERVER','',NULL,'3.5.4',NULL,NULL,'5795153174'),('8.0.0-adding-credential-columns','keycloak','META-INF/jpa-changelog-8.0.0.xml','2021-11-01 19:32:39',72,'EXECUTED','7:ec9707ae4d4f0b7452fee20128083879','addColumn tableName=CREDENTIAL; addColumn tableName=FED_USER_CREDENTIAL','',NULL,'3.5.4',NULL,NULL,'5795153174'),('8.0.0-updating-credential-data-not-oracle-fixed','keycloak','META-INF/jpa-changelog-8.0.0.xml','2021-11-01 19:32:39',73,'EXECUTED','7:3979a0ae07ac465e920ca696532fc736','update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=FED_USER_CREDENTIAL; update tableName=FED_USER_CREDENTIAL; update tableName=FED_USER_CREDENTIAL','',NULL,'3.5.4',NULL,NULL,'5795153174'),('8.0.0-updating-credential-data-oracle-fixed','keycloak','META-INF/jpa-changelog-8.0.0.xml','2021-11-01 19:32:39',74,'MARK_RAN','7:5abfde4c259119d143bd2fbf49ac2bca','update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=FED_USER_CREDENTIAL; update tableName=FED_USER_CREDENTIAL; update tableName=FED_USER_CREDENTIAL','',NULL,'3.5.4',NULL,NULL,'5795153174'),('8.0.0-credential-cleanup-fixed','keycloak','META-INF/jpa-changelog-8.0.0.xml','2021-11-01 19:32:39',75,'EXECUTED','7:b48da8c11a3d83ddd6b7d0c8c2219345','dropDefaultValue columnName=COUNTER, tableName=CREDENTIAL; dropDefaultValue columnName=DIGITS, tableName=CREDENTIAL; dropDefaultValue columnName=PERIOD, tableName=CREDENTIAL; dropDefaultValue columnName=ALGORITHM, tableName=CREDENTIAL; dropColumn ...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('8.0.0-resource-tag-support','keycloak','META-INF/jpa-changelog-8.0.0.xml','2021-11-01 19:32:39',76,'EXECUTED','7:a73379915c23bfad3e8f5c6d5c0aa4bd','addColumn tableName=MIGRATION_MODEL; createIndex indexName=IDX_UPDATE_TIME, tableName=MIGRATION_MODEL','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.0-always-display-client','keycloak','META-INF/jpa-changelog-9.0.0.xml','2021-11-01 19:32:39',77,'EXECUTED','7:39e0073779aba192646291aa2332493d','addColumn tableName=CLIENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.0-drop-constraints-for-column-increase','keycloak','META-INF/jpa-changelog-9.0.0.xml','2021-11-01 19:32:39',78,'MARK_RAN','7:81f87368f00450799b4bf42ea0b3ec34','dropUniqueConstraint constraintName=UK_FRSR6T700S9V50BU18WS5PMT, tableName=RESOURCE_SERVER_PERM_TICKET; dropUniqueConstraint constraintName=UK_FRSR6T700S9V50BU18WS5HA6, tableName=RESOURCE_SERVER_RESOURCE; dropPrimaryKey constraintName=CONSTRAINT_O...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.0-increase-column-size-federated-fk','keycloak','META-INF/jpa-changelog-9.0.0.xml','2021-11-01 19:32:39',79,'EXECUTED','7:20b37422abb9fb6571c618148f013a15','modifyDataType columnName=CLIENT_ID, tableName=FED_USER_CONSENT; modifyDataType columnName=CLIENT_REALM_CONSTRAINT, tableName=KEYCLOAK_ROLE; modifyDataType columnName=OWNER, tableName=RESOURCE_SERVER_POLICY; modifyDataType columnName=CLIENT_ID, ta...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.0-recreate-constraints-after-column-increase','keycloak','META-INF/jpa-changelog-9.0.0.xml','2021-11-01 19:32:39',80,'MARK_RAN','7:1970bb6cfb5ee800736b95ad3fb3c78a','addNotNullConstraint columnName=CLIENT_ID, tableName=OFFLINE_CLIENT_SESSION; addNotNullConstraint columnName=OWNER, tableName=RESOURCE_SERVER_PERM_TICKET; addNotNullConstraint columnName=REQUESTER, tableName=RESOURCE_SERVER_PERM_TICKET; addNotNull...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.1-add-index-to-client.client_id','keycloak','META-INF/jpa-changelog-9.0.1.xml','2021-11-01 19:32:39',81,'EXECUTED','7:45d9b25fc3b455d522d8dcc10a0f4c80','createIndex indexName=IDX_CLIENT_ID, tableName=CLIENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.1-KEYCLOAK-12579-drop-constraints','keycloak','META-INF/jpa-changelog-9.0.1.xml','2021-11-01 19:32:39',82,'MARK_RAN','7:890ae73712bc187a66c2813a724d037f','dropUniqueConstraint constraintName=SIBLING_NAMES, tableName=KEYCLOAK_GROUP','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.1-KEYCLOAK-12579-add-not-null-constraint','keycloak','META-INF/jpa-changelog-9.0.1.xml','2021-11-01 19:32:39',83,'EXECUTED','7:0a211980d27fafe3ff50d19a3a29b538','addNotNullConstraint columnName=PARENT_GROUP, tableName=KEYCLOAK_GROUP','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.1-KEYCLOAK-12579-recreate-constraints','keycloak','META-INF/jpa-changelog-9.0.1.xml','2021-11-01 19:32:39',84,'MARK_RAN','7:a161e2ae671a9020fff61e996a207377','addUniqueConstraint constraintName=SIBLING_NAMES, tableName=KEYCLOAK_GROUP','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.1-add-index-to-events','keycloak','META-INF/jpa-changelog-9.0.1.xml','2021-11-01 19:32:39',85,'EXECUTED','7:01c49302201bdf815b0a18d1f98a55dc','createIndex indexName=IDX_EVENT_TIME, tableName=EVENT_ENTITY','',NULL,'3.5.4',NULL,NULL,'5795153174'),('map-remove-ri','keycloak','META-INF/jpa-changelog-11.0.0.xml','2021-11-01 19:32:39',86,'EXECUTED','7:3dace6b144c11f53f1ad2c0361279b86','dropForeignKeyConstraint baseTableName=REALM, constraintName=FK_TRAF444KK6QRKMS7N56AIWQ5Y; dropForeignKeyConstraint baseTableName=KEYCLOAK_ROLE, constraintName=FK_KJHO5LE2C0RAL09FL8CM9WFW9','',NULL,'3.5.4',NULL,NULL,'5795153174'),('map-remove-ri','keycloak','META-INF/jpa-changelog-12.0.0.xml','2021-11-01 19:32:40',87,'EXECUTED','7:578d0b92077eaf2ab95ad0ec087aa903','dropForeignKeyConstraint baseTableName=REALM_DEFAULT_GROUPS, constraintName=FK_DEF_GROUPS_GROUP; dropForeignKeyConstraint baseTableName=REALM_DEFAULT_ROLES, constraintName=FK_H4WPD7W4HSOOLNI3H0SW7BTJE; dropForeignKeyConstraint baseTableName=CLIENT...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('12.1.0-add-realm-localization-table','keycloak','META-INF/jpa-changelog-12.0.0.xml','2021-11-01 19:32:40',88,'EXECUTED','7:c95abe90d962c57a09ecaee57972835d','createTable tableName=REALM_LOCALIZATIONS; addPrimaryKey tableName=REALM_LOCALIZATIONS','',NULL,'3.5.4',NULL,NULL,'5795153174'),('default-roles','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',89,'EXECUTED','7:f1313bcc2994a5c4dc1062ed6d8282d3','addColumn tableName=REALM; customChange','',NULL,'3.5.4',NULL,NULL,'5795153174'),('default-roles-cleanup','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',90,'EXECUTED','7:90d763b52eaffebefbcbde55f269508b','dropTable tableName=REALM_DEFAULT_ROLES; dropTable tableName=CLIENT_DEFAULT_ROLES','',NULL,'3.5.4',NULL,NULL,'5795153174'),('13.0.0-KEYCLOAK-16844','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',91,'EXECUTED','7:d554f0cb92b764470dccfa5e0014a7dd','createIndex indexName=IDX_OFFLINE_USS_PRELOAD, tableName=OFFLINE_USER_SESSION','',NULL,'3.5.4',NULL,NULL,'5795153174'),('map-remove-ri-13.0.0','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',92,'EXECUTED','7:73193e3ab3c35cf0f37ccea3bf783764','dropForeignKeyConstraint baseTableName=DEFAULT_CLIENT_SCOPE, constraintName=FK_R_DEF_CLI_SCOPE_SCOPE; dropForeignKeyConstraint baseTableName=CLIENT_SCOPE_CLIENT, constraintName=FK_C_CLI_SCOPE_SCOPE; dropForeignKeyConstraint baseTableName=CLIENT_SC...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('13.0.0-KEYCLOAK-17992-drop-constraints','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',93,'MARK_RAN','7:90a1e74f92e9cbaa0c5eab80b8a037f3','dropPrimaryKey constraintName=C_CLI_SCOPE_BIND, tableName=CLIENT_SCOPE_CLIENT; dropIndex indexName=IDX_CLSCOPE_CL, tableName=CLIENT_SCOPE_CLIENT; dropIndex indexName=IDX_CL_CLSCOPE, tableName=CLIENT_SCOPE_CLIENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('13.0.0-increase-column-size-federated','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',94,'EXECUTED','7:5b9248f29cd047c200083cc6d8388b16','modifyDataType columnName=CLIENT_ID, tableName=CLIENT_SCOPE_CLIENT; modifyDataType columnName=SCOPE_ID, tableName=CLIENT_SCOPE_CLIENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('13.0.0-KEYCLOAK-17992-recreate-constraints','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',95,'MARK_RAN','7:64db59e44c374f13955489e8990d17a1','addNotNullConstraint columnName=CLIENT_ID, tableName=CLIENT_SCOPE_CLIENT; addNotNullConstraint columnName=SCOPE_ID, tableName=CLIENT_SCOPE_CLIENT; addPrimaryKey constraintName=C_CLI_SCOPE_BIND, tableName=CLIENT_SCOPE_CLIENT; createIndex indexName=...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('json-string-accomodation-fixed','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',96,'EXECUTED','7:329a578cdb43262fff975f0a7f6cda60','addColumn tableName=REALM_ATTRIBUTE; update tableName=REALM_ATTRIBUTE; dropColumn columnName=VALUE, tableName=REALM_ATTRIBUTE; renameColumn newColumnName=VALUE, oldColumnName=VALUE_NEW, tableName=REALM_ATTRIBUTE','',NULL,'3.5.4',NULL,NULL,'5795153174'),('14.0.0-KEYCLOAK-11019','keycloak','META-INF/jpa-changelog-14.0.0.xml','2021-11-01 19:32:40',97,'EXECUTED','7:fae0de241ac0fd0bbc2b380b85e4f567','createIndex indexName=IDX_OFFLINE_CSS_PRELOAD, tableName=OFFLINE_CLIENT_SESSION; createIndex indexName=IDX_OFFLINE_USS_BY_USER, tableName=OFFLINE_USER_SESSION; createIndex indexName=IDX_OFFLINE_USS_BY_USERSESS, tableName=OFFLINE_USER_SESSION','',NULL,'3.5.4',NULL,NULL,'5795153174'),('14.0.0-KEYCLOAK-18286','keycloak','META-INF/jpa-changelog-14.0.0.xml','2021-11-01 19:32:40',98,'MARK_RAN','7:075d54e9180f49bb0c64ca4218936e81','createIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES','',NULL,'3.5.4',NULL,NULL,'5795153174'),('14.0.0-KEYCLOAK-18286-revert','keycloak','META-INF/jpa-changelog-14.0.0.xml','2021-11-01 19:32:40',99,'MARK_RAN','7:06499836520f4f6b3d05e35a59324910','dropIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES','',NULL,'3.5.4',NULL,NULL,'5795153174'),('14.0.0-KEYCLOAK-18286-supported-dbs','keycloak','META-INF/jpa-changelog-14.0.0.xml','2021-11-01 19:32:40',100,'EXECUTED','7:b558ad47ea0e4d3c3514225a49cc0d65','createIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES','',NULL,'3.5.4',NULL,NULL,'5795153174'),('14.0.0-KEYCLOAK-18286-unsupported-dbs','keycloak','META-INF/jpa-changelog-14.0.0.xml','2021-11-01 19:32:40',101,'MARK_RAN','7:3d2b23076e59c6f70bae703aa01be35b','createIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES','',NULL,'3.5.4',NULL,NULL,'5795153174'),('KEYCLOAK-17267-add-index-to-user-attributes','keycloak','META-INF/jpa-changelog-14.0.0.xml','2021-11-01 19:32:40',102,'EXECUTED','7:1a7f28ff8d9e53aeb879d76ea3d9341a','createIndex indexName=IDX_USER_ATTRIBUTE_NAME, tableName=USER_ATTRIBUTE','',NULL,'3.5.4',NULL,NULL,'5795153174'),('KEYCLOAK-18146-add-saml-art-binding-identifier','keycloak','META-INF/jpa-changelog-14.0.0.xml','2021-11-01 19:32:40',103,'EXECUTED','7:2fd554456fed4a82c698c555c5b751b6','customChange','',NULL,'3.5.4',NULL,NULL,'5795153174'),('15.0.0-KEYCLOAK-18467','keycloak','META-INF/jpa-changelog-15.0.0.xml','2021-11-01 19:32:40',104,'EXECUTED','7:b06356d66c2790ecc2ae54ba0458397a','addColumn tableName=REALM_LOCALIZATIONS; update tableName=REALM_LOCALIZATIONS; dropColumn columnName=TEXTS, tableName=REALM_LOCALIZATIONS; renameColumn newColumnName=TEXTS, oldColumnName=TEXTS_NEW, tableName=REALM_LOCALIZATIONS; addNotNullConstrai...','',NULL,'3.5.4',NULL,NULL,'5795153174');
/*!40000 ALTER TABLE `DATABASECHANGELOG` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `DATABASECHANGELOGLOCK`
--

DROP TABLE IF EXISTS `DATABASECHANGELOGLOCK`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `DATABASECHANGELOGLOCK` (
  `ID` int(11) NOT NULL,
  `LOCKED` bit(1) NOT NULL,
  `LOCKGRANTED` datetime DEFAULT NULL,
  `LOCKEDBY` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `DATABASECHANGELOGLOCK`
--

LOCK TABLES `DATABASECHANGELOGLOCK` WRITE;
/*!40000 ALTER TABLE `DATABASECHANGELOGLOCK` DISABLE KEYS */;
INSERT INTO `DATABASECHANGELOGLOCK` VALUES (1,'\0',NULL,NULL),(1000,'\0',NULL,NULL),(1001,'\0',NULL,NULL);
/*!40000 ALTER TABLE `DATABASECHANGELOGLOCK` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `DEFAULT_CLIENT_SCOPE`
--

DROP TABLE IF EXISTS `DEFAULT_CLIENT_SCOPE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `DEFAULT_CLIENT_SCOPE` (
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `SCOPE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `DEFAULT_SCOPE` bit(1) NOT NULL DEFAULT b'0',
  PRIMARY KEY (`REALM_ID`,`SCOPE_ID`),
  KEY `IDX_DEFCLS_REALM` (`REALM_ID`),
  KEY `IDX_DEFCLS_SCOPE` (`SCOPE_ID`),
  CONSTRAINT `FK_R_DEF_CLI_SCOPE_REALM` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `DEFAULT_CLIENT_SCOPE`
--

LOCK TABLES `DEFAULT_CLIENT_SCOPE` WRITE;
/*!40000 ALTER TABLE `DEFAULT_CLIENT_SCOPE` DISABLE KEYS */;
INSERT INTO `DEFAULT_CLIENT_SCOPE` VALUES ('master','13052fde-d239-4154-b80b-0f406ed76ded','\0'),('master','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','\0'),('master','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',''),('master','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f','\0'),('master','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','\0'),('master','cb25d275-eff3-4655-b032-e163a0a23c0f',''),('master','e4f019f4-8a8a-4682-bf50-8e883c89cd03',''),('master','f0e07760-3d3d-45d5-b651-403f8b19de35',''),('master','f55ceb89-6d3c-4bcb-882e-44c498d8b305','');
/*!40000 ALTER TABLE `DEFAULT_CLIENT_SCOPE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `EVENT_ENTITY`
--

DROP TABLE IF EXISTS `EVENT_ENTITY`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `EVENT_ENTITY` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `CLIENT_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `DETAILS_JSON` text COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ERROR` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `IP_ADDRESS` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REALM_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `SESSION_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `EVENT_TIME` bigint(20) DEFAULT NULL,
  `TYPE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `USER_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_EVENT_TIME` (`REALM_ID`,`EVENT_TIME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `EVENT_ENTITY`
--

LOCK TABLES `EVENT_ENTITY` WRITE;
/*!40000 ALTER TABLE `EVENT_ENTITY` DISABLE KEYS */;
/*!40000 ALTER TABLE `EVENT_ENTITY` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `FEDERATED_IDENTITY`
--

DROP TABLE IF EXISTS `FEDERATED_IDENTITY`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `FEDERATED_IDENTITY` (
  `IDENTITY_PROVIDER` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `FEDERATED_USER_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `FEDERATED_USERNAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `TOKEN` text COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `USER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`IDENTITY_PROVIDER`,`USER_ID`),
  KEY `IDX_FEDIDENTITY_USER` (`USER_ID`),
  KEY `IDX_FEDIDENTITY_FEDUSER` (`FEDERATED_USER_ID`),
  CONSTRAINT `FK404288B92EF007A6` FOREIGN KEY (`USER_ID`) REFERENCES `USER_ENTITY` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `FEDERATED_IDENTITY`
--

LOCK TABLES `FEDERATED_IDENTITY` WRITE;
/*!40000 ALTER TABLE `FEDERATED_IDENTITY` DISABLE KEYS */;
/*!40000 ALTER TABLE `FEDERATED_IDENTITY` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `FEDERATED_USER`
--

DROP TABLE IF EXISTS `FEDERATED_USER`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `FEDERATED_USER` (
  `ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `STORAGE_PROVIDER_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `FEDERATED_USER`
--

LOCK TABLES `FEDERATED_USER` WRITE;
/*!40000 ALTER TABLE `FEDERATED_USER` DISABLE KEYS */;
/*!40000 ALTER TABLE `FEDERATED_USER` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `FED_USER_ATTRIBUTE`
--

DROP TABLE IF EXISTS `FED_USER_ATTRIBUTE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `FED_USER_ATTRIBUTE` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `USER_ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `STORAGE_PROVIDER_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `VALUE` text COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_FU_ATTRIBUTE` (`USER_ID`,`REALM_ID`,`NAME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `FED_USER_ATTRIBUTE`
--

LOCK TABLES `FED_USER_ATTRIBUTE` WRITE;
/*!40000 ALTER TABLE `FED_USER_ATTRIBUTE` DISABLE KEYS */;
/*!40000 ALTER TABLE `FED_USER_ATTRIBUTE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `FED_USER_CONSENT`
--

DROP TABLE IF EXISTS `FED_USER_CONSENT`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `FED_USER_CONSENT` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `CLIENT_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `USER_ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `STORAGE_PROVIDER_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CREATED_DATE` bigint(20) DEFAULT NULL,
  `LAST_UPDATED_DATE` bigint(20) DEFAULT NULL,
  `CLIENT_STORAGE_PROVIDER` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `EXTERNAL_CLIENT_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_FU_CONSENT` (`USER_ID`,`CLIENT_ID`),
  KEY `IDX_FU_CONSENT_RU` (`REALM_ID`,`USER_ID`),
  KEY `IDX_FU_CNSNT_EXT` (`USER_ID`,`CLIENT_STORAGE_PROVIDER`,`EXTERNAL_CLIENT_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `FED_USER_CONSENT`
--

LOCK TABLES `FED_USER_CONSENT` WRITE;
/*!40000 ALTER TABLE `FED_USER_CONSENT` DISABLE KEYS */;
/*!40000 ALTER TABLE `FED_USER_CONSENT` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `FED_USER_CONSENT_CL_SCOPE`
--

DROP TABLE IF EXISTS `FED_USER_CONSENT_CL_SCOPE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `FED_USER_CONSENT_CL_SCOPE` (
  `USER_CONSENT_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `SCOPE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`USER_CONSENT_ID`,`SCOPE_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `FED_USER_CONSENT_CL_SCOPE`
--

LOCK TABLES `FED_USER_CONSENT_CL_SCOPE` WRITE;
/*!40000 ALTER TABLE `FED_USER_CONSENT_CL_SCOPE` DISABLE KEYS */;
/*!40000 ALTER TABLE `FED_USER_CONSENT_CL_SCOPE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `FED_USER_CREDENTIAL`
--

DROP TABLE IF EXISTS `FED_USER_CREDENTIAL`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `FED_USER_CREDENTIAL` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `SALT` tinyblob DEFAULT NULL,
  `TYPE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CREATED_DATE` bigint(20) DEFAULT NULL,
  `USER_ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `STORAGE_PROVIDER_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `USER_LABEL` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `SECRET_DATA` longtext COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CREDENTIAL_DATA` longtext COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `PRIORITY` int(11) DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_FU_CREDENTIAL` (`USER_ID`,`TYPE`),
  KEY `IDX_FU_CREDENTIAL_RU` (`REALM_ID`,`USER_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `FED_USER_CREDENTIAL`
--

LOCK TABLES `FED_USER_CREDENTIAL` WRITE;
/*!40000 ALTER TABLE `FED_USER_CREDENTIAL` DISABLE KEYS */;
/*!40000 ALTER TABLE `FED_USER_CREDENTIAL` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `FED_USER_GROUP_MEMBERSHIP`
--

DROP TABLE IF EXISTS `FED_USER_GROUP_MEMBERSHIP`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `FED_USER_GROUP_MEMBERSHIP` (
  `GROUP_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `USER_ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `STORAGE_PROVIDER_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`GROUP_ID`,`USER_ID`),
  KEY `IDX_FU_GROUP_MEMBERSHIP` (`USER_ID`,`GROUP_ID`),
  KEY `IDX_FU_GROUP_MEMBERSHIP_RU` (`REALM_ID`,`USER_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `FED_USER_GROUP_MEMBERSHIP`
--

LOCK TABLES `FED_USER_GROUP_MEMBERSHIP` WRITE;
/*!40000 ALTER TABLE `FED_USER_GROUP_MEMBERSHIP` DISABLE KEYS */;
/*!40000 ALTER TABLE `FED_USER_GROUP_MEMBERSHIP` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `FED_USER_REQUIRED_ACTION`
--

DROP TABLE IF EXISTS `FED_USER_REQUIRED_ACTION`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `FED_USER_REQUIRED_ACTION` (
  `REQUIRED_ACTION` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT ' ',
  `USER_ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `STORAGE_PROVIDER_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`REQUIRED_ACTION`,`USER_ID`),
  KEY `IDX_FU_REQUIRED_ACTION` (`USER_ID`,`REQUIRED_ACTION`),
  KEY `IDX_FU_REQUIRED_ACTION_RU` (`REALM_ID`,`USER_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `FED_USER_REQUIRED_ACTION`
--

LOCK TABLES `FED_USER_REQUIRED_ACTION` WRITE;
/*!40000 ALTER TABLE `FED_USER_REQUIRED_ACTION` DISABLE KEYS */;
/*!40000 ALTER TABLE `FED_USER_REQUIRED_ACTION` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `FED_USER_ROLE_MAPPING`
--

DROP TABLE IF EXISTS `FED_USER_ROLE_MAPPING`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `FED_USER_ROLE_MAPPING` (
  `ROLE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `USER_ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `STORAGE_PROVIDER_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ROLE_ID`,`USER_ID`),
  KEY `IDX_FU_ROLE_MAPPING` (`USER_ID`,`ROLE_ID`),
  KEY `IDX_FU_ROLE_MAPPING_RU` (`REALM_ID`,`USER_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `FED_USER_ROLE_MAPPING`
--

LOCK TABLES `FED_USER_ROLE_MAPPING` WRITE;
/*!40000 ALTER TABLE `FED_USER_ROLE_MAPPING` DISABLE KEYS */;
/*!40000 ALTER TABLE `FED_USER_ROLE_MAPPING` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `GROUP_ATTRIBUTE`
--

DROP TABLE IF EXISTS `GROUP_ATTRIBUTE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `GROUP_ATTRIBUTE` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'sybase-needs-something-here',
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(255) CHARACTER SET utf8mb3 DEFAULT NULL,
  `GROUP_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_GROUP_ATTR_GROUP` (`GROUP_ID`),
  CONSTRAINT `FK_GROUP_ATTRIBUTE_GROUP` FOREIGN KEY (`GROUP_ID`) REFERENCES `KEYCLOAK_GROUP` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `GROUP_ATTRIBUTE`
--

LOCK TABLES `GROUP_ATTRIBUTE` WRITE;
/*!40000 ALTER TABLE `GROUP_ATTRIBUTE` DISABLE KEYS */;
/*!40000 ALTER TABLE `GROUP_ATTRIBUTE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `GROUP_ROLE_MAPPING`
--

DROP TABLE IF EXISTS `GROUP_ROLE_MAPPING`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `GROUP_ROLE_MAPPING` (
  `ROLE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `GROUP_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`ROLE_ID`,`GROUP_ID`),
  KEY `IDX_GROUP_ROLE_MAPP_GROUP` (`GROUP_ID`),
  CONSTRAINT `FK_GROUP_ROLE_GROUP` FOREIGN KEY (`GROUP_ID`) REFERENCES `KEYCLOAK_GROUP` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `GROUP_ROLE_MAPPING`
--

LOCK TABLES `GROUP_ROLE_MAPPING` WRITE;
/*!40000 ALTER TABLE `GROUP_ROLE_MAPPING` DISABLE KEYS */;
/*!40000 ALTER TABLE `GROUP_ROLE_MAPPING` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `IDENTITY_PROVIDER`
--

DROP TABLE IF EXISTS `IDENTITY_PROVIDER`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `IDENTITY_PROVIDER` (
  `INTERNAL_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ENABLED` bit(1) NOT NULL DEFAULT b'0',
  `PROVIDER_ALIAS` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `PROVIDER_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `STORE_TOKEN` bit(1) NOT NULL DEFAULT b'0',
  `AUTHENTICATE_BY_DEFAULT` bit(1) NOT NULL DEFAULT b'0',
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ADD_TOKEN_ROLE` bit(1) NOT NULL DEFAULT b'1',
  `TRUST_EMAIL` bit(1) NOT NULL DEFAULT b'0',
  `FIRST_BROKER_LOGIN_FLOW_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `POST_BROKER_LOGIN_FLOW_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `PROVIDER_DISPLAY_NAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `LINK_ONLY` bit(1) NOT NULL DEFAULT b'0',
  PRIMARY KEY (`INTERNAL_ID`),
  UNIQUE KEY `UK_2DAELWNIBJI49AVXSRTUF6XJ33` (`PROVIDER_ALIAS`,`REALM_ID`),
  KEY `IDX_IDENT_PROV_REALM` (`REALM_ID`),
  CONSTRAINT `FK2B4EBC52AE5C3B34` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `IDENTITY_PROVIDER`
--

LOCK TABLES `IDENTITY_PROVIDER` WRITE;
/*!40000 ALTER TABLE `IDENTITY_PROVIDER` DISABLE KEYS */;
/*!40000 ALTER TABLE `IDENTITY_PROVIDER` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `IDENTITY_PROVIDER_CONFIG`
--

DROP TABLE IF EXISTS `IDENTITY_PROVIDER_CONFIG`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `IDENTITY_PROVIDER_CONFIG` (
  `IDENTITY_PROVIDER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` longtext COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`IDENTITY_PROVIDER_ID`,`NAME`),
  CONSTRAINT `FKDC4897CF864C4E43` FOREIGN KEY (`IDENTITY_PROVIDER_ID`) REFERENCES `IDENTITY_PROVIDER` (`INTERNAL_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `IDENTITY_PROVIDER_CONFIG`
--

LOCK TABLES `IDENTITY_PROVIDER_CONFIG` WRITE;
/*!40000 ALTER TABLE `IDENTITY_PROVIDER_CONFIG` DISABLE KEYS */;
/*!40000 ALTER TABLE `IDENTITY_PROVIDER_CONFIG` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `IDENTITY_PROVIDER_MAPPER`
--

DROP TABLE IF EXISTS `IDENTITY_PROVIDER_MAPPER`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `IDENTITY_PROVIDER_MAPPER` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `IDP_ALIAS` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `IDP_MAPPER_NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_ID_PROV_MAPP_REALM` (`REALM_ID`),
  CONSTRAINT `FK_IDPM_REALM` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `IDENTITY_PROVIDER_MAPPER`
--

LOCK TABLES `IDENTITY_PROVIDER_MAPPER` WRITE;
/*!40000 ALTER TABLE `IDENTITY_PROVIDER_MAPPER` DISABLE KEYS */;
/*!40000 ALTER TABLE `IDENTITY_PROVIDER_MAPPER` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `IDP_MAPPER_CONFIG`
--

DROP TABLE IF EXISTS `IDP_MAPPER_CONFIG`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `IDP_MAPPER_CONFIG` (
  `IDP_MAPPER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` longtext COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`IDP_MAPPER_ID`,`NAME`),
  CONSTRAINT `FK_IDPMCONFIG` FOREIGN KEY (`IDP_MAPPER_ID`) REFERENCES `IDENTITY_PROVIDER_MAPPER` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `IDP_MAPPER_CONFIG`
--

LOCK TABLES `IDP_MAPPER_CONFIG` WRITE;
/*!40000 ALTER TABLE `IDP_MAPPER_CONFIG` DISABLE KEYS */;
/*!40000 ALTER TABLE `IDP_MAPPER_CONFIG` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `KEYCLOAK_GROUP`
--

DROP TABLE IF EXISTS `KEYCLOAK_GROUP`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `KEYCLOAK_GROUP` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `NAME` varchar(255) CHARACTER SET utf8mb3 DEFAULT NULL,
  `PARENT_GROUP` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `SIBLING_NAMES` (`REALM_ID`,`PARENT_GROUP`,`NAME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `KEYCLOAK_GROUP`
--

LOCK TABLES `KEYCLOAK_GROUP` WRITE;
/*!40000 ALTER TABLE `KEYCLOAK_GROUP` DISABLE KEYS */;
/*!40000 ALTER TABLE `KEYCLOAK_GROUP` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `KEYCLOAK_ROLE`
--

DROP TABLE IF EXISTS `KEYCLOAK_ROLE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `KEYCLOAK_ROLE` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `CLIENT_REALM_CONSTRAINT` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CLIENT_ROLE` bit(1) DEFAULT NULL,
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb3 DEFAULT NULL,
  `NAME` varchar(255) CHARACTER SET utf8mb3 DEFAULT NULL,
  `REALM_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CLIENT` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REALM` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_J3RWUVD56ONTGSUHOGM184WW2-2` (`NAME`,`CLIENT_REALM_CONSTRAINT`),
  KEY `IDX_KEYCLOAK_ROLE_CLIENT` (`CLIENT`),
  KEY `IDX_KEYCLOAK_ROLE_REALM` (`REALM`),
  CONSTRAINT `FK_6VYQFE4CN4WLQ8R6KT5VDSJ5C` FOREIGN KEY (`REALM`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `KEYCLOAK_ROLE`
--

LOCK TABLES `KEYCLOAK_ROLE` WRITE;
/*!40000 ALTER TABLE `KEYCLOAK_ROLE` DISABLE KEYS */;
INSERT INTO `KEYCLOAK_ROLE` VALUES ('03484350-2ad2-4c70-bc2a-f04dd90d8d44','4e4977d6-eaa9-4245-ae4c-04d20f5436d9','','${role_view-applications}','view-applications','master','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',NULL),('0b1ad207-9fa7-4ed3-84ba-f800607e7d09','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_view-realm}','view-realm','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('16801c90-306a-4dd9-8f04-6a79dc56ce7a','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_create-client}','create-client','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('1aa723ff-209c-4637-b3c8-8159d72e9b09','4e4977d6-eaa9-4245-ae4c-04d20f5436d9','','${role_manage-account}','manage-account','master','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',NULL),('1b33555e-7a77-48cd-8c2a-5b7fae1318df','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_impersonation}','impersonation','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('1d4a0b4b-1f99-4a6e-a9e1-463eec8147fd','4e4977d6-eaa9-4245-ae4c-04d20f5436d9','','${role_view-profile}','view-profile','master','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',NULL),('26c52158-724b-4001-b2ee-e2eb3af6e34d','4e4977d6-eaa9-4245-ae4c-04d20f5436d9','','${role_delete-account}','delete-account','master','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',NULL),('2a2bb3c8-e26f-4adb-9a9e-bbfb1685745e','4e4977d6-eaa9-4245-ae4c-04d20f5436d9','','${role_manage-account-links}','manage-account-links','master','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',NULL),('3b8b7ca4-89c8-4038-a996-c7ce380efa2c','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_view-events}','view-events','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('4d227aaf-b4fa-4a86-9535-30210f612f2e','master','\0','${role_default-roles}','default-roles-master','master',NULL,NULL),('5827ab16-b5bc-4738-b05e-89406e065439','master','\0','${role_admin}','admin','master',NULL,NULL),('70e66b49-b385-466c-95a8-7bff51eee65f','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_query-groups}','query-groups','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('766d6f0c-d33b-480c-b193-98f7dd41b5e6','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_manage-realm}','manage-realm','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('7b199ec2-c91d-4300-a565-f21a8c16b647','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_manage-authorization}','manage-authorization','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('8340533b-d129-4b11-ae94-708efcb2a14a','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_query-clients}','query-clients','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('839188a4-115b-4eab-9fa5-9573ee95fbf0','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_view-users}','view-users','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('8462f2ed-3177-493c-88af-ee6dd888b246','4e4977d6-eaa9-4245-ae4c-04d20f5436d9','','${role_view-consent}','view-consent','master','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',NULL),('9066e812-fa7d-4d41-8b29-7c350d08bedf','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_manage-events}','manage-events','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('982e7b3f-64e8-4d41-9323-7d308fe31b8f','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_query-realms}','query-realms','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('a433b041-695e-4da9-a275-1b6abf9184c9','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_query-users}','query-users','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('b3794300-e7e9-4f3d-af84-7d032a01df6b','master','\0','${role_uma_authorization}','uma_authorization','master',NULL,NULL),('b52948b8-e6cc-45df-aeeb-bfa8210e9378','master','\0','${role_create-realm}','create-realm','master',NULL,NULL),('c9a3a64a-f4f7-46a4-aaf0-b1af52473ce8','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_manage-clients}','manage-clients','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('ce72ef0f-db0f-4eb7-bf60-642c203cc1e9','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_view-clients}','view-clients','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('ceab846f-983c-4c02-a89e-bfac58410427','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_manage-users}','manage-users','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('d3540296-65b3-4fa5-a007-fe0feb99d6d1','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_view-identity-providers}','view-identity-providers','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('d726a590-eea5-4040-85aa-b44f10f3bf58','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_manage-identity-providers}','manage-identity-providers','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('e06c7506-138d-4968-9186-cd958b29e577','master','\0','${role_offline-access}','offline_access','master',NULL,NULL),('e498256c-c0b0-4bfa-a7c3-f10f05de507e','e6b04c6f-e451-49ce-95b1-01b3325b77f7','','${role_view-authorization}','view-authorization','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('f660cd73-8ffe-4077-b73a-b26c6a24f149','4e4977d6-eaa9-4245-ae4c-04d20f5436d9','','${role_manage-consent}','manage-consent','master','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',NULL),('f876756a-cf00-444a-89e8-ea745420c10b','5b62e4f6-f646-4e0b-aa07-83a17a324137','','${role_read-token}','read-token','master','5b62e4f6-f646-4e0b-aa07-83a17a324137',NULL);
/*!40000 ALTER TABLE `KEYCLOAK_ROLE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `MIGRATION_MODEL`
--

DROP TABLE IF EXISTS `MIGRATION_MODEL`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `MIGRATION_MODEL` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VERSION` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `UPDATE_TIME` bigint(20) NOT NULL DEFAULT 0,
  PRIMARY KEY (`ID`),
  KEY `IDX_UPDATE_TIME` (`UPDATE_TIME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `MIGRATION_MODEL`
--

LOCK TABLES `MIGRATION_MODEL` WRITE;
/*!40000 ALTER TABLE `MIGRATION_MODEL` DISABLE KEYS */;
INSERT INTO `MIGRATION_MODEL` VALUES ('a59ul','15.0.2',1635795164);
/*!40000 ALTER TABLE `MIGRATION_MODEL` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `OFFLINE_CLIENT_SESSION`
--

DROP TABLE IF EXISTS `OFFLINE_CLIENT_SESSION`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `OFFLINE_CLIENT_SESSION` (
  `USER_SESSION_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `CLIENT_ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `OFFLINE_FLAG` varchar(4) COLLATE utf8mb4_unicode_ci NOT NULL,
  `TIMESTAMP` int(11) DEFAULT NULL,
  `DATA` longtext COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CLIENT_STORAGE_PROVIDER` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'local',
  `EXTERNAL_CLIENT_ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'local',
  PRIMARY KEY (`USER_SESSION_ID`,`CLIENT_ID`,`CLIENT_STORAGE_PROVIDER`,`EXTERNAL_CLIENT_ID`,`OFFLINE_FLAG`),
  KEY `IDX_US_SESS_ID_ON_CL_SESS` (`USER_SESSION_ID`),
  KEY `IDX_OFFLINE_CSS_PRELOAD` (`CLIENT_ID`,`OFFLINE_FLAG`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `OFFLINE_CLIENT_SESSION`
--

LOCK TABLES `OFFLINE_CLIENT_SESSION` WRITE;
/*!40000 ALTER TABLE `OFFLINE_CLIENT_SESSION` DISABLE KEYS */;
/*!40000 ALTER TABLE `OFFLINE_CLIENT_SESSION` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `OFFLINE_USER_SESSION`
--

DROP TABLE IF EXISTS `OFFLINE_USER_SESSION`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `OFFLINE_USER_SESSION` (
  `USER_SESSION_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `USER_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `CREATED_ON` int(11) NOT NULL,
  `OFFLINE_FLAG` varchar(4) COLLATE utf8mb4_unicode_ci NOT NULL,
  `DATA` longtext COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `LAST_SESSION_REFRESH` int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`USER_SESSION_ID`,`OFFLINE_FLAG`),
  KEY `IDX_OFFLINE_USS_CREATEDON` (`CREATED_ON`),
  KEY `IDX_OFFLINE_USS_PRELOAD` (`OFFLINE_FLAG`,`CREATED_ON`,`USER_SESSION_ID`),
  KEY `IDX_OFFLINE_USS_BY_USER` (`USER_ID`,`REALM_ID`,`OFFLINE_FLAG`),
  KEY `IDX_OFFLINE_USS_BY_USERSESS` (`REALM_ID`,`OFFLINE_FLAG`,`USER_SESSION_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `OFFLINE_USER_SESSION`
--

LOCK TABLES `OFFLINE_USER_SESSION` WRITE;
/*!40000 ALTER TABLE `OFFLINE_USER_SESSION` DISABLE KEYS */;
/*!40000 ALTER TABLE `OFFLINE_USER_SESSION` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `POLICY_CONFIG`
--

DROP TABLE IF EXISTS `POLICY_CONFIG`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `POLICY_CONFIG` (
  `POLICY_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` longtext COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`POLICY_ID`,`NAME`),
  CONSTRAINT `FKDC34197CF864C4E43` FOREIGN KEY (`POLICY_ID`) REFERENCES `RESOURCE_SERVER_POLICY` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `POLICY_CONFIG`
--

LOCK TABLES `POLICY_CONFIG` WRITE;
/*!40000 ALTER TABLE `POLICY_CONFIG` DISABLE KEYS */;
/*!40000 ALTER TABLE `POLICY_CONFIG` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `PROTOCOL_MAPPER`
--

DROP TABLE IF EXISTS `PROTOCOL_MAPPER`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `PROTOCOL_MAPPER` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `PROTOCOL` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `PROTOCOL_MAPPER_NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `CLIENT_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CLIENT_SCOPE_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_PROTOCOL_MAPPER_CLIENT` (`CLIENT_ID`),
  KEY `IDX_CLSCOPE_PROTMAP` (`CLIENT_SCOPE_ID`),
  CONSTRAINT `FK_CLI_SCOPE_MAPPER` FOREIGN KEY (`CLIENT_SCOPE_ID`) REFERENCES `CLIENT_SCOPE` (`ID`),
  CONSTRAINT `FK_PCM_REALM` FOREIGN KEY (`CLIENT_ID`) REFERENCES `CLIENT` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `PROTOCOL_MAPPER`
--

LOCK TABLES `PROTOCOL_MAPPER` WRITE;
/*!40000 ALTER TABLE `PROTOCOL_MAPPER` DISABLE KEYS */;
INSERT INTO `PROTOCOL_MAPPER` VALUES ('036972dc-32e5-4370-bc20-7bd787304f91','zoneinfo','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('0cde626f-2b4f-4d62-8b0b-e632ae2f6dc0','audience resolve','openid-connect','oidc-audience-resolve-mapper','54905dd0-4ade-494e-9c35-ab2d445a99f5',NULL),('0e5b6089-e999-49d4-b8d5-4fac64fad1d9','upn','openid-connect','oidc-usermodel-property-mapper',NULL,'abde17dd-48e0-4d26-a2b7-e75c04b1ac7f'),('1a0f9afc-d076-4cb2-a0a0-201fe9bbeb0d','middle name','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('1f7b81ef-4446-4b48-9b9f-e4c0c7a81abd','username','openid-connect','oidc-usermodel-property-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('2d044ac9-0899-47f7-86a1-f71fffd3f487','allowed web origins','openid-connect','oidc-allowed-origins-mapper',NULL,'e4f019f4-8a8a-4682-bf50-8e883c89cd03'),('467c2758-877f-44d4-ad79-d947350fc843','phone number','openid-connect','oidc-usermodel-attribute-mapper',NULL,'13052fde-d239-4154-b80b-0f406ed76ded'),('573c6159-a4f2-4767-957b-28bb898307d6','phone number verified','openid-connect','oidc-usermodel-attribute-mapper',NULL,'13052fde-d239-4154-b80b-0f406ed76ded'),('58b9f316-ed7a-451c-b8cd-f2deb33b62f7','gender','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('5abd0fe2-dfa1-4ce0-99b2-cd811e0c5df8','website','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('630d46bc-fe33-4462-a13d-0eb9031c1b6d','role list','saml','saml-role-list-mapper',NULL,'f55ceb89-6d3c-4bcb-882e-44c498d8b305'),('654a9342-21fa-47b3-93cf-42ee3dd56c0b','given name','openid-connect','oidc-usermodel-property-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('6b6a6347-32f4-4015-908b-6c1ae96bb533','family name','openid-connect','oidc-usermodel-property-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('8809bca8-02c2-45d3-8a84-9cc04fea1f2a','picture','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('93113378-38c5-468c-ba6e-582a88eee4f3','realm roles','openid-connect','oidc-usermodel-realm-role-mapper',NULL,'f0e07760-3d3d-45d5-b651-403f8b19de35'),('9fafbd6d-4ba4-4e9f-8346-acf7427d6597','locale','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('afda1f0b-3aa3-4559-8f94-f36276c9f3a2','updated at','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('b9f9ef85-c17a-4f48-8d79-df663027d674','address','openid-connect','oidc-address-mapper',NULL,'c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb'),('bd9c1acb-36e7-46e4-b06f-9eee0caca2f3','locale','openid-connect','oidc-usermodel-attribute-mapper','bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc',NULL),('bf71c52a-c4ac-4c40-884c-36eac49a5911','audience resolve','openid-connect','oidc-audience-resolve-mapper',NULL,'f0e07760-3d3d-45d5-b651-403f8b19de35'),('bfb8523f-4da7-4373-892c-8ad546873db9','groups','openid-connect','oidc-usermodel-realm-role-mapper',NULL,'abde17dd-48e0-4d26-a2b7-e75c04b1ac7f'),('c6d317fc-5cca-43d0-8a37-f65018b9967e','birthdate','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('ca4cc8b3-6882-4216-acdd-b491bb219658','nickname','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('da471592-2af0-4299-b6a5-b71c9e48991b','profile','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('e07d0a56-65b6-4a34-bc9c-124e9a85a344','full name','openid-connect','oidc-full-name-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('e85d523d-597b-4357-85ca-c3f8d98eb894','email verified','openid-connect','oidc-usermodel-property-mapper',NULL,'cb25d275-eff3-4655-b032-e163a0a23c0f'),('ec3d6c94-95ae-4baa-82c5-162893cc80d0','client roles','openid-connect','oidc-usermodel-client-role-mapper',NULL,'f0e07760-3d3d-45d5-b651-403f8b19de35'),('fda628cc-59f0-4fc3-8197-627cc7012b2a','email','openid-connect','oidc-usermodel-property-mapper',NULL,'cb25d275-eff3-4655-b032-e163a0a23c0f');
/*!40000 ALTER TABLE `PROTOCOL_MAPPER` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `PROTOCOL_MAPPER_CONFIG`
--

DROP TABLE IF EXISTS `PROTOCOL_MAPPER_CONFIG`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `PROTOCOL_MAPPER_CONFIG` (
  `PROTOCOL_MAPPER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` longtext COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`PROTOCOL_MAPPER_ID`,`NAME`),
  CONSTRAINT `FK_PMCONFIG` FOREIGN KEY (`PROTOCOL_MAPPER_ID`) REFERENCES `PROTOCOL_MAPPER` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `PROTOCOL_MAPPER_CONFIG`
--

LOCK TABLES `PROTOCOL_MAPPER_CONFIG` WRITE;
/*!40000 ALTER TABLE `PROTOCOL_MAPPER_CONFIG` DISABLE KEYS */;
INSERT INTO `PROTOCOL_MAPPER_CONFIG` VALUES ('036972dc-32e5-4370-bc20-7bd787304f91','true','access.token.claim'),('036972dc-32e5-4370-bc20-7bd787304f91','zoneinfo','claim.name'),('036972dc-32e5-4370-bc20-7bd787304f91','true','id.token.claim'),('036972dc-32e5-4370-bc20-7bd787304f91','String','jsonType.label'),('036972dc-32e5-4370-bc20-7bd787304f91','zoneinfo','user.attribute'),('036972dc-32e5-4370-bc20-7bd787304f91','true','userinfo.token.claim'),('0e5b6089-e999-49d4-b8d5-4fac64fad1d9','true','access.token.claim'),('0e5b6089-e999-49d4-b8d5-4fac64fad1d9','upn','claim.name'),('0e5b6089-e999-49d4-b8d5-4fac64fad1d9','true','id.token.claim'),('0e5b6089-e999-49d4-b8d5-4fac64fad1d9','String','jsonType.label'),('0e5b6089-e999-49d4-b8d5-4fac64fad1d9','username','user.attribute'),('0e5b6089-e999-49d4-b8d5-4fac64fad1d9','true','userinfo.token.claim'),('1a0f9afc-d076-4cb2-a0a0-201fe9bbeb0d','true','access.token.claim'),('1a0f9afc-d076-4cb2-a0a0-201fe9bbeb0d','middle_name','claim.name'),('1a0f9afc-d076-4cb2-a0a0-201fe9bbeb0d','true','id.token.claim'),('1a0f9afc-d076-4cb2-a0a0-201fe9bbeb0d','String','jsonType.label'),('1a0f9afc-d076-4cb2-a0a0-201fe9bbeb0d','middleName','user.attribute'),('1a0f9afc-d076-4cb2-a0a0-201fe9bbeb0d','true','userinfo.token.claim'),('1f7b81ef-4446-4b48-9b9f-e4c0c7a81abd','true','access.token.claim'),('1f7b81ef-4446-4b48-9b9f-e4c0c7a81abd','preferred_username','claim.name'),('1f7b81ef-4446-4b48-9b9f-e4c0c7a81abd','true','id.token.claim'),('1f7b81ef-4446-4b48-9b9f-e4c0c7a81abd','String','jsonType.label'),('1f7b81ef-4446-4b48-9b9f-e4c0c7a81abd','username','user.attribute'),('1f7b81ef-4446-4b48-9b9f-e4c0c7a81abd','true','userinfo.token.claim'),('467c2758-877f-44d4-ad79-d947350fc843','true','access.token.claim'),('467c2758-877f-44d4-ad79-d947350fc843','phone_number','claim.name'),('467c2758-877f-44d4-ad79-d947350fc843','true','id.token.claim'),('467c2758-877f-44d4-ad79-d947350fc843','String','jsonType.label'),('467c2758-877f-44d4-ad79-d947350fc843','phoneNumber','user.attribute'),('467c2758-877f-44d4-ad79-d947350fc843','true','userinfo.token.claim'),('573c6159-a4f2-4767-957b-28bb898307d6','true','access.token.claim'),('573c6159-a4f2-4767-957b-28bb898307d6','phone_number_verified','claim.name'),('573c6159-a4f2-4767-957b-28bb898307d6','true','id.token.claim'),('573c6159-a4f2-4767-957b-28bb898307d6','boolean','jsonType.label'),('573c6159-a4f2-4767-957b-28bb898307d6','phoneNumberVerified','user.attribute'),('573c6159-a4f2-4767-957b-28bb898307d6','true','userinfo.token.claim'),('58b9f316-ed7a-451c-b8cd-f2deb33b62f7','true','access.token.claim'),('58b9f316-ed7a-451c-b8cd-f2deb33b62f7','gender','claim.name'),('58b9f316-ed7a-451c-b8cd-f2deb33b62f7','true','id.token.claim'),('58b9f316-ed7a-451c-b8cd-f2deb33b62f7','String','jsonType.label'),('58b9f316-ed7a-451c-b8cd-f2deb33b62f7','gender','user.attribute'),('58b9f316-ed7a-451c-b8cd-f2deb33b62f7','true','userinfo.token.claim'),('5abd0fe2-dfa1-4ce0-99b2-cd811e0c5df8','true','access.token.claim'),('5abd0fe2-dfa1-4ce0-99b2-cd811e0c5df8','website','claim.name'),('5abd0fe2-dfa1-4ce0-99b2-cd811e0c5df8','true','id.token.claim'),('5abd0fe2-dfa1-4ce0-99b2-cd811e0c5df8','String','jsonType.label'),('5abd0fe2-dfa1-4ce0-99b2-cd811e0c5df8','website','user.attribute'),('5abd0fe2-dfa1-4ce0-99b2-cd811e0c5df8','true','userinfo.token.claim'),('630d46bc-fe33-4462-a13d-0eb9031c1b6d','Role','attribute.name'),('630d46bc-fe33-4462-a13d-0eb9031c1b6d','Basic','attribute.nameformat'),('630d46bc-fe33-4462-a13d-0eb9031c1b6d','false','single'),('654a9342-21fa-47b3-93cf-42ee3dd56c0b','true','access.token.claim'),('654a9342-21fa-47b3-93cf-42ee3dd56c0b','given_name','claim.name'),('654a9342-21fa-47b3-93cf-42ee3dd56c0b','true','id.token.claim'),('654a9342-21fa-47b3-93cf-42ee3dd56c0b','String','jsonType.label'),('654a9342-21fa-47b3-93cf-42ee3dd56c0b','firstName','user.attribute'),('654a9342-21fa-47b3-93cf-42ee3dd56c0b','true','userinfo.token.claim'),('6b6a6347-32f4-4015-908b-6c1ae96bb533','true','access.token.claim'),('6b6a6347-32f4-4015-908b-6c1ae96bb533','family_name','claim.name'),('6b6a6347-32f4-4015-908b-6c1ae96bb533','true','id.token.claim'),('6b6a6347-32f4-4015-908b-6c1ae96bb533','String','jsonType.label'),('6b6a6347-32f4-4015-908b-6c1ae96bb533','lastName','user.attribute'),('6b6a6347-32f4-4015-908b-6c1ae96bb533','true','userinfo.token.claim'),('8809bca8-02c2-45d3-8a84-9cc04fea1f2a','true','access.token.claim'),('8809bca8-02c2-45d3-8a84-9cc04fea1f2a','picture','claim.name'),('8809bca8-02c2-45d3-8a84-9cc04fea1f2a','true','id.token.claim'),('8809bca8-02c2-45d3-8a84-9cc04fea1f2a','String','jsonType.label'),('8809bca8-02c2-45d3-8a84-9cc04fea1f2a','picture','user.attribute'),('8809bca8-02c2-45d3-8a84-9cc04fea1f2a','true','userinfo.token.claim'),('93113378-38c5-468c-ba6e-582a88eee4f3','true','access.token.claim'),('93113378-38c5-468c-ba6e-582a88eee4f3','realm_access.roles','claim.name'),('93113378-38c5-468c-ba6e-582a88eee4f3','String','jsonType.label'),('93113378-38c5-468c-ba6e-582a88eee4f3','true','multivalued'),('93113378-38c5-468c-ba6e-582a88eee4f3','foo','user.attribute'),('9fafbd6d-4ba4-4e9f-8346-acf7427d6597','true','access.token.claim'),('9fafbd6d-4ba4-4e9f-8346-acf7427d6597','locale','claim.name'),('9fafbd6d-4ba4-4e9f-8346-acf7427d6597','true','id.token.claim'),('9fafbd6d-4ba4-4e9f-8346-acf7427d6597','String','jsonType.label'),('9fafbd6d-4ba4-4e9f-8346-acf7427d6597','locale','user.attribute'),('9fafbd6d-4ba4-4e9f-8346-acf7427d6597','true','userinfo.token.claim'),('afda1f0b-3aa3-4559-8f94-f36276c9f3a2','true','access.token.claim'),('afda1f0b-3aa3-4559-8f94-f36276c9f3a2','updated_at','claim.name'),('afda1f0b-3aa3-4559-8f94-f36276c9f3a2','true','id.token.claim'),('afda1f0b-3aa3-4559-8f94-f36276c9f3a2','String','jsonType.label'),('afda1f0b-3aa3-4559-8f94-f36276c9f3a2','updatedAt','user.attribute'),('afda1f0b-3aa3-4559-8f94-f36276c9f3a2','true','userinfo.token.claim'),('b9f9ef85-c17a-4f48-8d79-df663027d674','true','access.token.claim'),('b9f9ef85-c17a-4f48-8d79-df663027d674','true','id.token.claim'),('b9f9ef85-c17a-4f48-8d79-df663027d674','country','user.attribute.country'),('b9f9ef85-c17a-4f48-8d79-df663027d674','formatted','user.attribute.formatted'),('b9f9ef85-c17a-4f48-8d79-df663027d674','locality','user.attribute.locality'),('b9f9ef85-c17a-4f48-8d79-df663027d674','postal_code','user.attribute.postal_code'),('b9f9ef85-c17a-4f48-8d79-df663027d674','region','user.attribute.region'),('b9f9ef85-c17a-4f48-8d79-df663027d674','street','user.attribute.street'),('b9f9ef85-c17a-4f48-8d79-df663027d674','true','userinfo.token.claim'),('bd9c1acb-36e7-46e4-b06f-9eee0caca2f3','true','access.token.claim'),('bd9c1acb-36e7-46e4-b06f-9eee0caca2f3','locale','claim.name'),('bd9c1acb-36e7-46e4-b06f-9eee0caca2f3','true','id.token.claim'),('bd9c1acb-36e7-46e4-b06f-9eee0caca2f3','String','jsonType.label'),('bd9c1acb-36e7-46e4-b06f-9eee0caca2f3','locale','user.attribute'),('bd9c1acb-36e7-46e4-b06f-9eee0caca2f3','true','userinfo.token.claim'),('bfb8523f-4da7-4373-892c-8ad546873db9','true','access.token.claim'),('bfb8523f-4da7-4373-892c-8ad546873db9','groups','claim.name'),('bfb8523f-4da7-4373-892c-8ad546873db9','true','id.token.claim'),('bfb8523f-4da7-4373-892c-8ad546873db9','String','jsonType.label'),('bfb8523f-4da7-4373-892c-8ad546873db9','true','multivalued'),('bfb8523f-4da7-4373-892c-8ad546873db9','foo','user.attribute'),('c6d317fc-5cca-43d0-8a37-f65018b9967e','true','access.token.claim'),('c6d317fc-5cca-43d0-8a37-f65018b9967e','birthdate','claim.name'),('c6d317fc-5cca-43d0-8a37-f65018b9967e','true','id.token.claim'),('c6d317fc-5cca-43d0-8a37-f65018b9967e','String','jsonType.label'),('c6d317fc-5cca-43d0-8a37-f65018b9967e','birthdate','user.attribute'),('c6d317fc-5cca-43d0-8a37-f65018b9967e','true','userinfo.token.claim'),('ca4cc8b3-6882-4216-acdd-b491bb219658','true','access.token.claim'),('ca4cc8b3-6882-4216-acdd-b491bb219658','nickname','claim.name'),('ca4cc8b3-6882-4216-acdd-b491bb219658','true','id.token.claim'),('ca4cc8b3-6882-4216-acdd-b491bb219658','String','jsonType.label'),('ca4cc8b3-6882-4216-acdd-b491bb219658','nickname','user.attribute'),('ca4cc8b3-6882-4216-acdd-b491bb219658','true','userinfo.token.claim'),('da471592-2af0-4299-b6a5-b71c9e48991b','true','access.token.claim'),('da471592-2af0-4299-b6a5-b71c9e48991b','profile','claim.name'),('da471592-2af0-4299-b6a5-b71c9e48991b','true','id.token.claim'),('da471592-2af0-4299-b6a5-b71c9e48991b','String','jsonType.label'),('da471592-2af0-4299-b6a5-b71c9e48991b','profile','user.attribute'),('da471592-2af0-4299-b6a5-b71c9e48991b','true','userinfo.token.claim'),('e07d0a56-65b6-4a34-bc9c-124e9a85a344','true','access.token.claim'),('e07d0a56-65b6-4a34-bc9c-124e9a85a344','true','id.token.claim'),('e07d0a56-65b6-4a34-bc9c-124e9a85a344','true','userinfo.token.claim'),('e85d523d-597b-4357-85ca-c3f8d98eb894','true','access.token.claim'),('e85d523d-597b-4357-85ca-c3f8d98eb894','email_verified','claim.name'),('e85d523d-597b-4357-85ca-c3f8d98eb894','true','id.token.claim'),('e85d523d-597b-4357-85ca-c3f8d98eb894','boolean','jsonType.label'),('e85d523d-597b-4357-85ca-c3f8d98eb894','emailVerified','user.attribute'),('e85d523d-597b-4357-85ca-c3f8d98eb894','true','userinfo.token.claim'),('ec3d6c94-95ae-4baa-82c5-162893cc80d0','true','access.token.claim'),('ec3d6c94-95ae-4baa-82c5-162893cc80d0','resource_access.${client_id}.roles','claim.name'),('ec3d6c94-95ae-4baa-82c5-162893cc80d0','String','jsonType.label'),('ec3d6c94-95ae-4baa-82c5-162893cc80d0','true','multivalued'),('ec3d6c94-95ae-4baa-82c5-162893cc80d0','foo','user.attribute'),('fda628cc-59f0-4fc3-8197-627cc7012b2a','true','access.token.claim'),('fda628cc-59f0-4fc3-8197-627cc7012b2a','email','claim.name'),('fda628cc-59f0-4fc3-8197-627cc7012b2a','true','id.token.claim'),('fda628cc-59f0-4fc3-8197-627cc7012b2a','String','jsonType.label'),('fda628cc-59f0-4fc3-8197-627cc7012b2a','email','user.attribute'),('fda628cc-59f0-4fc3-8197-627cc7012b2a','true','userinfo.token.claim');
/*!40000 ALTER TABLE `PROTOCOL_MAPPER_CONFIG` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `REALM`
--

DROP TABLE IF EXISTS `REALM`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `REALM` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ACCESS_CODE_LIFESPAN` int(11) DEFAULT NULL,
  `USER_ACTION_LIFESPAN` int(11) DEFAULT NULL,
  `ACCESS_TOKEN_LIFESPAN` int(11) DEFAULT NULL,
  `ACCOUNT_THEME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ADMIN_THEME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `EMAIL_THEME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ENABLED` bit(1) NOT NULL DEFAULT b'0',
  `EVENTS_ENABLED` bit(1) NOT NULL DEFAULT b'0',
  `EVENTS_EXPIRATION` bigint(20) DEFAULT NULL,
  `LOGIN_THEME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NOT_BEFORE` int(11) DEFAULT NULL,
  `PASSWORD_POLICY` text COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REGISTRATION_ALLOWED` bit(1) NOT NULL DEFAULT b'0',
  `REMEMBER_ME` bit(1) NOT NULL DEFAULT b'0',
  `RESET_PASSWORD_ALLOWED` bit(1) NOT NULL DEFAULT b'0',
  `SOCIAL` bit(1) NOT NULL DEFAULT b'0',
  `SSL_REQUIRED` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `SSO_IDLE_TIMEOUT` int(11) DEFAULT NULL,
  `SSO_MAX_LIFESPAN` int(11) DEFAULT NULL,
  `UPDATE_PROFILE_ON_SOC_LOGIN` bit(1) NOT NULL DEFAULT b'0',
  `VERIFY_EMAIL` bit(1) NOT NULL DEFAULT b'0',
  `MASTER_ADMIN_CLIENT` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `LOGIN_LIFESPAN` int(11) DEFAULT NULL,
  `INTERNATIONALIZATION_ENABLED` bit(1) NOT NULL DEFAULT b'0',
  `DEFAULT_LOCALE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REG_EMAIL_AS_USERNAME` bit(1) NOT NULL DEFAULT b'0',
  `ADMIN_EVENTS_ENABLED` bit(1) NOT NULL DEFAULT b'0',
  `ADMIN_EVENTS_DETAILS_ENABLED` bit(1) NOT NULL DEFAULT b'0',
  `EDIT_USERNAME_ALLOWED` bit(1) NOT NULL DEFAULT b'0',
  `OTP_POLICY_COUNTER` int(11) DEFAULT 0,
  `OTP_POLICY_WINDOW` int(11) DEFAULT 1,
  `OTP_POLICY_PERIOD` int(11) DEFAULT 30,
  `OTP_POLICY_DIGITS` int(11) DEFAULT 6,
  `OTP_POLICY_ALG` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT 'HmacSHA1',
  `OTP_POLICY_TYPE` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT 'totp',
  `BROWSER_FLOW` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REGISTRATION_FLOW` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `DIRECT_GRANT_FLOW` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `RESET_CREDENTIALS_FLOW` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CLIENT_AUTH_FLOW` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `OFFLINE_SESSION_IDLE_TIMEOUT` int(11) DEFAULT 0,
  `REVOKE_REFRESH_TOKEN` bit(1) NOT NULL DEFAULT b'0',
  `ACCESS_TOKEN_LIFE_IMPLICIT` int(11) DEFAULT 0,
  `LOGIN_WITH_EMAIL_ALLOWED` bit(1) NOT NULL DEFAULT b'1',
  `DUPLICATE_EMAILS_ALLOWED` bit(1) NOT NULL DEFAULT b'0',
  `DOCKER_AUTH_FLOW` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REFRESH_TOKEN_MAX_REUSE` int(11) DEFAULT 0,
  `ALLOW_USER_MANAGED_ACCESS` bit(1) NOT NULL DEFAULT b'0',
  `SSO_MAX_LIFESPAN_REMEMBER_ME` int(11) NOT NULL,
  `SSO_IDLE_TIMEOUT_REMEMBER_ME` int(11) NOT NULL,
  `DEFAULT_ROLE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_ORVSDMLA56612EAEFIQ6WL5OI` (`NAME`),
  KEY `IDX_REALM_MASTER_ADM_CLI` (`MASTER_ADMIN_CLIENT`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `REALM`
--

LOCK TABLES `REALM` WRITE;
/*!40000 ALTER TABLE `REALM` DISABLE KEYS */;
INSERT INTO `REALM` VALUES ('master',60,300,60,NULL,NULL,NULL,'','\0',0,NULL,'master',0,NULL,'\0','\0','\0','\0','EXTERNAL',1800,36000,'\0','\0','e6b04c6f-e451-49ce-95b1-01b3325b77f7',1800,'\0',NULL,'\0','\0','\0','\0',0,1,30,6,'HmacSHA1','totp','5adc5270-4510-48ee-898c-cbae8f28a3cd','a1a80a32-2bfd-4d72-9702-b41896868e69','7083d693-5892-42b4-a592-f63c604dd8dc','9e586e10-7d4f-4ad6-977a-fd05dffd6ee6','d180151b-30b4-4438-8ec3-9033d2e60c38',2592000,'\0',900,'','\0','df0e3e20-eeea-4996-ba53-8ec117fcbec1',0,'\0',0,0,'4d227aaf-b4fa-4a86-9535-30210f612f2e');
/*!40000 ALTER TABLE `REALM` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `REALM_ATTRIBUTE`
--

DROP TABLE IF EXISTS `REALM_ATTRIBUTE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `REALM_ATTRIBUTE` (
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` longtext CHARACTER SET utf8mb3 DEFAULT NULL,
  PRIMARY KEY (`NAME`,`REALM_ID`),
  KEY `IDX_REALM_ATTR_REALM` (`REALM_ID`),
  CONSTRAINT `FK_8SHXD6L3E9ATQUKACXGPFFPTW` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `REALM_ATTRIBUTE`
--

LOCK TABLES `REALM_ATTRIBUTE` WRITE;
/*!40000 ALTER TABLE `REALM_ATTRIBUTE` DISABLE KEYS */;
INSERT INTO `REALM_ATTRIBUTE` VALUES ('_browser_header.contentSecurityPolicy','master','frame-src \'self\'; frame-ancestors \'self\'; object-src \'none\';'),('_browser_header.contentSecurityPolicyReportOnly','master',''),('_browser_header.strictTransportSecurity','master','max-age=31536000; includeSubDomains'),('_browser_header.xContentTypeOptions','master','nosniff'),('_browser_header.xFrameOptions','master','SAMEORIGIN'),('_browser_header.xRobotsTag','master','none'),('_browser_header.xXSSProtection','master','1; mode=block'),('bruteForceProtected','master','false'),('defaultSignatureAlgorithm','master','RS256'),('displayName','master','Keycloak'),('displayNameHtml','master','<div class=\"kc-logo-text\"><span>Keycloak</span></div>'),('failureFactor','master','30'),('maxDeltaTimeSeconds','master','43200'),('maxFailureWaitSeconds','master','900'),('minimumQuickLoginWaitSeconds','master','60'),('offlineSessionMaxLifespan','master','5184000'),('offlineSessionMaxLifespanEnabled','master','false'),('permanentLockout','master','false'),('quickLoginCheckMilliSeconds','master','1000'),('waitIncrementSeconds','master','60');
/*!40000 ALTER TABLE `REALM_ATTRIBUTE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `REALM_DEFAULT_GROUPS`
--

DROP TABLE IF EXISTS `REALM_DEFAULT_GROUPS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `REALM_DEFAULT_GROUPS` (
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `GROUP_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`REALM_ID`,`GROUP_ID`),
  UNIQUE KEY `CON_GROUP_ID_DEF_GROUPS` (`GROUP_ID`),
  KEY `IDX_REALM_DEF_GRP_REALM` (`REALM_ID`),
  CONSTRAINT `FK_DEF_GROUPS_REALM` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `REALM_DEFAULT_GROUPS`
--

LOCK TABLES `REALM_DEFAULT_GROUPS` WRITE;
/*!40000 ALTER TABLE `REALM_DEFAULT_GROUPS` DISABLE KEYS */;
/*!40000 ALTER TABLE `REALM_DEFAULT_GROUPS` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `REALM_ENABLED_EVENT_TYPES`
--

DROP TABLE IF EXISTS `REALM_ENABLED_EVENT_TYPES`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `REALM_ENABLED_EVENT_TYPES` (
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`REALM_ID`,`VALUE`),
  KEY `IDX_REALM_EVT_TYPES_REALM` (`REALM_ID`),
  CONSTRAINT `FK_H846O4H0W8EPX5NWEDRF5Y69J` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `REALM_ENABLED_EVENT_TYPES`
--

LOCK TABLES `REALM_ENABLED_EVENT_TYPES` WRITE;
/*!40000 ALTER TABLE `REALM_ENABLED_EVENT_TYPES` DISABLE KEYS */;
/*!40000 ALTER TABLE `REALM_ENABLED_EVENT_TYPES` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `REALM_EVENTS_LISTENERS`
--

DROP TABLE IF EXISTS `REALM_EVENTS_LISTENERS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `REALM_EVENTS_LISTENERS` (
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`REALM_ID`,`VALUE`),
  KEY `IDX_REALM_EVT_LIST_REALM` (`REALM_ID`),
  CONSTRAINT `FK_H846O4H0W8EPX5NXEV9F5Y69J` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `REALM_EVENTS_LISTENERS`
--

LOCK TABLES `REALM_EVENTS_LISTENERS` WRITE;
/*!40000 ALTER TABLE `REALM_EVENTS_LISTENERS` DISABLE KEYS */;
INSERT INTO `REALM_EVENTS_LISTENERS` VALUES ('master','jboss-logging');
/*!40000 ALTER TABLE `REALM_EVENTS_LISTENERS` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `REALM_LOCALIZATIONS`
--

DROP TABLE IF EXISTS `REALM_LOCALIZATIONS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `REALM_LOCALIZATIONS` (
  `REALM_ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `LOCALE` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `TEXTS` longtext CHARACTER SET utf8mb3 NOT NULL,
  PRIMARY KEY (`REALM_ID`,`LOCALE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `REALM_LOCALIZATIONS`
--

LOCK TABLES `REALM_LOCALIZATIONS` WRITE;
/*!40000 ALTER TABLE `REALM_LOCALIZATIONS` DISABLE KEYS */;
/*!40000 ALTER TABLE `REALM_LOCALIZATIONS` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `REALM_REQUIRED_CREDENTIAL`
--

DROP TABLE IF EXISTS `REALM_REQUIRED_CREDENTIAL`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `REALM_REQUIRED_CREDENTIAL` (
  `TYPE` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `FORM_LABEL` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `INPUT` bit(1) NOT NULL DEFAULT b'0',
  `SECRET` bit(1) NOT NULL DEFAULT b'0',
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`REALM_ID`,`TYPE`),
  CONSTRAINT `FK_5HG65LYBEVAVKQFKI3KPONH9V` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `REALM_REQUIRED_CREDENTIAL`
--

LOCK TABLES `REALM_REQUIRED_CREDENTIAL` WRITE;
/*!40000 ALTER TABLE `REALM_REQUIRED_CREDENTIAL` DISABLE KEYS */;
INSERT INTO `REALM_REQUIRED_CREDENTIAL` VALUES ('password','password','','','master');
/*!40000 ALTER TABLE `REALM_REQUIRED_CREDENTIAL` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `REALM_SMTP_CONFIG`
--

DROP TABLE IF EXISTS `REALM_SMTP_CONFIG`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `REALM_SMTP_CONFIG` (
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`REALM_ID`,`NAME`),
  CONSTRAINT `FK_70EJ8XDXGXD0B9HH6180IRR0O` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `REALM_SMTP_CONFIG`
--

LOCK TABLES `REALM_SMTP_CONFIG` WRITE;
/*!40000 ALTER TABLE `REALM_SMTP_CONFIG` DISABLE KEYS */;
/*!40000 ALTER TABLE `REALM_SMTP_CONFIG` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `REALM_SUPPORTED_LOCALES`
--

DROP TABLE IF EXISTS `REALM_SUPPORTED_LOCALES`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `REALM_SUPPORTED_LOCALES` (
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`REALM_ID`,`VALUE`),
  KEY `IDX_REALM_SUPP_LOCAL_REALM` (`REALM_ID`),
  CONSTRAINT `FK_SUPPORTED_LOCALES_REALM` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `REALM_SUPPORTED_LOCALES`
--

LOCK TABLES `REALM_SUPPORTED_LOCALES` WRITE;
/*!40000 ALTER TABLE `REALM_SUPPORTED_LOCALES` DISABLE KEYS */;
/*!40000 ALTER TABLE `REALM_SUPPORTED_LOCALES` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `REDIRECT_URIS`
--

DROP TABLE IF EXISTS `REDIRECT_URIS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `REDIRECT_URIS` (
  `CLIENT_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`CLIENT_ID`,`VALUE`),
  KEY `IDX_REDIR_URI_CLIENT` (`CLIENT_ID`),
  CONSTRAINT `FK_1BURS8PB4OUJ97H5WUPPAHV9F` FOREIGN KEY (`CLIENT_ID`) REFERENCES `CLIENT` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `REDIRECT_URIS`
--

LOCK TABLES `REDIRECT_URIS` WRITE;
/*!40000 ALTER TABLE `REDIRECT_URIS` DISABLE KEYS */;
INSERT INTO `REDIRECT_URIS` VALUES ('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','/realms/master/account/*'),('54905dd0-4ade-494e-9c35-ab2d445a99f5','/realms/master/account/*'),('5a059221-51fd-434f-84a6-40fa51cda5ce','https://app.localssl.dev/api/v1/oidc/redirect'),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','/admin/master/console/*');
/*!40000 ALTER TABLE `REDIRECT_URIS` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `REQUIRED_ACTION_CONFIG`
--

DROP TABLE IF EXISTS `REQUIRED_ACTION_CONFIG`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `REQUIRED_ACTION_CONFIG` (
  `REQUIRED_ACTION_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` longtext COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`REQUIRED_ACTION_ID`,`NAME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `REQUIRED_ACTION_CONFIG`
--

LOCK TABLES `REQUIRED_ACTION_CONFIG` WRITE;
/*!40000 ALTER TABLE `REQUIRED_ACTION_CONFIG` DISABLE KEYS */;
/*!40000 ALTER TABLE `REQUIRED_ACTION_CONFIG` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `REQUIRED_ACTION_PROVIDER`
--

DROP TABLE IF EXISTS `REQUIRED_ACTION_PROVIDER`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `REQUIRED_ACTION_PROVIDER` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ALIAS` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ENABLED` bit(1) NOT NULL DEFAULT b'0',
  `DEFAULT_ACTION` bit(1) NOT NULL DEFAULT b'0',
  `PROVIDER_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `PRIORITY` int(11) DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_REQ_ACT_PROV_REALM` (`REALM_ID`),
  CONSTRAINT `FK_REQ_ACT_REALM` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `REQUIRED_ACTION_PROVIDER`
--

LOCK TABLES `REQUIRED_ACTION_PROVIDER` WRITE;
/*!40000 ALTER TABLE `REQUIRED_ACTION_PROVIDER` DISABLE KEYS */;
INSERT INTO `REQUIRED_ACTION_PROVIDER` VALUES ('0c0818a4-641d-42e0-9b86-8bfeb7ee7368','CONFIGURE_TOTP','Configure OTP','master','','\0','CONFIGURE_TOTP',10),('1781b401-8336-4a8c-a102-4a092c723cd3','UPDATE_PASSWORD','Update Password','master','','\0','UPDATE_PASSWORD',30),('1bbbf0d1-e6f8-42b4-8741-19d1c59af15f','delete_account','Delete Account','master','\0','\0','delete_account',60),('49195d42-495c-42cf-828e-f736ff686b9b','update_user_locale','Update User Locale','master','','\0','update_user_locale',1000),('4a942b81-ccbb-49f9-a510-db0dda0d4ed9','VERIFY_EMAIL','Verify Email','master','','\0','VERIFY_EMAIL',50),('5508dda9-65dc-40de-9e49-baf5d918b980','terms_and_conditions','Terms and Conditions','master','\0','\0','terms_and_conditions',20),('556d4b2e-61e8-40db-924f-b0f0bdcae242','UPDATE_PROFILE','Update Profile','master','','\0','UPDATE_PROFILE',40);
/*!40000 ALTER TABLE `REQUIRED_ACTION_PROVIDER` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `RESOURCE_ATTRIBUTE`
--

DROP TABLE IF EXISTS `RESOURCE_ATTRIBUTE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `RESOURCE_ATTRIBUTE` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'sybase-needs-something-here',
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `RESOURCE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `FK_5HRM2VLF9QL5FU022KQEPOVBR` (`RESOURCE_ID`),
  CONSTRAINT `FK_5HRM2VLF9QL5FU022KQEPOVBR` FOREIGN KEY (`RESOURCE_ID`) REFERENCES `RESOURCE_SERVER_RESOURCE` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `RESOURCE_ATTRIBUTE`
--

LOCK TABLES `RESOURCE_ATTRIBUTE` WRITE;
/*!40000 ALTER TABLE `RESOURCE_ATTRIBUTE` DISABLE KEYS */;
/*!40000 ALTER TABLE `RESOURCE_ATTRIBUTE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `RESOURCE_POLICY`
--

DROP TABLE IF EXISTS `RESOURCE_POLICY`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `RESOURCE_POLICY` (
  `RESOURCE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `POLICY_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`RESOURCE_ID`,`POLICY_ID`),
  KEY `IDX_RES_POLICY_POLICY` (`POLICY_ID`),
  CONSTRAINT `FK_FRSRPOS53XCX4WNKOG82SSRFY` FOREIGN KEY (`RESOURCE_ID`) REFERENCES `RESOURCE_SERVER_RESOURCE` (`ID`),
  CONSTRAINT `FK_FRSRPP213XCX4WNKOG82SSRFY` FOREIGN KEY (`POLICY_ID`) REFERENCES `RESOURCE_SERVER_POLICY` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `RESOURCE_POLICY`
--

LOCK TABLES `RESOURCE_POLICY` WRITE;
/*!40000 ALTER TABLE `RESOURCE_POLICY` DISABLE KEYS */;
/*!40000 ALTER TABLE `RESOURCE_POLICY` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `RESOURCE_SCOPE`
--

DROP TABLE IF EXISTS `RESOURCE_SCOPE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `RESOURCE_SCOPE` (
  `RESOURCE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `SCOPE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`RESOURCE_ID`,`SCOPE_ID`),
  KEY `IDX_RES_SCOPE_SCOPE` (`SCOPE_ID`),
  CONSTRAINT `FK_FRSRPOS13XCX4WNKOG82SSRFY` FOREIGN KEY (`RESOURCE_ID`) REFERENCES `RESOURCE_SERVER_RESOURCE` (`ID`),
  CONSTRAINT `FK_FRSRPS213XCX4WNKOG82SSRFY` FOREIGN KEY (`SCOPE_ID`) REFERENCES `RESOURCE_SERVER_SCOPE` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `RESOURCE_SCOPE`
--

LOCK TABLES `RESOURCE_SCOPE` WRITE;
/*!40000 ALTER TABLE `RESOURCE_SCOPE` DISABLE KEYS */;
/*!40000 ALTER TABLE `RESOURCE_SCOPE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `RESOURCE_SERVER`
--

DROP TABLE IF EXISTS `RESOURCE_SERVER`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `RESOURCE_SERVER` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ALLOW_RS_REMOTE_MGMT` bit(1) NOT NULL DEFAULT b'0',
  `POLICY_ENFORCE_MODE` varchar(15) COLLATE utf8mb4_unicode_ci NOT NULL,
  `DECISION_STRATEGY` tinyint(4) NOT NULL DEFAULT 1,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `RESOURCE_SERVER`
--

LOCK TABLES `RESOURCE_SERVER` WRITE;
/*!40000 ALTER TABLE `RESOURCE_SERVER` DISABLE KEYS */;
/*!40000 ALTER TABLE `RESOURCE_SERVER` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `RESOURCE_SERVER_PERM_TICKET`
--

DROP TABLE IF EXISTS `RESOURCE_SERVER_PERM_TICKET`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `RESOURCE_SERVER_PERM_TICKET` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `OWNER` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REQUESTER` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CREATED_TIMESTAMP` bigint(20) NOT NULL,
  `GRANTED_TIMESTAMP` bigint(20) DEFAULT NULL,
  `RESOURCE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `SCOPE_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `RESOURCE_SERVER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `POLICY_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_FRSR6T700S9V50BU18WS5PMT` (`OWNER`,`REQUESTER`,`RESOURCE_SERVER_ID`,`RESOURCE_ID`,`SCOPE_ID`),
  KEY `FK_FRSRHO213XCX4WNKOG82SSPMT` (`RESOURCE_SERVER_ID`),
  KEY `FK_FRSRHO213XCX4WNKOG83SSPMT` (`RESOURCE_ID`),
  KEY `FK_FRSRHO213XCX4WNKOG84SSPMT` (`SCOPE_ID`),
  KEY `FK_FRSRPO2128CX4WNKOG82SSRFY` (`POLICY_ID`),
  CONSTRAINT `FK_FRSRHO213XCX4WNKOG82SSPMT` FOREIGN KEY (`RESOURCE_SERVER_ID`) REFERENCES `RESOURCE_SERVER` (`ID`),
  CONSTRAINT `FK_FRSRHO213XCX4WNKOG83SSPMT` FOREIGN KEY (`RESOURCE_ID`) REFERENCES `RESOURCE_SERVER_RESOURCE` (`ID`),
  CONSTRAINT `FK_FRSRHO213XCX4WNKOG84SSPMT` FOREIGN KEY (`SCOPE_ID`) REFERENCES `RESOURCE_SERVER_SCOPE` (`ID`),
  CONSTRAINT `FK_FRSRPO2128CX4WNKOG82SSRFY` FOREIGN KEY (`POLICY_ID`) REFERENCES `RESOURCE_SERVER_POLICY` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `RESOURCE_SERVER_PERM_TICKET`
--

LOCK TABLES `RESOURCE_SERVER_PERM_TICKET` WRITE;
/*!40000 ALTER TABLE `RESOURCE_SERVER_PERM_TICKET` DISABLE KEYS */;
/*!40000 ALTER TABLE `RESOURCE_SERVER_PERM_TICKET` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `RESOURCE_SERVER_POLICY`
--

DROP TABLE IF EXISTS `RESOURCE_SERVER_POLICY`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `RESOURCE_SERVER_POLICY` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb3 DEFAULT NULL,
  `TYPE` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `DECISION_STRATEGY` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `LOGIC` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `RESOURCE_SERVER_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `OWNER` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_FRSRPT700S9V50BU18WS5HA6` (`NAME`,`RESOURCE_SERVER_ID`),
  KEY `IDX_RES_SERV_POL_RES_SERV` (`RESOURCE_SERVER_ID`),
  CONSTRAINT `FK_FRSRPO213XCX4WNKOG82SSRFY` FOREIGN KEY (`RESOURCE_SERVER_ID`) REFERENCES `RESOURCE_SERVER` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `RESOURCE_SERVER_POLICY`
--

LOCK TABLES `RESOURCE_SERVER_POLICY` WRITE;
/*!40000 ALTER TABLE `RESOURCE_SERVER_POLICY` DISABLE KEYS */;
/*!40000 ALTER TABLE `RESOURCE_SERVER_POLICY` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `RESOURCE_SERVER_RESOURCE`
--

DROP TABLE IF EXISTS `RESOURCE_SERVER_RESOURCE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `RESOURCE_SERVER_RESOURCE` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `TYPE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ICON_URI` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `OWNER` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `RESOURCE_SERVER_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `OWNER_MANAGED_ACCESS` bit(1) NOT NULL DEFAULT b'0',
  `DISPLAY_NAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_FRSR6T700S9V50BU18WS5HA6` (`NAME`,`OWNER`,`RESOURCE_SERVER_ID`),
  KEY `IDX_RES_SRV_RES_RES_SRV` (`RESOURCE_SERVER_ID`),
  CONSTRAINT `FK_FRSRHO213XCX4WNKOG82SSRFY` FOREIGN KEY (`RESOURCE_SERVER_ID`) REFERENCES `RESOURCE_SERVER` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `RESOURCE_SERVER_RESOURCE`
--

LOCK TABLES `RESOURCE_SERVER_RESOURCE` WRITE;
/*!40000 ALTER TABLE `RESOURCE_SERVER_RESOURCE` DISABLE KEYS */;
/*!40000 ALTER TABLE `RESOURCE_SERVER_RESOURCE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `RESOURCE_SERVER_SCOPE`
--

DROP TABLE IF EXISTS `RESOURCE_SERVER_SCOPE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `RESOURCE_SERVER_SCOPE` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ICON_URI` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `RESOURCE_SERVER_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `DISPLAY_NAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_FRSRST700S9V50BU18WS5HA6` (`NAME`,`RESOURCE_SERVER_ID`),
  KEY `IDX_RES_SRV_SCOPE_RES_SRV` (`RESOURCE_SERVER_ID`),
  CONSTRAINT `FK_FRSRSO213XCX4WNKOG82SSRFY` FOREIGN KEY (`RESOURCE_SERVER_ID`) REFERENCES `RESOURCE_SERVER` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `RESOURCE_SERVER_SCOPE`
--

LOCK TABLES `RESOURCE_SERVER_SCOPE` WRITE;
/*!40000 ALTER TABLE `RESOURCE_SERVER_SCOPE` DISABLE KEYS */;
/*!40000 ALTER TABLE `RESOURCE_SERVER_SCOPE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `RESOURCE_URIS`
--

DROP TABLE IF EXISTS `RESOURCE_URIS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `RESOURCE_URIS` (
  `RESOURCE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`RESOURCE_ID`,`VALUE`),
  CONSTRAINT `FK_RESOURCE_SERVER_URIS` FOREIGN KEY (`RESOURCE_ID`) REFERENCES `RESOURCE_SERVER_RESOURCE` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `RESOURCE_URIS`
--

LOCK TABLES `RESOURCE_URIS` WRITE;
/*!40000 ALTER TABLE `RESOURCE_URIS` DISABLE KEYS */;
/*!40000 ALTER TABLE `RESOURCE_URIS` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ROLE_ATTRIBUTE`
--

DROP TABLE IF EXISTS `ROLE_ATTRIBUTE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ROLE_ATTRIBUTE` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ROLE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(255) CHARACTER SET utf8mb3 DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_ROLE_ATTRIBUTE` (`ROLE_ID`),
  CONSTRAINT `FK_ROLE_ATTRIBUTE_ID` FOREIGN KEY (`ROLE_ID`) REFERENCES `KEYCLOAK_ROLE` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ROLE_ATTRIBUTE`
--

LOCK TABLES `ROLE_ATTRIBUTE` WRITE;
/*!40000 ALTER TABLE `ROLE_ATTRIBUTE` DISABLE KEYS */;
/*!40000 ALTER TABLE `ROLE_ATTRIBUTE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `SCOPE_MAPPING`
--

DROP TABLE IF EXISTS `SCOPE_MAPPING`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `SCOPE_MAPPING` (
  `CLIENT_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ROLE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`CLIENT_ID`,`ROLE_ID`),
  KEY `IDX_SCOPE_MAPPING_ROLE` (`ROLE_ID`),
  CONSTRAINT `FK_OUSE064PLMLR732LXJCN1Q5F1` FOREIGN KEY (`CLIENT_ID`) REFERENCES `CLIENT` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `SCOPE_MAPPING`
--

LOCK TABLES `SCOPE_MAPPING` WRITE;
/*!40000 ALTER TABLE `SCOPE_MAPPING` DISABLE KEYS */;
INSERT INTO `SCOPE_MAPPING` VALUES ('54905dd0-4ade-494e-9c35-ab2d445a99f5','1aa723ff-209c-4637-b3c8-8159d72e9b09');
/*!40000 ALTER TABLE `SCOPE_MAPPING` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `SCOPE_POLICY`
--

DROP TABLE IF EXISTS `SCOPE_POLICY`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `SCOPE_POLICY` (
  `SCOPE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `POLICY_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`SCOPE_ID`,`POLICY_ID`),
  KEY `IDX_SCOPE_POLICY_POLICY` (`POLICY_ID`),
  CONSTRAINT `FK_FRSRASP13XCX4WNKOG82SSRFY` FOREIGN KEY (`POLICY_ID`) REFERENCES `RESOURCE_SERVER_POLICY` (`ID`),
  CONSTRAINT `FK_FRSRPASS3XCX4WNKOG82SSRFY` FOREIGN KEY (`SCOPE_ID`) REFERENCES `RESOURCE_SERVER_SCOPE` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `SCOPE_POLICY`
--

LOCK TABLES `SCOPE_POLICY` WRITE;
/*!40000 ALTER TABLE `SCOPE_POLICY` DISABLE KEYS */;
/*!40000 ALTER TABLE `SCOPE_POLICY` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USERNAME_LOGIN_FAILURE`
--

DROP TABLE IF EXISTS `USERNAME_LOGIN_FAILURE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USERNAME_LOGIN_FAILURE` (
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `USERNAME` varchar(255) CHARACTER SET utf8mb3 NOT NULL,
  `FAILED_LOGIN_NOT_BEFORE` int(11) DEFAULT NULL,
  `LAST_FAILURE` bigint(20) DEFAULT NULL,
  `LAST_IP_FAILURE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NUM_FAILURES` int(11) DEFAULT NULL,
  PRIMARY KEY (`REALM_ID`,`USERNAME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USERNAME_LOGIN_FAILURE`
--

LOCK TABLES `USERNAME_LOGIN_FAILURE` WRITE;
/*!40000 ALTER TABLE `USERNAME_LOGIN_FAILURE` DISABLE KEYS */;
/*!40000 ALTER TABLE `USERNAME_LOGIN_FAILURE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_ATTRIBUTE`
--

DROP TABLE IF EXISTS `USER_ATTRIBUTE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_ATTRIBUTE` (
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(255) CHARACTER SET utf8mb3 DEFAULT NULL,
  `USER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'sybase-needs-something-here',
  PRIMARY KEY (`ID`),
  KEY `IDX_USER_ATTRIBUTE` (`USER_ID`),
  KEY `IDX_USER_ATTRIBUTE_NAME` (`NAME`,`VALUE`),
  CONSTRAINT `FK_5HRM2VLF9QL5FU043KQEPOVBR` FOREIGN KEY (`USER_ID`) REFERENCES `USER_ENTITY` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_ATTRIBUTE`
--

LOCK TABLES `USER_ATTRIBUTE` WRITE;
/*!40000 ALTER TABLE `USER_ATTRIBUTE` DISABLE KEYS */;
/*!40000 ALTER TABLE `USER_ATTRIBUTE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_CONSENT`
--

DROP TABLE IF EXISTS `USER_CONSENT`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_CONSENT` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `CLIENT_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `USER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `CREATED_DATE` bigint(20) DEFAULT NULL,
  `LAST_UPDATED_DATE` bigint(20) DEFAULT NULL,
  `CLIENT_STORAGE_PROVIDER` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `EXTERNAL_CLIENT_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_JKUWUVD56ONTGSUHOGM8UEWRT` (`CLIENT_ID`,`CLIENT_STORAGE_PROVIDER`,`EXTERNAL_CLIENT_ID`,`USER_ID`),
  KEY `IDX_USER_CONSENT` (`USER_ID`),
  CONSTRAINT `FK_GRNTCSNT_USER` FOREIGN KEY (`USER_ID`) REFERENCES `USER_ENTITY` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_CONSENT`
--

LOCK TABLES `USER_CONSENT` WRITE;
/*!40000 ALTER TABLE `USER_CONSENT` DISABLE KEYS */;
/*!40000 ALTER TABLE `USER_CONSENT` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_CONSENT_CLIENT_SCOPE`
--

DROP TABLE IF EXISTS `USER_CONSENT_CLIENT_SCOPE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_CONSENT_CLIENT_SCOPE` (
  `USER_CONSENT_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `SCOPE_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`USER_CONSENT_ID`,`SCOPE_ID`),
  KEY `IDX_USCONSENT_CLSCOPE` (`USER_CONSENT_ID`),
  CONSTRAINT `FK_GRNTCSNT_CLSC_USC` FOREIGN KEY (`USER_CONSENT_ID`) REFERENCES `USER_CONSENT` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_CONSENT_CLIENT_SCOPE`
--

LOCK TABLES `USER_CONSENT_CLIENT_SCOPE` WRITE;
/*!40000 ALTER TABLE `USER_CONSENT_CLIENT_SCOPE` DISABLE KEYS */;
/*!40000 ALTER TABLE `USER_CONSENT_CLIENT_SCOPE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_ENTITY`
--

DROP TABLE IF EXISTS `USER_ENTITY`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_ENTITY` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `EMAIL` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `EMAIL_CONSTRAINT` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `EMAIL_VERIFIED` bit(1) NOT NULL DEFAULT b'0',
  `ENABLED` bit(1) NOT NULL DEFAULT b'0',
  `FEDERATION_LINK` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `FIRST_NAME` varchar(255) CHARACTER SET utf8mb3 DEFAULT NULL,
  `LAST_NAME` varchar(255) CHARACTER SET utf8mb3 DEFAULT NULL,
  `REALM_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `USERNAME` varchar(255) CHARACTER SET utf8mb3 DEFAULT NULL,
  `CREATED_TIMESTAMP` bigint(20) DEFAULT NULL,
  `SERVICE_ACCOUNT_CLIENT_LINK` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NOT_BEFORE` int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_DYKN684SL8UP1CRFEI6ECKHD7` (`REALM_ID`,`EMAIL_CONSTRAINT`),
  UNIQUE KEY `UK_RU8TT6T700S9V50BU18WS5HA6` (`REALM_ID`,`USERNAME`),
  KEY `IDX_USER_EMAIL` (`EMAIL`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_ENTITY`
--

LOCK TABLES `USER_ENTITY` WRITE;
/*!40000 ALTER TABLE `USER_ENTITY` DISABLE KEYS */;
INSERT INTO `USER_ENTITY` VALUES ('563bb06b-d712-48c1-9381-cd6473e18590',NULL,'87ff588a-ad31-4774-b5e2-988648080774','\0','',NULL,NULL,NULL,'master','admin',1635795167832,NULL,0),('744b396e-3cf9-4e9f-9493-61b7b188fb10','jane.doe@example.com','jane.doe@example.com','','',NULL,'Jane','Doe','master','user',1635845837109,NULL,0);
/*!40000 ALTER TABLE `USER_ENTITY` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_FEDERATION_CONFIG`
--

DROP TABLE IF EXISTS `USER_FEDERATION_CONFIG`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_FEDERATION_CONFIG` (
  `USER_FEDERATION_PROVIDER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`USER_FEDERATION_PROVIDER_ID`,`NAME`),
  CONSTRAINT `FK_T13HPU1J94R2EBPEKR39X5EU5` FOREIGN KEY (`USER_FEDERATION_PROVIDER_ID`) REFERENCES `USER_FEDERATION_PROVIDER` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_FEDERATION_CONFIG`
--

LOCK TABLES `USER_FEDERATION_CONFIG` WRITE;
/*!40000 ALTER TABLE `USER_FEDERATION_CONFIG` DISABLE KEYS */;
/*!40000 ALTER TABLE `USER_FEDERATION_CONFIG` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_FEDERATION_MAPPER`
--

DROP TABLE IF EXISTS `USER_FEDERATION_MAPPER`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_FEDERATION_MAPPER` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `FEDERATION_PROVIDER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `FEDERATION_MAPPER_TYPE` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_USR_FED_MAP_FED_PRV` (`FEDERATION_PROVIDER_ID`),
  KEY `IDX_USR_FED_MAP_REALM` (`REALM_ID`),
  CONSTRAINT `FK_FEDMAPPERPM_FEDPRV` FOREIGN KEY (`FEDERATION_PROVIDER_ID`) REFERENCES `USER_FEDERATION_PROVIDER` (`ID`),
  CONSTRAINT `FK_FEDMAPPERPM_REALM` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_FEDERATION_MAPPER`
--

LOCK TABLES `USER_FEDERATION_MAPPER` WRITE;
/*!40000 ALTER TABLE `USER_FEDERATION_MAPPER` DISABLE KEYS */;
/*!40000 ALTER TABLE `USER_FEDERATION_MAPPER` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_FEDERATION_MAPPER_CONFIG`
--

DROP TABLE IF EXISTS `USER_FEDERATION_MAPPER_CONFIG`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_FEDERATION_MAPPER_CONFIG` (
  `USER_FEDERATION_MAPPER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`USER_FEDERATION_MAPPER_ID`,`NAME`),
  CONSTRAINT `FK_FEDMAPPER_CFG` FOREIGN KEY (`USER_FEDERATION_MAPPER_ID`) REFERENCES `USER_FEDERATION_MAPPER` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_FEDERATION_MAPPER_CONFIG`
--

LOCK TABLES `USER_FEDERATION_MAPPER_CONFIG` WRITE;
/*!40000 ALTER TABLE `USER_FEDERATION_MAPPER_CONFIG` DISABLE KEYS */;
/*!40000 ALTER TABLE `USER_FEDERATION_MAPPER_CONFIG` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_FEDERATION_PROVIDER`
--

DROP TABLE IF EXISTS `USER_FEDERATION_PROVIDER`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_FEDERATION_PROVIDER` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `CHANGED_SYNC_PERIOD` int(11) DEFAULT NULL,
  `DISPLAY_NAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `FULL_SYNC_PERIOD` int(11) DEFAULT NULL,
  `LAST_SYNC` int(11) DEFAULT NULL,
  `PRIORITY` int(11) DEFAULT NULL,
  `PROVIDER_NAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REALM_ID` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_USR_FED_PRV_REALM` (`REALM_ID`),
  CONSTRAINT `FK_1FJ32F6PTOLW2QY60CD8N01E8` FOREIGN KEY (`REALM_ID`) REFERENCES `REALM` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_FEDERATION_PROVIDER`
--

LOCK TABLES `USER_FEDERATION_PROVIDER` WRITE;
/*!40000 ALTER TABLE `USER_FEDERATION_PROVIDER` DISABLE KEYS */;
/*!40000 ALTER TABLE `USER_FEDERATION_PROVIDER` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_GROUP_MEMBERSHIP`
--

DROP TABLE IF EXISTS `USER_GROUP_MEMBERSHIP`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_GROUP_MEMBERSHIP` (
  `GROUP_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `USER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`GROUP_ID`,`USER_ID`),
  KEY `IDX_USER_GROUP_MAPPING` (`USER_ID`),
  CONSTRAINT `FK_USER_GROUP_USER` FOREIGN KEY (`USER_ID`) REFERENCES `USER_ENTITY` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_GROUP_MEMBERSHIP`
--

LOCK TABLES `USER_GROUP_MEMBERSHIP` WRITE;
/*!40000 ALTER TABLE `USER_GROUP_MEMBERSHIP` DISABLE KEYS */;
/*!40000 ALTER TABLE `USER_GROUP_MEMBERSHIP` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_REQUIRED_ACTION`
--

DROP TABLE IF EXISTS `USER_REQUIRED_ACTION`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_REQUIRED_ACTION` (
  `USER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `REQUIRED_ACTION` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT ' ',
  PRIMARY KEY (`REQUIRED_ACTION`,`USER_ID`),
  KEY `IDX_USER_REQACTIONS` (`USER_ID`),
  CONSTRAINT `FK_6QJ3W1JW9CVAFHE19BWSIUVMD` FOREIGN KEY (`USER_ID`) REFERENCES `USER_ENTITY` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_REQUIRED_ACTION`
--

LOCK TABLES `USER_REQUIRED_ACTION` WRITE;
/*!40000 ALTER TABLE `USER_REQUIRED_ACTION` DISABLE KEYS */;
/*!40000 ALTER TABLE `USER_REQUIRED_ACTION` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_ROLE_MAPPING`
--

DROP TABLE IF EXISTS `USER_ROLE_MAPPING`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_ROLE_MAPPING` (
  `ROLE_ID` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `USER_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`ROLE_ID`,`USER_ID`),
  KEY `IDX_USER_ROLE_MAPPING` (`USER_ID`),
  CONSTRAINT `FK_C4FQV34P1MBYLLOXANG7B1Q3L` FOREIGN KEY (`USER_ID`) REFERENCES `USER_ENTITY` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_ROLE_MAPPING`
--

LOCK TABLES `USER_ROLE_MAPPING` WRITE;
/*!40000 ALTER TABLE `USER_ROLE_MAPPING` DISABLE KEYS */;
INSERT INTO `USER_ROLE_MAPPING` VALUES ('4d227aaf-b4fa-4a86-9535-30210f612f2e','563bb06b-d712-48c1-9381-cd6473e18590'),('4d227aaf-b4fa-4a86-9535-30210f612f2e','744b396e-3cf9-4e9f-9493-61b7b188fb10'),('5827ab16-b5bc-4738-b05e-89406e065439','563bb06b-d712-48c1-9381-cd6473e18590');
/*!40000 ALTER TABLE `USER_ROLE_MAPPING` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_SESSION`
--

DROP TABLE IF EXISTS `USER_SESSION`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_SESSION` (
  `ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `AUTH_METHOD` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `IP_ADDRESS` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `LAST_SESSION_REFRESH` int(11) DEFAULT NULL,
  `LOGIN_USERNAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REALM_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `REMEMBER_ME` bit(1) NOT NULL DEFAULT b'0',
  `STARTED` int(11) DEFAULT NULL,
  `USER_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `USER_SESSION_STATE` int(11) DEFAULT NULL,
  `BROKER_SESSION_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `BROKER_USER_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_SESSION`
--

LOCK TABLES `USER_SESSION` WRITE;
/*!40000 ALTER TABLE `USER_SESSION` DISABLE KEYS */;
/*!40000 ALTER TABLE `USER_SESSION` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_SESSION_NOTE`
--

DROP TABLE IF EXISTS `USER_SESSION_NOTE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_SESSION_NOTE` (
  `USER_SESSION` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` text COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`USER_SESSION`,`NAME`),
  CONSTRAINT `FK5EDFB00FF51D3472` FOREIGN KEY (`USER_SESSION`) REFERENCES `USER_SESSION` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_SESSION_NOTE`
--

LOCK TABLES `USER_SESSION_NOTE` WRITE;
/*!40000 ALTER TABLE `USER_SESSION_NOTE` DISABLE KEYS */;
/*!40000 ALTER TABLE `USER_SESSION_NOTE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `WEB_ORIGINS`
--

DROP TABLE IF EXISTS `WEB_ORIGINS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `WEB_ORIGINS` (
  `CLIENT_ID` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `VALUE` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`CLIENT_ID`,`VALUE`),
  KEY `IDX_WEB_ORIG_CLIENT` (`CLIENT_ID`),
  CONSTRAINT `FK_LOJPHO213XCX4WNKOG82SSRFY` FOREIGN KEY (`CLIENT_ID`) REFERENCES `CLIENT` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `WEB_ORIGINS`
--

LOCK TABLES `WEB_ORIGINS` WRITE;
/*!40000 ALTER TABLE `WEB_ORIGINS` DISABLE KEYS */;
INSERT INTO `WEB_ORIGINS` VALUES ('5a059221-51fd-434f-84a6-40fa51cda5ce','https://app.localssl.dev/'),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','+');
/*!40000 ALTER TABLE `WEB_ORIGINS` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;