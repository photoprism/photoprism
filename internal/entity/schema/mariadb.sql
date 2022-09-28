create table accounts
(
    id             int unsigned auto_increment
        primary key,
    acc_name       varchar(160)   null,
    acc_owner      varchar(160)   null,
    acc_url        varchar(255)   null,
    acc_type       varbinary(255) null,
    acc_key        varbinary(255) null,
    acc_user       varbinary(255) null,
    acc_pass       varbinary(255) null,
    acc_timeout    varbinary(16)  null,
    acc_error      varbinary(512) null,
    acc_errors     int            null,
    acc_share      tinyint(1)     null,
    acc_sync       tinyint(1)     null,
    retry_limit    int            null,
    share_path     varbinary(500) null,
    share_size     varbinary(16)  null,
    share_expires  int            null,
    sync_path      varbinary(500) null,
    sync_status    varbinary(16)  null,
    sync_interval  int            null,
    sync_date      datetime       null,
    sync_upload    tinyint(1)     null,
    sync_download  tinyint(1)     null,
    sync_filenames tinyint(1)     null,
    sync_raw       tinyint(1)     null,
    created_at     datetime       null,
    updated_at     datetime       null,
    deleted_at     datetime       null
);

create index idx_accounts_deleted_at
    on accounts (deleted_at);

create table addresses
(
    id              int auto_increment
        primary key,
    cell_id         varbinary(64) default 'zz' null,
    address_src     varbinary(8)               null,
    address_lat     float                      null,
    address_lng     float                      null,
    address_line1   varchar(255)               null,
    address_line2   varchar(255)               null,
    address_zip     varchar(32)                null,
    address_city    varchar(128)               null,
    address_state   varchar(128)               null,
    address_country varbinary(2)  default 'zz' null,
    address_notes   varchar(1024)              null,
    created_at      datetime                   null,
    updated_at      datetime                   null,
    deleted_at      datetime                   null
);

create index idx_addresses_address_lat
    on addresses (address_lat);

create index idx_addresses_address_lng
    on addresses (address_lng);

create index idx_addresses_cell_id
    on addresses (cell_id);

create index idx_addresses_deleted_at
    on addresses (deleted_at);

create table albums
(
    id                int unsigned auto_increment
        primary key,
    album_uid         varbinary(64)                   null,
    parent_uid        varbinary(64)   default ''      null,
    album_slug        varbinary(160)                  null,
    album_path        varbinary(500)                  null,
    album_type        varbinary(8)    default 'album' null,
    album_title       varchar(160)                    null,
    album_location    varchar(160)                    null,
    album_category    varchar(100)                    null,
    album_caption     varchar(1024)                   null,
    album_description varchar(2048)                   null,
    album_notes       varchar(1024)                   null,
    album_filter      varbinary(2048) default ''      null,
    album_order       varbinary(32)                   null,
    album_template    varbinary(255)                  null,
    album_state       varchar(100)                    null,
    album_country     varbinary(2)    default 'zz'    null,
    album_year        int                             null,
    album_month       int                             null,
    album_day         int                             null,
    album_favorite    tinyint(1)                      null,
    album_private     tinyint(1)                      null,
    thumb             varbinary(128)  default ''      null,
    thumb_src         varbinary(8)    default ''      null,
    created_at        datetime                        null,
    updated_at        datetime                        null,
    deleted_at        datetime                        null,
    constraint uix_albums_album_uid
        unique (album_uid)
);

create index idx_albums_album_category
    on albums (album_category);

create index idx_albums_album_filter
    on albums (album_filter(512));

create index idx_albums_album_path
    on albums (album_path);

create index idx_albums_album_slug
    on albums (album_slug);

create index idx_albums_album_state
    on albums (album_state);

create index idx_albums_album_title
    on albums (album_title);

create index idx_albums_country_year_month
    on albums (album_country, album_year, album_month);

create index idx_albums_deleted_at
    on albums (deleted_at);

create index idx_albums_thumb
    on albums (thumb);

