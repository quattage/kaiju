
#define OVERLAY_GRID_TOTAL_STEPS 8
#define OVERLAY_GRID_STEPS_DRAW 3

#define GRID_ALIGNED (1u << 0u)
#define GRID_PLANE_XY (1u << 1u)
#define GRID_PLANE_XZ (1u << 2u)
#define GRID_PLANE_YZ (1u << 3u)
#define GRID_AXIS_X (1u << 4u)
#define GRID_AXIS_Y (1u << 5u)
#define GRID_AXIS_Z (1u << 6u)

#define GRID_SHOWN GRID_PLANE_XY | GRID_PLANE_XZ | GRID_PLANE_YZ
#define GRID_AXES_SHOWN GRID_AXIS_X | GRID_AXIS_Y | GRID_AXIS_Z

