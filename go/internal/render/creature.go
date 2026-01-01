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
			// Arms swing opposite to legs
			armPhase := walkPhase + math.Pi
			if legIndex%2 == 1 {
				armPhase += math.Pi
			}
			armSwing := float32(math.Sin(float64(armPhase))) * 0.4

			// Apply attack swing override for right arm
			if legIndex%2 == 1 && creature.SwingPhase > 0 {
				// Overwrite arm swing with attack swing
				// Swing down: starts high, swings down rapidly
				swingProgress := creature.SwingPhase / math.Pi // 0 to 1
				armSwing = float32(math.Sin(float64(swingProgress*math.Pi))) * 2.0
				// Also rotate inward slightly?
			} else if creature.IsMoving {
				// Only apply walk swing if moving and not attacking (or for left arm)
			} else {
				armSwing = 0
			}

			model = model.Mul4(mgl32.Translate3D(0, part.Size.Y()*0.5, 0))
			model = model.Mul4(mgl32.HomogRotate3DX(armSwing))
			model = model.Mul4(mgl32.Translate3D(0, -part.Size.Y()*0.5, 0))

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

		model := mgl32.Translate3D(pos.X(), pos.Y(), pos.Z())
		// Rotate item slightly to look held
		model = model.Mul4(mgl32.HomogRotate3DY(creature.Rotation))

		// Apply swing rotation to item matches arm
		if creature.SwingPhase > 0 {
			swingProgress := creature.SwingPhase / math.Pi
			swingAngle := float32(math.Sin(float64(swingProgress*math.Pi))) * 2.0
			model = model.Mul4(mgl32.Translate3D(0, 0, 0)) // Pivot?
			model = model.Mul4(mgl32.HomogRotate3DX(swingAngle))
		}

		// Adjust orientation based on item type
		if creature.HeldItem == block.Pickaxe || creature.HeldItem == block.Axe || creature.HeldItem == block.Shovel || creature.HeldItem == block.Sword {
			// Tools tend to be held vertically or angled forward
			model = model.Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(45))) // Tilt forward
		} else {
			// Blocks
			model = model.Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(45)))
			// Scale blocks down a bit
			model = model.Mul4(mgl32.Scale3D(0.25, 0.25, 0.25))
		}

		// Set color based on block type
		color := creature.HeldItem.GetColor()

		// Use RenderItem
		cr.RenderItem(creature.HeldItem, model, color)
	}

	gl.BindVertexArray(0)
}

// RenderViewModel renders the held item in first-person view
func (cr *CreatureRenderer) RenderViewModel(item block.Type, animationTime float32, swingPhase float32, projection mgl32.Mat4, sunDir mgl32.Vec3) {
	if cr.shader == nil || item == block.Air {
		return
	}

	cr.shader.Use()
	// Use a fixed view matrix for the view model relative to the "camera"
	viewModelView := mgl32.LookAtV(mgl32.Vec3{0, 0, 2}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	cr.shader.SetMat4("uView", viewModelView)
	cr.shader.SetMat4("uProjection", projection)
	cr.shader.SetVec3("uSunDirection", sunDir)

	// Animation: subtle breathing and sway
	breathing := float32(math.Sin(float64(animationTime)*1.5)) * 0.015
	swayX := float32(math.Sin(float64(animationTime)*0.8)) * 0.01

	// Position: Bottom-Right of the screen
	// Adjusted for 3D models which might have different pivots
	itemPos := mgl32.Vec3{0.6 + swayX, -0.6 + breathing, 0.5}

	// Different scale/rotation for tools vs blocks
	isTool := item == block.Pickaxe || item == block.Axe || item == block.Shovel || item == block.Sword

	model := mgl32.Translate3D(itemPos.X(), itemPos.Y(), itemPos.Z())

	if isTool {
		// Tools: Held by handle, upright-ish
		model = model.Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(-15)))
		model = model.Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(15)))
		// model = model.Mul4(mgl32.Scale3D(0.4, 0.4, 0.4)) // Scale handled in RenderItem for tools
	} else {
		// Blocks: Held as cube
		model = model.Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(-25)))
		model = model.Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(20)))
		model = model.Mul4(mgl32.Scale3D(0.4, 0.4, 0.4))
	}

	// Apply swing animation
	if swingPhase > 0 {
		// Swing arc: rotate down and slightly in
		swingProgress := swingPhase / math.Pi
		swingAngle := float32(math.Sin(float64(swingProgress*math.Pi))) * 1.5

		// Move item forward/center during swing
		model = model.Mul4(mgl32.Translate3D(-0.3*swingProgress, -0.2*swingProgress, -0.3*swingProgress))
		model = model.Mul4(mgl32.HomogRotate3DX(swingAngle))
	}

	cr.shader.SetMat4("uModel", model)

	// Set color based on block type
	color := item.GetColor()
	cr.RenderItem(item, model, color)

	gl.Disable(gl.CULL_FACE)
	gl.BindVertexArray(cr.cubeVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 36)
	gl.Enable(gl.CULL_FACE)

	gl.BindVertexArray(0)
}

