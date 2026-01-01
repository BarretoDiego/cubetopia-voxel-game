#version 410 core

// Attributes (Per Vertex)
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec2 aTexCoord;

// Attributes (Per Instance)
layout (location = 2) in vec3 aInstancePos;
layout (location = 3) in vec4 aInstanceColor;
layout (location = 4) in float aInstanceSize;

// Uniforms
uniform mat4 uView;
uniform mat4 uProjection;

// Outputs
out vec2 vTexCoord;
out vec4 vColor;

void main() {
    vTexCoord = aTexCoord;
    vColor = aInstanceColor;

    // Billboarding:
    // Extract camera right and up vectors from view matrix
    // View matrix is [Right, Up, Forward, Pos] (transposed or not depending on row/col major)
    // Actually, usually:
    // [ R.x U.x F.x 0 ]
    // [ R.y U.y F.y 0 ]
    // [ R.z U.z F.z 0 ]
    // [ ... ... ... 1 ]
    // So Right is row 0 or col 0. GLM/MGL is usually Column Major.
    // So uView[0] is Right, uView[1] is Up.
    
    vec3 right = vec3(uView[0][0], uView[1][0], uView[2][0]);
    vec3 up = vec3(uView[0][1], uView[1][1], uView[2][1]);
    
    // Scale the quad
    vec3 vertexPos = (right * aPos.x + up * aPos.y) * aInstanceSize;
    
    // Move to instance position
    vec3 finalPos = vertexPos + aInstancePos;
    
    gl_Position = uProjection * uView * vec4(finalPos, 1.0);
}
