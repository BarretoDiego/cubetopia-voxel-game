// Package render provides post-processing effects
package render

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// PostProcess manages post-processing effects
type PostProcess struct {
	// Framebuffer
	fbo      uint32
	colorTex uint32
	depthTex uint32
	width    int
	height   int

	// Quad for fullscreen rendering
	quadVAO uint32
	quadVBO uint32

	// Shaders
	fxaaShader  *Shader
	bloomShader *Shader

	// Settings
	EnableFXAA    bool
	EnableBloom   bool
	BloomStrength float32
}

// NewPostProcess creates a post-processing pipeline
func NewPostProcess(width, height int) (*PostProcess, error) {
	pp := &PostProcess{
		width:         width,
		height:        height,
		EnableFXAA:    true,
		EnableBloom:   true,
		BloomStrength: 0.15,
	}

	// Create framebuffer
	if err := pp.createFramebuffer(); err != nil {
		return nil, err
	}

	// Create quad
	pp.createQuad()

	// Create shaders
	fxaa, err := NewShader(postVertexShader, fxaaFragmentShader)
	if err != nil {
		return nil, err
	}
	pp.fxaaShader = fxaa

	bloom, err := NewShader(postVertexShader, bloomFragmentShader)
	if err != nil {
		return nil, err
	}
	pp.bloomShader = bloom

	return pp, nil
}

func (pp *PostProcess) createFramebuffer() error {
	gl.GenFramebuffers(1, &pp.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, pp.fbo)

	// Color texture
	gl.GenTextures(1, &pp.colorTex)
	gl.BindTexture(gl.TEXTURE_2D, pp.colorTex)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, int32(pp.width), int32(pp.height), 0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, pp.colorTex, 0)

	// Depth texture
	gl.GenTextures(1, &pp.depthTex)
	gl.BindTexture(gl.TEXTURE_2D, pp.depthTex)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT24, int32(pp.width), int32(pp.height), 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, pp.depthTex, 0)

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	return nil
}

func (pp *PostProcess) createQuad() {
	vertices := []float32{
		-1, -1, 0, 0,
		1, -1, 1, 0,
		1, 1, 1, 1,
		-1, -1, 0, 0,
		1, 1, 1, 1,
		-1, 1, 0, 1,
	}

	gl.GenVertexArrays(1, &pp.quadVAO)
	gl.GenBuffers(1, &pp.quadVBO)

	gl.BindVertexArray(pp.quadVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, pp.quadVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointerWithOffset(0, 2, gl.FLOAT, false, 4*4, 0)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 4*4, 2*4)
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

// BeginScene begins rendering to the offscreen buffer
func (pp *PostProcess) BeginScene() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, pp.fbo)
	gl.Viewport(0, 0, int32(pp.width), int32(pp.height))
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

// EndScene ends rendering and applies post-processing
func (pp *PostProcess) EndScene() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, int32(pp.width), int32(pp.height))

	gl.Disable(gl.DEPTH_TEST)

	// Apply effects
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, pp.colorTex)

	if pp.EnableFXAA && pp.fxaaShader != nil {
		pp.fxaaShader.Use()
		pp.fxaaShader.SetInt("uTexture", 0)
		pp.fxaaShader.SetVec2("uTexelSize", mgl32.Vec2{1.0 / float32(pp.width), 1.0 / float32(pp.height)})
		pp.fxaaShader.SetFloat("uBloomStrength", pp.BloomStrength)
	} else if pp.EnableBloom && pp.bloomShader != nil {
		pp.bloomShader.Use()
		pp.bloomShader.SetInt("uTexture", 0)
		pp.bloomShader.SetFloat("uBloomStrength", pp.BloomStrength)
	}

	gl.BindVertexArray(pp.quadVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)

	gl.Enable(gl.DEPTH_TEST)
}

// Resize updates the framebuffer size
func (pp *PostProcess) Resize(width, height int) {
	pp.width = width
	pp.height = height

	// Recreate textures
	gl.BindTexture(gl.TEXTURE_2D, pp.colorTex)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, int32(width), int32(height), 0, gl.RGBA, gl.FLOAT, nil)

	gl.BindTexture(gl.TEXTURE_2D, pp.depthTex)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT24, int32(width), int32(height), 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)
}

// Cleanup releases resources
func (pp *PostProcess) Cleanup() {
	if pp.fbo != 0 {
		gl.DeleteFramebuffers(1, &pp.fbo)
	}
	if pp.colorTex != 0 {
		gl.DeleteTextures(1, &pp.colorTex)
	}
	if pp.depthTex != 0 {
		gl.DeleteTextures(1, &pp.depthTex)
	}
	if pp.quadVAO != 0 {
		gl.DeleteVertexArrays(1, &pp.quadVAO)
	}
	if pp.quadVBO != 0 {
		gl.DeleteBuffers(1, &pp.quadVBO)
	}
	if pp.fxaaShader != nil {
		pp.fxaaShader.Delete()
	}
	if pp.bloomShader != nil {
		pp.bloomShader.Delete()
	}
}

