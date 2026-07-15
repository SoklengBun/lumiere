package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"lumiere/internal/cache"
	lyricsmodel "lumiere/internal/lyrics"
	lyricsrepo "lumiere/internal/lyrics/repository"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	repo  lyricsrepo.LyricsRepo
	cache cache.Cache
}

func New(repo lyricsrepo.LyricsRepo, caches ...cache.Cache) *Service {
	var c cache.Cache
	if len(caches) > 0 {
		c = caches[0]
	}
	return &Service{repo: repo, cache: c}
}

const (
	lyricsCacheTTL     = 30 * 24 * time.Hour // 30 days
	lyricsListCacheTTL = 30 * 24 * time.Hour // 30 days
)

const lyricsListVersionKey = "lumiere:lyrics:list:version"

type lyricsListCacheEntry struct {
	Items []lyricsmodel.Lyrics `json:"items"`
	Total int64                `json:"total"`
}

func lyricsIDCacheKey(id uint) string {
	return fmt.Sprintf("lumiere:lyrics:id:%d", id)
}

func lyricsVideoCacheKey(videoID string) string {
	// Hash the caller-provided ID so cache keys remain compact and contain no
	// characters that have special meaning to Redis tooling.
	hash := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(videoID))))
	return fmt.Sprintf("lumiere:lyrics:video:%x", hash[:])
}

func (s *Service) cacheLyrics(ctx context.Context, l *lyricsmodel.Lyrics) {
	if s.cache == nil || l == nil {
		return
	}
	_ = s.cache.Set(ctx, lyricsIDCacheKey(l.ID), l, lyricsCacheTTL)
	_ = s.cache.Set(ctx, lyricsVideoCacheKey(l.VideoID), l, lyricsCacheTTL)
}

func (s *Service) invalidateLyrics(ctx context.Context, l *lyricsmodel.Lyrics) {
	if s.cache == nil || l == nil {
		return
	}
	_ = s.cache.Delete(ctx, lyricsIDCacheKey(l.ID), lyricsVideoCacheKey(l.VideoID))
}

func (s *Service) lyricsListVersion(ctx context.Context) string {
	if s.cache == nil {
		return ""
	}

	var version int64
	if err := s.cache.Get(ctx, lyricsListVersionKey, &version); err != nil || version < 0 {
		// Version zero is the initial namespace. The first write increments it
		// to one, invalidating pages cached before that write.
		version = 0
	}
	return strconv.FormatInt(version, 10)
}

func lyricsListCacheKey(version string, page, offset int) string {
	return fmt.Sprintf("lumiere:lyrics:list:v%s:page:%d:offset:%d", version, page, offset)
}

func (s *Service) invalidateLyricsList(ctx context.Context) {
	if s.cache == nil {
		return
	}
	if counter, ok := s.cache.(cache.Counter); ok {
		_, _ = counter.Increment(ctx, lyricsListVersionKey)
	}
}

func (s *Service) Add(ctx context.Context, l *lyricsmodel.Lyrics) (*lyricsmodel.Lyrics, error) {
	if err := s.repo.Create(ctx, l); err != nil {
		return nil, err
	}
	s.cacheLyrics(ctx, l)
	s.invalidateLyricsList(ctx)
	return l, nil
}

func (s *Service) Get(ctx context.Context, id uint) (*lyricsmodel.Lyrics, error) {
	if s.cache != nil {
		var l lyricsmodel.Lyrics
		if err := s.cache.Get(ctx, lyricsIDCacheKey(id), &l); err == nil {
			return &l, nil
		}
	}

	l, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.cacheLyrics(ctx, l)
	return l, nil
}

func (s *Service) GetByVideoID(ctx context.Context, videoID string) (*lyricsmodel.Lyrics, error) {
	videoID = strings.TrimSpace(videoID)
	if s.cache != nil {
		var l lyricsmodel.Lyrics
		if err := s.cache.Get(ctx, lyricsVideoCacheKey(videoID), &l); err == nil {
			return &l, nil
		}
	}

	l, err := s.repo.GetByVideoID(ctx, videoID)
	if err != nil {
		return nil, err
	}
	s.cacheLyrics(ctx, l)
	return l, nil
}

func (s *Service) List(ctx context.Context, page int, offset int) ([]lyricsmodel.Lyrics, int64, error) {
	version := s.lyricsListVersion(ctx)
	key := lyricsListCacheKey(version, page, offset)
	if s.cache != nil {
		var entry lyricsListCacheEntry
		if err := s.cache.Get(ctx, key, &entry); err == nil {
			return entry.Items, entry.Total, nil
		}
	}

	list, total, err := s.repo.List(ctx, page, offset)
	if err != nil {
		return nil, 0, err
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, key, lyricsListCacheEntry{Items: list, Total: total}, lyricsListCacheTTL)
	}
	return list, total, nil
}

func (s *Service) ListRandom(ctx context.Context, limit int) ([]lyricsmodel.Lyrics, error) {
	return s.repo.ListRandom(ctx, limit)
}

func (s *Service) ListByUser(ctx context.Context, userID uint) ([]lyricsmodel.Lyrics, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Search(ctx context.Context, q string) ([]lyricsmodel.Lyrics, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []lyricsmodel.Lyrics{}, nil
	}
	return s.repo.Search(ctx, q)
}

func (s *Service) Update(ctx context.Context, l *lyricsmodel.Lyrics) (*lyricsmodel.Lyrics, error) {
	var old *lyricsmodel.Lyrics
	if s.cache != nil {
		// Keep the old video ID available in case the update changes it.
		old, _ = s.repo.GetByID(ctx, l.ID)
	}
	if err := s.repo.Update(ctx, l); err != nil {
		return nil, err
	}
	if old != nil {
		s.invalidateLyrics(ctx, old)
	}
	s.invalidateLyrics(ctx, l)
	s.invalidateLyricsList(ctx)
	return l, nil
}
