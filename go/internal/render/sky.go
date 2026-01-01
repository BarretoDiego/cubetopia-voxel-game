// Package render provides sky rendering
package render

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Sky renders a procedural sky with gradient and sun
type Sky struct {
	shader   *Shader
	vao, vbo uint32

	// Time of day (0-24)
	TimeOfDay float32

	// Sun direction
	SunDirection mgl32.Vec3
}

// NewSky creates a new sky renderer
func NewSky() (*Sky, error) {
	s := &Sky{
		TimeOfDay: 12.0, // Noon
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

	return s, nil
}

// Update updates the sky based on time
func (s *Sky) Update(dt float32) {
	// Advance time (1 real second = 1 game minute)
	s.TimeOfDay += dt / 60.0
	if s.TimeOfDay >= 24.0 {
		s.TimeOfDay -= 24.0
	}

	s.updateSunDirection()
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
	s.shader.SetFloat("uTimeOfDay", s.TimeOfDay)

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
uniform float uTimeOfDay;

out vec4 fragColor;

void main() {
    vec3 rayDir = normalize(vRayDir);
    
    // Sky gradient based on height
    float height = rayDir.y * 0.5 + 0.5;
    
    // Day/night cycle
    float dayFactor = smoothstep(5.0, 7.0, uTimeOfDay) - smoothstep(17.0, 19.0, uTimeOfDay);
    
    // Horizon and zenith colors for day
    vec3 horizonDay = vec3(0.8, 0.85, 1.0);
    vec3 zenithDay = vec3(0.4, 0.6, 1.0);
    
    // Horizon and zenith colors for night
    vec3 horizonNight = vec3(0.1, 0.1, 0.15);
    vec3 zenithNight = vec3(0.02, 0.02, 0.05);
    
    // Sunrise/sunset colors
    vec3 sunsetColor = vec3(1.0, 0.5, 0.2);
    float sunsetFactor = 0.0;
    if (uTimeOfDay > 5.0 && uTimeOfDay < 8.0) {
        sunsetFactor = 1.0 - abs(uTimeOfDay - 6.5) / 1.5;
    } else if (uTimeOfDay > 16.0 && uTimeOfDay < 19.0) {
        sunsetFactor = 1.0 - abs(uTimeOfDay - 17.5) / 1.5;
    }
    
    // Interpolate sky color
    vec3 horizon = mix(horizonNight, horizonDay, dayFactor);
    vec3 zenith = mix(zenithNight, zenithDay, dayFactor);
    vec3 skyColor = mix(horizon, zenith, pow(height, 0.5));
    
    // Add sunset/sunrise glow
    skyColor = mix(skyColor, sunsetColor, sunsetFactor * (1.0 - height));
    
    // Sun disk
    float sunAngle = dot(rayDir, uSunDirection);
    float sunDisk = smoothstep(0.9995, 0.9999, sunAngle);
    vec3 sunColor = vec3(1.0, 0.95, 0.8) * dayFactor;
    
    // Sun glow
    float sunGlow = pow(max(0.0, sunAngle), 32.0) * 0.5 * dayFactor;
    
    skyColor += sunColor * sunDisk + vec3(1.0, 0.8, 0.5) * sunGlow;
    
    // Stars at night
    if (dayFactor < 0.3) {
        float starNoise = fract(sin(dot(rayDir.xz * 100.0, vec2(12.9898, 78.233))) * 43758.5453);
        float stars = step(0.998, starNoise) * (1.0 - dayFactor) * 2.0;
        skyColor += vec3(stars);
    }
    
    fragColor = vec4(skyColor, 1.0);
}
` + "\x00"
