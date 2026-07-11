package spotify

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	playlistItemsLimit    = 50
	maximumPlaylistTracks = 5000
)

// Playlist is one Spotify playlist snapshot.
type Playlist struct {
	ID         string
	Name       string
	ArtworkURL string
	SpotifyURL string
	Tracks     []Track
}

// GetPlaylist retrieves provider metadata and every track available to the account.
func (c *Client) GetPlaylist(ctx context.Context, credentials Credentials, playlistID string) (Playlist, error) {
	playlistID = strings.TrimSpace(playlistID)
	if !validSpotifyID(playlistID) {
		return Playlist{}, errors.New("Spotify playlist ID is invalid")
	}
	metadataRequest, err := c.authorizedRequest(ctx, http.MethodGet,
		webAPIEndpoint+"/playlists/"+url.PathEscape(playlistID)+"?fields=id,name,images,external_urls", credentials)
	if err != nil {
		return Playlist{}, err
	}
	var metadata struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		ExternalURLs struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
	}
	if err := c.doJSON(metadataRequest, &metadata, credentials.AccessToken); err != nil {
		return Playlist{}, err
	}
	if metadata.ID != playlistID || strings.TrimSpace(metadata.Name) == "" {
		return Playlist{}, errors.New("Spotify playlist response is incomplete")
	}
	playlist := Playlist{ID: metadata.ID, Name: strings.TrimSpace(metadata.Name), SpotifyURL: metadata.ExternalURLs.Spotify}
	if len(metadata.Images) > 0 {
		playlist.ArtworkURL = metadata.Images[0].URL
	}
	for offset := 0; offset < maximumPlaylistTracks; offset += playlistItemsLimit {
		page, err := c.getPlaylistItems(ctx, credentials, playlistID, offset)
		if err != nil {
			return Playlist{}, err
		}
		playlist.Tracks = append(playlist.Tracks, page.Tracks...)
		if !page.HasMore {
			return playlist, nil
		}
	}
	return Playlist{}, errors.New("Spotify playlist exceeds the supported track limit")
}

type playlistItemsPage struct {
	Tracks  []Track
	HasMore bool
}

func (c *Client) getPlaylistItems(ctx context.Context, credentials Credentials, playlistID string, offset int) (playlistItemsPage, error) {
	parameters := url.Values{
		"limit": {strconv.Itoa(playlistItemsLimit)}, "offset": {strconv.Itoa(offset)}, "additional_types": {"track"},
	}
	request, err := c.authorizedRequest(ctx, http.MethodGet,
		webAPIEndpoint+"/playlists/"+url.PathEscape(playlistID)+"/items?"+parameters.Encode(), credentials)
	if err != nil {
		return playlistItemsPage{}, err
	}
	var response struct {
		Items []struct {
			Track *searchTrack `json:"track"`
			Item  *searchTrack `json:"item"`
		} `json:"items"`
		Next  *string `json:"next"`
		Total int     `json:"total"`
	}
	if err := c.doJSON(request, &response, credentials.AccessToken); err != nil {
		return playlistItemsPage{}, err
	}
	if response.Total < 0 || response.Total > maximumPlaylistTracks || len(response.Items) > playlistItemsLimit {
		return playlistItemsPage{}, errors.New("Spotify playlist items response is invalid")
	}
	page := playlistItemsPage{HasMore: response.Next != nil}
	for _, entry := range response.Items {
		item := entry.Track
		if item == nil {
			item = entry.Item
		}
		if item == nil || !validSpotifyID(item.ID) || strings.TrimSpace(item.Name) == "" {
			continue
		}
		page.Tracks = append(page.Tracks, mapSearchTrack(*item))
	}
	return page, nil
}

func validSpotifyID(value string) bool {
	if len(value) < 1 || len(value) > 64 {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') {
			continue
		}
		return false
	}
	return true
}
