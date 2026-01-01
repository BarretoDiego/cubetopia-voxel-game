#version 410 core

layout(location = 0) in vec3 aPosition;
layout(location = 1) in vec3 aNormal;
layout(location = 2) in vec3 aColor; // Keeps support for color-tinted blocks
layout(location = 3) in float aAO;
layout(location = 4) in vec2 aTexCoord; // New: Texture Coordinates
layout(location = 5) in float aMaterialId; // New: Material ID for special effects

uniform mat4 uProjection;
uniform mat4 uView;
uniform float uTime;
uniform vec3 uWindDir;
uniform float uWindStrength;

out vec3 vColor;
out vec3 vNormal;
out float vAO;
out vec3 vWorldPos;
out vec2 vTexCoord;
out float vMaterialId;

// Simple hash function for random offsets
float hash(vec2 p) {
    return fract(sin(dot(p, vec2(12.9898, 78.233))) * 43758.5453);
}

void main() {
    vec3 pos = aPosition;
    
    // Wind Effect (Material ID 2 = Grass/Foliage, 5 = Leaves)
    // We'll define these IDs properly later, but for now assuming non-zero is potentially swayable
    // Ideally we pass a flag or use specific IDs.
    if (aMaterialId == 2.0 || aMaterialId == 5.0) { 
        // Only sway top vertices of grass/leaves
        // Simplified check: if not bottom vertices (assuming unit cube 0-1)
        // Adjust logic based on actual mesh coordinates if needed.
        // For standard cubes, y is integer.
        
        float sway = sin(uTime * 2.0 + pos.x * 0.5 + pos.z * 0.5) * 0.1 * uWindStrength;
        
        // Apply wind to top vertices
        if (fract(pos.y) > 0.01) { // If y is not an integer base (heuristic)
             pos.x += sway;
             pos.z += sway * 0.5;
        }
    }

    vColor = aColor;
    vNormal = aNormal;
    vAO = aAO;
    vTexCoord = aTexCoord;
    vMaterialId = aMaterialId;
    vWorldPos = pos;
    
    gl_Position = uProjection * uView * vec4(pos, 1.0);
}
