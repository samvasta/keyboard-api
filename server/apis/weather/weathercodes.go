package weather

type WeatherCodeDescription struct {
	Description string `json:"description"`
	ImageUrl    string `json:"imageUrl"`
}

type WeatherCodeDictEntry struct {
	Day   WeatherCodeDescription `json:"day"`
	Night WeatherCodeDescription `json:"night"`
}

type WeatherCodeDictionary map[int]WeatherCodeDictEntry

var WeatherCodes = WeatherCodeDictionary{
	0: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Sunny",
			ImageUrl:    "http://openweathermap.org/img/wn/01d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Clear",
			ImageUrl:    "http://openweathermap.org/img/wn/01n@2x.png",
		},
	},
	1: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Mainly Sunny",
			ImageUrl:    "http://openweathermap.org/img/wn/01d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Mainly Clear",
			ImageUrl:    "http://openweathermap.org/img/wn/01n@2x.png",
		},
	},
	2: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Partly Cloudy",
			ImageUrl:    "http://openweathermap.org/img/wn/02d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Partly Cloudy",
			ImageUrl:    "http://openweathermap.org/img/wn/02n@2x.png",
		},
	},
	3: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Cloudy",
			ImageUrl:    "http://openweathermap.org/img/wn/03d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Cloudy",
			ImageUrl:    "http://openweathermap.org/img/wn/03n@2x.png",
		},
	},
	45: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Foggy",
			ImageUrl:    "http://openweathermap.org/img/wn/50d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Foggy",
			ImageUrl:    "http://openweathermap.org/img/wn/50n@2x.png",
		},
	},
	48: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Rime Fog",
			ImageUrl:    "http://openweathermap.org/img/wn/50d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Rime Fog",
			ImageUrl:    "http://openweathermap.org/img/wn/50n@2x.png",
		},
	},
	51: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Light Drizzle",
			ImageUrl:    "http://openweathermap.org/img/wn/09d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Light Drizzle",
			ImageUrl:    "http://openweathermap.org/img/wn/09n@2x.png",
		},
	},
	53: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Drizzle",
			ImageUrl:    "http://openweathermap.org/img/wn/09d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Drizzle",
			ImageUrl:    "http://openweathermap.org/img/wn/09n@2x.png",
		},
	},
	55: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Heavy Drizzle",
			ImageUrl:    "http://openweathermap.org/img/wn/09d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Heavy Drizzle",
			ImageUrl:    "http://openweathermap.org/img/wn/09n@2x.png",
		},
	},
	56: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Light Freezing Drizzle",
			ImageUrl:    "http://openweathermap.org/img/wn/09d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Light Freezing Drizzle",
			ImageUrl:    "http://openweathermap.org/img/wn/09n@2x.png",
		},
	},
	57: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Freezing Drizzle",
			ImageUrl:    "http://openweathermap.org/img/wn/09d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Freezing Drizzle",
			ImageUrl:    "http://openweathermap.org/img/wn/09n@2x.png",
		},
	},
	61: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Light Rain",
			ImageUrl:    "http://openweathermap.org/img/wn/10d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Light Rain",
			ImageUrl:    "http://openweathermap.org/img/wn/10n@2x.png",
		},
	},
	63: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Rain",
			ImageUrl:    "http://openweathermap.org/img/wn/10d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Rain",
			ImageUrl:    "http://openweathermap.org/img/wn/10n@2x.png",
		},
	},
	65: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Heavy Rain",
			ImageUrl:    "http://openweathermap.org/img/wn/10d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Heavy Rain",
			ImageUrl:    "http://openweathermap.org/img/wn/10n@2x.png",
		},
	},
	66: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Light Freezing Rain",
			ImageUrl:    "http://openweathermap.org/img/wn/10d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Light Freezing Rain",
			ImageUrl:    "http://openweathermap.org/img/wn/10n@2x.png",
		},
	},
	67: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Freezing Rain",
			ImageUrl:    "http://openweathermap.org/img/wn/10d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Freezing Rain",
			ImageUrl:    "http://openweathermap.org/img/wn/10n@2x.png",
		},
	},
	71: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Light Snow",
			ImageUrl:    "http://openweathermap.org/img/wn/13d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Light Snow",
			ImageUrl:    "http://openweathermap.org/img/wn/13n@2x.png",
		},
	},
	73: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Snow",
			ImageUrl:    "http://openweathermap.org/img/wn/13d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Snow",
			ImageUrl:    "http://openweathermap.org/img/wn/13n@2x.png",
		},
	},
	75: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Heavy Snow",
			ImageUrl:    "http://openweathermap.org/img/wn/13d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Heavy Snow",
			ImageUrl:    "http://openweathermap.org/img/wn/13n@2x.png",
		},
	},
	77: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Snow Grains",
			ImageUrl:    "http://openweathermap.org/img/wn/13d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Snow Grains",
			ImageUrl:    "http://openweathermap.org/img/wn/13n@2x.png",
		},
	},
	80: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Light Showers",
			ImageUrl:    "http://openweathermap.org/img/wn/09d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Light Showers",
			ImageUrl:    "http://openweathermap.org/img/wn/09n@2x.png",
		},
	},
	81: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Showers",
			ImageUrl:    "http://openweathermap.org/img/wn/09d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Showers",
			ImageUrl:    "http://openweathermap.org/img/wn/09n@2x.png",
		},
	},
	82: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Heavy Showers",
			ImageUrl:    "http://openweathermap.org/img/wn/09d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Heavy Showers",
			ImageUrl:    "http://openweathermap.org/img/wn/09n@2x.png",
		},
	},
	85: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Light Snow Showers",
			ImageUrl:    "http://openweathermap.org/img/wn/13d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Light Snow Showers",
			ImageUrl:    "http://openweathermap.org/img/wn/13n@2x.png",
		},
	},
	86: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Snow Showers",
			ImageUrl:    "http://openweathermap.org/img/wn/13d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Snow Showers",
			ImageUrl:    "http://openweathermap.org/img/wn/13n@2x.png",
		},
	},
	95: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Thunderstorm",
			ImageUrl:    "http://openweathermap.org/img/wn/11d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Thunderstorm",
			ImageUrl:    "http://openweathermap.org/img/wn/11n@2x.png",
		},
	},
	96: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Light Thunderstorms With Hail",
			ImageUrl:    "http://openweathermap.org/img/wn/11d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Light Thunderstorms With Hail",
			ImageUrl:    "http://openweathermap.org/img/wn/11n@2x.png",
		},
	},
	99: WeatherCodeDictEntry{
		Day: WeatherCodeDescription{
			Description: "Thunderstorm With Hail",
			ImageUrl:    "http://openweathermap.org/img/wn/11d@2x.png",
		},
		Night: WeatherCodeDescription{
			Description: "Thunderstorm With Hail",
			ImageUrl:    "http://openweathermap.org/img/wn/11n@2x.png",
		},
	},
}

func WeatherCodeToString(code int, isDay bool) string {
	if isDay {
		return WeatherCodes[code].Day.Description
	} else {
		return WeatherCodes[code].Night.Description
	}
}