create index idx_albums_ymd
    on albums (album_day);

create table cameras
(
    id                 int unsigned auto_increment
        primary key,
    camera_slug        varbinary(160) null,
    camera_name        varchar(160)   null,
    camera_make        varchar(160)   null,
    camera_model       varchar(160)   null,
    camera_type        varchar(100)   null,
    camera_description varchar(2048)  null,
    camera_notes       varchar(1024)  null,
    created_at         datetime       null,
    updated_at         datetime       null,
    deleted_at         datetime       null,
    constraint uix_cameras_camera_slug
        unique (camera_slug)
);

create index idx_cameras_deleted_at
    on cameras (deleted_at);

create table categories
(
    label_id    int unsigned not null,
    category_id int unsigned not null,
    primary key (label_id, category_id)
);

create table cells
(
    id            varbinary(64)              not null
        primary key,
    cell_name     varchar(200)               null,
    cell_street   varchar(100)               null,
    cell_postcode varchar(50)                null,
    cell_category varchar(50)                null,
    place_id      varbinary(64) default 'zz' null,
    created_at    datetime                   null,
    updated_at    datetime                   null
);

create table countries
(
    id                  varbinary(2)   not null
        primary key,
    country_slug        varbinary(160) null,
    country_name        varchar(160)   null,
    country_description varchar(2048)  null,
    country_notes       varchar(1024)  null,
    country_photo_id    int unsigned   null,
    constraint uix_countries_country_slug
        unique (country_slug)
);

create table details
(
    photo_id      int unsigned  not null
        primary key,
    keywords      varchar(2048) null,
    keywords_src  varbinary(8)  null,
    notes         varchar(2048) null,
    notes_src     varbinary(8)  null,
    subject       varchar(1024) null,
    subject_src   varbinary(8)  null,
    artist        varchar(1024) null,
    artist_src    varbinary(8)  null,
    copyright     varchar(1024) null,
    copyright_src varbinary(8)  null,
    license       varchar(1024) null,
    license_src   varbinary(8)  null,
    software      varchar(1024) null,
    software_src  varbinary(8)  null,
    created_at    datetime      null,
    updated_at    datetime      null
);

create table duplicates
(
    file_name varbinary(755)             not null,
    file_root varbinary(16)  default '/' not null,
    file_hash varbinary(128) default ''  null,
    file_size bigint                     null,
    mod_time  bigint                     null,
    primary key (file_name, file_root)
);

create index idx_duplicates_file_hash
    on duplicates (file_hash);

create table errors
(
    id            int unsigned auto_increment
        primary key,
    error_time    datetime        null,
    error_level   varbinary(32)   null,
    error_message varbinary(2048) null
);

create index idx_errors_error_time
    on errors (error_time);

create table faces
(
    id               varbinary(64)            not null
        primary key,
    face_src         varbinary(8)             null,
    face_kind        int                      null,
    face_hidden      tinyint(1)               null,
    subj_uid         varbinary(64) default '' null,
    samples          int                      null,
    sample_radius    double                   null,
    collisions       int                      null,
    collision_radius double                   null,
    embedding_json   mediumblob               null,
    matched_at       datetime                 null,
    created_at       datetime                 null,
    updated_at       datetime                 null
);

create index idx_faces_subj_uid
    on faces (subj_uid);

