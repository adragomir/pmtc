package ui

import (
	"errors"
	"log"
	"sync"

	"github.com/adragomir/linuxcncgo/ui/glutil"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/inkyblackness/imgui-go/v4"
)

type GlPreview struct {
	glutil.ActiveGlArea

	helpers glutil.GlArea

	// ui interaction state
	modelChanged  bool
	sizeChanged   bool
	cameraChanged bool

	// draw state
	lock          sync.RWMutex
	hasModel      bool
	modelVertices []float32
	modelBbox     *glutil.BoundingBox

	// uniforms, camera
	camera        *glutil.Camera //mgl32.Mat4
	cameraControl *glutil.CameraControl

	model mgl32.Mat4

	active bool
}

func (p *GlPreview) buildPrograms() error {
	var vertexShader = `
	#version 330
	uniform mat4 projection;
	uniform mat4 view;
	uniform mat4 model;

	in vec3 vert;
	in float inLineType;

	flat out vec3 startPos;
	out vec3 vertPos;
	out float lineType; 

	void main() {
		vec4 pos = projection * view * model * vec4(vert.xyz, 1.0);
		gl_Position = pos;
		vertPos = pos.xyz / pos.w;
		startPos = vertPos;
		lineType = inLineType;
	}
	` + "\x00"

	var fragmentShader = `
	#version 330

	flat in vec3 startPos;
	in vec3 vertPos;
	in float lineType;

	out vec4 outputColor;

	uniform vec2 resolution;
	uniform float dashSize;
	uniform float gapSize;

	void main() {
    	// vec2  dir  = (vertPos.xy-startPos.xy) * resolution/2.0;
    	// float dist = length(dir);

    	// if (fract(dist / (dashSize + gapSize)) > dashSize/(dashSize + gapSize))
        	// discard; 
        if (lineType < 0.5) {
			outputColor = vec4(1.0, 0.0, 0.0, 1.0);
        } else {
			outputColor = vec4(0.0, 1.0, 0.0, 1.0);
        }
	}
	` + "\x00"

	// Configure the vertex and fragment shaders
	shader, err := glutil.NewShaderString(vertexShader, fragmentShader, "")
	if err != nil {
		log.Printf("Error building shader: %+v", err)
		return err
	}
	p.Shader = shader
	return nil
}

func (p *GlPreview) InitGL() error {
	err := p.buildPrograms()
	if err != nil {
		log.Printf("Error building OpenGl program: %+v ", err)
		return err
	}

	gl.UseProgram(p.Shader.Program)

	gl.Uniform1f(p.Shader.Uniforms["dashSize"], 5.0)
	gl.Uniform1f(p.Shader.Uniforms["gapSize"], 5.0)
	gl.Uniform2fv(p.Shader.Uniforms["resolution"], 1, &p.Size[0])

	gl.BindFragDataLocation(p.Shader.Program, 0, gl.Str("outputColor\x00"))

	// Configure the vertex data
	gl.GenVertexArrays(1, &p.Vao)
	gl.GenBuffers(1, &p.Vbo)

	gl.BindVertexArray(p.Vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, p.Vbo)

	gl.EnableVertexAttribArray(p.Shader.Attributes["vert"])
	gl.VertexAttribPointer(p.Shader.Attributes["vert"], 3, gl.FLOAT, false, 4*4, gl.PtrOffset(0))

	gl.EnableVertexAttribArray(p.Shader.Attributes["inLineType"])
	gl.VertexAttribPointer(p.Shader.Attributes["inLineType"], 1, gl.FLOAT, false, 4*4, gl.PtrOffset(3*4))

	return nil
}

func (p *GlPreview) NoData() {
	p.modelVertices = nil
	p.modelBbox = nil
	p.hasModel = false
}

func (p *GlPreview) SetData(vertices []float32, bbox *glutil.BoundingBox) {
	p.lock.Lock()
	p.modelVertices = vertices
	p.modelBbox = bbox
	p.modelChanged = true
	p.hasModel = true
	p.lock.Unlock()
	p.cameraChanged = true
	p.sizeChanged = true
}

func (p *GlPreview) Reshape(pos imgui.Vec2, size imgui.Vec2) {
	if p.Size[0] != size.X || p.Size[1] != size.Y {
		p.Size = [2]float32{size.X, size.Y}
		p.Bounds = [4]float32{
			pos.X,
			pos.X + size.X,
			pos.Y,
			pos.Y + size.Y,
		}
		p.sizeChanged = true
	}
}

