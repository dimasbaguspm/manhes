-- 000007_create_manga_langs.up.sql
CREATE TABLE manga_langs (
    slug       VARCHAR(255) NOT NULL,
    language   VARCHAR(10) NOT NULL,
    available  INT NOT NULL DEFAULT 0,
    downloaded INT NOT NULL DEFAULT 0,
    PRIMARY KEY (slug, language),
    CONSTRAINT fk_manga_langs_manga FOREIGN KEY (slug) REFERENCES manga(slug) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
