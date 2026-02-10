package notifications

import (
	"fmt"
	"time"

	"github.com/user/weather-notice-bot/internal/weather"
)

// FormatMorningMessage formats the morning weather notification with HTML
func FormatMorningMessage(w *weather.WeatherResponse, advice string) string {
	realtime := w.Result.Realtime
	daily := w.Result.Daily

	// Get today's forecast
	var todayTemp weather.DailyTempPoint
	if len(daily.Temperature) > 0 {
		todayTemp = daily.Temperature[0]
	}

	// Format message with HTML
	message := fmt.Sprintf(
		"🌅 <b>早安！今日天气预报</b>\n\n"+
			"%s <b>%s</b>\n"+
			"🌡️ 温度：%.1f°C (今日 %.1f°C ~ %.1f°C)\n"+
			"💧 湿度：%.0f%%\n"+
			"💨 风速：%.1f m/s\n",
		realtime.Skycon.Emoji(),
		realtime.Skycon.Chinese(),
		realtime.Temperature,
		todayTemp.Min,
		todayTemp.Max,
		realtime.Humidity*100,
		realtime.Wind.Speed,
	)

	// Add AQI if available
	if realtime.AirQuality.AQI.CN > 0 {
		message += fmt.Sprintf("🏭 空气质量：%s (AQI %d)\n",
			getAQILevel(realtime.AirQuality.AQI.CN),
			realtime.AirQuality.AQI.CN,
		)
	}

	// Add LLM advice
	if advice != "" {
		message += fmt.Sprintf("\n💡 <b>出行建议</b>\n%s", advice)
	}

	return message
}

// FormatEveningMessage formats the evening weather notification with HTML
func FormatEveningMessage(w *weather.WeatherResponse, advice string) string {
	realtime := w.Result.Realtime
	daily := w.Result.Daily

	// Get tomorrow's forecast
	var tomorrowTemp weather.DailyTempPoint
	var tomorrowSkycon weather.Skycon
	if len(daily.Temperature) > 1 {
		tomorrowTemp = daily.Temperature[1]
	}
	if len(daily.Skycon) > 1 {
		tomorrowSkycon = daily.Skycon[1].Value
	}

	// Format message with HTML
	message := fmt.Sprintf(
		"🌙 <b>晚安！今晚天气及明日预报</b>\n\n"+
			"<b>今晚</b>\n"+
			"%s <b>%s</b>\n"+
			"🌡️ 温度：%.1f°C\n"+
			"💧 湿度：%.0f%%\n\n"+
			"<b>明日预报</b>\n"+
			"%s <b>%s</b>\n"+
			"🌡️ 温度：%.1f°C ~ %.1f°C\n",
		realtime.Skycon.Emoji(),
		realtime.Skycon.Chinese(),
		realtime.Temperature,
		realtime.Humidity*100,
		tomorrowSkycon.Emoji(),
		tomorrowSkycon.Chinese(),
		tomorrowTemp.Min,
		tomorrowTemp.Max,
	)

	// Add LLM advice
	if advice != "" {
		message += fmt.Sprintf("\n💡 <b>温馨提示</b>\n%s", advice)
	}

	return message
}

// FormatChangeAlert formats a weather change alert with HTML
func FormatChangeAlert(changes string, w *weather.WeatherResponse, advice string) string {
	realtime := w.Result.Realtime

	// Format message with HTML
	message := fmt.Sprintf(
		"⚠️ <b>天气变化提醒</b>\n\n"+
			"<b>检测到显著变化：</b>\n%s\n\n"+
			"<b>当前天气</b>\n"+
			"%s <b>%s</b>\n"+
			"🌡️ 温度：%.1f°C\n"+
			"💧 湿度：%.0f%%\n"+
			"💨 风速：%.1f m/s\n",
		changes,
		realtime.Skycon.Emoji(),
		realtime.Skycon.Chinese(),
		realtime.Temperature,
		realtime.Humidity*100,
		realtime.Wind.Speed,
	)

	// Add AQI if available
	if realtime.AirQuality.AQI.CN > 0 {
		message += fmt.Sprintf("🏭 空气质量：%s (AQI %d)\n",
			getAQILevel(realtime.AirQuality.AQI.CN),
			realtime.AirQuality.AQI.CN,
		)
	}

	// Add LLM advice
	if advice != "" {
		message += fmt.Sprintf("\n💡 <b>应对建议</b>\n%s", advice)
	}

	return message
}

// getAQILevel returns the AQI level description in Chinese
func getAQILevel(aqi int) string {
	switch {
	case aqi <= 50:
		return "优"
	case aqi <= 100:
		return "良"
	case aqi <= 150:
		return "轻度污染"
	case aqi <= 200:
		return "中度污染"
	case aqi <= 300:
		return "重度污染"
	default:
		return "严重污染"
	}
}

// Helper function to format hourly forecast (for future use)
func formatHourlyForecast(hourly []weather.ValuePoint, skycon []weather.SkyconPoint) string {
	if len(hourly) == 0 {
		return ""
	}

	message := "\n📊 <b>未来24小时预报</b>\n"

	// Show forecast for next 24 hours (every 3 hours)
	now := time.Now()
	for i := 0; i < len(hourly) && i < 24; i += 3 {
		hour := hourly[i]
		if hour.Datetime.After(now) {
			skyconStr := "🌡️"
			if i < len(skycon) {
				skyconStr = skycon[i].Value.Emoji()
			}

			message += fmt.Sprintf("%s %s %.1f°C\n",
				hour.Datetime.Format("15:04"),
				skyconStr,
				hour.Value,
			)
		}
	}

	return message
}
