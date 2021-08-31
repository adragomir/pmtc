package mgl32

import "math"

const Pi = math.Pi
const degreeToRadiansFactor = math.Pi / 180
const radianToDegreesFactor = 180.0 / math.Pi

// DegToRad converts a number from degrees to radians
func DegToRad(degrees float32) float32 {
	return degrees * degreeToRadiansFactor
}

// RadToDeg converts a number from radians to degrees
func RadToDeg(radians float32) float32 {
	return radians * radianToDegreesFactor
}

// ClampInt clamps x to the provided closed interval [a, b]
func ClampInt(x, a, b int) int {
	if x < a {
		return a
	}
	if x > b {
		return b
	}
	return x
}

func Sincos(a float32) (float32, float32) {
	s, c := math.Sincos(float64(a))
	return float32(s), float32(c)
}

func Acos(v float32) float32 {
	return float32(math.Acos(float64(v)))
}

func Asin(v float32) float32 {
	return float32(math.Asin(float64(v)))
}

func Atan(v float32) float32 {
	return float32(math.Atan(float64(v)))
}

func Atan2(y, x float32) float32 {
	return float32(math.Atan2(float64(y), float64(x)))
}

func Ceil(v float32) float32 {
	return float32(math.Ceil(float64(v)))
}

func Cos(v float32) float32 {
	return float32(math.Cos(float64(v)))
}

func Floor(v float32) float32 {
	return float32(math.Floor(float64(v)))
}

func Inf(sign int) float32 {
	return float32(math.Inf(sign))
}

func IsNaN(v float32) bool {
	return math.IsNaN(float64(v))
}

func Sin(v float32) float32 {
	return float32(math.Sin(float64(v)))
}

func Sqrt(v float32) float32 {
	return float32(math.Sqrt(float64(v)))
}

func Modf(f float32) (int float32, frac float32) {
	int64, frac64 := math.Modf(float64(f))
	return float32(int64), float32(frac64)
}

func Max(a, b float32) float32 {
	return float32(math.Max(float64(a), float64(b)))
}

func Min(a, b float32) float32 {
	return float32(math.Min(float64(a), float64(b)))
}

func Mod(a, b float32) float32 {
	return float32(math.Mod(float64(a), float64(b)))
}

func Pow(a, b float32) float32 {
	return float32(math.Pow(float64(a), float64(b)))
}

func Tan(v float32) float32 {
	return float32(math.Tan(float64(v)))
}
