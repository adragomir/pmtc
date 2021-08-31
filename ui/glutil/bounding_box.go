package glutil

import (
	"github.com/go-gl/mathgl/mgl32"
)

type Range mgl32.Vec2

func NewRange() *Range {
	tmp := new(Range)
	tmp[0] = mgl32.Inf(+1)
	tmp[1] = mgl32.Inf(-1)
	return tmp
}

func (r *Range) Push(n float32) {
	r[0] = mgl32.Min(r[0], n)
	r[1] = mgl32.Max(r[1], n)
}

type BoundingBox struct {
	X *Range
	Y *Range
	Z *Range
}

func NewBoundingBox() *BoundingBox {
	return &BoundingBox{
		X: NewRange(),
		Y: NewRange(),
		Z: NewRange(),
	}
}

func (b *BoundingBox) PushPolyline(polyline [][]mgl32.Vec3) {
	for _, line := range polyline {
		for _, point := range line {
			b.PushPoint(point)
		}
	}
}

func (b *BoundingBox) PushPoint(point mgl32.Vec3) {
	b.PushCoord(point.X(), point.Y(), point.Z())
}

func (b *BoundingBox) PushCoord(x float32, y float32, z float32) {
	b.X.Push(x)
	b.Y.Push(y)
	b.Z.Push(z)
}

func (b *BoundingBox) Min() mgl32.Vec3 {
	return mgl32.Vec3{float32(b.X[0]), float32(b.Y[0]), float32(b.Z[0])}
}

func (b *BoundingBox) Max() mgl32.Vec3 {
	return mgl32.Vec3{float32(b.X[1]), float32(b.Y[1]), float32(b.Z[1])}
}

func (b *BoundingBox) Center() mgl32.Vec3 {
	return mgl32.Vec3{float32(b.X[1]+b.X[0]) / 2.0, float32(b.Y[1]+b.Y[0]) / 2.0, float32(b.Z[1]+b.Z[0]) / 2.0}
}
func (b *BoundingBox) SphereRadius() float32 {
	return mgl32.Sqrt(
		mgl32.Pow(b.X[1]-b.Center()[0], 2) +
			mgl32.Pow(b.Y[1]-b.Center()[1], 2) +
			mgl32.Pow(b.Z[1]-b.Center()[2], 2),
	)
}

func (b *BoundingBox) Scale() float32 {
	dx := b.X[1] - b.X[0]
	dy := b.Y[1] - b.Y[0]
	dz := b.Z[1] - b.Z[0]
	scale := mgl32.Inf(-1)
	if scale < dx {
		scale = dx
	}
	if scale < dy {
		scale = dy
	}
	if scale < dz {
		scale = dz
	}
	return scale
}
