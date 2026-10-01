/******************************************************************************/
/* fallback.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fallback

import "kaijuengine.com/editor/editor_areas"

// A non-exhaustive list of things that could go wrong with retrieving an Area:
// - Bad SessionRestore JSON syntax/names
// - Attempting to access a registry by ID before the registry has loaded
// - A typo in editor or plugin code
// - A plugin was uninstalled or updated

// The Fallback area is loaded whenever the workspace receives a request to open
// an AreaType by an ID that isn't mapped in the registry. The only significant
// feature of its UI content is that it'll display the plaintext ID whose
// aquisition failure prompted Fallback to be loaded in its place.

// Doing it this way allows session restores to be much more watertight in
// the event that the game or plugin developer breaks something unintentionally.
// If smething bad happens, it (theoretically) won't cause the session restore
// to fail.

func init() {
	editor_areas.Register(func() editor_areas.AreaType {
		return editor_areas.AreaType{
			ID:             "com.kaiju.area_fallback",
			Name:           "Fallback",
			Flags:          editor_areas.AreaCapabilityFlagDockable | editor_areas.AreaCapabilityFlagDraggable | editor_areas.AreaCapabilityFlagOverlayable,
			HandlerFactory: func() editor_areas.AreaHandler { return &Handler{} },
		}
	})
}

// TODO when the fallback area type gets serialized to the session restore file,
// serialize the failed ID instead so that the Workspace could be recovered
// completely in the event that the developer breaks and then fixes something
// between restarts.

type Handler struct {
	FailedID string
}

func (h *Handler) Open(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
}

func (h *Handler) Close(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
	area.TakedownLayout()
}

func (h *Handler) FocusInterface(editor editor_areas.EditorAreaInterface) {

}

func (h *Handler) BlurInterface(editor editor_areas.EditorAreaInterface) {

}

func (h *Handler) Update(area *editor_areas.Area, editor editor_areas.EditorAreaInterface, deltaTime float64, posx, posy, width, height float32) {

}

func (h *Handler) GetOffsets(ed editor_areas.EditorAreaInterface) (float32, float32, float32, float32) {
	return 0, 0, 0, 0
}

func (h *Handler) GetOverlayDimensions() (float32, float32) {
	return 256, 256
}

func (h *Handler) IsFocusedOnInput() bool {
	return false
}
