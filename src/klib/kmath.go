package klib

import "math"

const (
	fsignMask = 0x80000000
)

func FAbs(x float32) float32 {
	return math.Float32frombits(math.Float32bits(x) &^ fsignMask)
}
