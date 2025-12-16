#version 450

layout(location=0) out vec4 fragColor;

layout(location=0) in struct in_dto {
    vec2 texcoord;
} dto;

// model uniforms max 128 bytes
layout(push_constant) uniform push_constants {
    mat4 model; // 64 bytes
    vec4 color; // 16 bytes
    vec4 args4; // 16 bytes
} mu;

// standard shadertoy variables expected by main
float iTime;      // in seconds
vec2 iResolution; // screen resolution in pixels

// constants for the end game tada effect.
const float bright = 0.001; // ring brightness.
const float radius = 0.068; // size of the tada rings.

void main() {
    iResolution = vec2(mu.args4.xy);
    iTime = mu.args4.z;
    vec2 fragCoord = gl_FragCoord.xy;
    vec2 uv = fragCoord/iResolution.xy;
    float seed = mu.args4.w;

    // primary input game color
    vec4 c1 = mu.color;

    // create a randomized swirly pattern based on supplied seed and color.
    // The loop iterations makes it swirly.
    float seedX = seed*9.0; // larger can slightly smooth the effect.
    float scale = 0.0037;   // smaller zooms in on repeating pattern.
    vec2 p = fragCoord * scale;
    for (int cnt=1; cnt<11; cnt++) {
        p.x += 0.37/float(cnt)* sin(float(cnt)*3.1*p.y+seedX) + seedX;
        p.y += 0.37/float(cnt)* cos(float(cnt)*3.1*p.x+seedX) + seedX;
    }

    // skew the swirl towards the given color and lighten it up a bit.
    float squash = 0.33;             // scale down sin/cos range from -1:1.
    float lighten = 0.7 + seed*0.25; // lighten up the colors.
    float r=cos(sin(p.x) - sin(p.y)+seedX*2.0)*squash+lighten;
    float g=sin(cos(p.x) + cos(p.y)-seedX*2.0)*squash+lighten;
    float b=(sin(cos(p.x) * cos(p.y)+seedX*2.0) - cos(sin(p.x) * sin(p.y) + seedX*2.0))*squash+lighten;
    vec4 swirl = vec4(r, g, b, c1.w);

    // darken the outer edges by mixing a darker color
    // with a radial gradient.
    vec4 dark = vec4(c1.xyz*0.15, c1.w);
    vec2 center = vec2(0.5);
    float d = distance(uv, center);
    fragColor = mix(swirl, dark, sin(d*2.4));

    // add the end game tada effect when alpa is less than 1.0.
    // These are a group of rings that move according to an axis based on game seed.
    if (c1.w < 0.999) {
        float rings = 78.0; // higher for more rings.

        // convert seed from a 0:1 range into a 2D axis
        vec2 axis = vec2(160.0*seed, 160.0*seed+(seed*seed));

        // create the moving rings by combining cos/sin as a vec2.
        // This is a unit circle when x==y and a complex closed curve when x!=y.
        // See: Lissajous curve on wikipedia.
        float ring = -0.3;
        vec2 centr = 1.6 * (fragCoord.xy * 2.0 - iResolution) / min(iResolution.x, iResolution.y);
        centr.y -= 0.28; // lower to fit the empty space in a winning board.
        for (float cnt = 0.0; cnt < rings; cnt++) {
            float si = sin(iTime + cnt * 0.5 * axis.x);
            float co = cos(iTime + cnt * 0.5 * axis.y);
            ring += bright / abs(length(centr + vec2(si, co)) - radius);
        }
        vec4 tada = vec4(vec3(fragColor.xyz) * ring, 1.0);

        // combine with the normal background.
        fragColor = mix(tada, fragColor, c1.w);
    }
}
