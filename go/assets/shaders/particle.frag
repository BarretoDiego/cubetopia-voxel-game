#version 410 core

in vec2 vTexCoord;
in vec4 vColor;

out vec4 FragColor;

// uniform sampler2D uTexture;

void main() {
    // Basic Circle Shape / Soft Particle
    vec2 center = vec2(0.5, 0.5);
    float dist = length(vTexCoord - center);
    
    // Soft edge
    float alpha = smoothstep(0.5, 0.3, dist);
    
    // Combine with instance color
    FragColor = vColor * alpha;
    
    // Discard transparent
    if (FragColor.a < 0.1) discard;
}
