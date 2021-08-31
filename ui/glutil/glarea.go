package glutil

type GlArea struct {
	Shader Shader

	Vao, Vbo uint32
	Fbo      uint32
	Rbo      uint32
	FboTex   uint32

	Size [2]float32
}

type ActiveGlArea struct {
	GlArea
	// top left, width, height
	// minx maxx, miny, maxy
	Bounds [4]float32
}

func (a ActiveGlArea) In(x, y float32) bool {
	return x >= a.Bounds[0] && x <= a.Bounds[1] &&
		y >= a.Bounds[2] && y <= a.Bounds[3]
}
