-- 000019_update_chapters_fk.up.sql
-- Update chapters table FK from manga(slug) to manga(id).
ALTER TABLE chapters DROP FOREIGN KEY fk_chapters_manga;
ALTER TABLE chapters MODIFY COLUMN manga_id VARCHAR(36) NOT NULL;
ALTER TABLE chapters ADD FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE;
