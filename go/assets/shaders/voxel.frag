#version 410 core

in vec3 vColor;
in vec3 vNormal;
in float vAO;
in vec3 vWorldPos;
in vec2 vTexCoord;
in float vMaterialId;

uniform vec3 uSunDirection;
uniform vec3 uCameraPos;
uniform float uTime;

// Textures - functionality to be implemented
// uniform sampler2DArray uBlockAtlas; 
// uniform sampler2D uSkybox; 

out vec4 fragColor;

void main() {
    // Basic directional lighting
    float diffuse = max(dot(vNormal, uSunDirection), 0.0);
    float ambient = 0.4; // Slightly elevated ambient
    
    // Apply AO
    float ao = 1.0 - vAO * 0.4;
    
    vec3 objectColor = vColor;
    
    // Placeholder for texture sampling (requires texture system implementation)
    // vec4 texColor = texture(uBlockAtlas, vec3(vTexCoord, vMaterialId));
    // objectColor *= texColor.rgb;

    // Specular highlight (Blinn-Phong)
    vec3 viewDir = normalize(uCameraPos - vWorldPos);
    vec3 halfDir = normalize(uSunDirection + viewDir);
    float spec = pow(max(dot(vNormal, halfDir), 0.0), 32.0);
    
    // Material specific properties
    float specularStrength = 0.0;
    
    if (vMaterialId == 3.0) { // Ice/Glass
        specularStrength = 0.8;
    } else if (vMaterialId == 2.0) { // Water/Liquid
        specularStrength = 0.5;
    }
    
    vec3 specular = vec3(1.0) * spec * specularStrength;
    
    vec3 lighting = objectColor * (ambient + diffuse * 0.8) * ao + specular;
    
    // Distance fog (Linear)
    float dist = length(uCameraPos - vWorldPos);
    float fogStart = 40.0;
    float fogEnd = 90.0; 
    float fogFactor = clamp((dist - fogStart) / (fogEnd - fogStart), 0.0, 1.0);
    
    vec3 fogColor = vec3(0.6, 0.8, 1.0); // Match clear color
    
    vec3 finalColor = mix(lighting, fogColor, fogFactor);
    
    // Transparency for Ice/Water (Basic alpha for now)
    float alpha = 1.0;
    if (vMaterialId == 3.0) alpha = 0.7; // Ice
    if (vMaterialId == 2.0) alpha = 0.6; // Water
    
    fragColor = vec4(finalColor, alpha);
}