create table files
(
    id                 int unsigned auto_increment
        primary key,
    photo_id           int unsigned              null,
    photo_uid          varbinary(64)             null,
    photo_taken_at     datetime                  null,
    time_index         varbinary(48)             null,
    media_id           varbinary(32)             null,
    media_utc          bigint                    null,
    instance_id        varbinary(64)             null,
    file_uid           varbinary(64)             null,
    file_name          varbinary(755)            null,
    file_root          varbinary(16) default '/' null,
    original_name      varbinary(755)            null,
    file_hash          varbinary(128)            null,
    file_size          bigint                    null,
    file_codec         varbinary(32)             null,
    file_type          varbinary(16)             null,
    media_type         varbinary(16)             null,
    file_mime          varbinary(64)             null,
    file_primary       tinyint(1)                null,
    file_sidecar       tinyint(1)                null,
    file_missing       tinyint(1)                null,
    file_portrait      tinyint(1)                null,
    file_video         tinyint(1)                null,
    file_duration      bigint                    null,
    file_fps           double                    null,
    file_frames        int                       null,
    file_width         int                       null,
    file_height        int                       null,
    file_orientation   int                       null,
    file_projection    varbinary(64)             null,
    file_aspect_ratio  float                     null,
    file_hdr           tinyint(1)                null,
    file_watermark     tinyint(1)                null,
    file_color_profile varbinary(64)             null,
    file_main_color    varbinary(16)             null,
    file_colors        varbinary(9)              null,
    file_luminance     varbinary(9)              null,
    file_diff          int           default -1  null,
    file_chroma        smallint      default -1  null,
    file_software      varchar(64)               null,
    file_error         varbinary(512)            null,
    mod_time           bigint                    null,
    created_at         datetime                  null,
    created_in         bigint                    null,
    updated_at         datetime                  null,
    updated_in         bigint                    null,
    deleted_at         datetime                  null,
    constraint idx_files_name_root
        unique (file_name, file_root),
    constraint idx_files_search_media
        unique (media_id),
    constraint idx_files_search_timeline
        unique (time_index),
    constraint uix_files_file_uid
        unique (file_uid)
);

create index idx_files_deleted_at
    on files (deleted_at);

create index idx_files_file_hash
    on files (file_hash);

create index idx_files_file_main_color
    on files (file_main_color);

create index idx_files_instance_id
    on files (instance_id);

create index idx_files_media_utc
    on files (media_utc);

create index idx_files_missing_root
    on files (file_missing, file_root);

create index idx_files_photo_id
    on files (photo_id, file_primary);

create index idx_files_photo_taken_at
    on files (photo_taken_at);

create index idx_files_photo_uid
    on files (photo_uid);

create table files_share
(
    file_id     int unsigned   not null,
    account_id  int unsigned   not null,
    remote_name varbinary(255) not null,
    status      varbinary(16)  null,
    error       varbinary(512) null,
    errors      int            null,
    created_at  datetime       null,
    updated_at  datetime       null,
    primary key (file_id, account_id, remote_name)
);

create table files_sync
(
    remote_name varbinary(255) not null,
    account_id  int unsigned   not null,
    file_id     int unsigned   null,
    remote_date datetime       null,
    remote_size bigint         null,
    status      varbinary(16)  null,
    error       varbinary(512) null,
    errors      int            null,
    created_at  datetime       null,
    updated_at  datetime       null,
    primary key (remote_name, account_id)
);

create index idx_files_sync_file_id
    on files_sync (file_id);

create table folders
(
    path               varbinary(500)             null,
    root               varbinary(16) default ''   null,
    folder_uid         varbinary(64)              not null
        primary key,
    folder_type        varbinary(16)              null,
    folder_title       varchar(200)               null,
    folder_category    varchar(100)               null,
    folder_description varchar(2048)              null,
    folder_order       varbinary(32)              null,
    folder_country     varbinary(2)  default 'zz' null,
    folder_year        int                        null,
    folder_month       int                        null,
    folder_day         int                        null,
    folder_favorite    tinyint(1)                 null,
    folder_private     tinyint(1)                 null,
    folder_ignore      tinyint(1)                 null,
    folder_watch       tinyint(1)                 null,
    created_at         datetime                   null,
    updated_at         datetime                   null,
    modified_at        datetime                   null,
    deleted_at         datetime                   null,
    constraint idx_folders_path_root
        unique (path, root)
);

create index idx_folders_country_year_month
    on folders (folder_country, folder_year, folder_month);

create index idx_folders_deleted_at
    on folders (deleted_at);

create index idx_folders_folder_category
    on folders (folder_category);

create table keywords
(
    id      int unsigned auto_increment
        primary key,
    keyword varchar(64) null,
    skip    tinyint(1)  null
);

