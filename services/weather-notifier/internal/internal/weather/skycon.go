package weather

// Skycon represents weather conditions from Caiyun API
type Skycon string

const (
	ClearDay          Skycon = "CLEAR_DAY"
	ClearNight        Skycon = "CLEAR_NIGHT"
	PartlyCloudyDay   Skycon = "PARTLY_CLOUDY_DAY"
	PartlyCloudyNight Skycon = "PARTLY_CLOUDY_NIGHT"
	Cloudy            Skycon = "CLOUDY"
	LightHaze         Skycon = "LIGHT_HAZE"
	ModerateHaze      Skycon = "MODERATE_HAZE"
	HeavyHaze         Skycon = "HEAVY_HAZE"
	LightRain         Skycon = "LIGHT_RAIN"
	ModerateRain      Skycon = "MODERATE_RAIN"
	HeavyRain         Skycon = "HEAVY_RAIN"
	StormRain         Skycon = "STORM_RAIN"
	Fog               Skycon = "FOG"
	LightSnow         Skycon = "LIGHT_SNOW"
	ModerateSnow      Skycon = "MODERATE_SNOW"
	HeavySnow         Skycon = "HEAVY_SNOW"
	StormSnow         Skycon = "STORM_SNOW"
	Dust              Skycon = "DUST"
	Sand              Skycon = "SAND"
	WindySkycon       Skycon = "WIND"
)

// Chinese returns the Chinese translation of the weather condition
func (s Skycon) Chinese() string {
	translations := map[Skycon]string{
		ClearDay:          "晴天",
		ClearNight:        "晴夜",
		PartlyCloudyDay:   "多云",
		PartlyCloudyNight: "多云",
		Cloudy:            "阴天",
		LightHaze:         "轻度雾霾",
		ModerateHaze:      "中度雾霾",
		HeavyHaze:         "重度雾霾",
		LightRain:         "小雨",
		ModerateRain:      "中雨",
		HeavyRain:         "大雨",
		StormRain:         "暴雨",
		Fog:               "雾",
		LightSnow:         "小雪",
		ModerateSnow:      "中雪",
		HeavySnow:         "大雪",
		StormSnow:         "暴雪",
		Dust:              "浮尘",
		Sand:              "沙尘",
		WindySkycon:       "大风",
	}

	if cn, ok := translations[s]; ok {
		return cn
	}
	return string(s)
}

// Emoji returns an emoji representation of the weather condition
func (s Skycon) Emoji() string {
	emojis := map[Skycon]string{
		ClearDay:          "☀️",
		ClearNight:        "🌙",
		PartlyCloudyDay:   "⛅",
		PartlyCloudyNight: "☁️",
		Cloudy:            "☁️",
		LightHaze:         "🌫️",
		ModerateHaze:      "🌫️",
		HeavyHaze:         "🌫️",
		LightRain:         "🌧️",
		ModerateRain:      "🌧️",
		HeavyRain:         "🌧️",
		StormRain:         "⛈️",
		Fog:               "🌫️",
		LightSnow:         "🌨️",
		ModerateSnow:      "❄️",
		HeavySnow:         "❄️",
		StormSnow:         "❄️",
		Dust:              "💨",
		Sand:              "💨",
		WindySkycon:       "💨",
	}

	if emoji, ok := emojis[s]; ok {
		return emoji
	}
	return "🌡️"
}
