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

func dailyWeatherUrl(timezone string, latitude, longitude float64, numDays int) string {
	if numDays <= 0 {
		numDays = 0
	}
	if numDays > 7 {
		numDays = 7
	}

	startDate := time.Now().Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, numDays).Format("2006-01-02")

	return fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&hourly=temperature_2m,precipitation_probability,precipitation,weather_code&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timeformat=unixtime&timezone=%s&start_date=%s&end_date=%s", latitude, longitude, timezone, startDate, endDate)
}

type openMeteoDailyResponse struct {
	openMeteoResponseBase
	DailyUnits struct {
		Time             string `json:"time"`
		WeatherCode      string `json:"weather_code"`
		Temperature2mMax string `json:"temperature_2m_max"`
		Temperature2mMin string `json:"temperature_2m_min"`
		Sunrise          string `json:"sunrise"`
		Sunset           string `json:"sunset"`
		DaylightDuration string `json:"daylight_duration"`
		UvIndexMax       string `json:"uv_index_max"`
		PrecipitationSum string `json:"precipitation_sum"`
	} `json:"daily_units"`
	Daily struct {
		Time             []int64   `json:"time"`
		WeatherCode      []int     `json:"weather_code"`
		Temperature2mMax []float64 `json:"temperature_2m_max"`
		Temperature2mMin []float64 `json:"temperature_2m_min"`
		Sunrise          []int64   `json:"sunrise"`
		Sunset           []int64   `json:"sunset"`
		DaylightDuration []float64 `json:"daylight_duration"`
		UvIndexMax       []float64 `json:"uv_index_max"`
		PrecipitationSum []float64 `json:"precipitation_sum"`
	} `json:"daily"`
}

func (r *openMeteoDailyResponse) getDailyWeatherCode() (forecast []timeseriesPoint) {
	for index, value := range r.Daily.WeatherCode {
		point := timeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    WeatherCodeToString(value, r.Daily.Time[index] == 1),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *openMeteoDailyResponse) getDailyTemperatureMax() (forecast []timeseriesPoint) {
	for index, value := range r.Daily.Temperature2mMax {
		point := timeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    fmt.Sprintf("%g%s", value, unitsToString(r.DailyUnits.Temperature2mMax, value)),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *openMeteoDailyResponse) getDailyTemperatureMin() (forecast []timeseriesPoint) {
	for index, value := range r.Daily.Temperature2mMin {
		point := timeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    fmt.Sprintf("%g%s", value, unitsToString(r.DailyUnits.Temperature2mMin, value)),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *openMeteoDailyResponse) getDailySunrise() (forecast []timeseriesPoint) {
	for index, value := range r.Daily.Sunrise {
		point := timeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    hourlyUnixTimeToString(value + int64(r.UtcOffsetSeconds)),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *openMeteoDailyResponse) getDailySunset() (forecast []timeseriesPoint) {
	for index, value := range r.Daily.Sunset {
		point := timeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    hourlyUnixTimeToString(value + int64(r.UtcOffsetSeconds)),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func shortDur(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}

func (r *openMeteoDailyResponse) getDailyDaylightDuration() (forecast []timeseriesPoint) {
	for index, value := range r.Daily.DaylightDuration {
		point := timeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    shortDur(time.Duration(value) * time.Second),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *openMeteoDailyResponse) getDailyUvIndexMax() (forecast []timeseriesPoint) {
	for index, value := range r.Daily.UvIndexMax {
		point := timeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    fmt.Sprintf("%.2f", value),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *openMeteoDailyResponse) getDailyPrecipitationSum() (forecast []timeseriesPoint) {
	for index, value := range r.Daily.PrecipitationSum {
		point := timeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    fmt.Sprintf("%g%s", value, unitsToString(r.DailyUnits.PrecipitationSum, value)),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

type dayStatus struct {
	Unixtime         int64  `json:"unix_time"`
	TimeStr          string `json:"time_str"`
	WeatherCode      string `json:"weather_code"`
	TemperatureMax   string `json:"temperature_max"`
	TemperatureMin   string `json:"temperature_min"`
	Sunrise          string `json:"sunrise"`
	Sunset           string `json:"sunset"`
	DaylightDuration string `json:"daylight_duration"`
	UvIndexMax       string `json:"uv_index_max"`
	PrecipitationSum string `json:"precipitation_sum"`
}

type dailyStatus struct {
	Daily []dayStatus `json:"days"`
}

func dailyUnixTimeToString(unixtime int64) string {
	t := time.Unix(unixtime, 0)
	return t.Format("Mon")
}

func (r *openMeteoDailyResponse) translateToWeatherStatus() dailyStatus {
	var status dailyStatus

	dailyWeatherCode := r.getDailyWeatherCode()
	dailyTemperatureMax := r.getDailyTemperatureMax()
	dailyTemperatureMin := r.getDailyTemperatureMin()
	dailySunrise := r.getDailySunrise()
	dailySunset := r.getDailySunset()
	dailyDaylightDuration := r.getDailyDaylightDuration()
	dailyUvIndexMax := r.getDailyUvIndexMax()
	dailyPrecipitationSum := r.getDailyPrecipitationSum()

	for i := range dailyWeatherCode {
		status.Daily = append(status.Daily, dayStatus{
			Unixtime:         dailyWeatherCode[i].Unixtime,
			TimeStr:          dailyUnixTimeToString(dailyWeatherCode[i].Unixtime + int64(r.UtcOffsetSeconds)),
			WeatherCode:      dailyWeatherCode[i].Value,
			TemperatureMax:   dailyTemperatureMax[i].Value,
			TemperatureMin:   dailyTemperatureMin[i].Value,
			Sunrise:          dailySunrise[i].Value,
			Sunset:           dailySunset[i].Value,
			DaylightDuration: dailyDaylightDuration[i].Value,
			UvIndexMax:       dailyUvIndexMax[i].Value,
			PrecipitationSum: dailyPrecipitationSum[i].Value,
		})
	}

	return status
}

func DailyWeatherHandler(app *pocketbase.PocketBase) func(c echo.Context) error {
	return func(c echo.Context) error {
		latitude, longitude, timezone, err := parseLatLongTz(c)

		if err != nil {
			return err
		}

		numDaysRaw := c.QueryParam("numDays")

		numDays := 5

		if numDaysRaw != "" {
			numDays, _ = strconv.Atoi(numDaysRaw)
		}

		url := dailyWeatherUrl(timezone, latitude, longitude, numDays)

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

		var response openMeteoDailyResponse
		err = json.Unmarshal([]byte(bodyStr), &response)
		if err != nil {
			return apis.NewApiError(500, "Failed to parse weather data", err)
		}

		return c.JSON(200, response.translateToWeatherStatus())
	}
}
