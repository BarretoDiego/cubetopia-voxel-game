// Package render provides underwater and environmental effects
package render

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// UnderwaterEffect provides underwater visual effects
type UnderwaterEffect struct {
	shader  *Shader
	quadVAO uint32
	quadVBO uint32

	// State
	IsUnderwater bool
	WaterColor   mgl32.Vec3
	FogDensity   float32
	WaveTime     float32
}

// NewUnderwaterEffect creates underwater effect renderer
func NewUnderwaterEffect() (*UnderwaterEffect, error) {
	uw := &UnderwaterEffect{
		WaterColor: mgl32.Vec3{0.1, 0.3, 0.5},
		FogDensity: 0.1,
	}

	// Create shader
	shader, err := NewShader(underwaterVertShader, underwaterFragShader)
	if err != nil {
		return nil, err
	}
	uw.shader = shader

	// Create fullscreen quad
	vertices := []float32{
		-1, -1, 0, 0,
		1, -1, 1, 0,
		1, 1, 1, 1,
		-1, -1, 0, 0,
		1, 1, 1, 1,
		-1, 1, 0, 1,
	}

	gl.GenVertexArrays(1, &uw.quadVAO)
	gl.GenBuffers(1, &uw.quadVBO)

	gl.BindVertexArray(uw.quadVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, uw.quadVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointerWithOffset(0, 2, gl.FLOAT, false, 4*4, 0)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 4*4, 2*4)
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)

	return uw, nil
}

// Update updates the effect
func (uw *UnderwaterEffect) Update(dt float32) {
	if uw.IsUnderwater {
		uw.WaveTime += dt * 2.0
	}
}

// Render renders the underwater overlay
func (uw *UnderwaterEffect) Render(screenTexture uint32) {
	if !uw.IsUnderwater || uw.shader == nil {
		return
	}

	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	uw.shader.Use()
	uw.shader.SetInt("uTexture", 0)
	uw.shader.SetVec3("uWaterColor", uw.WaterColor)
	uw.shader.SetFloat("uFogDensity", uw.FogDensity)
	uw.shader.SetFloat("uTime", uw.WaveTime)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, screenTexture)

	gl.BindVertexArray(uw.quadVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)

	gl.Enable(gl.DEPTH_TEST)
}

// Cleanup releases resources
func (uw *UnderwaterEffect) Cleanup() {
	if uw.quadVAO != 0 {
		gl.DeleteVertexArrays(1, &uw.quadVAO)
	}
	if uw.quadVBO != 0 {
		gl.DeleteBuffers(1, &uw.quadVBO)
	}
	if uw.shader != nil {
		uw.shader.Delete()
	}
}

var underwaterVertShader = `
#version 410 core

layout(location = 0) in vec2 aPos;
layout(location = 1) in vec2 aTexCoord;

out vec2 vTexCoord;

void main() {
    gl_Position = vec4(aPos, 0.0, 1.0);
    vTexCoord = aTexCoord;
}
` + "\x00"

var underwaterFragShader = `
#version 410 core

in vec2 vTexCoord;

uniform sampler2D uTexture;
uniform vec3 uWaterColor;
uniform float uFogDensity;
uniform float uTime;

out vec4 fragColor;

void main() {
    // Wave distortion
    vec2 distortedUV = vTexCoord;
    distortedUV.x += sin(vTexCoord.y * 20.0 + uTime * 3.0) * 0.01;
    distortedUV.y += cos(vTexCoord.x * 15.0 + uTime * 2.0) * 0.008;
    
    vec3 sceneColor = texture(uTexture, distortedUV).rgb;
    
    // Underwater color tint
    sceneColor = mix(sceneColor, uWaterColor, 0.3);
    
    // Depth fog (more fog at edges)
    float depth = length(vTexCoord - vec2(0.5)) * 2.0;
    sceneColor = mix(sceneColor, uWaterColor * 0.5, depth * uFogDensity);
    
    // Caustics (light patterns)
    float caustic1 = sin(vTexCoord.x * 50.0 + uTime * 2.0) * sin(vTexCoord.y * 50.0 + uTime * 1.5);
    float caustic2 = sin(vTexCoord.x * 30.0 - uTime * 1.5) * sin(vTexCoord.y * 40.0 + uTime * 2.5);
    float caustics = (caustic1 + caustic2) * 0.5 + 0.5;
    sceneColor += vec3(caustics * 0.1);
    
    // Vignette
    float vignette = 1.0 - length(vTexCoord - vec2(0.5)) * 0.5;
    sceneColor *= vignette;
    
    // Blue-green tint
    sceneColor.r *= 0.7;
    sceneColor.g *= 0.9;
    
    fragColor = vec4(sceneColor, 1.0);
}
` + "\x00"
