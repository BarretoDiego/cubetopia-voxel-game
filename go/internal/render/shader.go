// Package render provides shader management
package render

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Shader represents an OpenGL shader program
type Shader struct {
	ID uint32
}

// NewShader creates a shader program from vertex and fragment source
func NewShader(vertexSource, fragmentSource string) (*Shader, error) {
	vertexShader, err := compileShader(vertexSource, gl.VERTEX_SHADER)
	if err != nil {
		return nil, fmt.Errorf("vertex shader: %w", err)
	}
	defer gl.DeleteShader(vertexShader)

	fragmentShader, err := compileShader(fragmentSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, fmt.Errorf("fragment shader: %w", err)
	}
	defer gl.DeleteShader(fragmentShader)

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("link error: %s", log)
	}

	return &Shader{ID: program}, nil
}

// Use activates the shader
func (s *Shader) Use() {
	gl.UseProgram(s.ID)
}

// Delete cleans up the shader
func (s *Shader) Delete() {
	gl.DeleteProgram(s.ID)
}

// SetBool sets a boolean uniform
func (s *Shader) SetBool(name string, value bool) {
	var v int32 = 0
	if value {
		v = 1
	}
	gl.Uniform1i(s.getUniformLocation(name), v)
}

// SetInt sets an integer uniform
func (s *Shader) SetInt(name string, value int32) {
	gl.Uniform1i(s.getUniformLocation(name), value)
}

// SetFloat sets a float uniform
func (s *Shader) SetFloat(name string, value float32) {
	gl.Uniform1f(s.getUniformLocation(name), value)
}

// SetVec2 sets a vec2 uniform
func (s *Shader) SetVec2(name string, value mgl32.Vec2) {
	gl.Uniform2fv(s.getUniformLocation(name), 1, &value[0])
}

// SetVec3 sets a vec3 uniform
func (s *Shader) SetVec3(name string, value mgl32.Vec3) {
	gl.Uniform3fv(s.getUniformLocation(name), 1, &value[0])
}

// SetVec4 sets a vec4 uniform
func (s *Shader) SetVec4(name string, value mgl32.Vec4) {
	gl.Uniform4fv(s.getUniformLocation(name), 1, &value[0])
}

// SetMat4 sets a mat4 uniform
func (s *Shader) SetMat4(name string, value mgl32.Mat4) {
	gl.UniformMatrix4fv(s.getUniformLocation(name), 1, false, &value[0])
}

func (s *Shader) getUniformLocation(name string) int32 {
	return gl.GetUniformLocation(s.ID, gl.Str(name+"\x00"))
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	// Ensure shader source is null-terminated for OpenGL
	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("compile error: %s", log)
	}

	return shader, nil
}