// Draw implements the draw method
func (p *GlPreview) Draw() {
	if !p.hasModel {
		return
	}
	cameraChanged := false
	if p.modelChanged {
		// move model to origin
		// FIXME: we should have the cnc table box there
		translate := mgl32.Vec3{0, 0, 0}.Sub(p.modelBbox.Center())
		p.model = mgl32.Translate3D(translate[0], translate[1], translate[2])

		radius := p.modelBbox.SphereRadius()

		distance := radius/mgl32.Sin(mgl32.DegToRad(45.0)) + 30

		p.camera = glutil.NewPerspective(60, 1, 0.1, 1000)
		p.camera.SetPosition(mgl32.Vec3{0, 10, distance})
		p.camera.LookAt(mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})

		p.cameraControl = glutil.NewCameraControl(p.camera)
		p.cameraControl.SetTarget(mgl32.Vec3{0, 0, 0})
		p.camera.SetAspect(p.Size[0] / p.Size[1])
		p.cameraControl.SetAreaSize(p.Size[0], p.Size[1])

		cameraChanged = true

		// bind buffers
		gl.BindVertexArray(p.Vao)
		gl.BindBuffer(gl.ARRAY_BUFFER, p.Vbo)
		p.lock.RLock()
		gl.BufferData(gl.ARRAY_BUFFER, len(p.modelVertices)*4, gl.Ptr(p.modelVertices), gl.DYNAMIC_DRAW)
		p.lock.RUnlock()
	}

	if p.sizeChanged {
		glutil.DeleteRenderTarget(p.Fbo, p.FboTex, p.Rbo)
		fbo, fboTex, rbo := glutil.NewRenderTarget(int32(p.Size[0]), int32(p.Size[1]))
		p.Fbo = fbo
		p.FboTex = fboTex
		p.Rbo = rbo
		p.camera.SetAspect(p.Size[0] / p.Size[1])
		p.cameraControl.SetAreaSize(p.Size[0], p.Size[1])
		cameraChanged = true
	}

	// Here, we have a framebuffer
	gl.BindFramebuffer(gl.FRAMEBUFFER, p.Fbo)

	gl.UseProgram(p.Shader.Program)
	if cameraChanged || p.cameraChanged {
		gl.UniformMatrix4fv(p.Shader.Uniforms["projection"], 1, false, p.camera.ProjMatrixGl())
		gl.UniformMatrix4fv(p.Shader.Uniforms["view"], 1, false, p.camera.ViewMatrixGl())
	}
	if p.modelChanged {
		gl.UniformMatrix4fv(p.Shader.Uniforms["model"], 1, false, &p.model[0])
	}
	p.modelChanged = false
	p.cameraChanged = false
	p.sizeChanged = false

	gl.Scissor(int32(0), int32(0), int32(p.Size[0]), int32(p.Size[1]))
	gl.Enable(gl.SCISSOR_TEST)
	gl.Viewport(int32(0), int32(0), int32(p.Size[0]), int32(p.Size[1]))

	gl.ClearColor(0.0, 0.0, 0.0, 1)

	gl.Enable(gl.DEPTH_TEST)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.BindVertexArray(p.Vao)
	p.lock.RLock()
	gl.DrawArrays(gl.LINE_STRIP, 0, int32(len(p.modelVertices)/4))
	p.lock.RUnlock()

	gl.Disable(gl.DEPTH_TEST)
	gl.Disable(gl.SCISSOR_TEST)

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func (p *GlPreview) GetImage() (uint32, error) {
	if !p.hasModel {
		return 0, errors.New("Error: No model loaded for 3d preview")
	}
	if p.FboTex == 0 {
		return 0, errors.New("Error: no texture???!?")
	}
	return p.FboTex, nil
}

func (p *GlPreview) Reset() {
	p.camera = glutil.NewPerspective(60, 1, 0.1, 1000)
	radius := p.modelBbox.SphereRadius()
	distance := radius/mgl32.Sin(mgl32.DegToRad(45.0)) + 30

	p.camera.SetPosition(mgl32.Vec3{0, 10, distance})
	p.camera.LookAt(mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	p.camera.SetAspect(p.Size[0] / p.Size[1])

	p.cameraControl = glutil.NewCameraControl(p.camera)
	p.cameraControl.SetTarget(mgl32.Vec3{0, 0, 0})
	p.cameraControl.SetAreaSize(p.Size[0], p.Size[1])

	p.cameraChanged = true
	p.sizeChanged = true
}

// ui interaction
func (p *GlPreview) onCursorPos(x, y float64) bool {
	if p.In(float32(x), float32(y)) {
		p.cameraControl.OnCursorPos(x, y)
		p.cameraChanged = true
		return true
	}
	return false
}

func (p *GlPreview) onMouseButton(x, y float64, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) bool {
	if p.In(float32(x), float32(y)) {
		p.cameraControl.OnMouseButton(button, action, mods)
		p.cameraChanged = true
		return true
	}
	return false
}

func (p *GlPreview) onScroll(x, y float64, xoff float64, yoff float64) bool {
	if p.In(float32(x), float32(y)) {
		p.cameraControl.OnMouseScroll(xoff, yoff)
		p.cameraChanged = true
		return true
	}
	return false
}

func (p *GlPreview) onKey(x, y float64, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) bool {
	if p.In(float32(x), float32(y)) {
		p.cameraControl.OnKey(key, scancode, action, mods)
		p.cameraChanged = true
		return true
	}
	return false
}
