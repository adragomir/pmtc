package ui

import "github.com/inkyblackness/imgui-go/v4"

func TextCenter(txt string) {
	windowWidth := imgui.WindowSize().X
	textWidth := imgui.CalcTextSize(txt, false, -1).X
	cp := imgui.CursorPos()
	imgui.SetCursorPos(imgui.Vec2{X: (windowWidth - textWidth) * 0.5, Y: cp.Y})
	imgui.Text(txt)
}

func ButtonDisabled(txt string, cond bool, cb func()) {
	if cond {
		imgui.BeginDisabled()
	}
	if imgui.Button(txt) {
		cb()
	}
	if cond {
		imgui.EndDisabled()
	}

}

// func Color(R, G, B, A int) imgui.Vec4 {
// 	return imgui.Vec4{X: float32(R), Y: float32(G), Z: float32(B), W: float32(A)}
// }

type Color imgui.Vec4

func RGB(r, g, b int) Color {
	return RGBA(r, g, b, 255)
}

func RGBUA(rgba uint32) Color {
	sc := float32(1.0 / 255.0)

	return Color{
		X: float32((rgba>>0)&0xFF) * sc,
		Y: float32((rgba>>8)&0xFF) * sc,
		Z: float32((rgba>>16)&0xFF) * sc,
		W: float32((rgba>>24)&0xFF) * sc,
	}
}

func RGBA(r, g, b, a int) Color {
	sc := float32(1.0 / 255.0)
	return Color{
		X: float32(r) * sc,
		Y: float32(g) * sc,
		Z: float32(b) * sc,
		W: float32(a) * sc,
	}
}

func (c Color) V() imgui.Vec4 {
	return imgui.Vec4(c)
}

func Saturate(f float32) float32 {
	if f < 0.0 {
		return 0.0
	} else if f > 1.0 {
		return 1.0
	} else {
		return f
	}
}

func (c Color) U() uint32 {
	var out uint32
	out = uint32(Saturate(c.X)*255.0+0.5) << 0
	out |= uint32(Saturate(c.Y)*255.0+0.5) << 8
	out |= uint32(Saturate(c.Z)*255.0+0.5) << 16
	out |= uint32(Saturate(c.W)*255.0+0.5) << 24
	return out
}
