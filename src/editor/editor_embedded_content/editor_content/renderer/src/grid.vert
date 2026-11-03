#version 460
#define VERTEX_SHADER

#include "kaiju.glsl"
#include "grid_shared.glsl"

// https://github.com/blender/blender/blob/main/source/blender/draw/engines/overlay/shaders/overlay_grid_vert.glsl
layout(set = 0, binding = 1) readonly buffer GridData {
	uint GridFeatureFlags;
	float GridLevel;
	float FarClip;
    uint GridLineCount;
	vec4 ClipRect;
	vec4 steps[OVERLAY_GRID_TOTAL_STEPS];
};

layout(location = 0) out vec3 pos;
layout(location = 1) out vec2 coord;
layout(location = 2) flat out float VertexEmphasis; 

struct LineData {
	vec2 P;	
	uint axis; 
	uint level;
};

LineData decodeVertex(uint vertID) {
	LineData line;
	uint side = vertID & 0x1u;
	vertID = vertID >> 1u;
	line.axis = vertID & 0x1u;
	vertID = vertID >> 1u;
	line.level = OVERLAY_GRID_STEPS_DRAW - 1 - int(vertID / GridLineCount);
	vertID %= GridLineCount;
	const float halfLine = max(float(GridLineCount >> 1u), 1.0f);
	const float lineID = (float(vertID) - halfLine) / halfLine;
	line.P.x = lineID;
	line.P.y = side == 0 ? -1.0f : 1.0f;
	line.P.xy = line.axis == 0 ? line.P.xy : line.P.yx;
	return line;
}

bool isOccluded(LineData line, uint level) {
	if (hasFlag(GridFeatureFlags, GRID_SHOWN) && (line.level < OVERLAY_GRID_STEPS_DRAW - 1) && (level < OVERLAY_GRID_TOTAL_STEPS - 1)) {
		float occ = abs(line.P[line.axis]) / steps[level + 1][line.axis];
		if (isApprox(floor(occ), occ, 1e-4f))
			return true;
	}
	return false;
}

void main() {
	gl_Position = vec4(FLT_NAN);
	LineData line = decodeVertex(gl_VertexIndex);
	int lineLevel = int(GridLevel) + int(line.level) - (hasFlag(GridFeatureFlags, GRID_ALIGNED) ? 0 : 1);
	lineLevel = clamp(lineLevel, 0, OVERLAY_GRID_TOTAL_STEPS - 1);
	float size = steps[lineLevel][line.axis];
	if (size <= 0.0f)
		return;
	vec2 offset = vec2(cameraPosition.x, cameraPosition.z);
	if (hasFlag(GridFeatureFlags, GRID_SHOWN)) {
		if (line.axis != 0)
			offset.y = round(offset.y / size) * size;
		else offset.x = round(offset.x / size) * size;
	}

	coord = line.P;
	VertexEmphasis = clamp(line.level + 1.0f - fract(GridLevel), 0, 1);

	if (cameraPosition.w > 0) {
		// TODO offset alpha on the distance from the center of the viewport
		// for orthographic projections
	}

	line.P *= max(float(GridLineCount >> 1), 1.0f);
	line.P = offset + size * line.P;
	vec2 clipMin = vec2(-FLT_MAX), clipMax = vec2(FLT_MAX);
	if (hasFlag(GridFeatureFlags, GRID_SHOWN)) {
		clipMin = offset - ClipRect.xy;
		clipMax = offset + ClipRect.xy;
	}
	// if (all(lessThan(line.P, clipMin)) || all(greaterThan(line.P, clipMax)))
	// 	return;
	// line.P = clamp(line.P, clipMin, clipMax);

	pos = vec3(0.0f);
	if (hasFlag(GridFeatureFlags, GRID_SHOWN)) {
		if (hasFlag(GridFeatureFlags, GRID_PLANE_XY))
			pos.xy = line.P;
		else if (hasFlag(GridFeatureFlags, GRID_PLANE_XZ))
			pos.xz = line.P;
		else if (hasFlag(GridFeatureFlags, GRID_PLANE_YZ))
			pos.yz = line.P;
	} else if (hasFlag(GridFeatureFlags, GRID_AXES_SHOWN)) {
		const uint axisFlags[3] = {GRID_AXIS_X, GRID_AXIS_Y, GRID_AXIS_Z};
		if (!hasFlag(GridFeatureFlags, axisFlags[line.axis]))
			return;
		pos[line.axis] = line.P.x;
	}

	if (isOccluded(line, lineLevel))
		return;
	gl_Position = projection * view * vec4(pos, 1.0);
	gl_Position.z = 1.0f;
}
