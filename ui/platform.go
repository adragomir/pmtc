package ui

import (
	"fmt"
	"math"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/inkyblackness/imgui-go/v4"
)

// Platform covers mouse/keyboard/gamepad inputs, cursor shape, timing, windowing.
type Platform interface {
	// ShouldStop is regularly called as the abort condition for the program loop.
	ShouldStop() bool
	// ProcessEvents is called once per render loop to dispatch any pending events.
	ProcessEvents()
	// DisplaySize returns the dimension of the display.
	DisplaySize() [2]float32
	// FramebufferSize returns the dimension of the framebuffer.
	FramebufferSize() [2]float32
	// NewFrame marks the begin of a render pass. It must update the imgui IO state according to user input (mouse, keyboard, ...)
	NewFrame()
	// PostRender marks the completion of one render pass. Typically this causes the display buffer to be swapped.
	PostRender()
	// ClipboardText returns the current text of the clipboard, if available.
	ClipboardText() (string, error)
	// SetClipboardText sets the text as the current text of the clipboard.
	SetClipboardText(text string)

	AddAppCallback(tp string, f interface{})
}

type clipboard struct {
	platform Platform
}

func (board clipboard) Text() (string, error) {
	return board.platform.ClipboardText()
}

func (board clipboard) SetText(text string) {
	board.platform.SetClipboardText(text)
}

// GLFW implements a platform based on github.com/go-gl/glfw (v3.2).
type GLFW struct {
	imguiIO imgui.IO

	window *glfw.Window

	time             float64
	mouseJustPressed [3]bool

	callbacks map[string][]interface{}
}

const (
	windowWidth  = 1280
	windowHeight = 720

	mouseButtonPrimary   = 0
	mouseButtonSecondary = 1
	mouseButtonTertiary  = 2
	mouseButtonCount     = 3
)

// NewGLFW attempts to initialize a GLFW context.
func NewGLFW(io imgui.IO) (*GLFW, error) {
	runtime.LockOSThread()

	err := glfw.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize glfw: %w", err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, 1)
	glfw.WindowHint(glfw.ScaleToMonitor, glfw.True)
	glfw.WindowHint(glfw.CocoaRetinaFramebuffer, glfw.True)

	window, err := glfw.CreateWindow(windowWidth, windowHeight, "ImGui-Go GLFW3 example", nil, nil)
	if err != nil {
		glfw.Terminate()
		return nil, fmt.Errorf("failed to create window: %w", err)
	}
	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	platform := &GLFW{
		imguiIO: io,
		window:  window,

		callbacks: make(map[string][]interface{}),
	}

	platform.setKeyMapping()
	platform.installCallbacks()

	return platform, nil
}

// Dispose cleans up the resources.
func (platform *GLFW) Dispose() {
	platform.window.Destroy()
	glfw.Terminate()
}

// ShouldStop returns true if the window is to be closed.
func (platform *GLFW) ShouldStop() bool {
	return platform.window.ShouldClose()
}

// ProcessEvents handles all pending window events.
func (platform *GLFW) ProcessEvents() {
	glfw.PollEvents()
}

// DisplaySize returns the dimension of the display.
func (platform *GLFW) DisplaySize() [2]float32 {
	w, h := platform.window.GetSize()
	return [2]float32{float32(w), float32(h)}
}

// FramebufferSize returns the dimension of the framebuffer.
func (platform *GLFW) FramebufferSize() [2]float32 {
	w, h := platform.window.GetFramebufferSize()
	return [2]float32{float32(w), float32(h)}
}

