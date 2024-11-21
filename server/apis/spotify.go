package apis

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"keyboard-api/images"
	"keyboard-api/utils"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/types"
)

const (
	SCOPES                          = "user-read-playback-state user-read-currently-playing user-modify-playback-state user-library-read"
	SPOTIFY_TOKEN_REFRESH_BUFFER_MS = 5000
)

type OauthState struct {
	UserId string `json:"user_id"`
}

type SpotifyCredentials struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	ExpiresAt    int64  `json:"expires_at"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

func SpotifyLoginUrlHandler(c echo.Context) error {
	record, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)

	if record == nil {
		return apis.NewForbiddenError("You must be logged in", nil)
	}

	state := OauthState{
		UserId: record.Id,
	}

	stateStr, _ := json.Marshal(state)

	url := fmt.Sprintf("https://accounts.spotify.com/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=%s&state=%s", url.QueryEscape(os.Getenv("SPOTIFY_CLIENT_ID")), url.QueryEscape(utils.ServerURL("/spotify/callback")), url.QueryEscape(SCOPES), url.QueryEscape(string(stateStr)))

	return c.JSON(200, struct {
		Url string `json:"url"`
	}{
		Url: url,
	})
}

func SpotifyCallbackHandler(app *pocketbase.PocketBase) func(c echo.Context) error {
	return func(c echo.Context) error {
		code := c.QueryParam("code")
		state := c.QueryParam("state")
		errQuery := c.QueryParam("error")

		if errQuery != "" {
			return apis.NewBadRequestError("Error during authentication", nil)
		}

		oauthState := OauthState{}
		err := json.Unmarshal([]byte(state), &oauthState)
		if err != nil {
			return apis.NewBadRequestError("Invalid state", nil)
		}

		user, err := app.Dao().FindRecordById("users", oauthState.UserId)
		if err != nil {
			return apis.NewBadRequestError("Could not find user", nil)
		}

		payload := url.Values{}
		payload.Set("grant_type", "authorization_code")
		payload.Set("code", code)
		payload.Set("redirect_uri", utils.ServerURL("/spotify/callback"))

		req, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(payload.Encode()))

		authorizationToken := base64.StdEncoding.EncodeToString([]byte(os.Getenv("SPOTIFY_CLIENT_ID") + ":" + os.Getenv("SPOTIFY_CLIENT_SECRET")))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", authorizationToken))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		bodyStr := string(body)

		_, err = saveSpotifyCredentials(app, user, bodyStr)

		if err != nil {
			return apis.NewBadRequestError("Failed to exchange code for token", nil)
		}

		return c.Redirect(307, utils.ServerURL("/spotify/callback/success"))
	}
}

func saveSpotifyCredentials(app *pocketbase.PocketBase, user *models.Record, credsStr string) (accessToken string, err error) {
	token := struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		Scope        string `json:"scope"`
		RefreshToken string `json:"refresh_token"`
	}{}
	err = json.Unmarshal([]byte(credsStr), &token)
	if err != nil {
		return "", errors.New("invalid response")
	}

	creds := SpotifyCredentials{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		Scope:        token.Scope,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
		ExpiresAt:    time.Now().UnixMilli() + token.ExpiresIn,
	}

	if creds.RefreshToken == "" {
		var existingCreds SpotifyCredentials
		existingCredsRaw := user.Get("spotify").(types.JsonRaw)
		err = json.Unmarshal(existingCredsRaw, &existingCreds)

		if err != nil {
			return "", err
		}
		creds.RefreshToken = existingCreds.RefreshToken
	}

	user.Set("spotify", creds)
	app.Dao().SaveRecord(user)

	return creds.AccessToken, nil
}

func refreshSpotifyToken(app *pocketbase.PocketBase, userId string) (token string, err error) {

	user, err := app.Dao().FindRecordById("users", userId)
	if err != nil {
		return "", errors.New("could not find user")
	}

	creds := user.Get("spotify").(types.JsonRaw)
	if creds == nil {
		return "", errors.New("could not find spotify credentials")
	}

	var spotify SpotifyCredentials
	err = json.Unmarshal(creds, &spotify)
	if err != nil {
		return "", err
	}

	if spotify.ExpiresAt > time.Now().UnixMilli()-SPOTIFY_TOKEN_REFRESH_BUFFER_MS {
		return spotify.AccessToken, nil
	}

	payload := url.Values{}
	payload.Set("grant_type", "refresh_token")
	payload.Set("refresh_token", spotify.RefreshToken)

	req, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(payload.Encode()))

	authorizationToken := base64.StdEncoding.EncodeToString([]byte(os.Getenv("SPOTIFY_CLIENT_ID") + ":" + os.Getenv("SPOTIFY_CLIENT_SECRET")))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", authorizationToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyStr := string(body)
	return saveSpotifyCredentials(app, user, bodyStr)

}

func getSpotifyToken(app *pocketbase.PocketBase, user *models.Record) (token string, err error) {
	creds := user.Get("spotify").(types.JsonRaw)
	if creds == nil {
		return "", errors.New("could not find spotify credentials")
	}

	var spotify SpotifyCredentials
	err = json.Unmarshal(creds, &spotify)

	if err != nil {
		return "", err
	}

	if spotify.ExpiresAt < time.Now().UnixMilli()-SPOTIFY_TOKEN_REFRESH_BUFFER_MS || spotify.AccessToken == "" {
		return refreshSpotifyToken(app, user.Id)
	}

	return spotify.AccessToken, nil
}

type SpotifyAlbumImagesResponse struct {
	Url    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type RawSpotifyCurrentlyPlayingResponse struct {
	IsPlaying bool `json:"is_playing"`
	Track     struct {
		Id         string `json:"id"`
		Name       string `json:"name"`
		Popularity int    `json:"popularity"`
		DurationMs int    `json:"duration_ms"`

		Album struct {
			Id     string                       `json:"id"`
			Name   string                       `json:"name"`
			Images []SpotifyAlbumImagesResponse `json:"images"`
		}

		Artists []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		}
	} `json:"item"`
	CurrentlyPlayingType string `json:"currently_playing_type"`
	ProgressMs           int    `json:"progress_ms"`
}

type SpotifyCurrentlyPlaying struct {
	IsPlaying bool `json:"is_playing"`

	TrackId    string `json:"track_id"`
	TrackName  string `json:"track_name"`
	Popularity int    `json:"popularity"`

	TrackLengthMs   int `json:"track_length_ms"`
	TrackProgressMs int `json:"track_progress_ms"`

	AlbumId     string `json:"album_id"`
	AlbumName   string `json:"album_name"`
	AlbumArtUrl string `json:"album_art_url"`

	Artists []struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"artists"`
}

