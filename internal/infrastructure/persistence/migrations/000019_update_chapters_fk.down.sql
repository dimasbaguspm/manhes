-- 000019_update_chapters_fk.down.sql
ALTER TABLE chapters DROP FOREIGN KEY fk_chapters_manga;
ALTER TABLE chapters MODIFY COLUMN manga_id VARCHAR(255) NOT NULL;
ALTER TABLE chapters ADD FOREIGN KEY (manga_id) REFERENCES manga(slug) ON DELETE CASCADE;