// NewFrame marks the begin of a render pass. It forwards all current state to imgui IO.
func (platform *GLFW) NewFrame() {
	// Setup display size (every frame to accommodate for window resizing)
	displaySize := platform.DisplaySize()
	platform.imguiIO.SetDisplaySize(imgui.Vec2{X: displaySize[0], Y: displaySize[1]})

	// Setup time step
	currentTime := glfw.GetTime()
	if platform.time > 0 {
		platform.imguiIO.SetDeltaTime(float32(currentTime - platform.time))
	}
	platform.time = currentTime

	// Setup inputs
	if platform.window.GetAttrib(glfw.Focused) != 0 {
		x, y := platform.window.GetCursorPos()
		platform.imguiIO.SetMousePosition(imgui.Vec2{X: float32(x), Y: float32(y)})
	} else {
		platform.imguiIO.SetMousePosition(imgui.Vec2{X: -math.MaxFloat32, Y: -math.MaxFloat32})
	}

	for i := 0; i < len(platform.mouseJustPressed); i++ {
		down := platform.mouseJustPressed[i] || (platform.window.GetMouseButton(glfwButtonIDByIndex[i]) == glfw.Press)
		platform.imguiIO.SetMouseButtonDown(i, down)
		platform.mouseJustPressed[i] = false
	}
}

// PostRender performs a buffer swap.
func (platform *GLFW) PostRender() {
	platform.window.SwapBuffers()
}

func (platform *GLFW) setKeyMapping() {
	// Keyboard mapping. ImGui will use those indices to peek into the io.KeysDown[] array.
	platform.imguiIO.KeyMap(imgui.KeyTab, int(glfw.KeyTab))
	platform.imguiIO.KeyMap(imgui.KeyLeftArrow, int(glfw.KeyLeft))
	platform.imguiIO.KeyMap(imgui.KeyRightArrow, int(glfw.KeyRight))
	platform.imguiIO.KeyMap(imgui.KeyUpArrow, int(glfw.KeyUp))
	platform.imguiIO.KeyMap(imgui.KeyDownArrow, int(glfw.KeyDown))
	platform.imguiIO.KeyMap(imgui.KeyPageUp, int(glfw.KeyPageUp))
	platform.imguiIO.KeyMap(imgui.KeyPageDown, int(glfw.KeyPageDown))
	platform.imguiIO.KeyMap(imgui.KeyHome, int(glfw.KeyHome))
	platform.imguiIO.KeyMap(imgui.KeyEnd, int(glfw.KeyEnd))
	platform.imguiIO.KeyMap(imgui.KeyInsert, int(glfw.KeyInsert))
	platform.imguiIO.KeyMap(imgui.KeyDelete, int(glfw.KeyDelete))
	platform.imguiIO.KeyMap(imgui.KeyBackspace, int(glfw.KeyBackspace))
	platform.imguiIO.KeyMap(imgui.KeySpace, int(glfw.KeySpace))
	platform.imguiIO.KeyMap(imgui.KeyEnter, int(glfw.KeyEnter))
	platform.imguiIO.KeyMap(imgui.KeyEscape, int(glfw.KeyEscape))
	platform.imguiIO.KeyMap(imgui.KeyA, int(glfw.KeyA))
	platform.imguiIO.KeyMap(imgui.KeyC, int(glfw.KeyC))
	platform.imguiIO.KeyMap(imgui.KeyV, int(glfw.KeyV))
	platform.imguiIO.KeyMap(imgui.KeyX, int(glfw.KeyX))
	platform.imguiIO.KeyMap(imgui.KeyY, int(glfw.KeyY))
	platform.imguiIO.KeyMap(imgui.KeyZ, int(glfw.KeyZ))
}

func (platform *GLFW) installCallbacks() {
	platform.window.SetCursorPosCallback(platform.cursorPosChange)
	platform.window.SetMouseButtonCallback(platform.mouseButtonChange)
	platform.window.SetScrollCallback(platform.mouseScrollChange)
	platform.window.SetKeyCallback(platform.keyChange)
	platform.window.SetCharCallback(platform.charChange)
}

var glfwButtonIndexByID = map[glfw.MouseButton]int{
	glfw.MouseButton1: mouseButtonPrimary,
	glfw.MouseButton2: mouseButtonSecondary,
	glfw.MouseButton3: mouseButtonTertiary,
}

