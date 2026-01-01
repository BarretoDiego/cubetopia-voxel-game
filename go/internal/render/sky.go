// Package render provides sky rendering
package render

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Sky renders a procedural sky with gradient, sun, moon, stars, and clouds
type Sky struct {
	shader   *Shader
	vao, vbo uint32

	// Time of day (0-24)
	TimeOfDay float32

	// Day counter for moon phase
	DayCount float32

	// Sun direction
	SunDirection mgl32.Vec3

	// Moon direction
	MoonDirection mgl32.Vec3

	// Moon phase (0-1, where 0.5 = full moon)
	MoonPhase float32

	// Cloud offset for animation
	CloudOffset mgl32.Vec2

	// Total elapsed time for animations
	TotalTime float32
}

// NewSky creates a new sky renderer
func NewSky() (*Sky, error) {
	s := &Sky{
		TimeOfDay: 12.0, // Noon
		DayCount:  0,
		MoonPhase: 0.5, // Start with full moon
	}

	// Create sky shader
	shader, err := NewShader(skyVertexShader, skyFragmentShader)
	if err != nil {
		return nil, err
	}
	s.shader = shader

	// Create fullscreen quad
	vertices := []float32{
		-1, -1,
		1, -1,
		1, 1,
		-1, -1,
		1, 1,
		-1, 1,
	}

	gl.GenVertexArrays(1, &s.vao)
	gl.GenBuffers(1, &s.vbo)

	gl.BindVertexArray(s.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, s.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointerWithOffset(0, 2, gl.FLOAT, false, 2*4, 0)
	gl.EnableVertexAttribArray(0)

	gl.BindVertexArray(0)

	s.updateSunDirection()
	s.updateMoonDirection()

	return s, nil
}

// Update updates the sky based on time
func (s *Sky) Update(dt float32) {
	// Advance time (1 real second = 1 game minute)
	s.TimeOfDay += dt / 60.0
	if s.TimeOfDay >= 24.0 {
		s.TimeOfDay -= 24.0
		s.DayCount += 1.0
		// Moon phase cycle (29.5 days)
		s.MoonPhase = float32(math.Mod(float64(s.DayCount)/29.5, 1.0))
	}

	// Update cloud animation
	s.CloudOffset[0] += dt * 0.01
	s.CloudOffset[1] += dt * 0.005

	// Total time for animations
	s.TotalTime += dt

	s.updateSunDirection()
	s.updateMoonDirection()
}

func (s *Sky) updateSunDirection() {
	// Calculate sun position based on time
	// 6:00 = sunrise (east), 12:00 = zenith, 18:00 = sunset (west)
	angle := float64(s.TimeOfDay-6.0) / 12.0 * math.Pi

	s.SunDirection = mgl32.Vec3{
		float32(math.Cos(angle)),
		float32(math.Sin(angle)),
		0.2, // Slight tilt
	}.Normalize()
}

func (s *Sky) updateMoonDirection() {
	// Moon is opposite to sun with slight offset
	angle := float64(s.TimeOfDay-6.0)/12.0*math.Pi + math.Pi

	s.MoonDirection = mgl32.Vec3{
		float32(math.Cos(angle)),
		float32(math.Sin(angle)),
		-0.1, // Slight tilt
	}.Normalize()
}

// Render draws the sky
func (s *Sky) Render(invViewProj mgl32.Mat4, cameraPos mgl32.Vec3) {
	if s.shader == nil {
		return
	}

	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)

	s.shader.Use()
	s.shader.SetMat4("uInvViewProj", invViewProj)
	s.shader.SetVec3("uCameraPos", cameraPos)
	s.shader.SetVec3("uSunDirection", s.SunDirection)
	s.shader.SetVec3("uMoonDirection", s.MoonDirection)
	s.shader.SetFloat("uTimeOfDay", s.TimeOfDay)
	s.shader.SetFloat("uMoonPhase", s.MoonPhase)
	s.shader.SetFloat("uTime", s.TotalTime)
	s.shader.SetVec2("uCloudOffset", s.CloudOffset)

	gl.BindVertexArray(s.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)

	gl.DepthMask(true)
	gl.Enable(gl.DEPTH_TEST)
}

// GetSunDirection returns the current sun direction
func (s *Sky) GetSunDirection() mgl32.Vec3 {
	return s.SunDirection
}

