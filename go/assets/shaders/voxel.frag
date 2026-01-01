#version 410 core

in vec3 vColor;
in vec3 vNormal;
in float vAO;
in vec3 vWorldPos;
in vec2 vTexCoord;
in float vMaterialId;
in float vTextureLayerId;

uniform vec3 uSunDirection;
uniform vec3 uCameraPos;
uniform float uTime;

// Time of Day uniforms
uniform float uSunIntensity;    // 0-1, brightness based on time
uniform vec3 uSkyColor;         // Dynamic sky color
uniform vec3 uAmbientColor;     // Ambient light color (warmer day, cooler night)
uniform vec3 uFogColor;         // Fog color matching sky

// Texture Array for block textures
uniform sampler2DArray uBlockAtlas;

out vec4 fragColor;

void main() {
    // Sample texture from the array using layer ID
    vec3 texCoord3D = vec3(vTexCoord, vTextureLayerId);
    vec4 texColor = texture(uBlockAtlas, texCoord3D);
    
    // Use texture color if available, otherwise fall back to vertex color
    vec3 objectColor = vColor;
    if (texColor.a > 0.01) {
        // Blend texture with vertex color (tint effect)
        objectColor = texColor.rgb * vColor;
    }
    
    // Enhanced Ambient Occlusion with non-linear curve
    float ao = 1.0 - pow(vAO, 1.5) * 0.5;
    
    // Directional lighting based on sun intensity
    float diffuse = max(dot(vNormal, uSunDirection), 0.0) * uSunIntensity;
    
    // Ambient light (uses uAmbientColor for warm/cool tint)
    float ambientStrength = 0.3 + (1.0 - uSunIntensity) * 0.15; // Slightly brighter ambient at night
    vec3 ambient = uAmbientColor * ambientStrength;
    
    // Specular highlight (Blinn-Phong)
    vec3 viewDir = normalize(uCameraPos - vWorldPos);
    vec3 halfDir = normalize(uSunDirection + viewDir);
    float spec = pow(max(dot(vNormal, halfDir), 0.0), 32.0) * uSunIntensity;
    
    // Material specific properties
    float specularStrength = 0.0;
    float fresnel = 0.0;
    float emissive = 0.0;
    
    if (vMaterialId == 3.0) { // Ice/Glass
        specularStrength = 0.9;
        // Fresnel effect for glass/ice
        fresnel = pow(1.0 - max(dot(vNormal, viewDir), 0.0), 3.0) * 0.4;
    } else if (vMaterialId == 2.0) { // Water/Liquid
        specularStrength = 0.6;
        fresnel = pow(1.0 - max(dot(vNormal, viewDir), 0.0), 2.0) * 0.3;
        
        if (vTextureLayerId == 11.0) { // Lava
            // Lava animation - slower and more viscous looking
            float lavaSpeed = 0.5;
            float pulse = sin(uTime * lavaSpeed + vWorldPos.x * 0.2 + vWorldPos.z * 0.2) * 0.1 + 0.9;
            objectColor.rgb *= pulse;
            // Add some "heat" variants
            objectColor.rgb += vec3(0.2, 0.05, 0.0) * (sin(uTime * 1.5 + vWorldPos.x * 0.8) * 0.5 + 0.5);
            emissive = 0.8 + (1.0 - uSunIntensity) * 0.2;
        } else { // Water
            // Water animation
            objectColor.rgb *= 0.95 + 0.05 * sin(uTime * 2.0 + vWorldPos.x * 0.5);
        }
    } else if (vMaterialId == 4.0) { // Stone (slight roughness)
        specularStrength = 0.15;
    }
    
    // Check for other emissive blocks (Campfire, Diamond ore, etc.)
    if (vTextureLayerId == 12.0) { // Campfire
        float flicker = sin(uTime * 10.0 + vWorldPos.x * 10.0) * 0.1 + 0.9;
        objectColor.rgb *= flicker;
        emissive = 0.7 + (1.0 - uSunIntensity) * 0.3;
    } else if (vTextureLayerId > 10.0 && vTextureLayerId != 11.0 && vTextureLayerId != 12.0) { // Diamond/Other ores
        emissive = 0.15 + (1.0 - uSunIntensity) * 0.2; // Glow more at night
    }
    
    vec3 specular = vec3(1.0) * spec * specularStrength;
    
    // Combine lighting with AO
    // Emissive blocks are less affected by ambient/diffuse light
    vec3 lighting = objectColor * (ambient + vec3(diffuse * 0.7)) * ao;
    lighting = mix(lighting, objectColor * ao, emissive * 0.8); // Blend to self-illumination
    
    lighting += specular * uSunIntensity; // Reduce specular at night
    lighting += fresnel * uSkyColor * 0.5; // Reflection tinted by sky
    lighting += objectColor * (emissive * 1.5); // Stronger glow effect
    
    // Distance fog (Atmospheric) with dynamic color
    float dist = length(uCameraPos - vWorldPos);
    float fogStart = 50.0;
    float fogEnd = 120.0; 
    float fogFactor = clamp((dist - fogStart) / (fogEnd - fogStart), 0.0, 1.0);
    fogFactor = fogFactor * fogFactor; // Quadratic falloff for more natural fog
    
    vec3 finalColor = mix(lighting, uFogColor, fogFactor);
    
    // Gamma correction for better color perception
    finalColor = pow(finalColor, vec3(1.0/2.2));
    
    // Transparency for Ice/Water/Glass
    float alpha = 1.0;
    if (vMaterialId == 3.0) { // Ice/Glass
        alpha = 0.75;
        // Small depth offset to prevent z-fighting between adjacent transparent surfaces
        gl_FragDepth = gl_FragCoord.z + 0.00001;
    } else if (vMaterialId == 2.0) { // Water
        alpha = 0.65;
        // Small depth offset to prevent z-fighting between adjacent water surfaces
        gl_FragDepth = gl_FragCoord.z + 0.00001;
    } else {
        gl_FragDepth = gl_FragCoord.z;
    }
    
    // Discard fully transparent pixels
    if (texColor.a < 0.1 && vMaterialId == 1.0) {
        discard;
    }
    
    fragColor = vec4(finalColor, alpha);
}
