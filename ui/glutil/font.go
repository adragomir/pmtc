package glutil

import (
	"encoding/json"
	"image"
	"log"
)

type Glyph struct {
	Char        rune    `json:"unicode"`
	Advance     float32 `json:"advance"`
	PlaneBounds struct {
		Left   float32 `json:"left"`
		Bottom float32 `json:"bottom"`
		Right  float32 `json:"right"`
		Top    float32 `json:"top"`
	} `json:"planeBounds"`
	AtlasBounds struct {
		Left   float32 `json:"left"`
		Bottom float32 `json:"bottom"`
		Right  float32 `json:"right"`
		Top    float32 `json:"top"`
	} `json:"atlasBounds"`
}

type MsdfFont struct {
	Atlas struct {
		Type          string `json:"type"`
		DistanceRange int    `json:"distanceRange"`
		Size          int    `json:"size"`
		Width         int    `json:"width"`
		Height        int    `json:"height"`
		YOrigin       string `json:"yOrigin"`
	} `json:"atlas"`
	Metrics struct {
		EmSize             float32 `json:"emSize"`
		LineHeight         float32 `json:"lineHeight"`
		Ascender           float32 `json:"ascender"`
		Descender          float32 `json:"descender"`
		UnderlineY         float32 `json:"underlineY"`
		UnderlineThickness float32 `json:"underlineThickness"`
	} `json:"metrics"`
	Glyphs  []Glyph `json:"glyphs"`
	Chars   map[rune]int
	Image   *image.RGBA
	TabSize int
}

func NewFont(desc []byte, image []byte) *MsdfFont {
	var font MsdfFont
	json.Unmarshal(desc, &font)

	var err error
	rgba, err := ImageBytesToPixelData(image)
	if err != nil {
		log.Printf("Error loading image: %+v", err)
	}
	font.Image = rgba

	font.Chars = make(map[rune]int)
	for idx, glyph := range font.Glyphs {
		font.Chars[glyph.Char] = idx
	}
	font.TabSize = 8
	return &font
}

type LineMap struct {
	VOffset float32 // the vertical offset negative from 0 where this line starts
	Start   int     // the index of the first vertex on the line
	End     int     // the index of the last vertex on the line
}

func (m *MsdfFont) Render(text []byte, scaleFactor float32, startY float32) ([]float32, float32, []LineMap, []float32) {
	texelWidth := float32(1.0) / float32(m.Atlas.Width)
	texelHeight := float32(1.0) / float32(m.Atlas.Height)
	fsScale := float32(1.0) / (m.Metrics.Ascender - m.Metrics.Descender)
	pixelScaleFactor := scaleFactor / fsScale
	x := float32(0.0)
	y := startY - fsScale*pixelScaleFactor*m.Metrics.Ascender
	vertices := make([]float32, 0)
	lineHeight := fsScale * pixelScaleFactor * m.Metrics.LineHeight
	lineMap := make([]LineMap, 0)
	lineIndex := 0
	startVertex := 0

	for _, bc := range text {
		c := rune(bc)
		if c == '\r' {
			continue
		}
		if c == '\n' {
			x = 0
			y -= fsScale * pixelScaleFactor * m.Metrics.LineHeight
			lineMap = append(lineMap, LineMap{
				VOffset: y,
				Start:   startVertex,
				End:     len(vertices),
			})
			startVertex = len(vertices) + 1
			lineIndex = lineIndex + 1
			continue
		}
		glyph := m.Glyphs[m.Chars[c]]
		if c != ' ' && c != '\t' {
			pl, pb, pr, pt := glyph.PlaneBounds.Left, glyph.PlaneBounds.Bottom, glyph.PlaneBounds.Right, glyph.PlaneBounds.Top
			il, ib, ir, it := glyph.AtlasBounds.Left, glyph.AtlasBounds.Bottom, glyph.AtlasBounds.Right, glyph.AtlasBounds.Top
			pl = (pl * fsScale * pixelScaleFactor) + x
			pb = (pb * fsScale * pixelScaleFactor) + y
			pr = (pr * fsScale * pixelScaleFactor) + x
			pt = (pt * fsScale * pixelScaleFactor) + y

			il = il * texelWidth
			// FIXME: this is needed but we can flip the image
			ib = 1.0 - (ib * texelHeight)
			ir = ir * texelWidth
			// FIXME: this is needed but we can flip the image
			it = 1.0 - (it * texelHeight)

			vertices = append(vertices,
				pl, pb, il, ib,
				pr, pb, ir, ib,
				pl, pt, il, it,
				pr, pt, ir, it,
				pl, pt, il, it,
				pr, pb, ir, ib,
			)
			x += fsScale * pixelScaleFactor * glyph.Advance
		} else {
			if c == ' ' {
				x += fsScale * pixelScaleFactor * glyph.Advance
			} else {
				x += float32(m.TabSize) * fsScale * pixelScaleFactor * glyph.Advance
			}
		}
	}
	lineMap = append(lineMap, LineMap{
		VOffset: y,
		Start:   startVertex,
		End:     len(vertices),
	})
	txRange := []float32{
		float32(m.Atlas.DistanceRange) * texelWidth,
		float32(m.Atlas.DistanceRange) * texelHeight,
	}
	return txRange, lineHeight, lineMap, vertices
}