var glfwButtonIDByIndex = map[int]glfw.MouseButton{
	mouseButtonPrimary:   glfw.MouseButton1,
	mouseButtonSecondary: glfw.MouseButton2,
	mouseButtonTertiary:  glfw.MouseButton3,
}

func (platform *GLFW) cursorPosChange(window *glfw.Window, xPos float64, yPos float64) {
	if cbs, ok := platform.callbacks["CursorPos"]; ok {
		if cbs != nil && len(cbs) > 0 {
			for _, cb := range cbs {
				treated := cb.(func(float64, float64) bool)(xPos, yPos)
				if treated {
					break
				}
			}
		}
	}
}

func (platform *GLFW) mouseButtonChange(window *glfw.Window, rawButton glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	buttonIndex, known := glfwButtonIndexByID[rawButton]

	if known && (action == glfw.Press) {
		platform.mouseJustPressed[buttonIndex] = true
	}

	x, y := platform.window.GetCursorPos()
	if cbs, ok := platform.callbacks["MouseButton"]; ok {
		if cbs != nil && len(cbs) > 0 {
			for _, cb := range cbs {
				treated := cb.(func(float64, float64, glfw.MouseButton, glfw.Action, glfw.ModifierKey) bool)(x, y, rawButton, action, mods)
				if treated {
					break
				}
			}
		}
	}
}

func (platform *GLFW) mouseScrollChange(window *glfw.Window, x, y float64) {
	platform.imguiIO.AddMouseWheelDelta(float32(x), float32(y))

	xc, yc := platform.window.GetCursorPos()
	if cbs, ok := platform.callbacks["Scroll"]; ok {
		if cbs != nil && len(cbs) > 0 {
			for _, cb := range cbs {
				treated := cb.(func(float64, float64, float64, float64) bool)(xc, yc, x, y)
				if treated {
					break
				}
			}
		}
	}
}

func (platform *GLFW) keyChange(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		platform.imguiIO.KeyPress(int(key))
	}
	if action == glfw.Release {
		platform.imguiIO.KeyRelease(int(key))
	}

	// Modifiers are not reliable across systems
	platform.imguiIO.KeyCtrl(int(glfw.KeyLeftControl), int(glfw.KeyRightControl))
	platform.imguiIO.KeyShift(int(glfw.KeyLeftShift), int(glfw.KeyRightShift))
	platform.imguiIO.KeyAlt(int(glfw.KeyLeftAlt), int(glfw.KeyRightAlt))
	platform.imguiIO.KeySuper(int(glfw.KeyLeftSuper), int(glfw.KeyRightSuper))

	x, y := platform.window.GetCursorPos()
	if cbs, ok := platform.callbacks["Key"]; ok {
		if cbs != nil && len(cbs) > 0 {
			for _, cb := range cbs {
				treated := cb.(func(float64, float64, glfw.Key, int, glfw.Action, glfw.ModifierKey) bool)(x, y, key, scancode, action, mods)
				if treated {
					break
				}
			}
		}
	}
}

func (platform *GLFW) charChange(window *glfw.Window, char rune) {
	platform.imguiIO.AddInputCharacters(string(char))
	if !platform.imguiIO.WantCaptureKeyboard() {
		if cbs, ok := platform.callbacks["Char"]; ok {
			if cbs != nil && len(cbs) > 0 {
				for _, cb := range cbs {
					treated := cb.(func(rune) bool)(char)
					if treated {
						break
					}
				}
			}
		}
	}
}

// ClipboardText returns the current clipboard text, if available.
func (platform *GLFW) ClipboardText() (string, error) {
	return platform.window.GetClipboardString(), nil
}

// SetClipboardText sets the text as the current clipboard text.
func (platform *GLFW) SetClipboardText(text string) {
	platform.window.SetClipboardString(text)
}

func (platform *GLFW) AddAppCallback(tp string, f interface{}) {
	if _, ok := platform.callbacks[tp]; !ok {
		platform.callbacks[tp] = []interface{}{f}
	} else {
		platform.callbacks[tp] = append(platform.callbacks[tp], f)
	}
}
