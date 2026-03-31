DELETE FROM article_votes v
WHERE NOT EXISTS (SELECT 1 FROM articles a WHERE a.id = v.article_id);

DELETE FROM article_votes v
WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = v.user_id);
