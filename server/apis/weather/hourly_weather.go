package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
)

func hourlyWeatherUrl(timezone string, latitude, longitude float64, numHours int) string {

	if numHours <= 0 {
		numHours = 0
	}
	if numHours > 72 {
		numHours = 72
	}

	startHour := startOfNextHour(time.Now()).UTC().Format("2006-01-02T15:04")
	endHour := startOfNextHour(time.Now().Add(time.Duration(numHours) * time.Hour)).UTC().Format("2006-01-02T15:04")

	return fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&hourly=temperature_2m,precipitation_probability,precipitation,weather_code&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timeformat=unixtime&timezone=%s&start_hour=%s&end_hour=%s", latitude, longitude, timezone, startHour, endHour)
}

type openMeteoHourlyResponse struct {
	openMeteoResponseBase
	HourlyUnits struct {
		Time                     string `json:"time"`
		Temperature2m            string `json:"temperature_2m"`
		PrecipitationProbability string `json:"precipitation_probability"`
		Precipitation            string `json:"precipitation"`
		WeatherCode              string `json:"weather_code"`
	} `json:"hourly_units"`
	Hourly struct {
		Time                     []int64   `json:"time"`
		Temperature2m            []float64 `json:"temperature_2m"`
		PrecipitationProbability []int     `json:"precipitation_probability"`
		Precipitation            []float64 `json:"precipitation"`
		WeatherCode              []int     `json:"weather_code"`
	} `json:"hourly"`
}

func (r *openMeteoHourlyResponse) getHourlyTemperature() (forecast []timeseriesPoint) {
	for index, value := range r.Hourly.Temperature2m {
		point := timeseriesPoint{
			Unixtime: r.Hourly.Time[index],
			Value:    fmt.Sprintf("%g%s", value, unitsToString(r.HourlyUnits.Temperature2m, value)),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *openMeteoHourlyResponse) getHourlyPrecipitationProbability() (forecast []timeseriesPoint) {
	for index, value := range r.Hourly.PrecipitationProbability {
		point := timeseriesPoint{
			Unixtime: r.Hourly.Time[index],
			Value:    fmt.Sprintf("%d%s", value, unitsToString(r.HourlyUnits.PrecipitationProbability, float64(value))),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *openMeteoHourlyResponse) getHourlyPrecipitation() (forecast []timeseriesPoint) {
	for index, value := range r.Hourly.Precipitation {
		point := timeseriesPoint{
			Unixtime: r.Hourly.Time[index],
			Value:    fmt.Sprintf("%g%s", value, unitsToString(r.HourlyUnits.Precipitation, value)),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *openMeteoHourlyResponse) getHourlyWeatherCode() (forecast []timeseriesPoint) {
	for index, value := range r.Hourly.WeatherCode {
		point := timeseriesPoint{
			Unixtime: r.Hourly.Time[index],
			Value:    WeatherCodeToString(value, r.Hourly.Time[index] == 1),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

type hourStatus struct {
	Unixtime                 int64  `json:"unix_time"`
	TimeStr                  string `json:"time_str"`
	Temperature              string `json:"temperature"`
	PrecipitationProbability string `json:"precipitation_probability"`
	Precipitation            string `json:"precipitation"`
	WeatherCode              string `json:"weather_code"`
}

type hourlyStatus struct {
	Hourly []hourStatus `json:"hours"`
}

func hourlyUnixTimeToString(unixtime int64) string {
	t := time.Unix(unixtime, 0)
	timeStr := strings.ToLower(t.Format("3:04PM"))
	digitsStr := strings.TrimSuffix(timeStr[:len(timeStr)-2], ":00")
	return digitsStr + timeStr[len(timeStr)-2:]
}

func (r *openMeteoHourlyResponse) translateToWeatherStatus() hourlyStatus {
	var status hourlyStatus

	hourlyTemperature := r.getHourlyTemperature()
	hourlyPrecipitationProbability := r.getHourlyPrecipitationProbability()
	hourlyPrecipitation := r.getHourlyPrecipitation()
	hourlyWeatherCode := r.getHourlyWeatherCode()

	for i := range hourlyTemperature {
		status.Hourly = append(status.Hourly, hourStatus{
			Unixtime:                 hourlyTemperature[i].Unixtime,
			TimeStr:                  hourlyUnixTimeToString(hourlyTemperature[i].Unixtime + int64(r.UtcOffsetSeconds)),
			Temperature:              hourlyTemperature[i].Value,
			PrecipitationProbability: hourlyPrecipitationProbability[i].Value,
			Precipitation:            hourlyPrecipitation[i].Value,
			WeatherCode:              hourlyWeatherCode[i].Value,
		})
	}

	return status
}

func HourlyWeatherHandler(app *pocketbase.PocketBase) func(c echo.Context) error {
	return func(c echo.Context) error {
		latitude, longitude, timezone, err := parseLatLongTz(c)

		if err != nil {
			return err
		}

		numHoursRaw := c.QueryParam("numHours")

		numHours := 8
		if numHoursRaw != "" {
			numHours, _ = strconv.Atoi(numHoursRaw)
		}

		url := hourlyWeatherUrl(timezone, latitude, longitude, numHours)

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

		var response openMeteoHourlyResponse
		err = json.Unmarshal([]byte(bodyStr), &response)
		if err != nil {
			return apis.NewApiError(500, "Failed to parse weather data", err)
		}

		return c.JSON(200, response.translateToWeatherStatus())
	}
}
