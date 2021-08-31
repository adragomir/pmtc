package ui

import (
	"bufio"
	"bytes"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/adragomir/linuxcncgo/gcode"
	"github.com/adragomir/linuxcncgo/machine"
	"github.com/adragomir/linuxcncgo/network"
	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/inkyblackness/imgui-go/v4"
	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type UiState int

const (
	StateLoading UiState = iota
	StateLauncher
	StateMachine
	StateFiles
)

func convertPositionToMap(pos []float64) map[string]string {
	if len(pos) > 0 {
		return map[string]string{
			"X": fmt.Sprintf("%07.3f", pos[0]),
			"Y": fmt.Sprintf("%07.3f", pos[1]),
			"Z": fmt.Sprintf("%07.3f", pos[2]),
		}
	} else {
		return map[string]string{
			"X": "NaN",
			"Y": "NaN",
			"Z": "NaN",
		}
	}
}

func StartUi(services *machine.Services) {
	context := imgui.CreateContext(nil)
	defer context.Destroy()
	io := imgui.CurrentIO()

	platform, err := NewGLFW(io)
	if err != nil {
		log.Fatalf("Error initializing platform for Imgui: %v\n", err)
		os.Exit(-1)
	}
	defer platform.Dispose()

	renderer, err := NewOpenGL3(io)
	if err != nil {
		log.Fatalf("Error initializing renderer for imgui: %v\n", err)
		os.Exit(-1)
	}
	defer renderer.Dispose()

	imgui.CurrentIO().SetClipboard(clipboard{platform: platform})

	ui := NewUi(platform, renderer, services)

	fragments, accumulator, stats := gcode.LoadHardcodedGcodeFile()
	ui.gcodePreview.SetData(gcode.BuildVertexData(fragments, accumulator, stats, true))
	ui.gcodePreview.SetData(gcode.BuildVertexData(fragments, accumulator, stats, true))

	ui.services.Start()
	ui.Start()
}

type Ui struct {
	platform Platform
	renderer Renderer

	state    UiState
	services *machine.Services

	// ui state
	increments         []float64
	jogVelocity        float32
	feedOverride       float32
	rapidOverride      float32
	maxVelocity        float32
	maxMachineVelocity float32

	// ui custom components
	focusMdi                bool
	mdiHistory              []string
	mdiHistorySelectedValue string

	dimensions map[string][2]imgui.Vec2

	gcodePreview *GlPreview

	programContents []byte
}

func NewUi(platform Platform, renderer Renderer, services *machine.Services) *Ui {
	tmpUi := &Ui{
		platform: platform,
		renderer: renderer,

		state:    StateLoading,
		services: services,

		feedOverride:  1.0,
		rapidOverride: 1.0,

		dimensions:   make(map[string][2]imgui.Vec2),
		gcodePreview: &GlPreview{},
	}
	services.AddCb("launcherUpdate", func(l *machine.Launcher) {
		// firce imgui to refresh
		// for uuid, l := range tmpUi.services.Launchers {
		// 	if l.Complete {
		// 	}
		// }
		tmpUi.state = StateLauncher
	})
	services.AddCb("machineReady", func(m *machine.Machine) {
		tmpUi.increments = m.Increments()
		if m.ConfigState != nil {
			tmpUi.maxMachineVelocity = float32(m.ConfigState.GetMaxVelocity())
			tmpUi.jogVelocity = tmpUi.maxMachineVelocity / 2
			tmpUi.maxVelocity = tmpUi.maxMachineVelocity / 2
		}
		m.AddCb("configUpdateIncrements", func(incs []float64) {
			tmpUi.increments = incs
		})
		m.AddCb("configUpdateMaxVelocity", func(maxVel float64) {
			tmpUi.maxMachineVelocity = float32(maxVel)
			tmpUi.jogVelocity = float32(maxVel / 2)
			tmpUi.maxVelocity = float32(maxVel)
		})
	})
	tmpUi.gcodePreview.InitGL()

	tmpUi.platform.AddAppCallback("Key", func(x, y float64, k glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) bool {
		treated := tmpUi.gcodePreview.onKey(x, y, k, scancode, action, mods)
		if !treated {
			if action == glfw.Press || action == glfw.Release {
				return tmpUi.keyJogEvent(k, scancode, action, mods)
			}
		}
		return false
	})
	tmpUi.platform.AddAppCallback("CursorPos", func(x, y float64) bool {
		return tmpUi.gcodePreview.onCursorPos(x, y)
	})
	tmpUi.platform.AddAppCallback("MouseButton", func(x, y float64, b glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) bool {
		return tmpUi.gcodePreview.onMouseButton(x, y, b, action, mods)
	})
	tmpUi.platform.AddAppCallback("Scroll", func(x, y float64, xd, yd float64) bool {
		return tmpUi.gcodePreview.onScroll(x, y, xd, yd)
	})
	return tmpUi
}

const (
	sleepDuration = time.Millisecond * 25
)

func (ui *Ui) Start() {
	clearColor := [3]float32{0.0, 0.0, 0.0}
	for !ui.platform.ShouldStop() {
		ui.platform.ProcessEvents()

		// Signal start of a new frame
		ui.platform.NewFrame()
		imgui.NewFrame()

		ui.Layout()
		ui.AfterLayout()

		imgui.Render()
		ui.renderer.PreRender(clearColor)
		// A this point, the application could perform its own rendering...
		ui.RenderCustomComponents()

		ui.renderer.Render(
			ui.platform.DisplaySize(),
			ui.platform.FramebufferSize(),
			imgui.RenderedDrawData(),
		)
		ui.platform.PostRender()
		// sleep to avoid 100% CPU usage for this demo
		<-time.After(sleepDuration)
	}
}

func (ui *Ui) keyJogEvent(k glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) bool {
	switch k {
	case glfw.KeyF4:
		ui.focusMdi = true
		return true
	case glfw.KeyEscape:
	// X jog
	case glfw.KeyLeft:
		if action == glfw.Press {
			ui.services.ActiveMachine.XJogAction.SetVelocity(-ui.services.ActiveMachine.JogVelocity)
			ui.services.ActiveMachine.XJogAction.Trigger()
		} else {
			if ui.services.ActiveMachine.XJogAction.Distance == 0.0 {
				ui.services.ActiveMachine.XJogAction.SetVelocity(0.0)
				ui.services.ActiveMachine.XJogAction.Trigger()
			}
		}
		return true
	case glfw.KeyRight:
		if action == glfw.Press {
			ui.services.ActiveMachine.XJogAction.SetVelocity(ui.services.ActiveMachine.JogVelocity)
			ui.services.ActiveMachine.XJogAction.Trigger()
		} else {
			if ui.services.ActiveMachine.XJogAction.Distance == 0.0 {
				ui.services.ActiveMachine.XJogAction.SetVelocity(0.0)
				ui.services.ActiveMachine.XJogAction.Trigger()
			}
		}
		return true
	// Y jog
	case glfw.KeyDown:
		if action == glfw.Press {
			ui.services.ActiveMachine.YJogAction.SetVelocity(-ui.services.ActiveMachine.JogVelocity)
			ui.services.ActiveMachine.YJogAction.Trigger()
		} else {
			if ui.services.ActiveMachine.YJogAction.Distance == 0.0 {
				ui.services.ActiveMachine.YJogAction.SetVelocity(0.0)
				ui.services.ActiveMachine.YJogAction.Trigger()
			}
		}
		return true
	case glfw.KeyUp:
		if action == glfw.Press {
			ui.services.ActiveMachine.YJogAction.SetVelocity(ui.services.ActiveMachine.JogVelocity)
			ui.services.ActiveMachine.YJogAction.Trigger()
		} else {
			if ui.services.ActiveMachine.YJogAction.Distance == 0.0 {
				ui.services.ActiveMachine.YJogAction.SetVelocity(0.0)
				ui.services.ActiveMachine.YJogAction.Trigger()
			}
		}
		return true
	// Z jog
	case glfw.KeyPageUp:
		if action == glfw.Press {
			ui.services.ActiveMachine.ZJogAction.SetVelocity(ui.services.ActiveMachine.JogVelocity)
			ui.services.ActiveMachine.ZJogAction.Trigger()
		} else {
			if ui.services.ActiveMachine.ZJogAction.Distance == 0.0 {
				ui.services.ActiveMachine.ZJogAction.SetVelocity(0.0)
				ui.services.ActiveMachine.ZJogAction.Trigger()
			}
		}
		return true
	case glfw.KeyPageDown:
		if action == glfw.Press {
			ui.services.ActiveMachine.ZJogAction.SetVelocity(-ui.services.ActiveMachine.JogVelocity)
			ui.services.ActiveMachine.ZJogAction.Trigger()
		} else {
			if ui.services.ActiveMachine.ZJogAction.Distance == 0.0 {
				ui.services.ActiveMachine.ZJogAction.SetVelocity(0.0)
				ui.services.ActiveMachine.ZJogAction.Trigger()
			}
		}
		return true
	}
	return false
}

func (ui *Ui) Layout() {
	switch ui.state {
	case StateLoading:
	case StateLauncher:
		ui.platform.(*GLFW).window.SetTitle("Launchers")
		ui.LayoutLauncher()
	case StateMachine:
		ui.platform.(*GLFW).window.SetTitle("Machine")
		ui.LayoutMachine()
	case StateFiles:
		ui.platform.(*GLFW).window.SetTitle("Files")
		ui.LayoutFiles()
	default:
	}
}

func (ui *Ui) AfterLayout() {
	// tmp := ui.dimensions["gcodePreview"]
	// pos, size := tmp[0], tmp[1]
	// ui.gcodePreview.Reshape(pos, size)
}

func (ui *Ui) RenderCustomComponents() {
	// ui.gcodePreview.Draw()
	// if tex, err := ui.gcodePreview.GetImage(); err == nil {
	// outImage := image.NewRGBA(
	// 	image.Rect(
	// 		0, 0,
	// 		int(ui.dimensions["gcodePreview"][1].X), int(ui.dimensions["gcodePreview"][1].Y),
	// 	),
	// )
	// gl.ActiveTexture(gl.TEXTURE0)
	// gl.BindTexture(gl.TEXTURE_2D, tex)

	// gl.GetTexImage(gl.TEXTURE_2D,
	// 	0,
	// 	gl.RGBA,
	// 	gl.UNSIGNED_BYTE,
	// 	unsafe.Pointer(&outImage.Pix[0]))
	// f, _ := os.Create("out.png")
	// png.Encode(f, outImage)

	// posMax := imgui.Vec2{
	// 	X: ui.dimensions["gcodePreview"][0].X + ui.dimensions["gcodePreview"][1].X,
	// 	Y: ui.dimensions["gcodePreview"][0].Y + ui.dimensions["gcodePreview"][1].Y,
	// }

	// imgui.BackgroundDrawList().AddImageV(
	// 	imgui.TextureID(tex),
	// 	ui.dimensions["gcodePreview"][0],
	// 	posMax,
	// 	imgui.Vec2{X: 0, Y: 1},
	// 	imgui.Vec2{X: 1, Y: 0},
	// 	imgui.Packed(color.White),
	// )
	// }
}

func (ui *Ui) LayoutLauncher() {
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})
	imgui.SetNextWindowBgAlpha(0.)
	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: 0})
	tmp := ui.platform.DisplaySize()
	imgui.SetNextWindowSize(imgui.Vec2{X: tmp[0], Y: tmp[1]})
	imgui.BeginV("cncui", nil,
		imgui.WindowFlagsNoNav|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoScrollWithMouse|imgui.WindowFlagsNoScrollbar,
	)

	{

		for uuid, l := range ui.services.Launchers {
			if l.Complete {
				imgui.AlignTextToFramePadding()
				imgui.Text(l.Name)
				imgui.SameLineV(0, 20)
				ButtonDisabled("Start", l.Running, func() {
					l.Start()
				})
				imgui.SameLineV(0, 10)
				ButtonDisabled("Stop", !(!l.Terminating && l.Running), func() {
					//l.Stop()
				})
				imgui.SameLineV(0, 10)
				ButtonDisabled("Activate", !(!l.Terminating && l.Running), func() {
					ui.state = StateMachine
					ui.services.ActiveMachine = ui.services.Machines[uuid]
					//l.Stop()
				})
			}
		}

	}

	imgui.End()
	imgui.PopStyleVar()
}

