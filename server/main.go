package main

import (
	"log"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	keyboard_apis "keyboard-api/apis"
	"keyboard-api/apis/weather"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	app := pocketbase.New()

	// loosely check if it was executed using "go run"
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Admin UI
		// (the isGoRun check is to enable it only during development)
		Automigrate: isGoRun,
	})

	// serves static files from the provided public dir (if exists)
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), false))

		e.Router.GET("/spotify/loginUrl", keyboard_apis.SpotifyLoginUrlHandler)
		e.Router.GET("/spotify/callback", keyboard_apis.SpotifyCallbackHandler(app))
		e.Router.GET("/spotify/currently-playing", keyboard_apis.SpotifyCurrentlyPlayingHandler(app))
		e.Router.GET("/spotify/currently-playing-art", keyboard_apis.SpotifyCurrentlyPlayingArtHandler(app))

		e.Router.GET("/weather/current", weather.CurrentWeatherHandler(app))

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
