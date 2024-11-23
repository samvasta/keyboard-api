package weather

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
)

func startOfNextHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), (t.Hour()+1)%24, 0, 0, 0, t.Location())
}

type openMeteoResponseBase struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	GenerationTimeMs     float64 `json:"generationtime_ms"`
	UtcOffsetSeconds     int     `json:"utc_offset_seconds"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	Elevation            float64 `json:"elevation"`
}

func unitsToString(units string, value float64) string {
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

type timeseriesPoint struct {
	Unixtime int64  `json:"unix_time"`
	Value    string `json:"value"`
}

func parseLatLongTz(c echo.Context) (latitude, longitude float64, timezone string, err error) {
	// record, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)

	// if record == nil {
	// 	return apis.NewForbiddenError("You must be logged in", nil)
	// }
	latitudeRaw := c.QueryParam("latitude")

	if latitudeRaw == "" {
		return 0, 0, "", apis.NewBadRequestError("latitude is required", nil)
	}

	longitudeRaw := c.QueryParam("longitude")

	if longitudeRaw == "" {
		return 0, 0, "", apis.NewBadRequestError("longitude is required", nil)
	}

	latitude, err = strconv.ParseFloat(latitudeRaw, 64)
	if err != nil {
		return 0, 0, "", apis.NewBadRequestError("latitude is not a valid number", nil)
	}
	if latitude < -90 || latitude > 90 {
		return 0, 0, "", apis.NewBadRequestError("latitude must be between -90 and 90", nil)
	}
	longitude, err = strconv.ParseFloat(longitudeRaw, 64)
	if err != nil {
		return 0, 0, "", apis.NewBadRequestError("longitude is not a valid number", nil)
	}
	if longitude < -180 || longitude > 180 {
		return 0, 0, "", apis.NewBadRequestError("longitude must be between -180 and 180", nil)
	}

	timezone = c.QueryParam("timezone")

	if timezone == "" {
		return 0, 0, "", apis.NewBadRequestError("timezone is required", nil)
	}

	return latitude, longitude, timezone, nil
}