func getBestFitSpotifyAlbumArtUrl(imgs []SpotifyAlbumImagesResponse, thumbnailWidth, thumbnailHeight int) string {
	if thumbnailWidth <= 0 || thumbnailHeight <= 0 {
		return ""
	}
	if len(imgs) == 0 {
		return ""
	}

	var bestFitImage SpotifyAlbumImagesResponse = imgs[0]

	// find the smallest image that is larger than the thumbnail
	for _, img := range imgs {
		fmt.Printf("checking image: %+v\n", img)
		if img.Width >= thumbnailWidth && img.Height >= thumbnailHeight {
			if bestFitImage.Width < img.Width && bestFitImage.Height < img.Height {
				continue
			}
			bestFitImage = img
			break
		}
	}
	return utils.ServerURL(fmt.Sprintf("/spotify/currently-playing-art?url=%s&thumbnailWidth=%d&thumbnailHeight=%d", url.QueryEscape(bestFitImage.Url), thumbnailWidth, thumbnailHeight))
}

func loadSpotifyAlbumArt(url string, thumbnailWidth, thumbnailHeight int) (albumArtBmp []byte, err error) {
	imgResponse, err := http.Get(url)

	if err != nil {
		return albumArtBmp, err
	}
	defer imgResponse.Body.Close()

	contentType := imgResponse.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return albumArtBmp, errors.New("album art is not an image")
	}
	img, _, err := image.Decode(imgResponse.Body)
	if err != nil {
		return albumArtBmp, err
	}

	var b bytes.Buffer
	writer := io.Writer(&b)
	images.ToBitmap(img, thumbnailWidth, thumbnailHeight, &writer)

	return b.Bytes(), nil
}

