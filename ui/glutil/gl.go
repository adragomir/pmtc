package glutil

import "C"
import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

const GL_FLOAT32_SIZE = 4

func readFile(file string) ([]byte, error) {
	reader, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	return data, err
}

func FullProgram(vertexShaderSource, fragmentShaderSource, geometryShaderSource string) (uint32, error) {
	return createProgram([]byte(vertexShaderSource), []byte(fragmentShaderSource), []byte(geometryShaderSource))
}

func BasicProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func NewShaderString(vertSrc, fragSrc, geomSrc string) (Shader, error) {
	var shader Shader

	p, err := createProgram([]byte(vertSrc), []byte(fragSrc), []byte(geomSrc))
	if err != nil {
		return shader, err
	}
	shader = setupShader(p)
	return shader, nil
}

func NewShader(vertFile, fragFile, geomFile string) (Shader, error) {
	var shader Shader
	vertSrc, err := readFile(vertFile)
	if err != nil {
		return shader, err
	}

	fragSrc, err := readFile(fragFile)
	if err != nil {
		return shader, err
	}

	var geomSrc []byte
	if geomFile != "" {
		geomSrc, err = readFile(geomFile)
		if err != nil {
			return shader, err
		}
	}

	p, err := createProgram(vertSrc, fragSrc, geomSrc)
	if err != nil {
		return shader, err
	}
	shader = setupShader(p)
	return shader, nil
}

func setupShader(program uint32) Shader {
	var (
		c int32
		i uint32
	)
	gl.UseProgram(program)
	uniforms := make(map[string]int32)
	attributes := map[string]uint32{} //make(map[string]uint32)

	gl.GetProgramiv(program, gl.ACTIVE_UNIFORMS, &c)
	for i = 0; i < uint32(c); i++ {
		var buf [256]byte
		gl.GetActiveUniform(program, i, 256, nil, nil, nil, &buf[0])
		loc := gl.GetUniformLocation(program, &buf[0])
		name := gl.GoStr(&buf[0])
		uniforms[name] = loc
	}

	gl.GetProgramiv(program, gl.ACTIVE_ATTRIBUTES, &c)
	for i = 0; i < uint32(c); i++ {
		var buf [256]byte
		gl.GetActiveAttrib(program, i, 256, nil, nil, nil, &buf[0])
		loc := gl.GetAttribLocation(program, &buf[0])
		name := gl.GoStr(&buf[0])
		attributes[name] = uint32(loc)
	}

	return Shader{
		Program:    program,
		Uniforms:   uniforms,
		Attributes: attributes,
	}
}

func createProgram(v, f, g []byte) (uint32, error) {
	var p, vertex, frag, geom uint32
	use_geom := false

	if val, err := compileShader(string(v)+"\x00", gl.VERTEX_SHADER); err != nil {
		return 0, err
	} else {
		vertex = val
		defer func(s uint32) { deleteShader(p, s) }(vertex)
	}

	if val, err := compileShader(string(f)+"\x00", gl.FRAGMENT_SHADER); err != nil {
		return 0, err
	} else {
		frag = val
		defer func(s uint32) { deleteShader(p, s) }(frag)
	}

	if len(g) > 0 {
		if val, err := compileShader(string(g)+"\x00", gl.GEOMETRY_SHADER); err != nil {
			return 0, err
		} else {
			geom = val
			defer func(s uint32) { deleteShader(p, s) }(geom)
		}
	}

	p, err := linkProgram(vertex, frag, geom, use_geom)
	if err != nil {
		return 0, err
	}

	return p, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func deleteShader(p, s uint32) {
	gl.DetachShader(p, s)
	gl.DeleteShader(s)
}

func linkProgram(v, f, g uint32, use_geom bool) (uint32, error) {
	program := gl.CreateProgram()
	gl.AttachShader(program, v)
	gl.AttachShader(program, f)
	if use_geom {
		gl.AttachShader(program, g)
	}

	gl.LinkProgram(program)
	// check for program linking errors
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	return program, nil
}

type Shader struct {
	Program    uint32
	Uniforms   map[string]int32
	Attributes map[string]uint32
}

func (s *Shader) Delete() {
	gl.DeleteProgram(s.Program)
}

type VertexArray struct {
	Data          []float32
	Indices       []uint32
	Stride        int32
	Normalized    bool
	DrawMode      uint32
	Attributes    AttributesMap
	Vao, Vbo, Ebo uint32
}

func (v *VertexArray) Setup() {
	gl.GenVertexArrays(1, &v.Vao)
	fillVbo := true
	// Vbo already set when VertexArray was instancied.
	// This is a secondary structure using the same Vbo and vertex
	// data but with a different shader and attributes
	if v.Vbo == 0 {
		gl.GenBuffers(1, &v.Vbo)
	} else {
		fillVbo = false
	}
	if len(v.Indices) > 0 {
		gl.GenBuffers(1, &v.Ebo)
	}

	gl.BindVertexArray(v.Vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, v.Vbo)
	if fillVbo {
		gl.BufferData(gl.ARRAY_BUFFER, len(v.Data)*GL_FLOAT32_SIZE, gl.Ptr(v.Data), v.DrawMode)
	}

	if len(v.Indices) > 0 {
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, v.Ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(v.Indices)*GL_FLOAT32_SIZE, gl.Ptr(v.Indices), v.DrawMode)
	}

	for loc, ss := range v.Attributes {
		gl.EnableVertexAttribArray(loc)
		gl.VertexAttribPointer(loc, int32(ss[0]), gl.FLOAT, v.Normalized, v.Stride*GL_FLOAT32_SIZE, gl.PtrOffset(ss[1]*GL_FLOAT32_SIZE))
	}
	gl.BindVertexArray(0)
}