create index idx_keywords_keyword
    on keywords (keyword);

create table labels
(
    id                int unsigned auto_increment
        primary key,
    label_uid         varbinary(64)             null,
    label_slug        varbinary(160)            null,
    custom_slug       varbinary(160)            null,
    label_name        varchar(160)              null,
    label_priority    int                       null,
    label_favorite    tinyint(1)                null,
    label_description varchar(2048)             null,
    label_notes       varchar(1024)             null,
    photo_count       int            default 1  null,
    thumb             varbinary(128) default '' null,
    thumb_src         varbinary(8)   default '' null,
    created_at        datetime                  null,
    updated_at        datetime                  null,
    deleted_at        datetime                  null,
    constraint uix_labels_label_slug
        unique (label_slug),
    constraint uix_labels_label_uid
        unique (label_uid)
);

create index idx_labels_custom_slug
    on labels (custom_slug);

create index idx_labels_deleted_at
    on labels (deleted_at);

create index idx_labels_thumb
    on labels (thumb);

create table lenses
(
    id               int unsigned auto_increment
        primary key,
    lens_slug        varbinary(160) null,
    lens_name        varchar(160)   null,
    lens_make        varchar(160)   null,
    lens_model       varchar(160)   null,
    lens_type        varchar(100)   null,
    lens_description varchar(2048)  null,
    lens_notes       varchar(1024)  null,
    created_at       datetime       null,
    updated_at       datetime       null,
    deleted_at       datetime       null,
    constraint uix_lenses_lens_slug
        unique (lens_slug)
);

create index idx_lenses_deleted_at
    on lenses (deleted_at);

create table links
(
    link_uid     varbinary(64)  not null
        primary key,
    share_uid    varbinary(64)  null,
    share_slug   varbinary(160) null,
    link_token   varbinary(160) null,
    link_expires int            null,
    link_views   int unsigned   null,
    max_views    int unsigned   null,
    has_password tinyint(1)     null,
    can_comment  tinyint(1)     null,
    can_edit     tinyint(1)     null,
    created_at   datetime       null,
    modified_at  datetime       null,
    constraint idx_links_uid_token
        unique (share_uid, link_token)
);

create index idx_links_share_slug
    on links (share_slug);

create table markers
(
    marker_uid      varbinary(64)             not null
        primary key,
    file_uid        varbinary(64)  default '' null,
    marker_type     varbinary(8)   default '' null,
    marker_src      varbinary(8)   default '' null,
    marker_name     varchar(160)              null,
    marker_review   tinyint(1)                null,
    marker_invalid  tinyint(1)                null,
    subj_uid        varbinary(64)             null,
    subj_src        varbinary(8)   default '' null,
    face_id         varbinary(64)             null,
    face_dist       double         default -1 null,
    embeddings_json mediumblob                null,
    landmarks_json  mediumblob                null,
    x               float                     null,
    y               float                     null,
    w               float                     null,
    h               float                     null,
    q               int                       null,
    size            int            default -1 null,
    score           smallint                  null,
    thumb           varbinary(128) default '' null,
    matched_at      datetime                  null,
    created_at      datetime                  null,
    updated_at      datetime                  null
);

create index idx_markers_face_id
    on markers (face_id);

create index idx_markers_file_uid
    on markers (file_uid);

create index idx_markers_matched_at
    on markers (matched_at);

create index idx_markers_subj_uid_src
    on markers (subj_uid, subj_src);

create index idx_markers_thumb
    on markers (thumb);

create table migrations
(
    id          varchar(16)  not null
        primary key,
    dialect     varchar(16)  null,
    error       varchar(255) null,
    source      varchar(16)  null,
    started_at  datetime     null,
    finished_at datetime     null
);

create table passwords
(
    uid        varbinary(255) not null
        primary key,
    hash       varbinary(255) null,
    created_at datetime       null,
    updated_at datetime       null
);

