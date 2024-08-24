package processor

import (
	"context"
	"strings"
	"time"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/pkg/autobrr"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"golift.io/starr/lidarr"
)

func (s *Service) lidarr(ctx context.Context, cfg *domain.ArrConfig, dryRun bool, brr *autobrr.Client) error {
	l := log.With().Str("type", "lidarr").Str("client", cfg.Name).Logger()

	l.Debug().Msgf("gathering titles...")

	titles, artists, err := s.processLidarr(ctx, cfg, &l)
	if err != nil {
		return err
	}

	l.Debug().Msgf("got %v filter titles", len(titles))

	// Process titles
	var processedTitles []string
	for _, title := range titles {
		processedTitles = append(processedTitles, processTitle(title, cfg.MatchRelease)...)
	}

	// Update filter based on MatchRelease
	var f autobrr.UpdateFilter
	if cfg.MatchRelease {
		joinedTitles := strings.Join(processedTitles, ",")
		if len(joinedTitles) == 0 {
			return nil
		}
		f = autobrr.UpdateFilter{MatchReleases: joinedTitles}
	} else {
		// Process artists only if MatchRelease is false
		var processedArtists []string
		for _, artist := range artists {
			processedArtists = append(processedArtists, processTitle(artist, cfg.MatchRelease)...)
		}

		joinedTitles := strings.Join(processedTitles, ",")
		joinedArtists := strings.Join(processedArtists, ",")
		if len(joinedTitles) == 0 && len(joinedArtists) == 0 {
			return nil
		}
		f = autobrr.UpdateFilter{Albums: joinedTitles, Artists: joinedArtists}
	}

	for _, filterID := range cfg.Filters {
		l.Debug().Msgf("updating filter: %v", filterID)

		if !dryRun {
			if err := brr.UpdateFilterByID(ctx, filterID, f); err != nil {
				l.Error().Err(err).Msgf("error updating filter: %v", filterID)
				return errors.Wrapf(err, "error updating filter: %v", filterID)
			}
		}

		l.Debug().Msgf("successfully updated filter: %v", filterID)
	}

	return nil
}

func (s *Service) processLidarr(ctx context.Context, cfg *domain.ArrConfig, logger *zerolog.Logger) ([]string, []string, error) {
	c := starr.New(cfg.Apikey, cfg.Host, 60*time.Second)

	if cfg.BasicAuth != nil {
		if cfg.BasicAuth.User != "" {
			c.HTTPUser = cfg.BasicAuth.User
		}
		if cfg.BasicAuth.Pass != "" {
			c.HTTPPass = cfg.BasicAuth.Pass
		}
	}

	r := lidarr.New(c)

	albums, err := r.GetAlbumContext(ctx, "")
	if err != nil {
		return nil, nil, err
	}

	var titles []string
	var artists []string
	seenArtists := make(map[string]struct{})

	for _, album := range albums {
		if !album.Monitored {
			continue // Skip unmonitored albums
		}

		// Fetch the artist details
		artist, err := r.GetArtistByIDContext(ctx, album.ArtistID)
		if err != nil {
			logger.Error().Err(err).Msgf("Error fetching artist details for album: %v", album.Title)
			continue // Skip this album if there's an error fetching the artist
		}

		if artist.Monitored {
			processedTitles := processTitle(album.Title, cfg.MatchRelease)
			titles = append(titles, processedTitles...)

			// Debug logging
			logger.Debug().Msgf("Processing artist: %s", artist.ArtistName)

			if _, exists := seenArtists[artist.ArtistName]; !exists {
				artists = append(artists, artist.ArtistName)
				seenArtists[artist.ArtistName] = struct{}{}
				logger.Debug().Msgf("Added artist: %s", artist.ArtistName) // Log when an artist is added
			}
		}
	}

	logger.Debug().Msgf("Processed %d monitored albums with monitored artists, created %d titles, found %d unique artists", len(titles), len(titles), len(artists))

	return titles, artists, nil
}
