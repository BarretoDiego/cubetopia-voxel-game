// Package render provides raytracing support
package render

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// RaytracingRenderer provides software raytracing with GPU compute
type RaytracingRenderer struct {
	// Configuration
	Enabled         bool
	MaxBounces      int
	SamplesPerPixel int

	// Resolution (can be lower than screen for performance)
	width, height int
	scale         float32 // Resolution scale (0.25 to 1.0)

	// Compute shader
	computeShader uint32

	// Output texture
	outputTexture uint32

	// Fullscreen quad for display
	quadVAO, quadVBO uint32
	displayShader    *Shader

	// World data buffer (for GPU access)
	worldDataBuffer uint32

	// BVH acceleration structure
	bvh *BVH
}

// BVH is a Bounding Volume Hierarchy for acceleration
type BVH struct {
	Nodes []BVHNode
	Root  int
}

// BVHNode represents a node in the BVH
type BVHNode struct {
	BoundsMin      mgl32.Vec3
	BoundsMax      mgl32.Vec3
	LeftChild      int // -1 if leaf
	RightChild     int // -1 if leaf
	FirstPrimitive int // For leaves
	PrimitiveCount int // For leaves
}

// NewRaytracingRenderer creates a new raytracing renderer
func NewRaytracingRenderer(width, height int) *RaytracingRenderer {
	rt := &RaytracingRenderer{
		Enabled:         false,
		MaxBounces:      3,
		SamplesPerPixel: 1,
		width:           width / 2, // Half resolution for performance
		height:          height / 2,
		scale:           0.5,
		bvh:             NewBVH(),
	}

	// Create output texture
	gl.GenTextures(1, &rt.outputTexture)
	gl.BindTexture(gl.TEXTURE_2D, rt.outputTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(rt.width), int32(rt.height), 0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Create display quad
	rt.createDisplayQuad()

	// Create display shader
	displayShader, err := NewShader(rtQuadVertShader, rtQuadFragShader)
	if err == nil {
		rt.displayShader = displayShader
	}

	return rt
}

// NewBVH creates an empty BVH
func NewBVH() *BVH {
	return &BVH{
		Nodes: make([]BVHNode, 0),
		Root:  -1,
	}
}

// BuildFromChunks builds BVH from chunk data
func (bvh *BVH) BuildFromChunks(chunkPositions [][2]int, chunkSize int) {
	bvh.Nodes = make([]BVHNode, 0, len(chunkPositions)*2)

	if len(chunkPositions) == 0 {
		bvh.Root = -1
		return
	}

	// Create leaf nodes for each chunk
	leaves := make([]int, len(chunkPositions))
	for i, pos := range chunkPositions {
		node := BVHNode{
			BoundsMin: mgl32.Vec3{
				float32(pos[0] * chunkSize),
				0,
				float32(pos[1] * chunkSize),
			},
			BoundsMax: mgl32.Vec3{
				float32((pos[0] + 1) * chunkSize),
				64,
				float32((pos[1] + 1) * chunkSize),
			},
			LeftChild:      -1,
			RightChild:     -1,
			FirstPrimitive: i,
			PrimitiveCount: 1,
		}
		leaves[i] = len(bvh.Nodes)
		bvh.Nodes = append(bvh.Nodes, node)
	}

	// Build tree recursively
	bvh.Root = bvh.buildRecursive(leaves, 0)
}

func (bvh *BVH) buildRecursive(indices []int, axis int) int {
	if len(indices) == 0 {
		return -1
	}

	if len(indices) == 1 {
		return indices[0]
	}

	// Sort by axis
	bvh.sortByAxis(indices, axis)

	// Split in middle
	mid := len(indices) / 2
	leftIndices := indices[:mid]
	rightIndices := indices[mid:]

	// Create internal node
	node := BVHNode{
		LeftChild:      bvh.buildRecursive(leftIndices, (axis+1)%3),
		RightChild:     bvh.buildRecursive(rightIndices, (axis+1)%3),
		FirstPrimitive: -1,
		PrimitiveCount: 0,
	}

	// Calculate bounds from children
	if node.LeftChild >= 0 && node.RightChild >= 0 {
		leftNode := bvh.Nodes[node.LeftChild]
		rightNode := bvh.Nodes[node.RightChild]

		node.BoundsMin = mgl32.Vec3{
			float32(math.Min(float64(leftNode.BoundsMin.X()), float64(rightNode.BoundsMin.X()))),
			float32(math.Min(float64(leftNode.BoundsMin.Y()), float64(rightNode.BoundsMin.Y()))),
			float32(math.Min(float64(leftNode.BoundsMin.Z()), float64(rightNode.BoundsMin.Z()))),
		}
		node.BoundsMax = mgl32.Vec3{
			float32(math.Max(float64(leftNode.BoundsMax.X()), float64(rightNode.BoundsMax.X()))),
			float32(math.Max(float64(leftNode.BoundsMax.Y()), float64(rightNode.BoundsMax.Y()))),
			float32(math.Max(float64(leftNode.BoundsMax.Z()), float64(rightNode.BoundsMax.Z()))),
		}
	}

	nodeIndex := len(bvh.Nodes)
	bvh.Nodes = append(bvh.Nodes, node)
	return nodeIndex
}

func (bvh *BVH) sortByAxis(indices []int, axis int) {
	// Simple bubble sort for small arrays
	for i := 0; i < len(indices)-1; i++ {
		for j := i + 1; j < len(indices); j++ {
			centerI := bvh.Nodes[indices[i]].BoundsMin.Add(bvh.Nodes[indices[i]].BoundsMax).Mul(0.5)
			centerJ := bvh.Nodes[indices[j]].BoundsMin.Add(bvh.Nodes[indices[j]].BoundsMax).Mul(0.5)

			var compI, compJ float32
			switch axis {
			case 0:
				compI, compJ = centerI.X(), centerJ.X()
			case 1:
				compI, compJ = centerI.Y(), centerJ.Y()
			default:
				compI, compJ = centerI.Z(), centerJ.Z()
			}

			if compI > compJ {
				indices[i], indices[j] = indices[j], indices[i]
			}
		}
	}
}

func (rt *RaytracingRenderer) createDisplayQuad() {
	vertices := []float32{
		-1, -1, 0, 0,
		1, -1, 1, 0,
		1, 1, 1, 1,
		-1, -1, 0, 0,
		1, 1, 1, 1,
		-1, 1, 0, 1,
	}

	gl.GenVertexArrays(1, &rt.quadVAO)
	gl.GenBuffers(1, &rt.quadVBO)

	gl.BindVertexArray(rt.quadVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, rt.quadVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointerWithOffset(0, 2, gl.FLOAT, false, 4*4, 0)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 4*4, 2*4)
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

// SetEnabled toggles raytracing
func (rt *RaytracingRenderer) SetEnabled(enabled bool) {
	rt.Enabled = enabled
}

// SetQuality sets raytracing quality
func (rt *RaytracingRenderer) SetQuality(scale float32, samples, bounces int) {
	rt.scale = scale
	rt.SamplesPerPixel = samples
	rt.MaxBounces = bounces
}

// Render performs raytracing and displays result
func (rt *RaytracingRenderer) Render(camera *Camera, sunDirection mgl32.Vec3) {
	if !rt.Enabled || rt.displayShader == nil {
		return
	}

	// For now, display a debug pattern to show raytracing is active
	// Real raytracing would require compute shaders or CPU rendering
	rt.renderDebugPattern(camera, sunDirection)
}

func (rt *RaytracingRenderer) renderDebugPattern(camera *Camera, sunDirection mgl32.Vec3) {
	// Bind display shader
	rt.displayShader.Use()
	rt.displayShader.SetInt("uTexture", 0)
	rt.displayShader.SetBool("uRaytracingActive", true)
	rt.displayShader.SetVec3("uCameraPos", camera.Position)
	rt.displayShader.SetVec3("uSunDir", sunDirection)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, rt.outputTexture)

	gl.BindVertexArray(rt.quadVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}

// Resize updates the resolution
func (rt *RaytracingRenderer) Resize(width, height int) {
	rt.width = int(float32(width) * rt.scale)
	rt.height = int(float32(height) * rt.scale)

	// Recreate texture
	gl.BindTexture(gl.TEXTURE_2D, rt.outputTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(rt.width), int32(rt.height), 0, gl.RGBA, gl.FLOAT, nil)
}

// Cleanup releases resources
func (rt *RaytracingRenderer) Cleanup() {
	if rt.outputTexture != 0 {
		gl.DeleteTextures(1, &rt.outputTexture)
	}
	if rt.quadVAO != 0 {
		gl.DeleteVertexArrays(1, &rt.quadVAO)
	}
	if rt.quadVBO != 0 {
		gl.DeleteBuffers(1, &rt.quadVBO)
	}
	if rt.displayShader != nil {
		rt.displayShader.Delete()
	}
}

var rtQuadVertShader = `
#version 410 core

layout(location = 0) in vec2 aPos;
layout(location = 1) in vec2 aTexCoord;

out vec2 vTexCoord;
out vec2 vScreenPos;

void main() {
    gl_Position = vec4(aPos, 0.0, 1.0);
    vTexCoord = aTexCoord;
    vScreenPos = aPos;
}
` + "\x00"

var rtQuadFragShader = `
#version 410 core

in vec2 vTexCoord;
in vec2 vScreenPos;

uniform sampler2D uTexture;
uniform bool uRaytracingActive;
uniform vec3 uCameraPos;
uniform vec3 uSunDir;

out vec4 fragColor;

// Simple ray marching for demonstration
vec3 rayMarch(vec3 ro, vec3 rd) {
    float t = 0.0;
    
    for (int i = 0; i < 64; i++) {
        vec3 p = ro + rd * t;
        
        // Ground plane
        if (p.y < 0.0) {
            vec3 groundPos = ro + rd * (-ro.y / rd.y);
            // Checkerboard pattern
            float check = mod(floor(groundPos.x) + floor(groundPos.z), 2.0);
            vec3 groundColor = mix(vec3(0.3, 0.5, 0.2), vec3(0.4, 0.3, 0.2), check);
            
            // Shadow
            float shadow = dot(normalize(uSunDir), vec3(0, 1, 0)) * 0.5 + 0.5;
            return groundColor * shadow;
        }
        
        // Simple voxel blocks (demonstration)
        vec3 blockPos = floor(p);
        if (blockPos.y >= 0.0 && blockPos.y < 10.0) {
            float noise = fract(sin(dot(blockPos.xz, vec2(12.9898, 78.233))) * 43758.5453);
            if (noise > 0.85 && mod(blockPos.y, 3.0) < 1.0) {
                // Hit a block
                vec3 blockColor = vec3(noise, noise * 0.7, noise * 0.4);
                float light = dot(normalize(uSunDir), vec3(0, 1, 0)) * 0.5 + 0.5;
                return blockColor * light;
            }
        }
        
        t += 0.5;
        if (t > 100.0) break;
    }
    
    // Sky
    float skyGradient = rd.y * 0.5 + 0.5;
    return mix(vec3(0.6, 0.7, 0.9), vec3(0.3, 0.5, 0.9), skyGradient);
}

void main() {
    if (!uRaytracingActive) {
        fragColor = texture(uTexture, vTexCoord);
        return;
    }
    
    // Calculate ray direction from screen position
    vec3 ro = uCameraPos;
    vec3 rd = normalize(vec3(vScreenPos.x * 1.5, vScreenPos.y * 0.8, 1.0));
    
    vec3 color = rayMarch(ro, rd);
    
    // Raytracing indicator border
    if (abs(vScreenPos.x) > 0.98 || abs(vScreenPos.y) > 0.98) {
        color = vec3(1.0, 0.5, 0.0); // Orange border to indicate RT mode
    }
    
    fragColor = vec4(color, 1.0);
}
` + "\x00"