create table photos
(
    id                 int unsigned auto_increment
        primary key,
    uuid               varbinary(64)                 null,
    taken_at           datetime                      null,
    taken_at_local     datetime                      null,
    taken_src          varbinary(8)                  null,
    photo_uid          varbinary(64)                 null,
    photo_type         varbinary(8)  default 'image' null,
    type_src           varbinary(8)                  null,
    photo_title        varchar(200)                  null,
    title_src          varbinary(8)                  null,
    photo_description  varchar(4096)                 null,
    description_src    varbinary(8)                  null,
    photo_path         varbinary(500)                null,
    photo_name         varbinary(255)                null,
    original_name      varbinary(755)                null,
    photo_stack        tinyint                       null,
    photo_favorite     tinyint(1)                    null,
    photo_private      tinyint(1)                    null,
    photo_scan         tinyint(1)                    null,
    photo_panorama     tinyint(1)                    null,
    time_zone          varbinary(64)                 null,
    place_id           varbinary(64) default 'zz'    null,
    place_src          varbinary(8)                  null,
    cell_id            varbinary(42) default 'zz'    null,
    cell_accuracy      int                           null,
    photo_altitude     int                           null,
    photo_lat          float                         null,
    photo_lng          float                         null,
    photo_country      varbinary(2)  default 'zz'    null,
    photo_year         int                           null,
    photo_month        int                           null,
    photo_day          int                           null,
    photo_iso          int                           null,
    photo_exposure     varbinary(64)                 null,
    photo_f_number     float                         null,
    photo_focal_length int                           null,
    photo_quality      smallint                      null,
    photo_faces        int                           null,
    photo_resolution   smallint                      null,
    photo_color        smallint      default -1      null,
    camera_id          int unsigned  default 1       null,
    camera_serial      varbinary(160)                null,
    camera_src         varbinary(8)                  null,
    lens_id            int unsigned  default 1       null,
    created_at         datetime                      null,
    updated_at         datetime                      null,
    edited_at          datetime                      null,
    checked_at         datetime                      null,
    estimated_at       datetime                      null,
    deleted_at         datetime                      null,
    constraint uix_photos_photo_uid
        unique (photo_uid)
);

create index idx_photos_camera_lens
    on photos (camera_id, lens_id);

create index idx_photos_cell_id
    on photos (cell_id);

create index idx_photos_checked_at
    on photos (checked_at);

create index idx_photos_country_year_month
    on photos (photo_country, photo_year, photo_month);

create index idx_photos_deleted_at
    on photos (deleted_at);

create index idx_photos_path_name
    on photos (photo_path, photo_name);

create index idx_photos_photo_lat
    on photos (photo_lat);

create index idx_photos_photo_lng
    on photos (photo_lng);

create index idx_photos_place_id
    on photos (place_id);

create index idx_photos_taken_uid
    on photos (taken_at, photo_uid);

create index idx_photos_uuid
    on photos (uuid);

create index idx_photos_ymd
    on photos (photo_day);

create table photos_albums
(
    photo_uid  varbinary(42) not null,
    album_uid  varbinary(42) not null,
    `order`    int           null,
    hidden     tinyint(1)    null,
    missing    tinyint(1)    null,
    created_at datetime      null,
    updated_at datetime      null,
    primary key (photo_uid, album_uid)
);

create index idx_photos_albums_album_uid
    on photos_albums (album_uid);

create table photos_keywords
(
    photo_id   int unsigned not null,
    keyword_id int unsigned not null,
    primary key (photo_id, keyword_id)
);

create index idx_photos_keywords_keyword_id
    on photos_keywords (keyword_id);

create table photos_labels
(
    photo_id    int unsigned not null,
    label_id    int unsigned not null,
    label_src   varbinary(8) null,
    uncertainty smallint     null,
    primary key (photo_id, label_id)
);

create index idx_photos_labels_label_id
    on photos_labels (label_id);

