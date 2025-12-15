#version 450

layout(location=0) out vec4 fragColor;

// Samplers
const int COLOR = 0;
layout(set = 1, binding = 0) uniform sampler2D samplers[1];

layout(location=0) in struct in_dto {
    vec2 texcoord;
    vec4 color;
} dto;

void main() {
    fragColor = texture(samplers[COLOR], dto.texcoord);
    fragColor = fragColor * dto.color; 
}
