// Package render provides creature mesh rendering
package render

import (
	"math"
	"voxelgame/internal/core/block"
	"voxelgame/internal/generation/entity"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// CreatureRenderer renders procedural creatures
type CreatureRenderer struct {
	shader *Shader

	// Cached cube mesh
	cubeVAO uint32
	cubeVBO uint32
}

// NewCreatureRenderer creates a new creature renderer
func NewCreatureRenderer() (*CreatureRenderer, error) {
	cr := &CreatureRenderer{}

	// Create shader
	shader, err := NewShader(creatureVertexShader, creatureFragmentShader)
	if err != nil {
		return nil, err
	}
	cr.shader = shader

	// Create cube mesh
	cr.createCubeMesh()

	return cr, nil
}

func (cr *CreatureRenderer) createCubeMesh() {
	// Unit cube vertices with normals
	vertices := []float32{
		// Positions        // Normals
		// Front face
		-0.5, -0.5, 0.5, 0, 0, 1,
		0.5, -0.5, 0.5, 0, 0, 1,
		0.5, 0.5, 0.5, 0, 0, 1,
		-0.5, -0.5, 0.5, 0, 0, 1,
		0.5, 0.5, 0.5, 0, 0, 1,
		-0.5, 0.5, 0.5, 0, 0, 1,

		// Back face
		0.5, -0.5, -0.5, 0, 0, -1,
		-0.5, -0.5, -0.5, 0, 0, -1,
		-0.5, 0.5, -0.5, 0, 0, -1,
		0.5, -0.5, -0.5, 0, 0, -1,
		-0.5, 0.5, -0.5, 0, 0, -1,
		0.5, 0.5, -0.5, 0, 0, -1,

		// Top face
		-0.5, 0.5, 0.5, 0, 1, 0,
		0.5, 0.5, 0.5, 0, 1, 0,
		0.5, 0.5, -0.5, 0, 1, 0,
		-0.5, 0.5, 0.5, 0, 1, 0,
		0.5, 0.5, -0.5, 0, 1, 0,
		-0.5, 0.5, -0.5, 0, 1, 0,

		// Bottom face
		-0.5, -0.5, -0.5, 0, -1, 0,
		0.5, -0.5, -0.5, 0, -1, 0,
		0.5, -0.5, 0.5, 0, -1, 0,
		-0.5, -0.5, -0.5, 0, -1, 0,
		0.5, -0.5, 0.5, 0, -1, 0,
		-0.5, -0.5, 0.5, 0, -1, 0,

		// Right face
		0.5, -0.5, 0.5, 1, 0, 0,
		0.5, -0.5, -0.5, 1, 0, 0,
		0.5, 0.5, -0.5, 1, 0, 0,
		0.5, -0.5, 0.5, 1, 0, 0,
		0.5, 0.5, -0.5, 1, 0, 0,
		0.5, 0.5, 0.5, 1, 0, 0,

		// Left face
		-0.5, -0.5, -0.5, -1, 0, 0,
		-0.5, -0.5, 0.5, -1, 0, 0,
		-0.5, 0.5, 0.5, -1, 0, 0,
		-0.5, -0.5, -0.5, -1, 0, 0,
		-0.5, 0.5, 0.5, -1, 0, 0,
		-0.5, 0.5, -0.5, -1, 0, 0,
	}

	gl.GenVertexArrays(1, &cr.cubeVAO)
	gl.GenBuffers(1, &cr.cubeVBO)

	gl.BindVertexArray(cr.cubeVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, cr.cubeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	stride := int32(6 * 4) // 3 pos + 3 normal
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, stride, 0)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, stride, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

// RenderCreature renders a single creature
func (cr *CreatureRenderer) RenderCreature(creature *entity.Creature, view, projection mgl32.Mat4, sunDir mgl32.Vec3) {
	if cr.shader == nil || creature == nil {
		return
	}

	cr.shader.Use()
	cr.shader.SetMat4("uView", view)
	cr.shader.SetMat4("uProjection", projection)
	cr.shader.SetVec3("uSunDirection", sunDir)

	// Calculate animation offsets
	walkPhase := creature.WalkPhase
	idleBreath := float32(math.Sin(float64(creature.AnimationTime)*2.0)) * 0.02
	legIndex := 0

	// Render each body part as a scaled cube
	for _, part := range creature.BodyParts {
		offset := part.Offset

		// Apply body bob when moving or idle breathing
		if part.Type == "torso" || part.Type == "body" || part.Type == "abdomen" {
			if creature.IsMoving {
				// Walking bob
				offset[1] += float32(math.Abs(math.Sin(float64(walkPhase)*2.0))) * 0.03 * creature.Size
			} else {
				// Idle breathing
				offset[1] += idleBreath * creature.Size
			}
		}

		// Head slight movement
		if part.Type == "head" {
			if creature.IsMoving {
				offset[1] += float32(math.Abs(math.Sin(float64(walkPhase)*2.0))) * 0.02 * creature.Size
			} else {
				offset[1] += idleBreath * creature.Size
			}
		}

		pos := creature.Position.Add(offset)

		// Model matrix: translate, rotate, scale
		model := mgl32.Translate3D(pos.X(), pos.Y(), pos.Z())
		model = model.Mul4(mgl32.HomogRotate3DY(creature.Rotation))

		// Apply part-specific animations
		switch part.Type {
		case "leg":
			if creature.IsMoving {
				// Alternate legs based on index
				legPhase := walkPhase
				if legIndex%2 == 1 {
					legPhase += math.Pi // Opposite phase
				}
				// Front vs back legs for quadrupeds
				if creature.Template == entity.TemplateQuadruped && legIndex >= 2 {
					legPhase += math.Pi * 0.5 // Offset for back legs
				}
				// Swing angle
				legSwing := float32(math.Sin(float64(legPhase))) * 0.5
				// Rotate around X axis at the hip (top of leg)
				model = model.Mul4(mgl32.Translate3D(0, part.Size.Y()*0.5, 0))
				model = model.Mul4(mgl32.HomogRotate3DX(legSwing))
				model = model.Mul4(mgl32.Translate3D(0, -part.Size.Y()*0.5, 0))
			}
			legIndex++

		case "arm":
			if creature.IsMoving {
				// Arms swing opposite to legs
				armPhase := walkPhase + math.Pi
				if legIndex%2 == 1 {
					armPhase += math.Pi
				}
				armSwing := float32(math.Sin(float64(armPhase))) * 0.4
				model = model.Mul4(mgl32.Translate3D(0, part.Size.Y()*0.5, 0))
				model = model.Mul4(mgl32.HomogRotate3DX(armSwing))
				model = model.Mul4(mgl32.Translate3D(0, -part.Size.Y()*0.5, 0))
			}
			legIndex++

		case "wing":
			// Wings always flap for flying creatures
			wingFlap := float32(math.Sin(float64(creature.AnimationTime)*8.0)) * 0.6
			// Flap around Z axis
			if offset.X() > 0 {
				model = model.Mul4(mgl32.HomogRotate3DZ(-wingFlap))
			} else {
				model = model.Mul4(mgl32.HomogRotate3DZ(wingFlap))
			}

		case "tail":
			// Tail wags slightly
			tailWag := float32(math.Sin(float64(creature.AnimationTime)*4.0)) * 0.2
			model = model.Mul4(mgl32.HomogRotate3DY(tailWag))

		case "blob":
			// Slime squish animation
			squish := float32(math.Sin(float64(creature.AnimationTime)*3.0))*0.1 + 1.0
			model = model.Mul4(mgl32.Scale3D(1.0/squish, squish, 1.0/squish))

		case "fin":
			// Fish fin flutter
			finFlutter := float32(math.Sin(float64(creature.AnimationTime)*6.0)) * 0.3
			model = model.Mul4(mgl32.HomogRotate3DZ(finFlutter))
		}

		model = model.Mul4(mgl32.Scale3D(part.Size.X(), part.Size.Y(), part.Size.Z()))

		cr.shader.SetMat4("uModel", model)

		// Choose color based on part type
		var color [3]float32
		switch part.Type {
		case "head", "torso", "body", "abdomen", "thorax", "blob":
			color = creature.PrimaryColor
		case "leg", "arm", "wing", "fin", "tail":
			color = creature.SecondaryColor
		default:
			color = creature.AccentColor
		}
		cr.shader.SetVec3("uColor", mgl32.Vec3{color[0], color[1], color[2]})

		gl.BindVertexArray(cr.cubeVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
	}

	// Render held item if any
	if creature.HeldItem != block.Air {
		// Calculate item position - assume right hand side
		// For biped, roughly: x=0.5, y=0.8, z=0.5
		var itemOffset mgl32.Vec3

		switch creature.Template {
		case entity.TemplateBiped:
			itemOffset = mgl32.Vec3{0.6, 0.9, 0.5}
		case entity.TemplateQuadruped:
			itemOffset = mgl32.Vec3{0.4, 0.8, 0.8} // Mouth?
		default:
			itemOffset = mgl32.Vec3{0.5, 0.5, 0.5}
		}

		pos := creature.Position.Add(itemOffset)

		// Item scale
		itemSize := float32(0.25)

		model := mgl32.Translate3D(pos.X(), pos.Y(), pos.Z())
		// Rotate item slightly to look held
		model = model.Mul4(mgl32.HomogRotate3DY(creature.Rotation))
		model = model.Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(45))) // Tilt forward
		model = model.Mul4(mgl32.Scale3D(itemSize, itemSize, itemSize))

		cr.shader.SetMat4("uModel", model)

		// Set color based on block type
		color := creature.HeldItem.GetColor()
		cr.shader.SetVec3("uColor", mgl32.Vec3{color[0], color[1], color[2]})

		gl.BindVertexArray(cr.cubeVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
	}

	gl.BindVertexArray(0)
}

// RenderCreatures renders multiple creatures
func (cr *CreatureRenderer) RenderCreatures(creatures []*entity.Creature, view, projection mgl32.Mat4, sunDir mgl32.Vec3) {
	if len(creatures) == 0 {
		return
	}

	for _, creature := range creatures {
		cr.RenderCreature(creature, view, projection, sunDir)
	}
}

// Cleanup releases resources
func (cr *CreatureRenderer) Cleanup() {
	if cr.cubeVAO != 0 {
		gl.DeleteVertexArrays(1, &cr.cubeVAO)
	}
	if cr.cubeVBO != 0 {
		gl.DeleteBuffers(1, &cr.cubeVBO)
	}
	if cr.shader != nil {
		cr.shader.Delete()
	}
}

var creatureVertexShader = `
#version 410 core

layout(location = 0) in vec3 aPosition;
layout(location = 1) in vec3 aNormal;

uniform mat4 uModel;
uniform mat4 uView;
uniform mat4 uProjection;

out vec3 vNormal;
out vec3 vWorldPos;

void main() {
    vec4 worldPos = uModel * vec4(aPosition, 1.0);
    vWorldPos = worldPos.xyz;
    vNormal = mat3(transpose(inverse(uModel))) * aNormal;
    gl_Position = uProjection * uView * worldPos;
}
` + "\x00"

var creatureFragmentShader = `
#version 410 core

in vec3 vNormal;
in vec3 vWorldPos;

uniform vec3 uColor;
uniform vec3 uSunDirection;

out vec4 fragColor;

void main() {
    vec3 normal = normalize(vNormal);
    
    // Lighting
    float ambient = 0.3;
    float diffuse = max(dot(normal, uSunDirection), 0.0) * 0.7;
    
    vec3 finalColor = uColor * (ambient + diffuse);
    
    fragColor = vec4(finalColor, 1.0);
}
` + "\x00"
