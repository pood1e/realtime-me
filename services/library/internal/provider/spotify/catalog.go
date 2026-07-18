package spotify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	defaultSearchLimit  = 10
	maximumSearchLimit  = 10
	maximumSearchOffset = 1000
)

type searchResponse struct {
	Tracks struct {
		Items  []searchTrack `json:"items"`
		Limit  int           `json:"limit"`
		Offset int           `json:"offset"`
		Total  int           `json:"total"`
		Next   *string       `json:"next"`
	} `json:"tracks"`
}

type searchTrack struct {
	ID           string `json:"id"`
	URI          string `json:"uri"`
	Name         string `json:"name"`
	DurationMS   int    `json:"duration_ms"`
	Explicit     bool   `json:"explicit"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	ExternalIDs struct {
		ISRC string `json:"isrc"`
	} `json:"external_ids"`
	Artists []struct {
		ID   string `json:"id"`
		URI  string `json:"uri"`
		Name string `json:"name"`
	} `json:"artists"`
	Album struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
	} `json:"album"`
}

func (c *Client) SearchTracks(
	ctx context.Context,
	credentials Credentials,
	search TrackSearchRequest,
) (TrackSearchPage, error) {
	parameters, err := searchParameters(search)
	if err != nil {
		return TrackSearchPage{}, err
	}

	request, err := c.authorizedRequest(
		ctx,
		http.MethodGet,
		webAPIEndpoint+"/search?"+parameters.Encode(),
		credentials,
	)
	if err != nil {
		return TrackSearchPage{}, err
	}

	var response searchResponse
	if err := c.doJSON(request, &response, credentials.AccessToken); err != nil {
		return TrackSearchPage{}, err
	}

	nextOffset, err := searchNextOffset(response.Tracks.Next)
	if err != nil {
		return TrackSearchPage{}, err
	}
	items := make([]Track, 0, len(response.Tracks.Items))
	for _, item := range response.Tracks.Items {
		items = append(items, mapSearchTrack(item))
	}

	return TrackSearchPage{
		Items:      items,
		Limit:      response.Tracks.Limit,
		Offset:     response.Tracks.Offset,
		Total:      response.Tracks.Total,
		NextOffset: nextOffset,
	}, nil
}

func searchParameters(search TrackSearchRequest) (url.Values, error) {
	query := strings.TrimSpace(search.Query)
	if query == "" {
		return nil, errors.New("Spotify track search query is required")
	}

	limit := search.Limit
	if limit == 0 {
		limit = defaultSearchLimit
	}
	if limit < 1 || limit > maximumSearchLimit {
		return nil, fmt.Errorf("Spotify track search limit must be between 1 and %d", maximumSearchLimit)
	}
	if search.Offset < 0 || search.Offset > maximumSearchOffset {
		return nil, fmt.Errorf("Spotify track search offset must be between 0 and %d", maximumSearchOffset)
	}

	parameters := url.Values{
		"q":      {query},
		"type":   {"track"},
		"limit":  {strconv.Itoa(limit)},
		"offset": {strconv.Itoa(search.Offset)},
	}
	if search.Market != "" {
		market := strings.ToUpper(strings.TrimSpace(search.Market))
		if len(market) != 2 || market[0] < 'A' || market[0] > 'Z' || market[1] < 'A' || market[1] > 'Z' {
			return nil, errors.New("Spotify track search market must be a two-letter country code")
		}
		parameters.Set("market", market)
	}
	return parameters, nil
}

func searchNextOffset(next *string) (*int, error) {
	if next == nil || *next == "" {
		return nil, nil
	}
	parsed, err := url.Parse(*next)
	if err != nil {
		return nil, errors.New("Spotify search response contains an invalid next page URL")
	}
	nextOffset, err := strconv.Atoi(parsed.Query().Get("offset"))
	if err != nil || nextOffset < 0 || nextOffset > maximumSearchOffset {
		return nil, errors.New("Spotify search response contains an invalid next page offset")
	}
	return &nextOffset, nil
}

func mapSearchTrack(source searchTrack) Track {
	artists := make([]Artist, 0, len(source.Artists))
	for _, artist := range source.Artists {
		artists = append(artists, Artist{
			ID:   artist.ID,
			URI:  artist.URI,
			Name: artist.Name,
		})
	}

	artworkURL := ""
	for _, image := range source.Album.Images {
		if image.URL != "" {
			artworkURL = image.URL
			break
		}
	}

	return Track{
		ID:                   source.ID,
		URI:                  source.URI,
		SpotifyURL:           source.ExternalURLs.Spotify,
		Name:                 source.Name,
		Artists:              artists,
		AlbumID:              source.Album.ID,
		AlbumName:            source.Album.Name,
		ArtworkURL:           artworkURL,
		ISRC:                 source.ExternalIDs.ISRC,
		DurationMilliseconds: source.DurationMS,
		Explicit:             source.Explicit,
	}
}
