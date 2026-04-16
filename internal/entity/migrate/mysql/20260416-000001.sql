ALTER TABLE auth_users_settings
    ADD COLUMN IF NOT EXISTS display_metadata_cards TEXT NULL,
    ADD COLUMN IF NOT EXISTS display_metadata_list TEXT NULL,
    ADD COLUMN IF NOT EXISTS display_metadata_lightbox TEXT NULL;