func (v *VertexArray) Delete() {
	gl.DeleteVertexArrays(1, &v.Vao)
	gl.DeleteBuffers(1, &v.Vbo)
	if len(v.Indices) > 0 {
		gl.DeleteBuffers(1, &v.Ebo)
	}
}

type AttributesMap map[uint32][2]int //map attrib loc to size / offset

func NewAttributesMap() AttributesMap {
	return make(AttributesMap)
}
func (am AttributesMap) Add(k uint32, size, offset int) {
	am[k] = [2]int{size, offset}
}

func ImageBytesToPixelData(image []byte) (*image.RGBA, error) {
	return ImageBufferToPixelData(bytes.NewBuffer(image))
}

func ImageFileToPixelData(file string) (*image.RGBA, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("texture %q not found on disk: %v", file, err)
	}
	defer imgFile.Close()
	return ImageBufferToPixelData(imgFile)

}

func ImageBufferToPixelData(r io.Reader) (*image.RGBA, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return nil, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)
	return rgba, nil
}

func NewRenderTarget(width, height int32) (uint32, uint32, uint32) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	// create a color attachment texture
	// FIXME: should be gl.RGB ????
	texture := NewTextureEmpty(width, height, gl.RGB, gl.LINEAR, gl.LINEAR)

	var rbo uint32
	gl.GenRenderbuffers(1, &rbo)
	gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH24_STENCIL8, width, height)

	// tie color
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)
	// tie renderbuffer
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, rbo)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		log.Fatalf("FRAMEBUFFER ERROR %d", gl.CheckFramebufferStatus(gl.FRAMEBUFFER))
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	return fbo, texture, rbo
}

func DeleteRenderTarget(fbo uint32, texture uint32, rbo uint32) {
	if fbo > 0 {
		gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
		// detach texture from fbo
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, 0, 0)
		// detach renderbuffer from fbo
		gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, 0)
		gl.DeleteTextures(1, &texture)
		gl.DeleteRenderbuffers(1, &rbo)
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.DeleteFramebuffers(1, &fbo)
	}
}

func NewTexture(wrap_s, wrap_t, min_f, mag_f int32, mipmap bool, file string) (uint32, error) {
	rgba, _ := ImageFileToPixelData(file)
	return NewTextureRGBA(wrap_s, wrap_t, min_f, mag_f, mipmap, rgba)
}

func NewTextureEmpty(width, height int32, format uint32, min_f, mag_f int32) uint32 {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, min_f)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, mag_f)
	gl.TexImage2D(
		gl.TEXTURE_2D, // target
		0,             // level
		int32(format), // internal format
		width,
		height,
		0,      // vorder
		format, // format
		gl.UNSIGNED_BYTE,
		nil,
	)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	return texture
}

func NewTextureRGBA(wrap_s, wrap_t, min_f, mag_f int32, mipmap bool, rgba *image.RGBA) (uint32, error) {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, wrap_s)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, wrap_t)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, min_f)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, mag_f)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))
	if mipmap {
		gl.GenerateMipmap(gl.TEXTURE_2D)
	}

	gl.BindTexture(gl.TEXTURE_2D, 0)

	return texture, nil
}

type Color struct {
	R, G, B, A float32
}

type Color32 struct {
	R, G, B, A float32
}

func (c *Color) To32() Color32 {
	return Color32{
		R: float32(c.R), G: float32(c.G),
		B: float32(c.B), A: float32(c.A)}
}

func RandColor() Color {
	return Color{rand.Float32(), rand.Float32(), rand.Float32(), 1.0}
}

func StepColor(c1, c2 Color, t, i int) Color {
	factorStep := 1 / (float32(t) - 1.0)
	c := interpolateColor(c1, c2, factorStep*float32(i))
	return Color{
		R: c.R,
		G: c.G,
		B: c.B,
	}
}

func interpolateColor(c1, c2 Color, factor float32) Color {
	result := new(Color)
	result.R = c1.R + factor*(c2.R-c1.R)
	result.G = c1.G + factor*(c2.G-c1.G)
	result.B = c1.B + factor*(c2.B-c1.B)
	return *result
}

func Rgb2Hex(c Color) string {
	rgb := []uint32{
		uint32(round(c.R*255, 0.5, 8)),
		uint32(round(c.G*255, 0.5, 8)),
		uint32(round(c.B*255, 0.5, 8)),
	}
	t := (rgb[0] << 16) + (rgb[1] << 8) + rgb[2]
	return "#" + fmt.Sprintf("%x", t)

}

