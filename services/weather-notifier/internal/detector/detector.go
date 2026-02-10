package detector

import (
	"fmt"

	"github.com/user/weather-notice-bot/internal/weather"
)

// WeatherChange represents a detected weather change
type WeatherChange struct {
	// Type of change detected
	Type ChangeType

	// Description of the change in English
	Description string

	// OldValue stores the previous value (type depends on change type)
	OldValue interface{}

	// NewValue stores the new value (type depends on change type)
	NewValue interface{}

	// Significance level (only for skycon changes)
	Significance SkyconSignificance
}

// ChangeDetector detects significant weather changes
type ChangeDetector struct {
	thresholds Thresholds
}

// NewChangeDetector creates a new ChangeDetector with the given thresholds
func NewChangeDetector(thresholds Thresholds) *ChangeDetector {
	return &ChangeDetector{
		thresholds: thresholds,
	}
}

// DetectChanges compares old and new weather states and returns detected changes
func (cd *ChangeDetector) DetectChanges(oldState WeatherState, newWeather weather.RealtimeWeather) []WeatherChange {
	var changes []WeatherChange

	// Check skycon changes
	if skyconChange := cd.checkSkyconChange(oldState.Skycon, newWeather.Skycon); skyconChange != nil {
		changes = append(changes, *skyconChange)
	}

	// Check temperature changes
	if tempChange := cd.checkTemperatureChange(oldState.Temperature, newWeather.Temperature); tempChange != nil {
		changes = append(changes, *tempChange)
	}

	// Check wind speed changes
	if windChange := cd.checkWindChange(oldState.WindSpeed, newWeather.Wind.Speed); windChange != nil {
		changes = append(changes, *windChange)
	}

	// Check visibility changes
	if visChange := cd.checkVisibilityChange(oldState.Visibility, newWeather.Visibility); visChange != nil {
		changes = append(changes, *visChange)
	}

	// Check AQI changes
	if aqiChange := cd.checkAQIChange(oldState.AQICN, oldState.AQIUSA, newWeather.AQI.CN, newWeather.AQI.USA); aqiChange != nil {
		changes = append(changes, *aqiChange)
	}

	return changes
}

// checkSkyconChange detects skycon (weather condition) changes
func (cd *ChangeDetector) checkSkyconChange(oldSkycon, newSkycon weather.Skycon) *WeatherChange {
	significance := GetSkyconSignificance(oldSkycon, newSkycon)

	// Only report if there's at least minor significance
	if significance == NoSignificance {
		return nil
	}

	return &WeatherChange{
		Type:         SkyconChange,
		Description:  fmt.Sprintf("Weather condition changed from %s to %s", oldSkycon.Chinese(), newSkycon.Chinese()),
		OldValue:     oldSkycon,
		NewValue:     newSkycon,
		Significance: significance,
	}
}

// checkTemperatureChange detects significant temperature changes
func (cd *ChangeDetector) checkTemperatureChange(oldTemp, newTemp float64) *WeatherChange {
	delta := newTemp - oldTemp
	absDelta := delta
	if absDelta < 0 {
		absDelta = -absDelta
	}

	if absDelta < cd.thresholds.TemperatureDelta {
		return nil
	}

	direction := "increased"
	if delta < 0 {
		direction = "decreased"
	}

	return &WeatherChange{
		Type:        TemperatureChange,
		Description: fmt.Sprintf("Temperature %s by %.1f°C (from %.1f°C to %.1f°C)", direction, absDelta, oldTemp, newTemp),
		OldValue:    oldTemp,
		NewValue:    newTemp,
	}
}

// checkWindChange detects significant wind speed changes
func (cd *ChangeDetector) checkWindChange(oldSpeed, newSpeed float64) *WeatherChange {
	delta := newSpeed - oldSpeed
	absDelta := delta
	if absDelta < 0 {
		absDelta = -absDelta
	}

	if absDelta < cd.thresholds.WindSpeedDelta {
		return nil
	}

	direction := "increased"
	if delta < 0 {
		direction = "decreased"
	}

	return &WeatherChange{
		Type:        WindChange,
		Description: fmt.Sprintf("Wind speed %s by %.1f m/s (from %.1f m/s to %.1f m/s)", direction, absDelta, oldSpeed, newSpeed),
		OldValue:    oldSpeed,
		NewValue:    newSpeed,
	}
}

// checkVisibilityChange detects significant visibility changes
func (cd *ChangeDetector) checkVisibilityChange(oldVis, newVis float64) *WeatherChange {
	delta := newVis - oldVis
	absDelta := delta
	if absDelta < 0 {
		absDelta = -absDelta
	}

	if absDelta < cd.thresholds.VisibilityDelta {
		return nil
	}

	direction := "improved"
	if delta < 0 {
		direction = "decreased"
	}

	return &WeatherChange{
		Type:        VisibilityChange,
		Description: fmt.Sprintf("Visibility %s by %.1f km (from %.1f km to %.1f km)", direction, absDelta, oldVis, newVis),
		OldValue:    oldVis,
		NewValue:    newVis,
	}
}

// checkAQIChange detects significant air quality changes
func (cd *ChangeDetector) checkAQIChange(oldCN, oldUSA, newCN, newUSA int) *WeatherChange {
	// Check China AQI
	cnDelta := newCN - oldCN
	absCNDelta := cnDelta
	if absCNDelta < 0 {
		absCNDelta = -absCNDelta
	}

	// Check USA AQI
	usaDelta := newUSA - oldUSA
	absUSADelta := usaDelta
	if absUSADelta < 0 {
		absUSADelta = -absUSADelta
	}

	// Report if either standard shows significant change
	if absCNDelta >= cd.thresholds.AQICNDelta {
		direction := "worsened"
		if cnDelta < 0 {
			direction = "improved"
		}

		return &WeatherChange{
			Type:        AQIChange,
			Description: fmt.Sprintf("Air quality %s (China AQI: %d → %d)", direction, oldCN, newCN),
			OldValue:    oldCN,
			NewValue:    newCN,
		}
	}

	if absUSADelta >= cd.thresholds.AQIUSADelta {
		direction := "worsened"
		if usaDelta < 0 {
			direction = "improved"
		}

		return &WeatherChange{
			Type:        AQIChange,
			Description: fmt.Sprintf("Air quality %s (USA AQI: %d → %d)", direction, oldUSA, newUSA),
			OldValue:    oldUSA,
			NewValue:    newUSA,
		}
	}

	return nil
}
