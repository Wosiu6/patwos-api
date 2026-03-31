ALTER TABLE comments
    ALTER COLUMN article_id TYPE TEXT USING article_id::TEXT;