func SpotifyCurrentlyPlayingHandler(app *pocketbase.PocketBase) func(c echo.Context) error {
	return func(c echo.Context) error {

		record, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)

		if record == nil {
			return apis.NewForbiddenError("You must be logged in", nil)
		}
		token, err := getSpotifyToken(app, record)
		if err != nil {
			return apis.NewBadRequestError("Could not get spotify token", nil)
		}

		req, _ := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/currently-playing", nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		bodyStr := string(body)

		var response RawSpotifyCurrentlyPlayingResponse
		err = json.Unmarshal([]byte(bodyStr), &response)
		if err != nil {
			return apis.NewBadRequestError("Could not parse response", nil)
		}

		if response.CurrentlyPlayingType != "track" {
			return c.JSON(200, SpotifyCurrentlyPlaying{
				IsPlaying:       response.IsPlaying,
				TrackId:         "",
				TrackName:       "",
				Popularity:      0,
				TrackLengthMs:   0,
				TrackProgressMs: 0,
				AlbumId:         "",
				AlbumName:       "",
				AlbumArtUrl:     "",
				Artists:         nil,
			})
		}

		thumbnailWidthRaw := c.QueryParam("thumbnailWidth")
		thumbnailHeightRaw := c.QueryParam("thumbnailHeight")

		thumbnailWidth := 0
		thumbnailHeight := 0
		if thumbnailWidthRaw != "" {
			thumbnailWidth, _ = strconv.Atoi(thumbnailWidthRaw)
		}
		if thumbnailHeightRaw != "" {
			thumbnailHeight, _ = strconv.Atoi(thumbnailHeightRaw)
		}

		albumArtUrl := getBestFitSpotifyAlbumArtUrl(response.Track.Album.Images, thumbnailWidth, thumbnailHeight)

		currentlyPlaying := SpotifyCurrentlyPlaying{
			IsPlaying: response.IsPlaying,

			TrackId:         response.Track.Id,
			TrackName:       response.Track.Name,
			Popularity:      response.Track.Popularity,
			TrackLengthMs:   response.Track.DurationMs,
			TrackProgressMs: response.ProgressMs,

			AlbumId:   response.Track.Album.Id,
			AlbumName: response.Track.Album.Name,

			AlbumArtUrl: albumArtUrl,

			Artists: response.Track.Artists,
		}

		return c.JSON(200, currentlyPlaying)
	}
}

func SpotifyCurrentlyPlayingArtHandler(app *pocketbase.PocketBase) func(c echo.Context) error {
	return func(c echo.Context) error {

		record, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)

		if record == nil {
			return apis.NewForbiddenError("You must be logged in", nil)
		}
		url := c.QueryParam("url")

		if url == "" {
			return apis.NewBadRequestError("url is required", nil)
		}

		thumbnailWidthRaw := c.QueryParam("thumbnailWidth")
		thumbnailHeightRaw := c.QueryParam("thumbnailHeight")

		thumbnailWidth := 0
		thumbnailHeight := 0
		if thumbnailWidthRaw != "" {
			thumbnailWidth, _ = strconv.Atoi(thumbnailWidthRaw)
		}
		if thumbnailHeightRaw != "" {
			thumbnailHeight, _ = strconv.Atoi(thumbnailHeightRaw)
		}

		if thumbnailWidth <= 0 || thumbnailHeight <= 0 {
			return apis.NewBadRequestError("thumbnailWidth and thumbnailHeight must be greater than 0", nil)
		}
		if thumbnailWidth > 320 || thumbnailHeight > 320 {
			return apis.NewBadRequestError("thumbnailWidth and thumbnailHeight must be less than 320", nil)
		}

		albumArtBmp, thumbnailErr := loadSpotifyAlbumArt(url, thumbnailWidth, thumbnailHeight)

		if thumbnailErr != nil {
			fmt.Println("error loading album art")
			fmt.Println(thumbnailErr)
		}

		return c.Blob(200, "image/bmp", albumArtBmp)
	}
}
