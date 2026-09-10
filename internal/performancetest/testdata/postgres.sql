--
-- PostgreSQL database dump
--

\restrict TwLC0hwfzvYiQveep1GpacaAqZFS6tGTIGOWQbXWzdcxTJV5Mw33X7den7q4GM3

-- Dumped from database version 18.4
-- Dumped by pg_dump version 18.4 (Ubuntu 18.4-0ubuntu0.26.04.1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: pg_database_owner
--

CREATE SCHEMA public;


ALTER SCHEMA public OWNER TO pg_database_owner;

--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: pg_database_owner
--

COMMENT ON SCHEMA public IS 'standard public schema';


--
-- Name: caseinsensitive; Type: COLLATION; Schema: public; Owner: migrate
--

CREATE COLLATION public.caseinsensitive (provider = icu, deterministic = false, locale = 'und');


ALTER COLLATION public.caseinsensitive OWNER TO migrate;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: albums; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.albums (
    id bigint NOT NULL,
    album_uid bytea,
    parent_uid bytea DEFAULT '\x'::bytea,
    album_slug bytea,
    album_path bytea,
    album_type bytea DEFAULT '\x616c62756d'::bytea,
    album_title character varying(160),
    album_location character varying(160),
    album_category character varying(100),
    album_caption character varying(1024),
    album_description character varying(2048),
    album_notes character varying(1024),
    album_filter bytea,
    album_order bytea,
    album_template bytea,
    album_state character varying(100),
    album_country bytea DEFAULT '\x7a7a'::bytea,
    album_year bigint,
    album_month bigint,
    album_day bigint,
    album_favorite boolean,
    album_private boolean,
    thumb bytea DEFAULT '\x'::bytea,
    thumb_src bytea DEFAULT '\x'::bytea,
    created_by bytea,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    published_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE public.albums OWNER TO migrate;

--
-- Name: albums_id_seq; Type: SEQUENCE; Schema: public; Owner: migrate
--

CREATE SEQUENCE public.albums_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.albums_id_seq OWNER TO migrate;

--
-- Name: albums_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: migrate
--

ALTER SEQUENCE public.albums_id_seq OWNED BY public.albums.id;


--
-- Name: albums_users; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.albums_users (
    uid bytea NOT NULL,
    user_uid bytea NOT NULL,
    team_uid bytea,
    perm bigint
);


ALTER TABLE public.albums_users OWNER TO migrate;

--
-- Name: auth_clients; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.auth_clients (
    client_uid bytea NOT NULL,
    node_uuid bytea DEFAULT '\x'::bytea,
    user_uid bytea DEFAULT '\x'::bytea,
    user_name character varying(200),
    app_name character varying(64),
    app_version character varying(64),
    client_name character varying(200),
    display_name character varying(200),
    name_src bytea DEFAULT '\x'::bytea,
    client_role character varying(64) DEFAULT ''::character varying,
    client_type bytea,
    client_url bytea DEFAULT '\x'::bytea,
    callback_url bytea DEFAULT '\x'::bytea,
    auth_provider bytea DEFAULT '\x'::bytea,
    auth_method bytea DEFAULT '\x'::bytea,
    auth_scope character varying(1024) DEFAULT ''::character varying,
    auth_expires bigint,
    auth_tokens bigint,
    auth_enabled boolean,
    refresh_token bytea DEFAULT '\x'::bytea,
    id_token bytea DEFAULT '\x'::bytea,
    data_json bytea,
    last_active bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.auth_clients OWNER TO migrate;

--
-- Name: auth_sessions; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.auth_sessions (
    id bytea NOT NULL,
    user_uid bytea DEFAULT '\x'::bytea,
    user_name character varying(200),
    client_uid bytea DEFAULT '\x'::bytea,
    client_name character varying(200) DEFAULT ''::character varying,
    client_ip character varying(64),
    auth_provider bytea DEFAULT '\x'::bytea,
    auth_method bytea DEFAULT '\x'::bytea,
    auth_issuer bytea DEFAULT '\x'::bytea,
    auth_id bytea DEFAULT '\x'::bytea,
    auth_scope character varying(1024) DEFAULT ''::character varying,
    grant_type bytea DEFAULT '\x'::bytea,
    last_active bigint,
    sess_expires bigint,
    sess_timeout bigint,
    preview_token bytea DEFAULT '\x'::bytea,
    download_token bytea DEFAULT '\x'::bytea,
    access_token bytea DEFAULT '\x'::bytea,
    refresh_token bytea DEFAULT '\x'::bytea,
    id_token bytea DEFAULT '\x'::bytea,
    user_agent character varying(512),
    data_json bytea,
    ref_id bytea DEFAULT '\x'::bytea,
    login_ip character varying(64),
    login_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.auth_sessions OWNER TO migrate;

--
-- Name: auth_users; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.auth_users (
    id bigint NOT NULL,
    user_uuid bytea,
    user_uid bytea,
    auth_provider bytea DEFAULT '\x'::bytea,
    auth_method bytea DEFAULT '\x'::bytea,
    auth_issuer bytea DEFAULT '\x'::bytea,
    auth_id bytea DEFAULT '\x'::bytea,
    user_name character varying(200),
    display_name character varying(200),
    user_email character varying(255),
    backup_email character varying(255),
    user_role character varying(64) DEFAULT ''::character varying,
    user_scope character varying(1024) DEFAULT '*'::character varying,
    user_attr character varying(1024) DEFAULT ''::character varying,
    super_admin boolean,
    can_login boolean,
    login_at timestamp with time zone,
    expires_at timestamp with time zone,
    webdav boolean,
    base_path bytea,
    upload_path bytea,
    can_invite boolean,
    invite_token bytea,
    invited_by character varying(64),
    verify_token bytea,
    verified_at timestamp with time zone,
    consent_at timestamp with time zone,
    born_at timestamp with time zone,
    reset_token bytea,
    preview_token bytea,
    download_token bytea,
    thumb bytea DEFAULT '\x'::bytea,
    thumb_src bytea DEFAULT '\x'::bytea,
    ref_id bytea,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE public.auth_users OWNER TO migrate;

--
-- Name: auth_users_details; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.auth_users_details (
    user_uid bytea NOT NULL,
    subj_uid bytea,
    subj_src bytea DEFAULT '\x'::bytea,
    place_id bytea DEFAULT '\x7a7a'::bytea,
    place_src bytea,
    cell_id bytea DEFAULT '\x7a7a'::bytea,
    birth_year bigint DEFAULT '-1'::integer,
    birth_month bigint DEFAULT '-1'::integer,
    birth_day bigint DEFAULT '-1'::integer,
    name_title character varying(32),
    given_name character varying(64),
    middle_name character varying(64),
    family_name character varying(64),
    name_suffix character varying(32),
    nick_name character varying(64),
    name_src bytea,
    user_gender character varying(16),
    user_about character varying(512),
    user_bio character varying(2048),
    user_location character varying(512),
    user_country bytea DEFAULT '\x7a7a'::bytea,
    user_phone character varying(32),
    site_url bytea,
    profile_url bytea,
    feed_url bytea,
    avatar_url bytea,
    org_title character varying(64),
    org_name character varying(128),
    org_email character varying(255),
    org_phone character varying(32),
    org_url bytea,
    id_url bytea,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.auth_users_details OWNER TO migrate;

--
-- Name: auth_users_id_seq; Type: SEQUENCE; Schema: public; Owner: migrate
--

CREATE SEQUENCE public.auth_users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.auth_users_id_seq OWNER TO migrate;

--
-- Name: auth_users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: migrate
--

ALTER SEQUENCE public.auth_users_id_seq OWNED BY public.auth_users.id;


--
-- Name: auth_users_settings; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.auth_users_settings (
    user_uid bytea NOT NULL,
    ui_theme bytea,
    ui_start_page character varying(64) DEFAULT 'default'::character varying,
    ui_language bytea,
    ui_time_zone bytea,
    maps_style bytea,
    maps_animate bigint DEFAULT 0,
    index_path bytea,
    index_rescan bigint DEFAULT 0,
    import_path bytea,
    import_move bigint DEFAULT 0,
    download_originals bigint DEFAULT 0,
    download_media_raw bigint DEFAULT 0,
    download_media_sidecar bigint DEFAULT 0,
    search_list_view bigint DEFAULT 0,
    search_show_titles bigint DEFAULT 0,
    search_show_captions bigint DEFAULT 0,
    upload_path bytea,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.auth_users_settings OWNER TO migrate;

--
-- Name: auth_users_shares; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.auth_users_shares (
    user_uid bytea NOT NULL,
    share_uid bytea NOT NULL,
    link_uid bytea,
    expires_at timestamp with time zone,
    comment character varying(512),
    perm bigint,
    ref_id bytea,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.auth_users_shares OWNER TO migrate;

--
-- Name: cameras; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.cameras (
    id bigint NOT NULL,
    camera_slug bytea,
    camera_name character varying(160),
    camera_make character varying(160),
    camera_model character varying(160),
    camera_type character varying(100),
    camera_description character varying(2048),
    camera_notes character varying(1024),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE public.cameras OWNER TO migrate;

--
-- Name: cameras_id_seq; Type: SEQUENCE; Schema: public; Owner: migrate
--

CREATE SEQUENCE public.cameras_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.cameras_id_seq OWNER TO migrate;

--
-- Name: cameras_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: migrate
--

ALTER SEQUENCE public.cameras_id_seq OWNED BY public.cameras.id;


--
-- Name: categories; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.categories (
    label_id bigint NOT NULL,
    category_id bigint NOT NULL
);


ALTER TABLE public.categories OWNER TO migrate;

--
-- Name: cells; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.cells (
    id bytea NOT NULL,
    cell_name character varying(200),
    cell_street character varying(100),
    cell_postcode character varying(50),
    cell_category character varying(50),
    place_id bytea DEFAULT '\x7a7a'::bytea,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.cells OWNER TO migrate;

--
-- Name: countries; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.countries (
    id bytea NOT NULL,
    country_slug bytea,
    country_name character varying(160),
    country_description character varying(2048),
    country_notes character varying(1024),
    country_photo_id bigint
);


ALTER TABLE public.countries OWNER TO migrate;

--
-- Name: details; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.details (
    photo_id bigint NOT NULL,
    keywords character varying(2048),
    keywords_src bytea,
    notes character varying(2048),
    notes_src bytea,
    subject character varying(1024),
    subject_src bytea,
    artist character varying(1024),
    artist_src bytea,
    copyright character varying(1024),
    copyright_src bytea,
    license character varying(1024),
    license_src bytea,
    software character varying(1024),
    software_src bytea,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.details OWNER TO migrate;

--
-- Name: duplicates; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.duplicates (
    file_name bytea NOT NULL,
    file_root bytea DEFAULT '\x2f'::bytea NOT NULL,
    file_hash bytea DEFAULT '\x'::bytea,
    file_size bigint,
    mod_time bigint
);


ALTER TABLE public.duplicates OWNER TO migrate;

--
-- Name: errors; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.errors (
    id bigint NOT NULL,
    error_time timestamp with time zone,
    error_level bytea,
    error_message bytea
);


ALTER TABLE public.errors OWNER TO migrate;

--
-- Name: errors_id_seq; Type: SEQUENCE; Schema: public; Owner: migrate
--

CREATE SEQUENCE public.errors_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.errors_id_seq OWNER TO migrate;

--
-- Name: errors_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: migrate
--

ALTER SEQUENCE public.errors_id_seq OWNED BY public.errors.id;


--
-- Name: faces; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.faces (
    id bytea NOT NULL,
    face_src bytea,
    face_kind bigint,
    face_hidden boolean,
    subj_uid bytea DEFAULT '\x'::bytea,
    samples bigint,
    sample_radius numeric,
    collisions bigint,
    collision_radius numeric,
    merge_retry smallint DEFAULT 0,
    merge_notes character varying(255) DEFAULT ''::character varying,
    embedding_json bytea,
    matched_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.faces OWNER TO migrate;

--
-- Name: files; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.files (
    id bigint NOT NULL,
    photo_id bigint,
    photo_uid bytea,
    photo_taken_at timestamp with time zone,
    time_index bytea,
    media_id bytea,
    media_utc bigint,
    instance_id bytea,
    file_uid bytea,
    file_name bytea,
    file_root bytea DEFAULT '\x2f'::bytea,
    original_name bytea,
    file_hash bytea,
    file_size bigint,
    file_codec bytea,
    file_type bytea,
    media_type bytea,
    file_mime bytea,
    file_primary boolean,
    file_sidecar boolean,
    file_missing boolean,
    file_portrait boolean,
    file_video boolean,
    file_duration bigint,
    file_fps numeric,
    file_frames bigint,
    file_pages bigint DEFAULT 0,
    file_width bigint,
    file_height bigint,
    file_orientation bigint,
    file_orientation_src bytea DEFAULT '\x'::bytea,
    file_projection bytea,
    file_aspect_ratio numeric,
    file_hdr boolean,
    file_watermark boolean,
    file_color_profile bytea,
    file_main_color bytea,
    file_colors bytea,
    file_luminance bytea,
    file_diff bigint,
    file_chroma smallint,
    file_software character varying(64),
    file_error bytea,
    mod_time bigint,
    created_at timestamp with time zone,
    created_in bigint,
    updated_at timestamp with time zone,
    updated_in bigint,
    published_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE public.files OWNER TO migrate;

--
-- Name: files_id_seq; Type: SEQUENCE; Schema: public; Owner: migrate
--

CREATE SEQUENCE public.files_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.files_id_seq OWNER TO migrate;

--
-- Name: files_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: migrate
--

ALTER SEQUENCE public.files_id_seq OWNED BY public.files.id;


--
-- Name: files_share; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.files_share (
    file_id bigint NOT NULL,
    service_id bigint NOT NULL,
    remote_name bytea NOT NULL,
    status bytea,
    error bytea,
    errors bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.files_share OWNER TO migrate;

--
-- Name: files_sync; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.files_sync (
    remote_name bytea NOT NULL,
    service_id bigint NOT NULL,
    file_id bigint,
    remote_date timestamp with time zone,
    remote_size bigint,
    status bytea,
    error bytea,
    errors bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.files_sync OWNER TO migrate;

--
-- Name: folders; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.folders (
    path bytea,
    root bytea DEFAULT '\x'::bytea,
    folder_uid bytea NOT NULL,
    folder_type bytea,
    folder_title character varying(200),
    folder_category character varying(100),
    folder_description character varying(2048),
    folder_order bytea,
    folder_country bytea DEFAULT '\x7a7a'::bytea,
    folder_year bigint,
    folder_month bigint,
    folder_day bigint,
    folder_favorite boolean,
    folder_private boolean,
    folder_ignore boolean,
    folder_watch boolean,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    modified_at timestamp with time zone,
    published_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE public.folders OWNER TO migrate;

--
-- Name: keywords; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.keywords (
    id bigint NOT NULL,
    keyword character varying(64),
    skip boolean
);


ALTER TABLE public.keywords OWNER TO migrate;

--
-- Name: keywords_id_seq; Type: SEQUENCE; Schema: public; Owner: migrate
--

CREATE SEQUENCE public.keywords_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.keywords_id_seq OWNER TO migrate;

--
-- Name: keywords_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: migrate
--

ALTER SEQUENCE public.keywords_id_seq OWNED BY public.keywords.id;


--
-- Name: labels; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.labels (
    id bigint NOT NULL,
    label_uid bytea,
    label_slug bytea,
    custom_slug bytea,
    label_name character varying(160),
    label_favorite boolean DEFAULT false,
    label_priority bigint DEFAULT 0,
    label_nsfw boolean DEFAULT false,
    label_description character varying(2048),
    label_notes character varying(1024),
    photo_count bigint DEFAULT 1,
    thumb bytea DEFAULT '\x'::bytea,
    thumb_src bytea DEFAULT '\x'::bytea,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    published_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE public.labels OWNER TO migrate;

--
-- Name: labels_id_seq; Type: SEQUENCE; Schema: public; Owner: migrate
--

CREATE SEQUENCE public.labels_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.labels_id_seq OWNER TO migrate;

--
-- Name: labels_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: migrate
--

ALTER SEQUENCE public.labels_id_seq OWNED BY public.labels.id;


--
-- Name: lenses; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.lenses (
    id bigint NOT NULL,
    lens_slug bytea,
    lens_name character varying(160),
    lens_make character varying(160),
    lens_model character varying(160),
    lens_type character varying(100),
    lens_description character varying(2048),
    lens_notes character varying(1024),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE public.lenses OWNER TO migrate;

--
-- Name: lenses_id_seq; Type: SEQUENCE; Schema: public; Owner: migrate
--

CREATE SEQUENCE public.lenses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.lenses_id_seq OWNER TO migrate;

--
-- Name: lenses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: migrate
--

ALTER SEQUENCE public.lenses_id_seq OWNED BY public.lenses.id;


--
-- Name: links; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.links (
    link_uid bytea NOT NULL,
    share_uid bytea,
    share_slug bytea,
    link_token bytea,
    link_expires bigint,
    link_views bigint,
    max_views bigint,
    has_password boolean,
    comment character varying(512),
    perm bigint,
    ref_id bytea,
    created_by bytea,
    created_at timestamp with time zone,
    modified_at timestamp with time zone
);


ALTER TABLE public.links OWNER TO migrate;

--
-- Name: markers; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.markers (
    marker_uid bytea NOT NULL,
    file_uid bytea DEFAULT '\x'::bytea,
    marker_type bytea DEFAULT '\x'::bytea,
    marker_src bytea DEFAULT '\x'::bytea,
    marker_name character varying(160),
    marker_review boolean,
    marker_invalid boolean,
    subj_uid bytea,
    subj_src bytea DEFAULT '\x'::bytea,
    face_id bytea,
    face_dist numeric DEFAULT '-1'::integer,
    embeddings_json bytea,
    landmarks_json bytea,
    x numeric,
    y numeric,
    w numeric,
    h numeric,
    q bigint,
    size bigint DEFAULT '-1'::integer,
    score smallint,
    thumb bytea DEFAULT '\x'::bytea,
    matched_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.markers OWNER TO migrate;

--
-- Name: migrations; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.migrations (
    id character varying(16) NOT NULL,
    dialect character varying(16),
    stage character varying(16),
    error character varying(255),
    source character varying(16),
    started_at timestamp with time zone,
    finished_at timestamp with time zone
);


ALTER TABLE public.migrations OWNER TO migrate;

--
-- Name: passcodes; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.passcodes (
    uid bytea NOT NULL,
    key_type character varying(64) DEFAULT ''::character varying NOT NULL,
    key_url character varying(2048) DEFAULT ''::character varying,
    recovery_code character varying(255) DEFAULT ''::character varying,
    verified_at timestamp with time zone,
    activated_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.passcodes OWNER TO migrate;

--
-- Name: passwords; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.passwords (
    uid bytea NOT NULL,
    hash bytea,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.passwords OWNER TO migrate;

--
-- Name: photos; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.photos (
    id bigint NOT NULL,
    uuid bytea,
    taken_at timestamp with time zone,
    taken_at_local timestamp with time zone,
    taken_src bytea,
    photo_uid bytea,
    photo_type bytea DEFAULT '\x696d616765'::bytea,
    type_src bytea,
    photo_title character varying(200),
    title_src bytea,
    photo_caption character varying(4096),
    caption_src bytea,
    photo_path bytea,
    photo_name bytea,
    original_name bytea,
    photo_stack smallint,
    photo_favorite boolean,
    photo_private boolean,
    photo_scan boolean,
    photo_panorama boolean,
    time_zone bytea DEFAULT '\x4c6f63616c'::bytea,
    place_id bytea DEFAULT '\x7a7a'::bytea,
    place_src bytea,
    cell_id bytea DEFAULT '\x7a7a'::bytea,
    cell_accuracy bigint,
    photo_altitude bigint,
    photo_lat numeric,
    photo_lng numeric,
    photo_country bytea DEFAULT '\x7a7a'::bytea,
    photo_year bigint,
    photo_month bigint,
    photo_day bigint,
    photo_iso bigint,
    photo_exposure bytea,
    photo_f_number numeric,
    photo_focal_length bigint,
    photo_quality smallint,
    photo_faces bigint,
    photo_resolution smallint,
    photo_duration bigint,
    photo_color smallint,
    camera_id bigint DEFAULT 1,
    camera_serial bytea,
    camera_src bytea,
    lens_id bigint DEFAULT 1,
    created_by bytea,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    edited_at timestamp with time zone,
    published_at timestamp with time zone,
    indexed_at timestamp with time zone,
    checked_at timestamp with time zone,
    estimated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE public.photos OWNER TO migrate;

--
-- Name: photos_albums; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.photos_albums (
    photo_uid bytea NOT NULL,
    album_uid bytea NOT NULL,
    "order" bigint,
    hidden boolean,
    missing boolean,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.photos_albums OWNER TO migrate;

--
-- Name: photos_id_seq; Type: SEQUENCE; Schema: public; Owner: migrate
--

CREATE SEQUENCE public.photos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.photos_id_seq OWNER TO migrate;

--
-- Name: photos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: migrate
--

ALTER SEQUENCE public.photos_id_seq OWNED BY public.photos.id;


--
-- Name: photos_keywords; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.photos_keywords (
    photo_id bigint NOT NULL,
    keyword_id bigint NOT NULL
);


ALTER TABLE public.photos_keywords OWNER TO migrate;

--
-- Name: photos_labels; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.photos_labels (
    photo_id bigint NOT NULL,
    label_id bigint NOT NULL,
    label_src bytea,
    uncertainty smallint,
    topicality smallint DEFAULT 0,
    nsfw smallint DEFAULT 0
);


ALTER TABLE public.photos_labels OWNER TO migrate;

--
-- Name: photos_users; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.photos_users (
    uid bytea NOT NULL,
    user_uid bytea NOT NULL,
    team_uid bytea,
    perm bigint
);


ALTER TABLE public.photos_users OWNER TO migrate;

--
-- Name: places; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.places (
    id bytea NOT NULL,
    place_label character varying(400),
    place_district character varying(100),
    place_city character varying(100),
    place_state character varying(100),
    place_country bytea,
    place_keywords character varying(300),
    place_favorite boolean,
    photo_count bigint DEFAULT 1,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.places OWNER TO migrate;

--
-- Name: reactions; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.reactions (
    uid bytea NOT NULL,
    user_uid bytea NOT NULL,
    reaction bytea NOT NULL,
    reacted bigint,
    reacted_at timestamp with time zone
);


ALTER TABLE public.reactions OWNER TO migrate;

--
-- Name: services; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.services (
    id bigint NOT NULL,
    acc_name character varying(160),
    acc_owner character varying(160),
    acc_url character varying(255),
    acc_type bytea,
    acc_key bytea,
    acc_user bytea,
    acc_pass bytea,
    acc_timeout bytea,
    acc_error bytea,
    acc_errors bigint,
    acc_share boolean,
    acc_sync boolean,
    retry_limit bigint,
    share_path bytea,
    share_size bytea,
    share_expires bigint,
    sync_path bytea,
    sync_status bytea,
    sync_interval bigint,
    sync_date timestamp with time zone,
    sync_upload boolean,
    sync_download boolean,
    sync_filenames boolean,
    sync_raw boolean,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE public.services OWNER TO migrate;

--
-- Name: services_id_seq; Type: SEQUENCE; Schema: public; Owner: migrate
--

CREATE SEQUENCE public.services_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.services_id_seq OWNER TO migrate;

--
-- Name: services_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: migrate
--

ALTER SEQUENCE public.services_id_seq OWNED BY public.services.id;


--
-- Name: subjects; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.subjects (
    subj_uid bytea NOT NULL,
    subj_type bytea DEFAULT '\x'::bytea,
    subj_src bytea DEFAULT '\x'::bytea,
    subj_slug bytea DEFAULT '\x'::bytea,
    subj_name character varying(160) DEFAULT ''::character varying,
    subj_alias character varying(160) DEFAULT ''::character varying,
    subj_about character varying(512),
    subj_bio character varying(2048),
    subj_notes character varying(1024),
    subj_favorite boolean DEFAULT false,
    subj_hidden boolean DEFAULT false,
    subj_private boolean DEFAULT false,
    subj_excluded boolean DEFAULT false,
    file_count bigint DEFAULT 0,
    photo_count bigint DEFAULT 0,
    thumb bytea DEFAULT '\x'::bytea,
    thumb_src bytea DEFAULT '\x'::bytea,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


ALTER TABLE public.subjects OWNER TO migrate;

--
-- Name: versions; Type: TABLE; Schema: public; Owner: migrate
--

CREATE TABLE public.versions (
    id bigint NOT NULL,
    version character varying(255),
    edition character varying(255),
    error character varying(255),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    migrated_at timestamp with time zone
);


ALTER TABLE public.versions OWNER TO migrate;

--
-- Name: versions_id_seq; Type: SEQUENCE; Schema: public; Owner: migrate
--

CREATE SEQUENCE public.versions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.versions_id_seq OWNER TO migrate;

--
-- Name: versions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: migrate
--

ALTER SEQUENCE public.versions_id_seq OWNED BY public.versions.id;


--
-- Name: albums id; Type: DEFAULT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.albums ALTER COLUMN id SET DEFAULT nextval('public.albums_id_seq'::regclass);


--
-- Name: auth_users id; Type: DEFAULT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.auth_users ALTER COLUMN id SET DEFAULT nextval('public.auth_users_id_seq'::regclass);


--
-- Name: cameras id; Type: DEFAULT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.cameras ALTER COLUMN id SET DEFAULT nextval('public.cameras_id_seq'::regclass);


--
-- Name: errors id; Type: DEFAULT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.errors ALTER COLUMN id SET DEFAULT nextval('public.errors_id_seq'::regclass);


--
-- Name: files id; Type: DEFAULT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.files ALTER COLUMN id SET DEFAULT nextval('public.files_id_seq'::regclass);


--
-- Name: keywords id; Type: DEFAULT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.keywords ALTER COLUMN id SET DEFAULT nextval('public.keywords_id_seq'::regclass);


--
-- Name: labels id; Type: DEFAULT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.labels ALTER COLUMN id SET DEFAULT nextval('public.labels_id_seq'::regclass);


--
-- Name: lenses id; Type: DEFAULT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.lenses ALTER COLUMN id SET DEFAULT nextval('public.lenses_id_seq'::regclass);


--
-- Name: photos id; Type: DEFAULT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos ALTER COLUMN id SET DEFAULT nextval('public.photos_id_seq'::regclass);


--
-- Name: services id; Type: DEFAULT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.services ALTER COLUMN id SET DEFAULT nextval('public.services_id_seq'::regclass);


--
-- Name: versions id; Type: DEFAULT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.versions ALTER COLUMN id SET DEFAULT nextval('public.versions_id_seq'::regclass);


--
-- Name: albums albums_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.albums
    ADD CONSTRAINT albums_pkey PRIMARY KEY (id);


--
-- Name: albums_users albums_users_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.albums_users
    ADD CONSTRAINT albums_users_pkey PRIMARY KEY (uid, user_uid);


--
-- Name: auth_clients auth_clients_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.auth_clients
    ADD CONSTRAINT auth_clients_pkey PRIMARY KEY (client_uid);


--
-- Name: auth_sessions auth_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.auth_sessions
    ADD CONSTRAINT auth_sessions_pkey PRIMARY KEY (id);


--
-- Name: auth_users_details auth_users_details_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.auth_users_details
    ADD CONSTRAINT auth_users_details_pkey PRIMARY KEY (user_uid);


--
-- Name: auth_users auth_users_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.auth_users
    ADD CONSTRAINT auth_users_pkey PRIMARY KEY (id);


--
-- Name: auth_users_settings auth_users_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.auth_users_settings
    ADD CONSTRAINT auth_users_settings_pkey PRIMARY KEY (user_uid);


--
-- Name: auth_users_shares auth_users_shares_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.auth_users_shares
    ADD CONSTRAINT auth_users_shares_pkey PRIMARY KEY (user_uid, share_uid);


--
-- Name: cameras cameras_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.cameras
    ADD CONSTRAINT cameras_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (label_id, category_id);


--
-- Name: cells cells_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.cells
    ADD CONSTRAINT cells_pkey PRIMARY KEY (id);


--
-- Name: countries countries_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.countries
    ADD CONSTRAINT countries_pkey PRIMARY KEY (id);


--
-- Name: details details_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.details
    ADD CONSTRAINT details_pkey PRIMARY KEY (photo_id);


--
-- Name: duplicates duplicates_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.duplicates
    ADD CONSTRAINT duplicates_pkey PRIMARY KEY (file_name, file_root);


--
-- Name: errors errors_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.errors
    ADD CONSTRAINT errors_pkey PRIMARY KEY (id);


--
-- Name: faces faces_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.faces
    ADD CONSTRAINT faces_pkey PRIMARY KEY (id);


--
-- Name: files files_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_pkey PRIMARY KEY (id);


--
-- Name: files_share files_share_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.files_share
    ADD CONSTRAINT files_share_pkey PRIMARY KEY (file_id, service_id, remote_name);


--
-- Name: files_sync files_sync_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.files_sync
    ADD CONSTRAINT files_sync_pkey PRIMARY KEY (remote_name, service_id);


--
-- Name: folders folders_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.folders
    ADD CONSTRAINT folders_pkey PRIMARY KEY (folder_uid);


--
-- Name: keywords keywords_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.keywords
    ADD CONSTRAINT keywords_pkey PRIMARY KEY (id);


--
-- Name: labels labels_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.labels
    ADD CONSTRAINT labels_pkey PRIMARY KEY (id);


--
-- Name: lenses lenses_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.lenses
    ADD CONSTRAINT lenses_pkey PRIMARY KEY (id);


--
-- Name: links links_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.links
    ADD CONSTRAINT links_pkey PRIMARY KEY (link_uid);


--
-- Name: markers markers_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.markers
    ADD CONSTRAINT markers_pkey PRIMARY KEY (marker_uid);


--
-- Name: migrations migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.migrations
    ADD CONSTRAINT migrations_pkey PRIMARY KEY (id);


--
-- Name: passcodes passcodes_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.passcodes
    ADD CONSTRAINT passcodes_pkey PRIMARY KEY (uid, key_type);


--
-- Name: passwords passwords_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.passwords
    ADD CONSTRAINT passwords_pkey PRIMARY KEY (uid);


--
-- Name: photos_albums photos_albums_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos_albums
    ADD CONSTRAINT photos_albums_pkey PRIMARY KEY (photo_uid, album_uid);


--
-- Name: photos_keywords photos_keywords_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos_keywords
    ADD CONSTRAINT photos_keywords_pkey PRIMARY KEY (photo_id, keyword_id);


--
-- Name: photos_labels photos_labels_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos_labels
    ADD CONSTRAINT photos_labels_pkey PRIMARY KEY (photo_id, label_id);


--
-- Name: photos photos_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos
    ADD CONSTRAINT photos_pkey PRIMARY KEY (id);


--
-- Name: photos_users photos_users_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos_users
    ADD CONSTRAINT photos_users_pkey PRIMARY KEY (uid, user_uid);


--
-- Name: places places_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.places
    ADD CONSTRAINT places_pkey PRIMARY KEY (id);


--
-- Name: reactions reactions_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.reactions
    ADD CONSTRAINT reactions_pkey PRIMARY KEY (uid, user_uid, reaction);


--
-- Name: services services_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.services
    ADD CONSTRAINT services_pkey PRIMARY KEY (id);


--
-- Name: subjects subjects_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.subjects
    ADD CONSTRAINT subjects_pkey PRIMARY KEY (subj_uid);


--
-- Name: versions versions_pkey; Type: CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.versions
    ADD CONSTRAINT versions_pkey PRIMARY KEY (id);


--
-- Name: idx_albums_album_category; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_albums_album_category ON public.albums USING btree (album_category);


--
-- Name: idx_albums_album_filter; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_albums_album_filter ON public.albums USING btree (album_filter);


--
-- Name: idx_albums_album_path; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_albums_album_path ON public.albums USING btree (album_path);


--
-- Name: idx_albums_album_slug; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_albums_album_slug ON public.albums USING btree (album_slug);


--
-- Name: idx_albums_album_state; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_albums_album_state ON public.albums USING btree (album_state);


--
-- Name: idx_albums_album_title; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_albums_album_title ON public.albums USING btree (album_title);


--
-- Name: idx_albums_album_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_albums_album_uid ON public.albums USING btree (album_uid);


--
-- Name: idx_albums_country_year_month; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_albums_country_year_month ON public.albums USING btree (album_country, album_year, album_month);


--
-- Name: idx_albums_created_by; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_albums_created_by ON public.albums USING btree (created_by);


--
-- Name: idx_albums_deleted_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_albums_deleted_at ON public.albums USING btree (deleted_at);


--
-- Name: idx_albums_published_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_albums_published_at ON public.albums USING btree (published_at);


--
-- Name: idx_albums_thumb; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_albums_thumb ON public.albums USING btree (thumb);


--
-- Name: idx_albums_users_team_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_albums_users_team_uid ON public.albums_users USING btree (team_uid);


--
-- Name: idx_albums_users_user_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_albums_users_user_uid ON public.albums_users USING btree (user_uid);


--
-- Name: idx_albums_ymd; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_albums_ymd ON public.albums USING btree (album_year, album_month, album_day);


--
-- Name: idx_auth_clients_node_uuid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_clients_node_uuid ON public.auth_clients USING btree (node_uuid);


--
-- Name: idx_auth_clients_user_name; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_clients_user_name ON public.auth_clients USING btree (user_name);


--
-- Name: idx_auth_clients_user_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_clients_user_uid ON public.auth_clients USING btree (user_uid);


--
-- Name: idx_auth_sessions_auth_id; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_sessions_auth_id ON public.auth_sessions USING btree (auth_id);


--
-- Name: idx_auth_sessions_client_ip; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_sessions_client_ip ON public.auth_sessions USING btree (client_ip);


--
-- Name: idx_auth_sessions_client_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_sessions_client_uid ON public.auth_sessions USING btree (client_uid);


--
-- Name: idx_auth_sessions_sess_expires; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_sessions_sess_expires ON public.auth_sessions USING btree (sess_expires);


--
-- Name: idx_auth_sessions_user_name; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_sessions_user_name ON public.auth_sessions USING btree (user_name);


--
-- Name: idx_auth_sessions_user_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_sessions_user_uid ON public.auth_sessions USING btree (user_uid);


--
-- Name: idx_auth_users_auth_id; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_auth_id ON public.auth_users USING btree (auth_id);


--
-- Name: idx_auth_users_born_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_born_at ON public.auth_users USING btree (born_at);


--
-- Name: idx_auth_users_deleted_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_deleted_at ON public.auth_users USING btree (deleted_at);


--
-- Name: idx_auth_users_details_cell_id; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_details_cell_id ON public.auth_users_details USING btree (cell_id);


--
-- Name: idx_auth_users_details_org_email; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_details_org_email ON public.auth_users_details USING btree (org_email);


--
-- Name: idx_auth_users_details_place_id; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_details_place_id ON public.auth_users_details USING btree (place_id);


--
-- Name: idx_auth_users_details_subj_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_details_subj_uid ON public.auth_users_details USING btree (subj_uid);


--
-- Name: idx_auth_users_expires_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_expires_at ON public.auth_users USING btree (expires_at);


--
-- Name: idx_auth_users_invite_token; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_invite_token ON public.auth_users USING btree (invite_token);


--
-- Name: idx_auth_users_shares_expires_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_shares_expires_at ON public.auth_users_shares USING btree (expires_at);


--
-- Name: idx_auth_users_shares_share_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_shares_share_uid ON public.auth_users_shares USING btree (share_uid);


--
-- Name: idx_auth_users_thumb; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_thumb ON public.auth_users USING btree (thumb);


--
-- Name: idx_auth_users_user_email; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_user_email ON public.auth_users USING btree (user_email);


--
-- Name: idx_auth_users_user_name; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_user_name ON public.auth_users USING btree (user_name);


--
-- Name: idx_auth_users_user_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_auth_users_user_uid ON public.auth_users USING btree (user_uid);


--
-- Name: idx_auth_users_uuid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_auth_users_uuid ON public.auth_users USING btree (user_uuid);


--
-- Name: idx_cameras_camera_slug; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_cameras_camera_slug ON public.cameras USING btree (camera_slug);


--
-- Name: idx_cameras_deleted_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_cameras_deleted_at ON public.cameras USING btree (deleted_at);


--
-- Name: idx_countries_country_slug; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_countries_country_slug ON public.countries USING btree (country_slug);


--
-- Name: idx_duplicates_file_hash; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_duplicates_file_hash ON public.duplicates USING btree (file_hash);


--
-- Name: idx_errors_error_time; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_errors_error_time ON public.errors USING btree (error_time);


--
-- Name: idx_faces_subj_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_faces_subj_uid ON public.faces USING btree (subj_uid);


--
-- Name: idx_files_deleted_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_files_deleted_at ON public.files USING btree (deleted_at);


--
-- Name: idx_files_file_error; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_files_file_error ON public.files USING btree (file_error);


--
-- Name: idx_files_file_hash; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_files_file_hash ON public.files USING btree (file_hash);


--
-- Name: idx_files_file_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_files_file_uid ON public.files USING btree (file_uid);


--
-- Name: idx_files_instance_id; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_files_instance_id ON public.files USING btree (instance_id);


--
-- Name: idx_files_media_utc; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_files_media_utc ON public.files USING btree (media_utc);


--
-- Name: idx_files_missing_root; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_files_missing_root ON public.files USING btree (file_missing, file_root);


--
-- Name: idx_files_name_root; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_files_name_root ON public.files USING btree (file_name, file_root);


--
-- Name: idx_files_photo_id; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_files_photo_id ON public.files USING btree (photo_id, file_primary);


--
-- Name: idx_files_photo_taken_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_files_photo_taken_at ON public.files USING btree (photo_taken_at);


--
-- Name: idx_files_photo_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_files_photo_uid ON public.files USING btree (photo_uid);


--
-- Name: idx_files_published_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_files_published_at ON public.files USING btree (published_at);


--
-- Name: idx_files_search_media; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_files_search_media ON public.files USING btree (media_id);


--
-- Name: idx_files_search_timeline; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_files_search_timeline ON public.files USING btree (time_index);


--
-- Name: idx_files_sync_file_id; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_files_sync_file_id ON public.files_sync USING btree (file_id);


--
-- Name: idx_folders_country_year_month; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_folders_country_year_month ON public.folders USING btree (folder_country, folder_year, folder_month);


--
-- Name: idx_folders_deleted_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_folders_deleted_at ON public.folders USING btree (deleted_at);


--
-- Name: idx_folders_folder_category; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_folders_folder_category ON public.folders USING btree (folder_category);


--
-- Name: idx_folders_path_root; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_folders_path_root ON public.folders USING btree (path, root);


--
-- Name: idx_folders_published_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_folders_published_at ON public.folders USING btree (published_at);


--
-- Name: idx_keywords_keyword; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_keywords_keyword ON public.keywords USING btree (keyword);


--
-- Name: idx_labels_custom_slug; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_labels_custom_slug ON public.labels USING btree (custom_slug);


--
-- Name: idx_labels_deleted_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_labels_deleted_at ON public.labels USING btree (deleted_at);


--
-- Name: idx_labels_label_slug; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_labels_label_slug ON public.labels USING btree (label_slug);


--
-- Name: idx_labels_label_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_labels_label_uid ON public.labels USING btree (label_uid);


--
-- Name: idx_labels_published_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_labels_published_at ON public.labels USING btree (published_at);


--
-- Name: idx_labels_thumb; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_labels_thumb ON public.labels USING btree (thumb);


--
-- Name: idx_lenses_deleted_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_lenses_deleted_at ON public.lenses USING btree (deleted_at);


--
-- Name: idx_lenses_lens_slug; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_lenses_lens_slug ON public.lenses USING btree (lens_slug);


--
-- Name: idx_links_created_by; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_links_created_by ON public.links USING btree (created_by);


--
-- Name: idx_links_share_slug; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_links_share_slug ON public.links USING btree (share_slug);


--
-- Name: idx_links_uid_token; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_links_uid_token ON public.links USING btree (share_uid, link_token);


--
-- Name: idx_markers_face_id; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_markers_face_id ON public.markers USING btree (face_id);


--
-- Name: idx_markers_file_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_markers_file_uid ON public.markers USING btree (file_uid);


--
-- Name: idx_markers_matched_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_markers_matched_at ON public.markers USING btree (matched_at);


--
-- Name: idx_markers_subj_uid_src; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_markers_subj_uid_src ON public.markers USING btree (subj_uid, subj_src);


--
-- Name: idx_markers_thumb; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_markers_thumb ON public.markers USING btree (thumb);


--
-- Name: idx_photos_camera_lens; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_camera_lens ON public.photos USING btree (camera_id, lens_id);


--
-- Name: idx_photos_cell_id; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_cell_id ON public.photos USING btree (cell_id);


--
-- Name: idx_photos_checked_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_checked_at ON public.photos USING btree (checked_at);


--
-- Name: idx_photos_country_year_month; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_country_year_month ON public.photos USING btree (photo_country, photo_year, photo_month);


--
-- Name: idx_photos_created_by; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_created_by ON public.photos USING btree (created_by);


--
-- Name: idx_photos_deleted_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_deleted_at ON public.photos USING btree (deleted_at);


--
-- Name: idx_photos_keywords_keyword_id; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_keywords_keyword_id ON public.photos_keywords USING btree (keyword_id);


--
-- Name: idx_photos_labels_label_id; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_labels_label_id ON public.photos_labels USING btree (label_id);


--
-- Name: idx_photos_path_name; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_path_name ON public.photos USING btree (photo_path, photo_name);


--
-- Name: idx_photos_photo_lat; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_photo_lat ON public.photos USING btree (photo_lat);


--
-- Name: idx_photos_photo_lng; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_photo_lng ON public.photos USING btree (photo_lng);


--
-- Name: idx_photos_photo_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_photos_photo_uid ON public.photos USING btree (photo_uid);


--
-- Name: idx_photos_place_id; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_place_id ON public.photos USING btree (place_id);


--
-- Name: idx_photos_published_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_published_at ON public.photos USING btree (published_at);


--
-- Name: idx_photos_taken_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_taken_uid ON public.photos USING btree (taken_at, photo_uid);


--
-- Name: idx_photos_users_team_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_users_team_uid ON public.photos_users USING btree (team_uid);


--
-- Name: idx_photos_users_user_uid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_users_user_uid ON public.photos_users USING btree (user_uid);


--
-- Name: idx_photos_uuid; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_uuid ON public.photos USING btree (uuid);


--
-- Name: idx_photos_ymd; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_photos_ymd ON public.photos USING btree (photo_year, photo_month, photo_day);


--
-- Name: idx_places_place_city; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_places_place_city ON public.places USING btree (place_city);


--
-- Name: idx_places_place_district; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_places_place_district ON public.places USING btree (place_district);


--
-- Name: idx_places_place_state; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_places_place_state ON public.places USING btree (place_state);


--
-- Name: idx_reactions_reacted_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_reactions_reacted_at ON public.reactions USING btree (reacted_at);


--
-- Name: idx_services_deleted_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_services_deleted_at ON public.services USING btree (deleted_at);


--
-- Name: idx_subjects_deleted_at; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_subjects_deleted_at ON public.subjects USING btree (deleted_at);


--
-- Name: idx_subjects_subj_name; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_subjects_subj_name ON public.subjects USING btree (subj_name);


--
-- Name: idx_subjects_subj_slug; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_subjects_subj_slug ON public.subjects USING btree (subj_slug);


--
-- Name: idx_subjects_thumb; Type: INDEX; Schema: public; Owner: migrate
--

CREATE INDEX idx_subjects_thumb ON public.subjects USING btree (thumb);


--
-- Name: idx_version_edition; Type: INDEX; Schema: public; Owner: migrate
--

CREATE UNIQUE INDEX idx_version_edition ON public.versions USING btree (version, edition);


--
-- Name: photos_albums fk_albums_photos; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos_albums
    ADD CONSTRAINT fk_albums_photos FOREIGN KEY (album_uid) REFERENCES public.albums(album_uid);


--
-- Name: auth_users_details fk_auth_users_user_details; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.auth_users_details
    ADD CONSTRAINT fk_auth_users_user_details FOREIGN KEY (user_uid) REFERENCES public.auth_users(user_uid) ON DELETE CASCADE;


--
-- Name: auth_users_settings fk_auth_users_user_settings; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.auth_users_settings
    ADD CONSTRAINT fk_auth_users_user_settings FOREIGN KEY (user_uid) REFERENCES public.auth_users(user_uid) ON DELETE CASCADE;


--
-- Name: auth_users_shares fk_auth_users_user_shares; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.auth_users_shares
    ADD CONSTRAINT fk_auth_users_user_shares FOREIGN KEY (user_uid) REFERENCES public.auth_users(user_uid);


--
-- Name: categories fk_categories_category; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT fk_categories_category FOREIGN KEY (category_id) REFERENCES public.labels(id);


--
-- Name: categories fk_categories_label; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT fk_categories_label FOREIGN KEY (label_id) REFERENCES public.labels(id);


--
-- Name: categories fk_categories_label_categories; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT fk_categories_label_categories FOREIGN KEY (category_id) REFERENCES public.labels(id);


--
-- Name: cells fk_cells_place; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.cells
    ADD CONSTRAINT fk_cells_place FOREIGN KEY (place_id) REFERENCES public.places(id);


--
-- Name: countries fk_countries_country_photo; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.countries
    ADD CONSTRAINT fk_countries_country_photo FOREIGN KEY (country_photo_id) REFERENCES public.photos(id);


--
-- Name: files_share fk_files_share; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.files_share
    ADD CONSTRAINT fk_files_share FOREIGN KEY (file_id) REFERENCES public.files(id);


--
-- Name: files_share fk_files_share_account; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.files_share
    ADD CONSTRAINT fk_files_share_account FOREIGN KEY (service_id) REFERENCES public.services(id);


--
-- Name: files_sync fk_files_sync; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.files_sync
    ADD CONSTRAINT fk_files_sync FOREIGN KEY (file_id) REFERENCES public.files(id);


--
-- Name: files_sync fk_files_sync_account; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.files_sync
    ADD CONSTRAINT fk_files_sync_account FOREIGN KEY (service_id) REFERENCES public.services(id);


--
-- Name: photos_albums fk_photos_albums_album; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos_albums
    ADD CONSTRAINT fk_photos_albums_album FOREIGN KEY (album_uid) REFERENCES public.albums(album_uid);


--
-- Name: photos_albums fk_photos_albums_photo; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos_albums
    ADD CONSTRAINT fk_photos_albums_photo FOREIGN KEY (photo_uid) REFERENCES public.photos(photo_uid);


--
-- Name: photos fk_photos_camera; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos
    ADD CONSTRAINT fk_photos_camera FOREIGN KEY (camera_id) REFERENCES public.cameras(id);


--
-- Name: photos fk_photos_cell; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos
    ADD CONSTRAINT fk_photos_cell FOREIGN KEY (cell_id) REFERENCES public.cells(id);


--
-- Name: details fk_photos_details; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.details
    ADD CONSTRAINT fk_photos_details FOREIGN KEY (photo_id) REFERENCES public.photos(id);


--
-- Name: files fk_photos_files; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT fk_photos_files FOREIGN KEY (photo_id) REFERENCES public.photos(id);


--
-- Name: photos_keywords fk_photos_keywords_keyword; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos_keywords
    ADD CONSTRAINT fk_photos_keywords_keyword FOREIGN KEY (keyword_id) REFERENCES public.keywords(id);


--
-- Name: photos_keywords fk_photos_keywords_photo; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos_keywords
    ADD CONSTRAINT fk_photos_keywords_photo FOREIGN KEY (photo_id) REFERENCES public.photos(id);


--
-- Name: photos_labels fk_photos_labels; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos_labels
    ADD CONSTRAINT fk_photos_labels FOREIGN KEY (photo_id) REFERENCES public.photos(id);


--
-- Name: photos_labels fk_photos_labels_label; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos_labels
    ADD CONSTRAINT fk_photos_labels_label FOREIGN KEY (label_id) REFERENCES public.labels(id);


--
-- Name: photos fk_photos_lens; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos
    ADD CONSTRAINT fk_photos_lens FOREIGN KEY (lens_id) REFERENCES public.lenses(id);


--
-- Name: photos fk_photos_place; Type: FK CONSTRAINT; Schema: public; Owner: migrate
--

ALTER TABLE ONLY public.photos
    ADD CONSTRAINT fk_photos_place FOREIGN KEY (place_id) REFERENCES public.places(id);


--
-- PostgreSQL database dump complete
--

\unrestrict TwLC0hwfzvYiQveep1GpacaAqZFS6tGTIGOWQbXWzdcxTJV5Mw33X7den7q4GM3

