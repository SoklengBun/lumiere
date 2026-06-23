package main

import (
	"log"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"lumiere/internal/artist"
	"lumiere/internal/config"
	"lumiere/internal/database"
	"lumiere/internal/lyrics"
	"lumiere/internal/models"
	"lumiere/internal/playlist"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.NewFromEnv()
	if err != nil {
		log.Fatal(err.Error())
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := migrateLyricsIDToString(db); err != nil {
		log.Fatal(err.Error())
	}

	if err := migrateArtistCVToSingleFK(db); err != nil {
		log.Fatal(err.Error())
	}

	if err := migrateArtistNormalizedToAltName(db); err != nil {
		log.Fatal(err.Error())
	}

	if err := normalizeLyricCoversPrimaryKey(db); err != nil {
		log.Fatal(err.Error())
	}

	if err := migrateLyricsPrimaryTitle(db); err != nil {
		log.Fatal(err.Error())
	}

	if err := migrateLyricTitlesToArray(db); err != nil {
		log.Fatal(err.Error())
	}

	if err := dropLegacyLyricReferences(db); err != nil {
		log.Fatal(err.Error())
	}

	err = db.AutoMigrate(
		&models.User{},
		&artist.Artist{}, // also creates artist_cvs join table
		&lyrics.Lyrics{},
		&lyrics.LyricCover{},
		&lyrics.LyricContent{},
		&playlist.Playlist{},
		&playlist.PlaylistItem{},
	)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func dropLegacyLyricReferences(db *gorm.DB) error {
	return db.Exec(`DROP TABLE IF EXISTS lyric_references CASCADE;`).Error
}

func migrateLyricTitlesToArray(db *gorm.DB) error {
	// Legacy schema stored alt titles in lyric_titles. Keep only the strings on lyrics.alt_titles.
	return db.Exec(`DO $$
BEGIN
	ALTER TABLE IF EXISTS lyrics ADD COLUMN IF NOT EXISTS alt_titles JSONB NOT NULL DEFAULT '[]'::jsonb;

	IF EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = CURRENT_SCHEMA()
			AND table_name = 'lyric_titles'
	) THEN
		UPDATE lyrics l
		SET alt_titles = x.titles
		FROM (
			SELECT lyrics_id, jsonb_agg(title ORDER BY id) AS titles
			FROM lyric_titles
			WHERE COALESCE(TRIM(title), '') <> ''
			GROUP BY lyrics_id
		) x
		WHERE l.id = x.lyrics_id
			AND COALESCE(l.alt_titles, '[]'::jsonb) = '[]'::jsonb;

		DROP TABLE IF EXISTS lyric_titles CASCADE;
	END IF;
END $$;`).Error
}

func migrateLyricsPrimaryTitle(db *gorm.DB) error {
	// Ensure lyrics.title exists and try to backfill from legacy summary/title rows.
	return db.Exec(`DO $$
BEGIN
		ALTER TABLE IF EXISTS lyrics ADD COLUMN IF NOT EXISTS title TEXT;

		IF EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = CURRENT_SCHEMA()
				AND table_name = 'lyrics'
				AND column_name = 'summary'
		) THEN
			UPDATE lyrics
			SET title = summary
			WHERE COALESCE(TRIM(title), '') = ''
				AND COALESCE(TRIM(summary), '') <> '';
		END IF;

		IF EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = CURRENT_SCHEMA()
				AND table_name = 'lyric_titles'
		) THEN
			UPDATE lyrics l
			SET title = x.title
			FROM (
				SELECT DISTINCT ON (lyrics_id) lyrics_id, title
				FROM lyric_titles
				WHERE COALESCE(TRIM(title), '') <> ''
				ORDER BY lyrics_id, id ASC
			) x
			WHERE l.id = x.lyrics_id
				AND COALESCE(TRIM(l.title), '') = '';
		END IF;
END $$;`).Error
}

func migrateLyricsIDToString(db *gorm.DB) error {
	// Legacy schema used bigint lyrics IDs. Convert to varchar(64) so YouTube IDs can be used as PK.
	return db.Exec(`DO $$
BEGIN
	ALTER TABLE IF EXISTS lyric_covers DROP CONSTRAINT IF EXISTS fk_lyric_covers_covers;

	IF EXISTS (
		SELECT 1
		FROM information_schema.columns
		WHERE table_schema = 'public'
			AND table_name = 'lyrics'
			AND column_name = 'id'
			AND udt_name IN ('int8', 'int4')
	) THEN
		ALTER TABLE IF EXISTS lyric_contents DROP CONSTRAINT IF EXISTS fk_lyrics_contents;
		ALTER TABLE IF EXISTS lyric_titles DROP CONSTRAINT IF EXISTS fk_lyrics_titles;
		ALTER TABLE IF EXISTS lyrics_artists DROP CONSTRAINT IF EXISTS fk_lyrics_artists_lyrics;
		ALTER TABLE IF EXISTS playlist_items DROP CONSTRAINT IF EXISTS fk_playlist_items_lyrics;
		ALTER TABLE IF EXISTS lyric_references DROP CONSTRAINT IF EXISTS fk_lyrics_references;
		ALTER TABLE IF EXISTS lyric_covers DROP CONSTRAINT IF EXISTS fk_lyric_covers_lyrics;
		ALTER TABLE IF EXISTS lyric_covers DROP CONSTRAINT IF EXISTS fk_lyric_covers_covers;

		ALTER TABLE IF EXISTS lyric_contents ALTER COLUMN lyrics_id TYPE varchar(64) USING lyrics_id::text;
		ALTER TABLE IF EXISTS lyric_titles ALTER COLUMN lyrics_id TYPE varchar(64) USING lyrics_id::text;
		ALTER TABLE IF EXISTS lyrics_artists ALTER COLUMN lyrics_id TYPE varchar(64) USING lyrics_id::text;
		ALTER TABLE IF EXISTS playlist_items ALTER COLUMN lyrics_id TYPE varchar(64) USING lyrics_id::text;
		ALTER TABLE IF EXISTS lyric_references ALTER COLUMN lyrics_id TYPE varchar(64) USING lyrics_id::text;
		ALTER TABLE IF EXISTS lyric_covers ALTER COLUMN lyrics_id TYPE varchar(64) USING lyrics_id::text;
		ALTER TABLE IF EXISTS lyric_covers ALTER COLUMN cover_id TYPE varchar(64) USING cover_id::text;
		ALTER TABLE lyrics ALTER COLUMN id TYPE varchar(64) USING id::text;
	END IF;
END $$;`).Error
}

func normalizeLyricCoversPrimaryKey(db *gorm.DB) error {
	// Ensure lyric_covers.id is the primary key so many2many FKs can reference it.
	return db.Exec(`DO $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = CURRENT_SCHEMA()
			AND table_name = 'lyric_covers'
	) THEN
		ALTER TABLE lyric_covers ADD COLUMN IF NOT EXISTS id BIGINT;

		CREATE SEQUENCE IF NOT EXISTS lyric_covers_id_seq;
		ALTER SEQUENCE lyric_covers_id_seq OWNED BY lyric_covers.id;
		ALTER TABLE lyric_covers ALTER COLUMN id SET DEFAULT nextval('lyric_covers_id_seq');

		UPDATE lyric_covers
		SET id = nextval('lyric_covers_id_seq')
		WHERE id IS NULL;

		WITH duplicated AS (
			SELECT ctid,
				ROW_NUMBER() OVER (PARTITION BY id ORDER BY ctid) AS rn
			FROM lyric_covers
			WHERE id IS NOT NULL
		)
		UPDATE lyric_covers lc
		SET id = nextval('lyric_covers_id_seq')
		FROM duplicated d
		WHERE lc.ctid = d.ctid AND d.rn > 1;

		IF EXISTS (SELECT 1 FROM lyric_covers) THEN
			PERFORM setval('lyric_covers_id_seq', (SELECT MAX(id) FROM lyric_covers), true);
		ELSE
			PERFORM setval('lyric_covers_id_seq', 1, false);
		END IF;

		IF NOT EXISTS (
			SELECT 1
			FROM pg_constraint c
			JOIN pg_class t ON c.conrelid = t.oid
			JOIN unnest(c.conkey) WITH ORDINALITY AS k(attnum, ord) ON true
			JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = k.attnum
			WHERE t.relname = 'lyric_covers' AND c.contype = 'p'
			GROUP BY c.oid
			HAVING array_agg(a.attname::text ORDER BY k.ord) = ARRAY['id']::text[]
		) THEN
			ALTER TABLE lyric_covers DROP CONSTRAINT IF EXISTS lyric_covers_pkey CASCADE;
			ALTER TABLE lyric_covers ADD CONSTRAINT lyric_covers_pkey PRIMARY KEY (id);
		END IF;

		CREATE UNIQUE INDEX IF NOT EXISTS idx_lyric_covers_lyrics_cover_unique
			ON lyric_covers (lyrics_id, cover_id);
	END IF;
END $$;`).Error
}

func migrateArtistCVToSingleFK(db *gorm.DB) error {
	// Drop the old many-to-many artist_cvs join table and replace with a single cv_id FK column.
	return db.Exec(`DO $$
BEGIN
	ALTER TABLE IF EXISTS artists ADD COLUMN IF NOT EXISTS cv_id BIGINT REFERENCES artists(id) ON DELETE SET NULL;

	IF EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = CURRENT_SCHEMA()
			AND table_name = 'artist_cvs'
	) THEN
		UPDATE artists a
		SET cv_id = (
			SELECT cv_id FROM artist_cvs WHERE artist_id = a.id LIMIT 1
		)
		WHERE cv_id IS NULL;

		DROP TABLE IF EXISTS artist_cvs;
	END IF;
END $$;`).Error
}

func migrateArtistNormalizedToAltName(db *gorm.DB) error {
	return db.Exec(`DO $$
BEGIN
	IF EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = CURRENT_SCHEMA()
			AND table_name = 'artists'
			AND column_name = 'normalized_name'
	) THEN
		ALTER TABLE artists RENAME COLUMN normalized_name TO alt_name;
	ELSE
		ALTER TABLE artists ADD COLUMN IF NOT EXISTS alt_name varchar(255);
	END IF;
	CREATE INDEX IF NOT EXISTS idx_artists_name_lower ON artists (LOWER(name));
	CREATE INDEX IF NOT EXISTS idx_artists_alt_name_lower ON artists (LOWER(alt_name));
END $$;`).Error
}