func (ui *Ui) LayoutMachine() {
	machine := ui.services.ActiveMachine
	state := BuildMachineState(machine)

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})
	imgui.SetNextWindowBgAlpha(0.)
	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: 0})
	tmp := ui.platform.DisplaySize()
	imgui.SetNextWindowSize(imgui.Vec2{X: tmp[0], Y: tmp[1]})
	imgui.BeginV("cncui", nil,
		imgui.WindowFlagsNoNav|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsAlwaysAutoResize,
	)

	{
		if imgui.Button(">L") {
			ui.state = StateLauncher
		}
		imgui.SameLineV(0, 2)

		if state.estop {
			imgui.PushStyleColor(imgui.StyleColorButton, RGB(87, 153, 61).V())
			imgui.PushStyleColor(imgui.StyleColorButtonHovered, RGB(89, 179, 54).V())
			imgui.PushStyleColor(imgui.StyleColorButtonActive, RGB(87, 153, 61).V())
		} else {
			imgui.PushStyleColor(imgui.StyleColorButton, RGB(153, 61, 61).V())
			imgui.PushStyleColor(imgui.StyleColorButtonHovered, RGB(179, 54, 54).V())
			imgui.PushStyleColor(imgui.StyleColorButtonActive, RGB(153, 61, 61).V())
		}
		if imgui.Button("E-Stop") {
			if machine != nil {
				machine.ToggleEstopReset()
			}
		}
		imgui.PopStyleColorV(3)
		if imgui.IsItemHovered() {
			imgui.SetTooltip("F1 - toggle E-stop")
		}
		imgui.SameLineV(0, 2)

		if state.on {
			imgui.PushStyleColor(imgui.StyleColorButton, RGB(87, 153, 61).V())
			imgui.PushStyleColor(imgui.StyleColorButtonHovered, RGB(89, 179, 54).V())
			imgui.PushStyleColor(imgui.StyleColorButtonActive, RGB(87, 153, 61).V())
		} else {
			imgui.PushStyleColor(imgui.StyleColorButton, RGB(153, 61, 61).V())
			imgui.PushStyleColor(imgui.StyleColorButtonHovered, RGB(179, 54, 54).V())
			imgui.PushStyleColor(imgui.StyleColorButtonActive, RGB(153, 61, 61).V())
		}
		if imgui.Button("Power") {
			if machine != nil {
				machine.TogglePower()
			}
		}
		imgui.PopStyleColorV(3)
		if imgui.IsItemHovered() {
			imgui.SetTooltip("F2 - toggle Power")
		}

		imgui.SameLineV(0, 10)
		if imgui.Button(">Ftp") {
			cmd := exec.Command("open", fmt.Sprintf("ftp://anonymous:anonymous@%s", state.ftp))
			err := cmd.Run()
			if err != nil {
				log.Printf("Open FTP client error: %+v", err)
			}
		}
		if imgui.IsItemHovered() {
			imgui.SetTooltip("F3 - Open Transmit to FTP files to machine")
		}
		imgui.SameLineV(0, 2)
		if imgui.Button("Open Program") {
			ui.state = StateFiles
		}
		if state.program != "" {
			imgui.SameLineV(0, 2)
			if imgui.Button("CLOSE") {
				machine.CloseProgram()
				ui.programContents = []byte{}
				ui.gcodePreview.NoData()
			}
		}
		imgui.SameLineV(0, 20)

		ButtonDisabled("RUN", state.canRun, func() {
			ui.services.ActiveMachine.RunProgram("execute", 0)
		})
		imgui.SameLineV(0, 2)
		ButtonDisabled("PAUSE", state.canPause, func() {
			ui.services.ActiveMachine.PauseProgram("execute")
		})
		imgui.SameLineV(0, 2)
		ButtonDisabled("< STEP >", state.canStep, func() {
			ui.services.ActiveMachine.StepProgram("execute")
		})
		imgui.SameLineV(0, 2)
		ButtonDisabled("RESUME", state.canResume, func() {
			ui.services.ActiveMachine.ResumeProgram("execute")
		})
		imgui.SameLineV(0, 2)
		ButtonDisabled("STOP", state.canStop, func() {
			ui.services.ActiveMachine.Abort("execute")
		})

		imgui.SameLineV(0, 10)
		if imgui.Button("DEBUG") {
			log.Printf("DEBUG %+v", ui.dimensions)
		}
		imgui.SameLineV(0, 10)
		imgui.AlignTextToFramePadding()

		var text string
		if state.program != "" {
			imgui.PushStyleColor(imgui.StyleColorText, RGBA(255, 255, 0, 255).V())
			text = fmt.Sprintf("Program: %s (progress %f %%)", state.program, state.progress)
		} else {
			imgui.PushStyleColor(imgui.StyleColorText, RGBA(255, 255, 255, 255).V())
			text = "Program: NO PROGRAM"
		}
		imgui.Text(text)
		imgui.PopStyleColor()
	}
	// top 3 horizontal splits
	{
		// left X Y Z
		leftWidth := float32(150.0)
		mdiWidth := float32(300.0)
		actionsWidth := float32(150.0)
		threedWidth := imgui.ContentRegionAvail().X - leftWidth - mdiWidth - actionsWidth - 9

		imgui.BeginChildV("left", imgui.Vec2{X: float32(leftWidth), Y: imgui.ContentRegionAvail().Y - 200}, true, imgui.WindowFlagsNoScrollbar|imgui.WindowFlagsNoScrollWithMouse)
		imgui.BeginGroup()
		{

			nums := map[string]map[string]string{
				"Position":       state.pos,
				"Distance to go": state.dtg,
				"G92 Offset":     state.g92Offset,
				"G54 Offset":     state.g5XOffset,
			}
			for _, k := range []string{
				"Position", "Distance to go", "G92 Offset", "G54 Offset",
			} {
				v := nums[k]
				TextCenter(k)

				imgui.AlignTextToFramePadding()
				imgui.Text("X")
				imgui.SameLineV(0, 30)
				imgui.BeginDisabled()
				imgui.Button(v["X"])
				imgui.EndDisabled()

				imgui.Text("Y")
				imgui.SameLineV(0, 30)
				imgui.BeginDisabled()
				imgui.Button(v["Y"])
				imgui.EndDisabled()

				imgui.Text("Z")
				imgui.SameLineV(0, 30)
				imgui.BeginDisabled()
				imgui.Button(v["Z"])
				imgui.EndDisabled()
			}

		}
		imgui.EndGroup()
		imgui.EndChild()

		imgui.SameLineV(0, 2)
		// 3d view
		gcodePreviewSize := imgui.Vec2{X: float32(threedWidth), Y: imgui.ContentRegionAvail().Y - 200}
		if ui.gcodePreview.active {
			imgui.PushStyleColor(imgui.StyleColorBorder, RGBA(255, 0, 0, 255).V())
		} else {
			imgui.PushStyleColor(imgui.StyleColorBorder, RGBA(100, 100, 100, 255).V())
		}
		imgui.BeginChildV("3d", gcodePreviewSize, true, imgui.WindowFlagsNoScrollbar|imgui.WindowFlagsNoScrollWithMouse)
		{
			ui.gcodePreview.active = false
			if imgui.IsWindowHovered() {
				ui.gcodePreview.active = true
			}
			imgui.BeginGroup()
			ui.dimensions["gcodePreview"] = [2]imgui.Vec2{
				imgui.WindowPos(),
				imgui.WindowSize(),
			}

			ui.gcodePreview.Reshape(imgui.WindowPos(), imgui.WindowSize())
			ui.gcodePreview.Draw()
			if tex, err := ui.gcodePreview.GetImage(); err == nil {
				posMax := imgui.Vec2{
					X: ui.dimensions["gcodePreview"][0].X + ui.dimensions["gcodePreview"][1].X,
					Y: ui.dimensions["gcodePreview"][0].Y + ui.dimensions["gcodePreview"][1].Y,
				}

				imgui.WindowDrawList().AddImageV(
					imgui.TextureID(tex),
					ui.dimensions["gcodePreview"][0],
					posMax,
					imgui.Vec2{X: 0, Y: 1},
					imgui.Vec2{X: 1, Y: 0},
					imgui.Packed(color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
				)
				if imgui.Button("Reset") {
					ui.gcodePreview.Reset()
				}
				pos := ui.gcodePreview.camera.Position()
				imgui.Text(fmt.Sprintf("camera: %f %f %f", pos[0], pos[1], pos[2]))
				target := ui.gcodePreview.cameraControl.Target()
				imgui.Text(fmt.Sprintf("target: %f %f %f", target[0], target[1], target[2]))
			} else {
				imgui.Text("NO PROGRAM")
			}
			imgui.EndGroup()
		}
		imgui.EndChild()
		imgui.PopStyleColor()
		imgui.SameLineV(0, 2)
		// mdi history
		imgui.BeginChildV("mdi", imgui.Vec2{X: float32(mdiWidth), Y: imgui.ContentRegionAvail().Y - 200}, true, imgui.WindowFlagsNoScrollbar|imgui.WindowFlagsNoScrollWithMouse)
		imgui.BeginGroup()
		{
			cp := imgui.CursorPos()
			imgui.SetCursorPos(imgui.Vec2{X: cp.X + 5, Y: cp.Y})
			imgui.AlignTextToFramePadding()
			imgui.Text("MDI History(F4 to focus)")
			imgui.SameLineV(0, 10)
			if imgui.Button("Clear") {
			}
			imgui.Separator()
			heightToReserve := imgui.CurrentStyle().ItemSpacing().Y + imgui.FrameHeightWithSpacing()

			imgui.BeginChildV("mdiHistory", imgui.Vec2{X: 0, Y: -heightToReserve}, false, imgui.WindowFlagsHorizontalScrollbar)
			for _, command := range ui.mdiHistory {
				if imgui.Selectable(command) {
					ui.mdiHistorySelectedValue = command
					ui.focusMdi = true
				}
			}

			if imgui.ScrollY() >= imgui.ScrollMaxY() {
				imgui.SetScrollHereY(1.0)
			}
			imgui.EndChild()
		}
		imgui.Separator()
		var command string
		if ui.mdiHistorySelectedValue != "" {
			command = ui.mdiHistorySelectedValue
		}
		imgui.SetNextItemWidth(imgui.ContentRegionAvail().X - 10)
		if ui.focusMdi == true {
			imgui.SetKeyboardFocusHere()
			ui.focusMdi = false
		}
		if imgui.InputTextV("##mdiCommand", &command, imgui.InputTextFlagsEnterReturnsTrue|imgui.InputTextFlagsCallbackCompletion|imgui.InputTextFlagsCallbackHistory, func(d imgui.InputTextCallbackData) int32 {
			return 0
		}) {
			ui.mdiHistory = append(ui.mdiHistory, command)
			if machine != nil {
				machine.ExecuteMdi("execute", command)
			}
			ui.focusMdi = true
			ui.mdiHistorySelectedValue = ""
		}
		imgui.EndGroup()
		imgui.EndChild()
		imgui.SameLineV(0, 2)
		// right buttons
		imgui.BeginChildV("Right", imgui.Vec2{X: float32(actionsWidth), Y: imgui.ContentRegionAvail().Y - 200}, true, imgui.WindowFlagsNoScrollbar|imgui.WindowFlagsNoScrollWithMouse)
		imgui.BeginGroup()
		{
			for _, k := range []string{"X", "Y", "Z"} {
				TextCenter(k)
				if imgui.ButtonV("Home##home"+k, imgui.Vec2{X: 70}) {
					if machine != nil {
						machine.HomeAxis(k)
					}
				}
				imgui.SameLineV(0, 2)
				if imgui.ButtonV("Unhome##unhome"+k, imgui.Vec2{X: 70}) {
					if machine != nil {
						machine.UnhomeAxis(k)
					}
				}
			}

			cp := imgui.CursorPos()
			imgui.SetCursorPos(imgui.Vec2{X: cp.X, Y: cp.Y + 20})
			if imgui.ButtonV("OVERRIDE", imgui.Vec2{X: 140}) {
				if machine != nil {
					machine.OverrideLimits()
				}
			}

			TextCenter("Jog Distance")
			// FIXME

			style := imgui.CurrentStyle()
			windowVisibleX2 := imgui.WindowPos().X + imgui.WindowContentRegionMax().X
			for i, val := range ui.increments {
				imgui.PushID(fmt.Sprintf("%d", i))

				name := fmt.Sprintf("%.2f", val)
				if val == 0.0 {
					name = "Inf"
				}

				ButtonDisabled(name, machine.JogDistance == val, func() {
					machine.SetJogDistance(ui.increments[i])
				})
				lastButtonX2 := imgui.ItemRectMax().X
				nextButtonX2 := lastButtonX2 + style.ItemSpacing().X + 40
				if i+1 < len(ui.increments) && nextButtonX2 < windowVisibleX2 {
					imgui.SameLine()
				}
				imgui.PopID()
			}

			TextCenter("Jog Velocity")
			imgui.SetNextItemWidth(imgui.ContentRegionAvail().X - 10)
			imgui.SliderFloatV("##jogvel", &ui.jogVelocity, 0.0, ui.maxVelocity, "%03.0f mm/s", imgui.SliderFlagsAlwaysClamp)
			if imgui.IsItemDeactivated() {
				if machine != nil {
					machine.JogVelocity = float64(ui.jogVelocity)
				}
			}

			TextCenter("Feed Override")
			imgui.SetNextItemWidth(imgui.ContentRegionAvail().X - 10)

			imgui.SliderFloatV("##fo", &ui.feedOverride, state.minFo, state.maxFo, "%.1f", imgui.SliderFlagsAlwaysClamp)

			TextCenter("Rapid Override")
			rapidOverride := float32(1.0)
			imgui.SetNextItemWidth(imgui.ContentRegionAvail().X - 10)
			imgui.SliderFloatV("##ro", &rapidOverride, state.minFo, state.maxFo, "%.1f", imgui.SliderFlagsAlwaysClamp)

			TextCenter("Max Velocity")
			imgui.SetNextItemWidth(imgui.ContentRegionAvail().X - 10)
			imgui.SliderFloatV("##maxvel", &ui.maxVelocity, 0.0, ui.maxMachineVelocity, "%03.0f mm/s", imgui.SliderFlagsAlwaysClamp)
			if imgui.IsItemDeactivated() {
				ui.jogVelocity = float32(ui.maxVelocity / 2)
			}

			TextCenter("Set LCS")
			for _, name := range []string{"G54", "G55", "G56", "G57", "G58", "G59", "G59.1", "G59.2", "G59.3"} {
				if imgui.ButtonV("Set "+name, imgui.Vec2{X: imgui.ContentRegionAvail().X - 10}) {
					log.Printf("Set lcs %s to current position", name)
					if machine != nil {
						machine.SetLcsToCurrent(name)
					}
				}
			}
		}
		imgui.EndGroup()
		imgui.EndChild()
	}
	// bottom multi line
	{
	}
	imgui.BeginChildV("Bottom", imgui.Vec2{X: 0, Y: imgui.ContentRegionAvail().Y}, true, imgui.WindowFlagsNone)
	{
		if ui.programContents != nil && len(ui.programContents) > 0 {
			scanner := bufio.NewScanner(bytes.NewReader(ui.programContents))
			lines := bytes.Count(ui.programContents, []byte{'\n'})
			lineNoLength := len(fmt.Sprintf("%d", lines))
			fstr := "% " + fmt.Sprintf("%d", lineNoLength+1) + "d"
			i := 1
			for scanner.Scan() {
				if state.currentLine == i {
					imgui.PushStyleColor(imgui.StyleColorText, RGBA(255, 0, 0, 255).V())
				}
				imgui.Text(fmt.Sprintf(fstr, i))
				imgui.SameLineV(0, 20)
				imgui.Text(scanner.Text())
				if state.currentLine == i {
					imgui.SetScrollHereY(0.0)
					imgui.PopStyleColor()
				}
				i++
			}
		} else {
			imgui.Text("NO PROGRAM")
		}
	}
	imgui.EndChild()
	imgui.End()
	imgui.PopStyleVar()
}