// IsNight returns true if it's night time
func (s *Sky) IsNight() bool {
	return s.TimeOfDay < 6 || s.TimeOfDay > 18
}

// GetTimeOfDayNormalized returns time of day as a value from 0.0 to 1.0
// where 0.0/1.0 = midnight, 0.5 = noon
func (s *Sky) GetTimeOfDayNormalized() float32 {
	return s.TimeOfDay / 24.0
}

// Cleanup releases resources
func (s *Sky) Cleanup() {
	if s.vao != 0 {
		gl.DeleteVertexArrays(1, &s.vao)
	}
	if s.vbo != 0 {
		gl.DeleteBuffers(1, &s.vbo)
	}
	if s.shader != nil {
		s.shader.Delete()
	}
}

var skyVertexShader = `
#version 410 core

layout(location = 0) in vec2 aPos;

uniform mat4 uInvViewProj;

out vec3 vRayDir;

void main() {
    gl_Position = vec4(aPos, 0.999, 1.0);
    
    // Calculate ray direction
    vec4 nearPoint = uInvViewProj * vec4(aPos, -1.0, 1.0);
    vec4 farPoint = uInvViewProj * vec4(aPos, 1.0, 1.0);
    nearPoint /= nearPoint.w;
    farPoint /= farPoint.w;
    
    vRayDir = normalize(farPoint.xyz - nearPoint.xyz);
}
` + "\x00"

