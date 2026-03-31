ALTER TABLE comments
    ALTER COLUMN article_id TYPE BIGINT USING article_id::BIGINT;
