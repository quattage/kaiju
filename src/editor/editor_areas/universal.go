package editor_areas

import (
	"fmt"
	"log/slog"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
)

var commonUI map[string]universalUIElement

type universalUIElement struct {
	doc *document.Document
	elm *document.Element
}

// GetUniversalUIElement is a helper that maintains an internal map of commonly
// used UI sub-elements that can be freely duplicated and used wherever. This
// function expects an HTML file called [id].html and an element in that html file
// tagged with "id=[id]" for proper retrieval. Once an element has been parsed
// from HTML once, it's stored and duplicated for subsequent requests.

// TODO this currently does not function because Kaiju currently lacks the
// ability to load a document from HTML without providing a manager and
// parenting the resulting document in-place. In order for this code to work as
// intended, there'd need to be a method similar to DocumentFromHTMLString that
// does the UI parsing, but doesn't prepare the elements or add them to a
// manager. This would enable this universal system to create orphaned documents
// that don't belong to any active UI so that they can be freely handed out to
// requesters. At the moment, the original copy participates in the UI that
// initially requested it, which makes this function essentially pointless.
func GetUniversalUIElement(man *ui.Manager, ed EditorAreaInterface, id string, withData any, functions ...func(*document.Element)) (*ui.UI, error) {
	if commonUI == nil {
		commonUI = make(map[string]universalUIElement)
	}
	preex, ok := commonUI[id]
	// if we've already loaded the HTML we hand over a copy of the mapped element
	if ok && preex.doc != nil && preex.elm != nil {
		if preex.doc.IsManagedBy(man) {
			dupe := preex.doc.DuplicateElement(preex.elm)
			return dupe.UI, nil
		}
	}
	css := ed.Theme().GetGlobalCSS(ed.Host())
	text, err := ed.Host().AssetDatabase().ReadText(id + ".html")
	if err != nil {
		return nil, err
	}
	var doc *document.Document
	if len(functions) <= 0 {
		doc = markup.DocumentFromHTMLString(man, text, css, withData, nil, nil)
	} else {
		var funcMap map[string]func(*document.Element) = mapFunctions(functions)
		doc = markup.DocumentFromHTMLString(man, text, css, withData, funcMap, nil)
	}
	elm, ok := doc.GetElementById(id)
	if !ok {
		return nil, fmt.Errorf("Couldn't find node called '%s' in %s.html", id, id)
	}
	commonUI[id] = universalUIElement{
		doc: doc,
		elm: elm,
	}
	// dupe even on first load so that the original copy is untouched
	dupe := doc.DuplicateElement(elm)
	slog.Debug(fmt.Sprintf("Loaded bespoke element '%s' from %s.html", id, id))
	return dupe.UI, nil
}
