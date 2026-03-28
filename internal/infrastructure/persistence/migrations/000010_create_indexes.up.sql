-- 000010_create_indexes.up.sql
CREATE INDEX idx_manga_chapters_pending ON manga_chapters(slug, language, chapter_num, uploaded);