func (ui *Ui) LayoutFiles() {
	var internal func([]network.FileEntry)
	flags := imgui.TreeNodeFlagsNone

	internal = func(files []network.FileEntry) {
		for _, f := range files {
			if f.Children == nil || len(f.Children) == 0 {
				imgui.TreeNodeV(f.Name, flags|imgui.TreeNodeFlagsLeaf)
				if imgui.IsItemHovered() && imgui.IsMouseDoubleClicked(0) {
					ui.selectRemoteFile(f.Path)
				}
				imgui.TreePop()
			} else {
				if imgui.TreeNodeV(f.Name, flags) {
					internal(f.Children)
					imgui.TreePop()
				}
			}
		}
	}
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})
	imgui.SetNextWindowBgAlpha(0.)
	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: 0})
	tmp := ui.platform.DisplaySize()
	imgui.SetNextWindowSize(imgui.Vec2{X: tmp[0], Y: tmp[1]})
	imgui.BeginV("cncui", nil,
		imgui.WindowFlagsNoNav|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoScrollWithMouse|imgui.WindowFlagsNoScrollbar,
	)

	{
		imgui.BeginChildV("left", imgui.Vec2{X: imgui.ContentRegionAvail().X, Y: imgui.ContentRegionAvail().Y}, true, 0)
		imgui.BeginGroup()
		if imgui.Button("< BACK") {
			ui.state = StateMachine
		}
		if ui.services.ActiveMachine != nil {
			files := ui.services.ActiveMachine.Files()
			internal(files)
		}

		// 		if imgui.TreeNodeV("one", flags) {
		// 			imgui.TreeNodeV("two", flags|imgui.TreeNodeFlagsLeaf)
		// 			imgui.TreePop()

		// 			imgui.TreePop()
		// 		}
		// if imgui.TreeNodeV("one", flags) {
		// 	imgui.TreeNodeV("two", flags|imgui.TreeNodeFlagsLeaf)
		// 	imgui.TreePop()
		// 	imgui.TreeNodeV("three", flags|imgui.TreeNodeFlagsLeaf)
		// 	imgui.TreePop()

		// 	imgui.TreeNodeV("four", flags|imgui.TreeNodeFlagsLeaf)
		// 	{
		// 		imgui.TreeNodeV("five", flags|imgui.TreeNodeFlagsLeaf)
		// 		imgui.TreePop()
		// 		imgui.TreeNodeV("six", flags|imgui.TreeNodeFlagsLeaf)
		// 		imgui.TreePop()
		// 		imgui.TreeNodeV("seven", flags|imgui.TreeNodeFlagsLeaf)
		// 		imgui.TreePop()
		// 	}
		// 	imgui.TreePop()

		// }
		// imgui.TreePop()
		imgui.EndGroup()
		imgui.EndChild()
	}

	imgui.End()
	imgui.PopStyleVar()
}

