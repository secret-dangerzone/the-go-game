package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"math"
)

type V2 struct {
	X float64
	Y float64
}

func (a V2) Add(b V2) (result V2) {
	result.X = a.X + b.X
	result.Y = a.Y + b.Y

	return result
}

func (a V2) Subtract(b V2) (result V2) {
	result.X = a.X - b.X
	result.Y = a.Y - b.Y

	return result
}

func (a V2) Multiply(b float64) (result V2) {
	result.X = a.X * b
	result.Y = a.Y * b

	return result
}

func (a V2) Rotate(r float64) (result V2) {
	result.X = a.X*math.Cos(r) - a.Y*math.Sin(r)
	result.Y = a.Y*math.Cos(r) + a.X*math.Sin(r)

	return result
}

func (a V2) ToPoint() sdl.Point {
	return sdl.Point{int32(a.X), int32(a.Y)}
}

func (a V2) ToPointOffset(offset V2) sdl.Point {
	return sdl.Point{int32(a.X + offset.X), int32(a.Y + offset.Y)}
}

type V2s []V2

func (a V2s) Rotate(r float64) (result V2s) {
	for _, v := range a {

		result = append(result,
			V2{
				X: v.X*math.Cos(r) - v.Y*math.Sin(r),
				Y: v.Y*math.Cos(r) + v.X*math.Sin(r),
			})
	}

	return result
}

func (a V2s) ToPointsOffset(offset V2) (result []sdl.Point) {
	for _, v := range a {
		result = append(result, v.ToPointOffset(offset))
	}

	return result
}

func (a V2s) Merge(b V2s) (result V2s) {
	for _, v := range a {
		result = append(result, v)
	}

	for _, v := range b {
		result = append(result, v)
	}

	return result
}
