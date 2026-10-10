UPDATE faces SET face_kind = 1 WHERE face_kind = 0 AND embedding_json IS NOT NULL AND LENGTH(embedding_json) > 0;
