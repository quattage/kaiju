/******************************************************************************/
/* units.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

type ConvertableUnits interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
	// The only thing that differentiates this generic from Number
	// is the omission of Complex
}

type Unit float64

type UnitConversionProvider func() float64

var ScreenUnitsDPMM UnitConversionProvider = func() float64 {
	return 3.7795276
}
var ScreenUnitsReferenceDPMM UnitConversionProvider = func() float64 {
	return 3.7795276
}
var ScreenUnitsScaleFactor UnitConversionProvider = func() float64 {
	return 1
}

func (u Unit) Pix() float64 {
	return MM2Pix(float64(u))
}

func (u Unit) FPix() float32 {
	return MM2Pix(float32(u))
}

func Pix2MMU(px float64) Unit {
	return Unit(Pix2MM(px))
}

func MM2Pix[T ConvertableUnits](mm T) T {
	scale := ScreenUnitsDPMM() / ScreenUnitsReferenceDPMM()
	scale *= ScreenUnitsScaleFactor()
	return mm * T(scale)
}

func Pix2MM[T ConvertableUnits](px T) T {
	scale := ScreenUnitsDPMM() / ScreenUnitsReferenceDPMM()
	scale *= ScreenUnitsScaleFactor()
	return px / T(scale)
}
