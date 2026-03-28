-- 000009_create_chapter_pages.up.sql
CREATE TABLE chapter_pages (
    slug        VARCHAR(255) NOT NULL,
    language    VARCHAR(10) NOT NULL,
    chapter_num VARCHAR(50) NOT NULL,
    page_index  INT NOT NULL,
    s3_url      VARCHAR(1000) NOT NULL,
    PRIMARY KEY (slug, language, chapter_num, page_index),
    CONSTRAINT fk_chapter_pages_chapter FOREIGN KEY (slug, language, chapter_num) REFERENCES manga_chapters(slug, language, chapter_num) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