func round(val float32, roundOn float32, places int) (newVal float32) {
	var round float32
	pow := mgl32.Pow(10, float32(places))
	digit := pow * val
	_, div := mgl32.Modf(digit)
	if div >= roundOn {
		round = mgl32.Ceil(digit)
	} else {
		round = mgl32.Floor(digit)
	}
	newVal = round / pow
	return
}

var Magenta = Color{R: 236.0 / 255.0, G: 0, B: 140.0 / 255.0}
var Black = Color{R: 0, G: 0, B: 0}
var White = Color{R: 1.0, G: 1.0, B: 1.0}

type OpenGLState struct {
	program   int32
	scissor   [4]int32
	viewport  [4]int32
	texture   int32
	lineWidth float32

	depthWriteMask bool
	depthFunc      int32

	polygonMode [2]int32

	cullFaceMode int32
	frontFace    int32

	blendSrcRgb   int32
	blendDstRgb   int32
	blendSrcAlpha int32
	blendDstAlpha int32

	scissorTest bool
	cullTest    bool
	depthTest   bool
	blend       bool
}

func SetOpenGLState(s *OpenGLState) {
	gl.UseProgram(uint32(s.program))
	gl.Scissor(s.scissor[0], s.scissor[1], s.scissor[2], s.scissor[3])
	gl.Viewport(s.viewport[0], s.viewport[1], s.viewport[2], s.viewport[3])
	gl.LineWidth(s.lineWidth)

	gl.DepthMask(s.depthWriteMask)
	gl.DepthFunc(uint32(s.depthFunc))

	gl.PolygonMode(uint32(s.polygonMode[0]), uint32(s.polygonMode[1]))
	gl.CullFace(uint32(s.cullFaceMode))
	gl.FrontFace(uint32(s.frontFace))

	gl.BlendFuncSeparate(uint32(s.blendSrcRgb), uint32(s.blendDstRgb), uint32(s.blendSrcAlpha), uint32(s.blendDstAlpha))
	if s.cullTest {
		gl.Enable(gl.CULL_FACE)
	} else {
		gl.Disable(gl.CULL_FACE)
	}
	if s.scissorTest {
		gl.Enable(gl.SCISSOR_TEST)
	} else {
		gl.Disable(gl.SCISSOR_TEST)
	}
	if s.blend {
		gl.Enable(gl.BLEND)
	} else {
		gl.Disable(gl.BLEND)
	}
	if s.depthTest {
		gl.Enable(gl.DEPTH_TEST)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}
}

func GetOpenGlState() *OpenGLState {
	var state OpenGLState
	gl.GetIntegerv(gl.CURRENT_PROGRAM, &state.program)

	gl.GetIntegerv(gl.SCISSOR_BOX, &state.scissor[0])
	gl.GetIntegerv(gl.VIEWPORT, &state.viewport[0])

	gl.GetIntegerv(gl.ACTIVE_TEXTURE, &state.texture)
	gl.GetFloatv(gl.LINE_WIDTH, &state.lineWidth)

	state.depthWriteMask = gl.IsEnabled(gl.DEPTH_WRITEMASK)
	//gl.GetIntegerv(gl.DEPTH_WRITEMASK, &state.depthWriteMask)
	gl.GetIntegerv(gl.DEPTH_FUNC, &state.depthFunc)

	gl.GetIntegerv(gl.POLYGON_MODE, &state.polygonMode[0])

	gl.GetIntegerv(gl.CULL_FACE_MODE, &state.cullFaceMode)
	gl.GetIntegerv(gl.FRONT_FACE, &state.frontFace)

	gl.GetIntegerv(gl.BLEND_SRC_RGB, &state.blendSrcRgb)
	gl.GetIntegerv(gl.BLEND_DST_RGB, &state.blendDstRgb)
	gl.GetIntegerv(gl.BLEND_SRC_ALPHA, &state.blendSrcAlpha)
	gl.GetIntegerv(gl.BLEND_DST_ALPHA, &state.blendDstAlpha)

	state.cullTest = gl.IsEnabled(gl.CULL_FACE)
	state.scissorTest = gl.IsEnabled(gl.SCISSOR_TEST)
	state.blend = gl.IsEnabled(gl.BLEND)
	state.depthTest = gl.IsEnabled(gl.DEPTH_TEST)
	return &state
}

func GetOpenGlBoundBuffers() (int32, int32) {
	var vab, abb int32
	gl.GetIntegerv(gl.VERTEX_ARRAY_BINDING, &vab)
	gl.GetIntegerv(gl.ARRAY_BUFFER_BINDING, &abb)
	return vab, abb
}

func SetOpenGlBoundBuffers(vab, abb int32) {
	gl.BindVertexArray(uint32(vab))
	gl.BindBuffer(gl.ARRAY_BUFFER, uint32(abb))
}

func GetOpenGLProgram() int32 {
	var program int32
	gl.GetIntegerv(gl.CURRENT_PROGRAM, &program)
	return program
}