var skyFragmentShader = `
#version 410 core

in vec3 vRayDir;

uniform vec3 uSunDirection;
uniform vec3 uMoonDirection;
uniform float uTimeOfDay;
uniform float uMoonPhase;
uniform float uTime;
uniform vec2 uCloudOffset;

out vec4 fragColor;

// Hash function for procedural noise
float hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453123);
}

float hash3(vec3 p) {
    return fract(sin(dot(p, vec3(127.1, 311.7, 74.7))) * 43758.5453123);
}

// Value noise
float noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    f = f * f * (3.0 - 2.0 * f);
    
    float a = hash(i);
    float b = hash(i + vec2(1.0, 0.0));
    float c = hash(i + vec2(0.0, 1.0));
    float d = hash(i + vec2(1.0, 1.0));
    
    return mix(mix(a, b, f.x), mix(c, d, f.x), f.y);
}

// FBM for clouds
float fbm(vec2 p) {
    float value = 0.0;
    float amplitude = 0.5;
    float frequency = 1.0;
    
    for (int i = 0; i < 5; i++) {
        value += amplitude * noise(p * frequency);
        frequency *= 2.0;
        amplitude *= 0.5;
    }
    return value;
}

// Voronoi for stars
float voronoi(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    
    float minDist = 1.0;
    for (int y = -1; y <= 1; y++) {
        for (int x = -1; x <= 1; x++) {
            vec2 neighbor = vec2(float(x), float(y));
            vec2 point = hash(i + neighbor) * 0.5 + 0.25;
            vec2 diff = neighbor + point - f;
            float dist = length(diff);
            minDist = min(minDist, dist);
        }
    }
    return minDist;
}

// Clouds
vec4 getClouds(vec3 rayDir, float dayFactor) {
    if (rayDir.y < 0.0) return vec4(0.0);
    
    // Project ray onto cloud plane
    float cloudHeight = 0.15;
    vec2 cloudUV = rayDir.xz / (rayDir.y + 0.1) * 2.0;
    cloudUV += uCloudOffset;
    
    // Multi-layer clouds
    float cloud1 = fbm(cloudUV * 0.5);
    float cloud2 = fbm(cloudUV * 1.0 + vec2(100.0));
    float cloud3 = fbm(cloudUV * 2.0 + vec2(200.0));
    
    float clouds = cloud1 * 0.5 + cloud2 * 0.3 + cloud3 * 0.2;
    clouds = smoothstep(0.4, 0.7, clouds);
    
    // Fade clouds at horizon and overhead
    float horizonFade = smoothstep(0.0, 0.2, rayDir.y);
    float overheadFade = smoothstep(0.9, 0.5, rayDir.y);
    clouds *= horizonFade * overheadFade;
    
    // Cloud color based on time of day
    vec3 cloudColorDay = vec3(1.0, 1.0, 1.0);
    vec3 cloudColorSunset = vec3(1.0, 0.7, 0.5);
    vec3 cloudColorNight = vec3(0.15, 0.15, 0.2);
    
    // Sunset factor
    float sunsetFactor = 0.0;
    if (uTimeOfDay > 5.0 && uTimeOfDay < 8.0) {
        sunsetFactor = 1.0 - abs(uTimeOfDay - 6.5) / 1.5;
    } else if (uTimeOfDay > 16.0 && uTimeOfDay < 19.0) {
        sunsetFactor = 1.0 - abs(uTimeOfDay - 17.5) / 1.5;
    }
    
    vec3 cloudColor = mix(cloudColorNight, cloudColorDay, dayFactor);
    cloudColor = mix(cloudColor, cloudColorSunset, sunsetFactor * 0.7);
    
    // Cloud lighting from sun
    float sunLight = max(0.0, dot(uSunDirection, vec3(0.0, 1.0, 0.0))) * dayFactor;
    cloudColor *= 0.7 + sunLight * 0.3;
    
    return vec4(cloudColor, clouds * 0.9);
}

// Stars
float getStars(vec3 rayDir, float nightFactor, float moonBrightness) {
    if (nightFactor < 0.1) return 0.0;
    
    // Create star field
    vec2 starUV = rayDir.xz / (abs(rayDir.y) + 0.001) * 50.0;
    
    // Multiple star layers
    float stars = 0.0;
    
    // Small dim stars
    float starNoise1 = hash(floor(starUV * 20.0));
    stars += step(0.992, starNoise1) * 0.3;
    
    // Medium stars
    float starNoise2 = hash(floor(starUV * 10.0));
    stars += step(0.985, starNoise2) * 0.6;
    
    // Bright stars with twinkle
    float starNoise3 = hash(floor(starUV * 5.0));
    float twinkle = sin(uTime * 3.0 + starNoise3 * 100.0) * 0.5 + 0.5;
    stars += step(0.975, starNoise3) * (0.8 + twinkle * 0.4);
    
    // Colored stars
    vec3 starPos = floor(starUV.xyy * 5.0);
    float colorSeed = hash3(starPos);
    
    // Fade stars based on moon brightness and vertical position
    float verticalFade = smoothstep(0.0, 0.3, rayDir.y);
    stars *= nightFactor * (1.0 - moonBrightness * 0.5) * verticalFade;
    
    return stars;
}

// Milky Way
float getMilkyWay(vec3 rayDir, float nightFactor) {
    if (nightFactor < 0.3) return 0.0;
    
    // Milky way band
    vec3 milkyDir = normalize(vec3(0.3, 0.5, 1.0));
    float bandDist = abs(dot(rayDir, milkyDir));
    float band = smoothstep(0.4, 0.0, bandDist);
    
    // Add noise to the band
    vec2 milkyUV = rayDir.xz * 30.0;
    float milkyNoise = fbm(milkyUV * 0.5) * fbm(milkyUV * 2.0);
    
    float milky = band * milkyNoise * 0.3;
    milky *= smoothstep(0.0, 0.5, rayDir.y);
    
    return milky * nightFactor;
}

// Moon
vec3 getMoon(vec3 rayDir, float moonAngle, float moonPhase) {
    // Moon disk
    float moonDisk = smoothstep(0.997, 0.999, moonAngle);
    
    if (moonDisk < 0.01) return vec3(0.0);
    
    // Moon phase shadow
    vec3 moonRight = normalize(cross(uMoonDirection, vec3(0.0, 1.0, 0.0)));
    float phaseAngle = (moonPhase - 0.5) * 2.0; // -1 to 1
    
    // Calculate position on moon disk for phase
    vec3 moonLocal = rayDir - uMoonDirection * moonAngle;
    float moonX = dot(moonLocal, moonRight);
    
    // Phase shadow
    float phaseShadow = 1.0;
    if (abs(phaseAngle) > 0.1) {
        phaseShadow = smoothstep(-phaseAngle * 0.003, phaseAngle * 0.003, moonX);
        if (phaseAngle < 0.0) phaseShadow = 1.0 - phaseShadow;
    }
    
    // Moon surface texture (simple craters)
    vec2 moonUV = rayDir.xz * 100.0;
    float craters = noise(moonUV * 5.0) * 0.2;
    
    // Moon color
    vec3 moonColor = vec3(0.95, 0.93, 0.88) - craters;
    moonColor *= phaseShadow;
    
    // Moon glow
    float moonGlow = smoothstep(0.99, 0.997, moonAngle);
    vec3 glowColor = vec3(0.4, 0.45, 0.6) * moonGlow;
    
    return moonColor * moonDisk + glowColor;
}

void main() {
    vec3 rayDir = normalize(vRayDir);
    
    // Sky height factor
    float height = rayDir.y * 0.5 + 0.5;
    
    // Day/night cycle
    float dayFactor = smoothstep(5.0, 7.0, uTimeOfDay) - smoothstep(17.0, 19.0, uTimeOfDay);
    float nightFactor = 1.0 - dayFactor;
    
    // === SKY GRADIENT ===
    
    // Day colors with atmospheric scattering
    vec3 zenithDay = vec3(0.2, 0.4, 0.9);      // Deep blue zenith
    vec3 horizonDay = vec3(0.7, 0.85, 1.0);    // Light blue horizon
    
    // Night colors
    vec3 zenithNight = vec3(0.01, 0.01, 0.03); // Nearly black
    vec3 horizonNight = vec3(0.05, 0.05, 0.12); // Dark blue horizon
    
    // Interpolate based on height
    vec3 skyDay = mix(horizonDay, zenithDay, pow(height, 0.6));
    vec3 skyNight = mix(horizonNight, zenithNight, pow(height, 0.4));
    
    vec3 skyColor = mix(skyNight, skyDay, dayFactor);
    
    // === SUNRISE/SUNSET ===
    
    float sunsetFactor = 0.0;
    vec3 sunsetColor = vec3(1.0, 0.4, 0.1);
    vec3 sunriseColor = vec3(1.0, 0.6, 0.3);
    
    if (uTimeOfDay > 5.0 && uTimeOfDay < 8.0) {
        sunsetFactor = 1.0 - abs(uTimeOfDay - 6.5) / 1.5;
        skyColor = mix(skyColor, sunriseColor, sunsetFactor * (1.0 - pow(height, 0.3)));
    } else if (uTimeOfDay > 16.0 && uTimeOfDay < 19.0) {
        sunsetFactor = 1.0 - abs(uTimeOfDay - 17.5) / 1.5;
        skyColor = mix(skyColor, sunsetColor, sunsetFactor * (1.0 - pow(height, 0.3)));
    }
    
    // === SUN ===
    
    float sunAngle = dot(rayDir, uSunDirection);
    
    // Sun disk
    float sunDisk = smoothstep(0.9995, 0.9999, sunAngle);
    vec3 sunColor = vec3(1.0, 0.98, 0.9) * sunDisk * dayFactor;
    
    // Sun corona
    float sunCorona = pow(max(0.0, sunAngle), 64.0) * 0.8 * dayFactor;
    vec3 coronaColor = vec3(1.0, 0.9, 0.7) * sunCorona;
    
    // Sun glow
    float sunGlow = pow(max(0.0, sunAngle), 8.0) * 0.4 * dayFactor;
    vec3 glowColor = vec3(1.0, 0.8, 0.5) * sunGlow;
    
    skyColor += sunColor + coronaColor + glowColor;
    
    // === MOON ===
    
    float moonAngle = dot(rayDir, uMoonDirection);
    
    // Moon brightness affects night sky
    float moonBrightness = (1.0 - abs(uMoonPhase - 0.5) * 2.0);
    moonBrightness *= smoothstep(0.0, 0.5, uMoonDirection.y); // Only when visible
    
    vec3 moonColor = getMoon(rayDir, moonAngle, uMoonPhase);
    skyColor += moonColor * nightFactor;
    
    // === STARS ===
    
    float stars = getStars(rayDir, nightFactor, moonBrightness);
    skyColor += vec3(stars);
    
    // === MILKY WAY ===
    
    float milkyWay = getMilkyWay(rayDir, nightFactor * (1.0 - moonBrightness));
    skyColor += vec3(0.6, 0.65, 0.8) * milkyWay;
    
    // === CLOUDS ===
    
    vec4 clouds = getClouds(rayDir, dayFactor);
    skyColor = mix(skyColor, clouds.rgb, clouds.a);
    
    // === HORIZON HAZE ===
    
    float haze = pow(1.0 - abs(rayDir.y), 8.0) * 0.3;
    vec3 hazeColor = mix(vec3(0.1, 0.1, 0.15), vec3(0.7, 0.75, 0.85), dayFactor);
    skyColor = mix(skyColor, hazeColor, haze);
    
    fragColor = vec4(skyColor, 1.0);
}
` + "\x00"
