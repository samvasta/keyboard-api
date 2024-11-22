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

func startOfNextHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), (t.Hour()+1)%24, 0, 0, 0, t.Location())
}

func RequestUrl(timezone string, latitude, longitude float64, numHours, numDays int) string {

	if numHours <= 0 {
		numHours = 0
	}
	if numHours > 72 {
		numHours = 72
	}

	if numDays <= 0 {
		numDays = 0
	}
	if numDays > 7 {
		numDays = 7
	}

	startDate := time.Now().Format("2006-01-02")
	startHour := startOfNextHour(time.Now()).UTC().Format("2006-01-02T15:04")
	endDate := time.Now().AddDate(0, 0, numDays).Format("2006-01-02")
	endHour := startOfNextHour(time.Now().Add(time.Duration(numHours) * time.Hour)).UTC().Format("2006-01-02T15:04")

	return fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m,apparent_temperature,is_day,precipitation,rain,showers,snowfall,weather_code,cloud_cover,surface_pressure,wind_speed_10m,wind_direction_10m&hourly=temperature_2m,precipitation_probability,precipitation,weather_code&daily=weather_code,temperature_2m_max,temperature_2m_min,sunrise,sunset,daylight_duration,uv_index_max,precipitation_sum&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timeformat=unixtime&timezone=%s&start_date=%s&end_date=%s&start_hour=%s&end_hour=%s", latitude, longitude, timezone, startDate, endDate, startHour, endHour)
}

