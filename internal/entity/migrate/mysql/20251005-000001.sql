UPDATE labels SET label_nsfw = 0 WHERE label_nsfw IS NULL;
UPDATE photos_labels SET nsfw = 0 WHERE nsfw IS NULL;
UPDATE photos_labels SET topicality = 0 WHERE topicality IS NULL;