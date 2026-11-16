#version 460
#define FRAGMENT_SHADER

#include "kaiju.glsl"
#include "Grid_shared.glsl"

layout(set = 0, binding = 1) readonly buffer GridData {
	uint GridFeatureFlags;
	float GridLevel;
	float FarClip;
    uint GridLineCount;
	vec4 ClipRect;
	vec4 Steps[OVERLAY_GRID_TOTAL_STEPS];
	vec4 Colors[4];
};

layout(location = 0) in vec3 pos;
layout(location = 1) in vec2 coord;                           
layout(location = 2) flat in float LineAlpha; 
layout(location = 3) flat in float LineEmphasis; 

void main() {
	bvec3 axes = bvec3(
		hasFlag(GridFeatureFlags, GRID_AXIS_X) && vec2IsApprox(pos.yz, vec2(0.0), 2e-3f),
		hasFlag(GridFeatureFlags, GRID_AXIS_Y) && vec2IsApprox(pos.xz, vec2(0.0), 2e-3f),
		hasFlag(GridFeatureFlags, GRID_AXIS_Z) && vec2IsApprox(pos.xy, vec2(0.0), 2e-3f)
	);

	if (!any(axes)) {
		outColor = Colors[0];
		outColor.a *= LineAlpha;
		outColor.a *= (0.85f * LineEmphasis + 0.15)
		float fade = 1.0f - min(1.0f, length(coord));
		outColor.a *= fade;
		vec3 V = cameraPosition.xyz - pos;
		float dist = length(V);
		V /= dist;
		if (pos.y == 0.0f)
			outColor.a *= 1.0f - pow(1.0f - abs(V.y), 3);
		outColor.a *= 1.0f - smoothstep(0.5f * FarClip, FarClip, dist);
		return;
	} else if (axes.x) outColor = Colors[1];
	else if (axes.y) outColor = Colors[2];
	else outColor = Colors[3];

	outColor.a *= LineEmphasis;
	vec3 V = cameraPosition.xyz - pos;
	float dist = length(V);
	V /= dist;
	outColor.a *= 1.0f - smoothstep(0.65f * FarClip, FarClip, dist);
}