type OpenMeteoResponse struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	GenerationTimeMs     float64 `json:"generationtime_ms"`
	UtcOffsetSeconds     int     `json:"utc_offset_seconds"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	Elevation            float64 `json:"elevation"`
	CurrentUnits         struct {
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

func UnitsToString(units string, value float64) string {
	switch units {
	case "s":
		if value != 1 {
			return " seconds"
		}
		return " second"
	case "inch":
		if value != 1 {
			return " inches"
		}
		return " inch"
	case "unixtime":
		return ""
	case "wmo code":
		return ""
	}
	return units
}

func (r *OpenMeteoResponse) GetCurrentWeatherCode() string {
	return WeatherCodeToString(r.Current.WeatherCode, r.Current.IsDay == 1)
}

func (r *OpenMeteoResponse) GetCurrentTemperature() string {
	return fmt.Sprintf("%g%s", r.Current.Temperature2m, UnitsToString(r.CurrentUnits.Temperature2m, r.Current.Temperature2m))
}

func (r *OpenMeteoResponse) GetCurrentRelativeHumidity() string {
	return fmt.Sprintf("%d%s", r.Current.RelativeHumidity2m, UnitsToString(r.CurrentUnits.RelativeHumidity2m, float64(r.Current.RelativeHumidity2m)))
}

func (r *OpenMeteoResponse) GetCurrentApparentTemperature() string {
	return fmt.Sprintf("%g%s", r.Current.ApparentTemperature, UnitsToString(r.CurrentUnits.ApparentTemperature, r.Current.ApparentTemperature))
}

func (r *OpenMeteoResponse) GetCurrentDayStatus() bool {
	return r.Current.IsDay == 1
}

func (r *OpenMeteoResponse) GetCurrentPrecipitation() string {
	return fmt.Sprintf("%g%s", r.Current.Precipitation, UnitsToString(r.CurrentUnits.Precipitation, r.Current.Precipitation))
}

func (r *OpenMeteoResponse) GetCurrentRain() string {
	return fmt.Sprintf("%g%s", r.Current.Rain, UnitsToString(r.CurrentUnits.Rain, r.Current.Rain))
}

func (r *OpenMeteoResponse) GetCurrentShowers() string {
	return fmt.Sprintf("%g%s", r.Current.Showers, UnitsToString(r.CurrentUnits.Showers, r.Current.Showers))
}

func (r *OpenMeteoResponse) GetCurrentSnowfall() string {
	return fmt.Sprintf("%g%s", r.Current.Snowfall, UnitsToString(r.CurrentUnits.Snowfall, r.Current.Snowfall))
}

func (r *OpenMeteoResponse) GetCurrentCloudCover() string {
	return fmt.Sprintf("%d%s", r.Current.CloudCover, UnitsToString(r.CurrentUnits.CloudCover, float64(r.Current.CloudCover)))
}

func (r *OpenMeteoResponse) GetCurrentSurfacePressure() string {
	return fmt.Sprintf("%g%s", r.Current.SurfacePressure, UnitsToString(r.CurrentUnits.SurfacePressure, r.Current.SurfacePressure))
}

func (r *OpenMeteoResponse) GetCurrentWindSpeed() string {
	return fmt.Sprintf("%g%s", r.Current.WindSpeed10m, UnitsToString(r.CurrentUnits.WindSpeed10m, r.Current.WindSpeed10m))
}

func (r *OpenMeteoResponse) GetCurrentWindDirection() string {
	return fmt.Sprintf("%d%s", r.Current.WindDirection10m, UnitsToString(r.CurrentUnits.WindDirection10m, float64(r.Current.WindDirection10m)))
}

type TimeseriesPoint struct {
	Unixtime int64  `json:"unix_time"`
	Value    string `json:"value"`
}

func (r *OpenMeteoResponse) GetHourlyTemperature() (forecast []TimeseriesPoint) {
	for index, value := range r.Hourly.Temperature2m {
		point := TimeseriesPoint{
			Unixtime: r.Hourly.Time[index],
			Value:    fmt.Sprintf("%g%s", value, UnitsToString(r.HourlyUnits.Temperature2m, value)),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *OpenMeteoResponse) GetHourlyPrecipitationProbability() (forecast []TimeseriesPoint) {
	for index, value := range r.Hourly.PrecipitationProbability {
		point := TimeseriesPoint{
			Unixtime: r.Hourly.Time[index],
			Value:    fmt.Sprintf("%d%s", value, UnitsToString(r.HourlyUnits.PrecipitationProbability, float64(value))),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *OpenMeteoResponse) GetHourlyPrecipitation() (forecast []TimeseriesPoint) {
	for index, value := range r.Hourly.Precipitation {
		point := TimeseriesPoint{
			Unixtime: r.Hourly.Time[index],
			Value:    fmt.Sprintf("%g%s", value, UnitsToString(r.HourlyUnits.Precipitation, value)),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *OpenMeteoResponse) GetHourlyWeatherCode() (forecast []TimeseriesPoint) {
	for index, value := range r.Hourly.WeatherCode {
		point := TimeseriesPoint{
			Unixtime: r.Hourly.Time[index],
			Value:    WeatherCodeToString(value, r.Hourly.Time[index] == 1),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *OpenMeteoResponse) GetDailyWeatherCode() (forecast []TimeseriesPoint) {
	for index, value := range r.Daily.WeatherCode {
		point := TimeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    WeatherCodeToString(value, r.Daily.Time[index] == 1),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *OpenMeteoResponse) GetDailyTemperatureMax() (forecast []TimeseriesPoint) {
	for index, value := range r.Daily.Temperature2mMax {
		point := TimeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    fmt.Sprintf("%g%s", value, UnitsToString(r.DailyUnits.Temperature2mMax, value)),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *OpenMeteoResponse) GetDailyTemperatureMin() (forecast []TimeseriesPoint) {
	for index, value := range r.Daily.Temperature2mMin {
		point := TimeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    fmt.Sprintf("%g%s", value, UnitsToString(r.DailyUnits.Temperature2mMin, value)),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *OpenMeteoResponse) GetDailySunrise() (forecast []TimeseriesPoint) {
	for index, value := range r.Daily.Sunrise {
		point := TimeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    hourlyUnixTimeToString(value + int64(r.UtcOffsetSeconds)),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *OpenMeteoResponse) GetDailySunset() (forecast []TimeseriesPoint) {
	for index, value := range r.Daily.Sunset {
		point := TimeseriesPoint{
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

func (r *OpenMeteoResponse) GetDailyDaylightDuration() (forecast []TimeseriesPoint) {
	for index, value := range r.Daily.DaylightDuration {
		point := TimeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    shortDur(time.Duration(value) * time.Second),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *OpenMeteoResponse) GetDailyUvIndexMax() (forecast []TimeseriesPoint) {
	for index, value := range r.Daily.UvIndexMax {
		point := TimeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    fmt.Sprintf("%.2f", value),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

func (r *OpenMeteoResponse) GetDailyPrecipitationSum() (forecast []TimeseriesPoint) {
	for index, value := range r.Daily.PrecipitationSum {
		point := TimeseriesPoint{
			Unixtime: r.Daily.Time[index],
			Value:    fmt.Sprintf("%g%s", value, UnitsToString(r.DailyUnits.PrecipitationSum, value)),
		}
		forecast = append(forecast, point)
	}
	return forecast
}

type CurrentStatus struct {
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

type HourStatus struct {
	Unixtime                 int64  `json:"unix_time"`
	TimeStr                  string `json:"time_str"`
	Temperature              string `json:"temperature"`
	PrecipitationProbability string `json:"precipitation_probability"`
	Precipitation            string `json:"precipitation"`
	WeatherCode              string `json:"weather_code"`
}

type DayStatus struct {
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

type WeatherStatus struct {
	Current CurrentStatus `json:"current"`
	Hourly  []HourStatus  `json:"hourly"`
	Daily   []DayStatus   `json:"daily"`
}

func hourlyUnixTimeToString(unixtime int64) string {
	t := time.Unix(unixtime, 0)
	timeStr := strings.ToLower(t.Format("3:04PM"))
	digitsStr := strings.TrimSuffix(timeStr[:len(timeStr)-2], ":00")
	return digitsStr + timeStr[len(timeStr)-2:]
}

func dailyUnixTimeToString(unixtime int64) string {
	t := time.Unix(unixtime, 0)
	return t.Format("Mon")
}

func (r *OpenMeteoResponse) TranslateToWeatherStatus() WeatherStatus {
	var status WeatherStatus

	status.Current.WeatherCode = r.GetCurrentWeatherCode()
	status.Current.Temperature = r.GetCurrentTemperature()
	status.Current.RelativeHumidity = r.GetCurrentRelativeHumidity()
	status.Current.ApparentTemperature = r.GetCurrentApparentTemperature()
	status.Current.IsDay = r.GetCurrentDayStatus()
	status.Current.CurrentPrecipitation = r.GetCurrentPrecipitation()
	status.Current.CurrentRain = r.GetCurrentRain()
	status.Current.CurrentShowers = r.GetCurrentShowers()
	status.Current.CurrentSnowfall = r.GetCurrentSnowfall()
	status.Current.CurrentCloudCover = r.GetCurrentCloudCover()
	status.Current.CurrentSurfacePressure = r.GetCurrentSurfacePressure()
	status.Current.CurrentWindSpeed = r.GetCurrentWindSpeed()
	status.Current.CurrentWindDirection = r.GetCurrentWindDirection()

	hourlyTemperature := r.GetHourlyTemperature()
	hourlyPrecipitationProbability := r.GetHourlyPrecipitationProbability()
	hourlyPrecipitation := r.GetHourlyPrecipitation()
	hourlyWeatherCode := r.GetHourlyWeatherCode()

	for i := range hourlyTemperature {
		status.Hourly = append(status.Hourly, HourStatus{
			Unixtime:                 hourlyTemperature[i].Unixtime,
			TimeStr:                  hourlyUnixTimeToString(hourlyTemperature[i].Unixtime + int64(r.UtcOffsetSeconds)),
			Temperature:              hourlyTemperature[i].Value,
			PrecipitationProbability: hourlyPrecipitationProbability[i].Value,
			Precipitation:            hourlyPrecipitation[i].Value,
			WeatherCode:              hourlyWeatherCode[i].Value,
		})
	}

	dailyWeatherCode := r.GetDailyWeatherCode()
	dailyTemperatureMax := r.GetDailyTemperatureMax()
	dailyTemperatureMin := r.GetDailyTemperatureMin()
	dailySunrise := r.GetDailySunrise()
	dailySunset := r.GetDailySunset()
	dailyDaylightDuration := r.GetDailyDaylightDuration()
	dailyUvIndexMax := r.GetDailyUvIndexMax()
	dailyPrecipitationSum := r.GetDailyPrecipitationSum()

	for i := range dailyWeatherCode {
		status.Daily = append(status.Daily, DayStatus{
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

func CurrentWeatherHandler(app *pocketbase.PocketBase) func(c echo.Context) error {
	return func(c echo.Context) error {

		// record, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)

		// if record == nil {
		// 	return apis.NewForbiddenError("You must be logged in", nil)
		// }
		latitudeRaw := c.QueryParam("latitude")

		if latitudeRaw == "" {
			return apis.NewBadRequestError("latitude is required", nil)
		}

		longitudeRaw := c.QueryParam("longitude")

		if longitudeRaw == "" {
			return apis.NewBadRequestError("longitude is required", nil)
		}

		latitude, err := strconv.ParseFloat(latitudeRaw, 64)
		if err != nil {
			return apis.NewBadRequestError("latitude is not a valid number", nil)
		}
		if latitude < -90 || latitude > 90 {
			return apis.NewBadRequestError("latitude must be between -90 and 90", nil)
		}
		longitude, err := strconv.ParseFloat(longitudeRaw, 64)
		if err != nil {
			return apis.NewBadRequestError("longitude is not a valid number", nil)
		}
		if longitude < -180 || longitude > 180 {
			return apis.NewBadRequestError("longitude must be between -180 and 180", nil)
		}

		timezone := c.QueryParam("timezone")

		if timezone == "" {
			return apis.NewBadRequestError("timezone is required", nil)
		}

		numHoursRaw := c.QueryParam("numHours")
		numDaysRaw := c.QueryParam("numDays")

		numHours := 8
		numDays := 5
		if numHoursRaw != "" {
			numHours, _ = strconv.Atoi(numHoursRaw)
		}
		if numDaysRaw != "" {
			numDays, _ = strconv.Atoi(numDaysRaw)
		}

		url := RequestUrl(timezone, latitude, longitude, numHours, numDays)

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

		var response OpenMeteoResponse
		err = json.Unmarshal([]byte(bodyStr), &response)
		if err != nil {
			return apis.NewApiError(500, "Failed to parse weather data", err)
		}

		return c.JSON(200, response.TranslateToWeatherStatus())
	}
}