create table places
(
    id             varbinary(42) not null
        primary key,
    place_label    varchar(400)  null,
    place_district varchar(100)  null,
    place_city     varchar(100)  null,
    place_state    varchar(100)  null,
    place_country  varbinary(2)  null,
    place_keywords varchar(300)  null,
    place_favorite tinyint(1)    null,
    photo_count    int default 1 null,
    created_at     datetime      null,
    updated_at     datetime      null
);

create index idx_places_place_city
    on places (place_city);

create index idx_places_place_district
    on places (place_district);

create index idx_places_place_state
    on places (place_state);

create table subjects
(
    subj_uid      varbinary(42)             not null
        primary key,
    subj_type     varbinary(8)   default '' null,
    subj_src      varbinary(8)   default '' null,
    subj_slug     varbinary(160) default '' null,
    subj_name     varchar(160)   default '' null,
    subj_alias    varchar(160)   default '' null,
    subj_bio      varchar(2048)             null,
    subj_notes    varchar(1024)             null,
    subj_favorite tinyint(1)     default 0  null,
    subj_hidden   tinyint(1)     default 0  null,
    subj_private  tinyint(1)     default 0  null,
    subj_excluded tinyint(1)     default 0  null,
    file_count    int            default 0  null,
    photo_count   int            default 0  null,
    thumb         varbinary(128) default '' null,
    thumb_src     varbinary(8)   default '' null,
    metadata_json mediumblob                null,
    created_at    datetime                  null,
    updated_at    datetime                  null,
    deleted_at    datetime                  null,
    constraint uix_subjects_subj_name
        unique (subj_name)
);

create index idx_subjects_deleted_at
    on subjects (deleted_at);

create index idx_subjects_subj_slug
    on subjects (subj_slug);

create index idx_subjects_thumb
    on subjects (thumb);

create table users
(
    id              int auto_increment
        primary key,
    address_id      int default 1  null,
    user_uid        varbinary(42)  null,
    mother_uid      varbinary(42)  null,
    father_uid      varbinary(42)  null,
    global_uid      varbinary(42)  null,
    full_name       varchar(128)   null,
    nick_name       varchar(64)    null,
    maiden_name     varchar(64)    null,
    artist_name     varchar(64)    null,
    user_name       varchar(64)    null,
    user_status     varchar(32)    null,
    user_disabled   tinyint(1)     null,
    user_settings   longtext       null,
    primary_email   varchar(255)   null,
    email_confirmed tinyint(1)     null,
    backup_email    varchar(255)   null,
    person_url      varbinary(255) null,
    person_phone    varchar(32)    null,
    person_status   varchar(32)    null,
    person_avatar   varbinary(255) null,
    person_location varchar(128)   null,
    person_bio      text           null,
    person_accounts longtext       null,
    business_url    varbinary(255) null,
    business_phone  varchar(32)    null,
    business_email  varchar(255)   null,
    company_name    varchar(128)   null,
    department_name varchar(128)   null,
    job_title       varchar(64)    null,
    birth_year      int            null,
    birth_month     int            null,
    birth_day       int            null,
    terms_accepted  tinyint(1)     null,
    is_artist       tinyint(1)     null,
    is_subject      tinyint(1)     null,
    role_admin      tinyint(1)     null,
    role_guest      tinyint(1)     null,
    role_child      tinyint(1)     null,
    role_family     tinyint(1)     null,
    role_friend     tinyint(1)     null,
    webdav          tinyint(1)     null,
    storage_path    varbinary(500) null,
    can_invite      tinyint(1)     null,
    invite_token    varbinary(32)  null,
    invited_by      varbinary(32)  null,
    confirm_token   varbinary(64)  null,
    reset_token     varbinary(64)  null,
    api_token       varbinary(128) null,
    api_secret      varbinary(128) null,
    login_attempts  int            null,
    login_at        datetime       null,
    created_at      datetime       null,
    updated_at      datetime       null,
    deleted_at      datetime       null,
    constraint uix_users_user_uid
        unique (user_uid)
);

create index idx_users_deleted_at
    on users (deleted_at);

create index idx_users_global_uid
    on users (global_uid);

create index idx_users_primary_email
    on users (primary_email);

