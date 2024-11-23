package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
)

func currentWeatherUrl(timezone string, latitude, longitude float64) string {
	return fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m,apparent_temperature,is_day,precipitation,rain,showers,snowfall,weather_code,cloud_cover,surface_pressure,wind_speed_10m,wind_direction_10m&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timeformat=unixtime&timezone=%s", latitude, longitude, timezone)
}

type openMeteoCurrentResponse struct {
	openMeteoResponseBase
	CurrentUnits struct {
		Time                string `json:"time"`
		Interval            string `json:"interval"`
		Temperature2m       string `json:"temperature_2m"`
		RelativeHumidity2m  string `json:"relative_humidity_2m"`
		ApparentTemperature string `json:"apparent_temperature"`
		IsDay               string `json:"is_day"`
		Precipitation       string `json:"precipitation"`
		Rain                string `json:"rain"`
		Showers             string `json:"showers"`
		Snowfall            string `json:"snowfall"`
		WeatherCode         string `json:"weather_code"`
		CloudCover          string `json:"cloud_cover"`
		SurfacePressure     string `json:"surface_pressure"`
		WindSpeed10m        string `json:"wind_speed_10m"`
		WindDirection10m    string `json:"wind_direction_10m"`
	} `json:"current_units"`
	Current struct {
		Time                int64   `json:"time"`
		Interval            int     `json:"interval"`
		Temperature2m       float64 `json:"temperature_2m"`
		RelativeHumidity2m  int     `json:"relative_humidity_2m"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		IsDay               int     `json:"is_day"`
		Precipitation       float64 `json:"precipitation"`
		Rain                float64 `json:"rain"`
		Showers             float64 `json:"showers"`
		Snowfall            float64 `json:"snowfall"`
		WeatherCode         int     `json:"weather_code"`
		CloudCover          int     `json:"cloud_cover"`
		SurfacePressure     float64 `json:"surface_pressure"`
		WindSpeed10m        float64 `json:"wind_speed_10m"`
		WindDirection10m    int     `json:"wind_direction_10m"`
	} `json:"current"`
}

func (r *openMeteoCurrentResponse) getCurrentWeatherCode() string {
	return WeatherCodeToString(r.Current.WeatherCode, r.Current.IsDay == 1)
}

func (r *openMeteoCurrentResponse) getCurrentTemperature() string {
	return fmt.Sprintf("%g%s", r.Current.Temperature2m, unitsToString(r.CurrentUnits.Temperature2m, r.Current.Temperature2m))
}

func (r *openMeteoCurrentResponse) getCurrentRelativeHumidity() string {
	return fmt.Sprintf("%d%s", r.Current.RelativeHumidity2m, unitsToString(r.CurrentUnits.RelativeHumidity2m, float64(r.Current.RelativeHumidity2m)))
}

func (r *openMeteoCurrentResponse) getCurrentApparentTemperature() string {
	return fmt.Sprintf("%g%s", r.Current.ApparentTemperature, unitsToString(r.CurrentUnits.ApparentTemperature, r.Current.ApparentTemperature))
}

func (r *openMeteoCurrentResponse) getCurrentDayStatus() bool {
	return r.Current.IsDay == 1
}

func (r *openMeteoCurrentResponse) getCurrentPrecipitation() string {
	return fmt.Sprintf("%g%s", r.Current.Precipitation, unitsToString(r.CurrentUnits.Precipitation, r.Current.Precipitation))
}

func (r *openMeteoCurrentResponse) getCurrentRain() string {
	return fmt.Sprintf("%g%s", r.Current.Rain, unitsToString(r.CurrentUnits.Rain, r.Current.Rain))
}

func (r *openMeteoCurrentResponse) getCurrentShowers() string {
	return fmt.Sprintf("%g%s", r.Current.Showers, unitsToString(r.CurrentUnits.Showers, r.Current.Showers))
}

func (r *openMeteoCurrentResponse) getCurrentSnowfall() string {
	return fmt.Sprintf("%g%s", r.Current.Snowfall, unitsToString(r.CurrentUnits.Snowfall, r.Current.Snowfall))
}

func (r *openMeteoCurrentResponse) getCurrentCloudCover() string {
	return fmt.Sprintf("%d%s", r.Current.CloudCover, unitsToString(r.CurrentUnits.CloudCover, float64(r.Current.CloudCover)))
}

func (r *openMeteoCurrentResponse) getCurrentSurfacePressure() string {
	return fmt.Sprintf("%g%s", r.Current.SurfacePressure, unitsToString(r.CurrentUnits.SurfacePressure, r.Current.SurfacePressure))
}

func (r *openMeteoCurrentResponse) getCurrentWindSpeed() string {
	return fmt.Sprintf("%g%s", r.Current.WindSpeed10m, unitsToString(r.CurrentUnits.WindSpeed10m, r.Current.WindSpeed10m))
}

func (r *openMeteoCurrentResponse) getCurrentWindDirection() string {
	return fmt.Sprintf("%d%s", r.Current.WindDirection10m, unitsToString(r.CurrentUnits.WindDirection10m, float64(r.Current.WindDirection10m)))
}

type currentStatus struct {
	WeatherCode            string `json:"weather_code"`
	Temperature            string `json:"temperature"`
	RelativeHumidity       string `json:"relative_humidity"`
	ApparentTemperature    string `json:"apparent_temperature"`
	IsDay                  bool   `json:"is_day"`
	CurrentPrecipitation   string `json:"current_precipitation"`
	CurrentRain            string `json:"current_rain"`
	CurrentShowers         string `json:"current_showers"`
	CurrentSnowfall        string `json:"current_snowfall"`
	CurrentCloudCover      string `json:"current_cloud_cover"`
	CurrentSurfacePressure string `json:"current_surface_pressure"`
	CurrentWindSpeed       string `json:"current_wind_speed"`
	CurrentWindDirection   string `json:"current_wind_direction"`
}

func (r *openMeteoCurrentResponse) translateToWeatherStatus() currentStatus {
	var status currentStatus

	status.WeatherCode = r.getCurrentWeatherCode()
	status.Temperature = r.getCurrentTemperature()
	status.RelativeHumidity = r.getCurrentRelativeHumidity()
	status.ApparentTemperature = r.getCurrentApparentTemperature()
	status.IsDay = r.getCurrentDayStatus()
	status.CurrentPrecipitation = r.getCurrentPrecipitation()
	status.CurrentRain = r.getCurrentRain()
	status.CurrentShowers = r.getCurrentShowers()
	status.CurrentSnowfall = r.getCurrentSnowfall()
	status.CurrentCloudCover = r.getCurrentCloudCover()
	status.CurrentSurfacePressure = r.getCurrentSurfacePressure()
	status.CurrentWindSpeed = r.getCurrentWindSpeed()
	status.CurrentWindDirection = r.getCurrentWindDirection()

	return status
}

func CurrentWeatherHandler(app *pocketbase.PocketBase) func(c echo.Context) error {
	return func(c echo.Context) error {
		latitude, longitude, timezone, err := parseLatLongTz(c)

		if err != nil {
			return err
		}

		url := currentWeatherUrl(timezone, latitude, longitude)

		resp, err := http.Get(url)
		if err != nil {
			return apis.NewApiError(500, "Failed to fetch weather data", err)
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return apis.NewApiError(500, "Failed to read weather data", err)
		}

		bodyStr := string(body)

		var response openMeteoCurrentResponse
		err = json.Unmarshal([]byte(bodyStr), &response)
		if err != nil {
			return apis.NewApiError(500, "Failed to parse weather data", err)
		}

		return c.JSON(200, response.translateToWeatherStatus())
	}
}
