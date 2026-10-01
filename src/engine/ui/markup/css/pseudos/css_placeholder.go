/******************************************************************************/
/* css_placeholder.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pseudos

import (
	"errors"

	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p Placeholder) Process(elm *document.Element, value rules.SelectorPart) ([]*document.Element, error) {
	if !elm.IsInput() {
		return []*document.Element{}, errors.New("placeholder can only be applied to input elements")
	}
	return []*document.Element{elm}, nil
}