// RenderItem renders a specific item in 3D
func (cr *CreatureRenderer) RenderItem(item block.Type, baseModel mgl32.Mat4, mainColor [3]float32) {
	// Wood color for handles
	handleColor := [3]float32{0.55, 0.35, 0.17}

	switch item {
	case block.Pickaxe:
		// Handle
		cr.renderCubePart(baseModel, mgl32.Vec3{0.08, 0.7, 0.08}, mgl32.Vec3{0, 0, 0}, handleColor)

		// Head (Curved top)
		// Center piece
		cr.renderCubePart(baseModel, mgl32.Vec3{0.2, 0.12, 0.12}, mgl32.Vec3{0, 0.35, 0}, mainColor)
		// Left spike
		cr.renderCubePart(baseModel, mgl32.Vec3{0.15, 0.1, 0.1}, mgl32.Vec3{-0.15, 0.32, 0}, mainColor)
		// Right spike
		cr.renderCubePart(baseModel, mgl32.Vec3{0.15, 0.1, 0.1}, mgl32.Vec3{0.15, 0.32, 0}, mainColor)
		// Tip L
		cr.renderCubePart(baseModel, mgl32.Vec3{0.1, 0.08, 0.08}, mgl32.Vec3{-0.25, 0.28, 0}, mainColor)
		// Tip R
		cr.renderCubePart(baseModel, mgl32.Vec3{0.1, 0.08, 0.08}, mgl32.Vec3{0.25, 0.28, 0}, mainColor)

	case block.Sword:
		// Handle
		cr.renderCubePart(baseModel, mgl32.Vec3{0.08, 0.25, 0.08}, mgl32.Vec3{0, -0.3, 0}, handleColor)

		// Guard
		cr.renderCubePart(baseModel, mgl32.Vec3{0.3, 0.05, 0.1}, mgl32.Vec3{0, -0.15, 0}, mainColor)

		// Blade
		cr.renderCubePart(baseModel, mgl32.Vec3{0.1, 0.7, 0.05}, mgl32.Vec3{0, 0.25, 0}, mainColor)

	case block.Axe:
		// Handle
		cr.renderCubePart(baseModel, mgl32.Vec3{0.08, 0.7, 0.08}, mgl32.Vec3{0, 0, 0}, handleColor)

		// Head
		// Connection
		cr.renderCubePart(baseModel, mgl32.Vec3{0.15, 0.15, 0.1}, mgl32.Vec3{0.05, 0.3, 0}, mainColor)
		// Blade part
		cr.renderCubePart(baseModel, mgl32.Vec3{0.1, 0.3, 0.1}, mgl32.Vec3{0.15, 0.3, 0}, mainColor)
		// Back
		cr.renderCubePart(baseModel, mgl32.Vec3{0.08, 0.1, 0.1}, mgl32.Vec3{-0.05, 0.3, 0}, mainColor)

	case block.Shovel:
		// Handle
		cr.renderCubePart(baseModel, mgl32.Vec3{0.08, 0.7, 0.08}, mgl32.Vec3{0, 0.1, 0}, handleColor)

		// Spade
		cr.renderCubePart(baseModel, mgl32.Vec3{0.25, 0.3, 0.05}, mgl32.Vec3{0, -0.35, 0}, mainColor)

	default:
		// Normal Cube Block
		cr.shader.SetMat4("uModel", baseModel)
		cr.shader.SetVec3("uColor", mgl32.Vec3{mainColor[0], mainColor[1], mainColor[2]})

		gl.BindVertexArray(cr.cubeVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
	}
}

// renderCubePart is a helper to render a scaled and translated part relative to base model
func (cr *CreatureRenderer) renderCubePart(baseModel mgl32.Mat4, size mgl32.Vec3, offset mgl32.Vec3, color [3]float32) {
	model := baseModel.Mul4(mgl32.Translate3D(offset.X(), offset.Y(), offset.Z()))
	model = model.Mul4(mgl32.Scale3D(size.X(), size.Y(), size.Z()))

	cr.shader.SetMat4("uModel", model)
	cr.shader.SetVec3("uColor", mgl32.Vec3{color[0], color[1], color[2]})

	gl.BindVertexArray(cr.cubeVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 36)
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
