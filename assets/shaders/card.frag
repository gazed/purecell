#version 450

layout(location=0) out vec4 fragColor;

// samplers
const int COLOR = 0;
layout(set = 1, binding = 0) uniform sampler2D samplers[1];

layout(location=0) in struct in_dto {
    vec2 texcoord;
} dto;

// model uniforms max 128 bytes
layout(push_constant) uniform push_constants {
    mat4 model; // 64 bytes
	vec4 color; // 16 bytes
} mu;

void main() {
    vec2 uv = dto.texcoord;
    vec4 bgColor = texture(samplers[COLOR], uv);

    // create a gradient for the foreground color
    float dist = distance(uv, vec2(0.25, 0.5)) * 2.0;
    vec4 fgColor = mix(vec4(1.0,1.0,1.0,1.0), mu.color, dist);
    fragColor = bgColor * fgColor;
}
