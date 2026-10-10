UPDATE files SET file_chroma = CASE WHEN file_colors <> '' THEN 1 ELSE -1 END WHERE file_chroma = 0;
