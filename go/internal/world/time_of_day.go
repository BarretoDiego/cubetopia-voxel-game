// Package world provides time of day management
package world

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// TimeOfDay manages the game's day/night cycle
type TimeOfDay struct {
	// Current time in hours (0-24)
	CurrentHour float32

	// Duration of a full day in real seconds (default: 600 = 10 minutes)
	DayDurationSeconds float32

	// Minimum brightness during night (0.1-0.5)
	NightBrightness float32

	// Accumulated real time since start
	elapsedTime float32
}

// NewTimeOfDay creates a new time of day system
func NewTimeOfDay() *TimeOfDay {
	return &TimeOfDay{
		CurrentHour:        12.0, // Start at noon
		DayDurationSeconds: 600.0,
		NightBrightness:    0.15,
		elapsedTime:        0,
	}
}

// Update advances time based on delta time (in seconds)
func (t *TimeOfDay) Update(deltaTime float32) {
	t.elapsedTime += deltaTime

	// Calculate hours per real second
	hoursPerSecond := 24.0 / t.DayDurationSeconds
	t.CurrentHour += deltaTime * hoursPerSecond

	// Wrap around at 24 hours
	for t.CurrentHour >= 24.0 {
		t.CurrentHour -= 24.0
	}
}

// SetTime sets the current time directly (0-24)
func (t *TimeOfDay) SetTime(hour float32) {
	t.CurrentHour = float32(math.Mod(float64(hour), 24.0))
	if t.CurrentHour < 0 {
		t.CurrentHour += 24.0
	}
}

// GetSunDirection returns the sun's direction vector based on current time
func (t *TimeOfDay) GetSunDirection() mgl32.Vec3 {
	// Sun rises at 6:00, sets at 18:00
	// At 6:00, sun is at horizon (east)
	// At 12:00, sun is overhead
	// At 18:00, sun is at horizon (west)

	// Convert hour to angle (6:00 = 0°, 12:00 = 90°, 18:00 = 180°)
	dayProgress := (t.CurrentHour - 6.0) / 12.0 // 0 at sunrise, 1 at sunset
	if dayProgress < 0 {
		dayProgress = 0
	}
	if dayProgress > 1 {
		dayProgress = 1
	}

	angle := dayProgress * math.Pi // 0 to π

	// Sun travels in arc
	sunY := float32(math.Sin(float64(angle)))
	sunX := float32(math.Cos(float64(angle)))

	// During night, sun is below horizon
	if t.CurrentHour < 6.0 || t.CurrentHour > 18.0 {
		sunY = -0.3
		// Moon position (opposite of sun)
		nightProgress := 0.0
		if t.CurrentHour >= 18.0 {
			nightProgress = (float64(t.CurrentHour) - 18.0) / 6.0
		} else {
			nightProgress = (float64(t.CurrentHour) + 6.0) / 6.0
		}
		sunX = float32(-math.Cos(nightProgress * math.Pi))
	}

	return mgl32.Vec3{sunX, sunY, 0.3}.Normalize()
}

// GetSunIntensity returns the sun's light intensity (0-1)
func (t *TimeOfDay) GetSunIntensity() float32 {
	// Full intensity from 8:00 to 16:00
	// Gradual transition during dawn/dusk

	if t.CurrentHour >= 8.0 && t.CurrentHour <= 16.0 {
		return 1.0
	}

	if t.CurrentHour >= 6.0 && t.CurrentHour < 8.0 {
		// Dawn
		return (t.CurrentHour - 6.0) / 2.0
	}

	if t.CurrentHour > 16.0 && t.CurrentHour <= 18.0 {
		// Dusk
		return (18.0 - t.CurrentHour) / 2.0
	}

	// Night
	return t.NightBrightness
}