var postVertexShader = `
#version 410 core
layout(location = 0) in vec2 aPos;
layout(location = 1) in vec2 aTexCoord;

out vec2 vTexCoord;

void main() {
    gl_Position = vec4(aPos, 0.0, 1.0);
    vTexCoord = aTexCoord;
}
` + "\x00"

var fxaaFragmentShader = `
#version 410 core

in vec2 vTexCoord;
out vec4 fragColor;

uniform sampler2D uTexture;
uniform vec2 uTexelSize;
uniform float uBloomStrength;

// FXAA constants
const float FXAA_REDUCE_MIN = 1.0/128.0;
const float FXAA_REDUCE_MUL = 1.0/8.0;
const float FXAA_SPAN_MAX = 8.0;

void main() {
    vec3 rgbNW = texture(uTexture, vTexCoord + vec2(-1.0, -1.0) * uTexelSize).rgb;
    vec3 rgbNE = texture(uTexture, vTexCoord + vec2(1.0, -1.0) * uTexelSize).rgb;
    vec3 rgbSW = texture(uTexture, vTexCoord + vec2(-1.0, 1.0) * uTexelSize).rgb;
    vec3 rgbSE = texture(uTexture, vTexCoord + vec2(1.0, 1.0) * uTexelSize).rgb;
    vec3 rgbM = texture(uTexture, vTexCoord).rgb;
    
    vec3 luma = vec3(0.299, 0.587, 0.114);
    float lumaNW = dot(rgbNW, luma);
    float lumaNE = dot(rgbNE, luma);
    float lumaSW = dot(rgbSW, luma);
    float lumaSE = dot(rgbSE, luma);
    float lumaM = dot(rgbM, luma);
    
    float lumaMin = min(lumaM, min(min(lumaNW, lumaNE), min(lumaSW, lumaSE)));
    float lumaMax = max(lumaM, max(max(lumaNW, lumaNE), max(lumaSW, lumaSE)));
    
    vec2 dir;
    dir.x = -((lumaNW + lumaNE) - (lumaSW + lumaSE));
    dir.y = ((lumaNW + lumaSW) - (lumaNE + lumaSE));
    
    float dirReduce = max((lumaNW + lumaNE + lumaSW + lumaSE) * (0.25 * FXAA_REDUCE_MUL), FXAA_REDUCE_MIN);
    float rcpDirMin = 1.0 / (min(abs(dir.x), abs(dir.y)) + dirReduce);
    dir = min(vec2(FXAA_SPAN_MAX), max(vec2(-FXAA_SPAN_MAX), dir * rcpDirMin)) * uTexelSize;
    
    vec3 rgbA = 0.5 * (
        texture(uTexture, vTexCoord + dir * (1.0/3.0 - 0.5)).rgb +
        texture(uTexture, vTexCoord + dir * (2.0/3.0 - 0.5)).rgb
    );
    vec3 rgbB = rgbA * 0.5 + 0.25 * (
        texture(uTexture, vTexCoord + dir * -0.5).rgb +
        texture(uTexture, vTexCoord + dir * 0.5).rgb
    );
    
    float lumaB = dot(rgbB, luma);
    vec3 finalColor;
    if(lumaB < lumaMin || lumaB > lumaMax) {
        finalColor = rgbA;
    } else {
        finalColor = rgbB;
    }
    
    // Simple bloom extraction (bright parts)
    float brightness = dot(finalColor, vec3(0.2126, 0.7152, 0.0722));
    vec3 bloom = max(vec3(0.0), finalColor - vec3(0.8)) * uBloomStrength;
    finalColor += bloom;
    
    fragColor = vec4(finalColor, 1.0);
}
` + "\x00"

var bloomFragmentShader = `
#version 410 core

in vec2 vTexCoord;
out vec4 fragColor;

uniform sampler2D uTexture;
uniform float uBloomStrength;

void main() {
    vec3 color = texture(uTexture, vTexCoord).rgb;
    
    // Extract bright parts
    float brightness = dot(color, vec3(0.2126, 0.7152, 0.0722));
    vec3 bloom = max(vec3(0.0), color - vec3(0.7)) * uBloomStrength;
    
    // Tone mapping
    color = color + bloom;
    color = color / (color + vec3(1.0)); // Reinhard
    
    // Gamma correction
    color = pow(color, vec3(1.0/2.2));
    
    fragColor = vec4(color, 1.0);
}
` + "\x00"