type MachineState struct {
	exists       bool
	ftp          string
	estop        bool
	on           bool
	program      string
	running      bool
	paused       bool
	motionExec   bool
	motionPaused bool

	canRun    bool
	canPause  bool
	canStep   bool
	canResume bool
	canStop   bool

	pos       map[string]string
	dtg       map[string]string
	g92Offset map[string]string
	g5XOffset map[string]string

	minFo float32
	maxFo float32

	currentLine int
	totalLines  int
	progress    float32
}

func BuildMachineState(m *machine.Machine) MachineState {
	tmp := MachineState{
		pos:       map[string]string{"X": "0.0", "Y": "0.0", "Z": "0.0"},
		dtg:       map[string]string{"X": "0.0", "Y": "0.0", "Z": "0.0"},
		g92Offset: map[string]string{"X": "0.0", "Y": "0.0", "Z": "0.0"},
		g5XOffset: map[string]string{"X": "0.0", "Y": "0.0", "Z": "0.0"},

		minFo: 0.3,
		maxFo: 1.4,

		currentLine: -1,
		totalLines:  -1,
		progress:    0,
	}

	if m == nil {
		return tmp
	} else {
		tmp.exists = true
		tmp.ftp = m.Dsn["file"][6:]

		if taskState := m.TaskState; taskState != nil {
			tmp.estop = taskState.GetTaskState() == pb.EmcTaskStateType_EMC_TASK_STATE_ESTOP
			tmp.on = taskState.GetTaskState() == pb.EmcTaskStateType_EMC_TASK_STATE_ON
			tmp.paused = taskState.GetTaskPaused() == 1
			tmp.totalLines = int(m.TaskState.GetTotalLines())
		}
		if motionState := m.MotionState; motionState != nil {
			tmp.motionExec = m.MotionState.GetState() == pb.RCS_STATUS_RCS_EXEC
			tmp.motionPaused = m.MotionState.GetPaused()
			tmp.pos = convertPositionToMap(m.GetPosition())
			tmp.dtg = convertPositionToMap(m.GetDtg())
			tmp.g92Offset = convertPositionToMap(m.GetG92Offset())
			tmp.g5XOffset = convertPositionToMap(m.GetG5XOffset())
			tmp.currentLine = int(m.MotionState.GetMotionLine())
		}
		if configState := m.ConfigState; configState != nil {
			tmp.minFo = float32(m.ConfigState.GetMinFeedOverride())
			tmp.maxFo = float32(m.ConfigState.GetMaxFeedOverride())
		}

		tmp.program = m.CurrentProgram
		tmp.running = m.Running()

		tmp.canRun = !(tmp.program != "" && tmp.on && !tmp.running && !tmp.motionExec)
		tmp.canPause = !(tmp.on && tmp.running && !tmp.motionPaused)
		tmp.canStep = !(tmp.program != "" && tmp.on && tmp.running && (tmp.paused && tmp.motionPaused))
		tmp.canResume = !(tmp.program != "" && tmp.on && tmp.running && tmp.paused)
		tmp.canStop = !(tmp.on && tmp.motionExec)

		if tmp.program != "" && tmp.on && tmp.motionExec {
			tmp.progress = float32(tmp.currentLine) / float32(tmp.totalLines) * 100.0
		}

		return tmp
	}
}

func (ui *Ui) loadRemoteFile(path string) {
	contents, err := ui.services.ActiveMachine.DownloadRemoteFile(path)
	if err != nil {
		log.Printf("Error downloading file: %+v", err)
		return
	}
	ui.programContents = contents
	fragments, accumulator, stats := gcode.SimulateGCode(string(contents))
	vertices, bbox := gcode.BuildVertexData(fragments, accumulator, stats, true)
	ui.gcodePreview.SetData(vertices, bbox)
}

func (ui *Ui) selectRemoteFile(path string) {
	rp := ui.services.ActiveMachine.GetRemotePath()
	ui.services.ActiveMachine.ExecuteProgram(rp + path)
	go ui.loadRemoteFile(path)
}