// GetSkyColor returns the sky color based on time of day
func (t *TimeOfDay) GetSkyColor() mgl32.Vec3 {
	// Define key colors
	dayColor := mgl32.Vec3{0.53, 0.81, 0.98}     // Light blue
	sunriseColor := mgl32.Vec3{0.98, 0.6, 0.4}   // Orange/pink
	sunsetColor := mgl32.Vec3{0.95, 0.45, 0.35}  // Deep orange
	nightColor := mgl32.Vec3{0.05, 0.05, 0.15}   // Dark blue
	twilightColor := mgl32.Vec3{0.2, 0.15, 0.35} // Purple

	hour := t.CurrentHour

	// Night (0:00 - 5:00)
	if hour < 5.0 {
		return nightColor
	}

	// Dawn (5:00 - 7:00)
	if hour < 7.0 {
		blend := (hour - 5.0) / 2.0
		if hour < 6.0 {
			return lerpVec3(nightColor, twilightColor, blend*2)
		}
		return lerpVec3(twilightColor, sunriseColor, (blend-0.5)*2)
	}

	// Morning transition (7:00 - 9:00)
	if hour < 9.0 {
		blend := (hour - 7.0) / 2.0
		return lerpVec3(sunriseColor, dayColor, blend)
	}

	// Day (9:00 - 16:00)
	if hour < 16.0 {
		return dayColor
	}

	// Afternoon transition (16:00 - 18:00)
	if hour < 18.0 {
		blend := (hour - 16.0) / 2.0
		return lerpVec3(dayColor, sunsetColor, blend)
	}

	// Dusk (18:00 - 20:00)
	if hour < 20.0 {
		blend := (hour - 18.0) / 2.0
		return lerpVec3(sunsetColor, twilightColor, blend)
	}

	// Night transition (20:00 - 21:00)
	if hour < 21.0 {
		blend := hour - 20.0
		return lerpVec3(twilightColor, nightColor, blend)
	}

	// Night (21:00 - 24:00)
	return nightColor
}

// GetAmbientColor returns the ambient light color
func (t *TimeOfDay) GetAmbientColor() mgl32.Vec3 {
	intensity := t.GetSunIntensity()

	// Day: warm white ambient
	// Night: cool blue ambient
	dayAmbient := mgl32.Vec3{1.0, 0.95, 0.9}
	nightAmbient := mgl32.Vec3{0.3, 0.35, 0.5}

	return lerpVec3(nightAmbient, dayAmbient, intensity)
}

// GetFogColor returns fog color based on time
func (t *TimeOfDay) GetFogColor() mgl32.Vec3 {
	sky := t.GetSkyColor()
	// Fog is slightly lighter than sky
	return mgl32.Vec3{
		sky.X()*0.9 + 0.1,
		sky.Y()*0.9 + 0.1,
		sky.Z()*0.9 + 0.1,
	}
}

// IsNight returns true if it's night time
func (t *TimeOfDay) IsNight() bool {
	return t.CurrentHour < 6.0 || t.CurrentHour >= 18.0
}

// GetTimeString returns a formatted time string (e.g., "12:30 PM")
func (t *TimeOfDay) GetTimeString() string {
	hour := int(t.CurrentHour)
	minute := int((t.CurrentHour - float32(hour)) * 60)

	period := "AM"
	displayHour := hour

	if hour >= 12 {
		period = "PM"
		if hour > 12 {
			displayHour = hour - 12
		}
	}
	if hour == 0 {
		displayHour = 12
	}

	return formatTime(displayHour, minute, period)
}

func formatTime(hour, minute int, period string) string {
	hourStr := itoa(hour)
	minuteStr := itoa(minute)
	if minute < 10 {
		minuteStr = "0" + minuteStr
	}
	return hourStr + ":" + minuteStr + " " + period
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}

// GetDayProgress returns progress through the day (0-1, 0=midnight, 0.5=noon)
func (t *TimeOfDay) GetDayProgress() float32 {
	return t.CurrentHour / 24.0
}

// Helper function to linearly interpolate between two Vec3
func lerpVec3(a, b mgl32.Vec3, t float32) mgl32.Vec3 {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return mgl32.Vec3{
		a.X() + (b.X()-a.X())*t,
		a.Y() + (b.Y()-a.Y())*t,
		a.Z() + (b.Z()-a.Z())*t,
	}
}
