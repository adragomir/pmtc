package machine

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adragomir/linuxcncgo/network"
	"github.com/adragomir/linuxcncgo/util"
	"github.com/machinekit/machinetalk_go/application"
	"github.com/machinekit/machinetalk_go/pathview"

	pb "github.com/machinekit/machinetalk_protobuf_go"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

type JogType int

const (
	StopJog JogType = iota
	ContinuousJog
	IncrementJog
)

type JogAction struct {
	triggered bool
	axis      uint32
	velocity  float64

	Distance        float64
	safeDistance    float64
	safeJogInterval time.Duration

	timer     *time.Ticker
	timerDone chan bool

	machine *Machine
}

func NewJogAction(m *Machine, axis uint32, absoluteVelocity float64) *JogAction {
	x := int64(50)
	sji := time.Duration(time.Duration(x) * time.Millisecond)
	return &JogAction{
		machine:         m,
		axis:            axis,
		velocity:        0.0,
		Distance:        0.0,
		safeJogInterval: sji,
		safeDistance:    0.0, //absoluteVelocity * float64(float64(x)/1000.0) * 1.2,
	}
}

func (j *JogAction) SetVelocity(vel float64) {
	j.velocity = vel
	j.safeDistance = 0.0 //math.Abs(vel) * 0.1 * 1.2
}

func (j *JogAction) restartTimer() {
	j.timer = time.NewTicker(j.safeJogInterval)
	go func() {
		for {
			select {
			case <-j.timerDone:
				return
			case <-j.timer.C:
				j.machine.jog(IncrementJog, j.axis, j.velocity, j.safeDistance)
			}
		}
	}()
}

func (j *JogAction) Trigger() {
	if j.velocity != 0.0 {
		if j.Distance == 0.0 {
			if j.safeDistance == 0.0 {
				if !j.triggered {
					j.machine.jog(ContinuousJog, j.axis, j.velocity, 0.0)
					j.triggered = true
				}
			} else {
				j.restartTimer()
			}
		} else {
			j.machine.jog(IncrementJog, j.axis, j.velocity, j.Distance)
		}
	} else {
		if j.timer != nil {
			j.timer.Stop()
		}
		j.machine.jog(StopJog, j.axis, 0.0, 0.0)
		j.triggered = false
	}
}

func (j *JogAction) Stop() {
	j.timerDone <- true
	if j.timer != nil {
		j.timer.Stop()
	}
	j.machine.jog(StopJog, j.axis, 0.0, 0.0)
	j.triggered = false
}

const (
	MotionChannel = 0x1
	ConfigChannel = 0x2
	IoChannel     = 0x4
	TaskChannel   = 0x8
	InterpChannel = 0x10
)

type Machine struct {
	*util.Callbacker
	uuid string

	Complete bool

	Dsn map[string]string

	config  *application.ConfigBase
	command *application.CommandBase
	err     *application.ErrorBase
	status  *application.StatusBase
	preview *pathview.PreviewClientBase

	ftp    *network.FtpConn
	files  []network.FileEntry
	fMutex sync.RWMutex

	TaskState   *pb.EmcStatusTask
	MotionState *pb.EmcStatusMotion
	IoState     *pb.EmcStatusIo
	InterpState *pb.EmcStatusInterp
	ConfigState *pb.EmcStatusConfig
	UiState     *pb.EmcStatusUI

	syncedChannels int
	running        bool

	remoteFilePath string
	CurrentProgram string

	increments []float64

	XJogAction *JogAction
	YJogAction *JogAction
	ZJogAction *JogAction

	JogDistance float64
	JogVelocity float64
}

func buildIncrements(tmp string) []float64 {
	out := make([]float64, 0)
	for _, v := range strings.Split(tmp, " ") {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			out = append(out, n)
		}
	}
	return out
}

func (m *Machine) TryBuilding(s *Services) {
	if len(m.Dsn) < 11 {
		log.Printf("Machine not ready, still %d services to go ....", 11-len(m.Dsn))
		return
	}
	log.Printf("Machine ready %s", m.uuid)

	m.command = application.NewCommandBase(0, "command")
	m.command.OnCommandMsgReceived = append(m.command.OnCommandMsgReceived, func(rx *pb.Container, rest ...interface{}) {
		//log.Printf("COMMAND message: %s", prototext.Format(rx))
	})
	m.command.SetCommandUri(m.Dsn["command"])
	m.command.Start()

	m.err = application.NewErrorBase(0, "error")
	m.err.ErrorChannel.AddSocketTopic("error")
	m.err.ErrorChannel.AddSocketTopic("text")
	m.err.ErrorChannel.AddSocketTopic("display")
	m.err.ErrorChannel.SocketUri = m.Dsn["error"]
	m.err.OnErrorMsgReceived = append(m.err.OnErrorMsgReceived, func(rx *pb.Container, rest ...interface{}) {
		log.Printf("COMMAND ERROR %s", prototext.Format(rx))
	})
	m.err.Start()

	// application configs, etc - uninteresting
	m.config = application.NewConfigBase(0, "config")
	m.config.SetConfigUri(m.Dsn["config"])
	m.config.OnConfigMsgReceived = append(m.config.OnConfigMsgReceived, func(rx *pb.Container, rest ...interface{}) {
	})
	m.config.Start()

	m.status = application.NewStatusBase(0, "status")
	m.status.AddStatusTopic("io")
	m.status.AddStatusTopic("task")
	m.status.AddStatusTopic("interp")
	m.status.AddStatusTopic("motion")
	m.status.AddStatusTopic("config")
	m.status.AddStatusTopic("ui")
	m.status.SetStatusUri(m.Dsn["status"])
	m.status.OnStatusMsgReceived = append(m.status.OnStatusMsgReceived, func(rx *pb.Container, rest ...interface{}) {
		if *rx.Type == pb.ContainerType_MT_EMCSTAT_FULL_UPDATE {
			switch rest[0].(string) {
			case "task":
				m.TaskState = rx.GetEmcStatusTask()
				m.syncedChannels |= TaskChannel
			case "motion":
				m.MotionState = rx.GetEmcStatusMotion()
				m.syncedChannels |= MotionChannel
			case "io":
				m.IoState = rx.GetEmcStatusIo()
				m.syncedChannels |= IoChannel
			case "interp":
				m.InterpState = rx.GetEmcStatusInterp()
				m.syncedChannels |= InterpChannel
			case "ui":
				m.UiState = rx.GetEmcStatusUi()
			case "config":
				m.ConfigState = rx.GetEmcStatusConfig()
				m.syncedChannels |= ConfigChannel

				m.remoteFilePath = m.ConfigState.GetRemotePath()
				m.increments = buildIncrements(m.ConfigState.GetIncrements())
				m.increments = append(m.increments, 0)
				m.RunCbs("configUpdateIncrements", m.increments)
				m.RunCbs("configUpdateMaxVelocity", m.ConfigState.GetMaxVelocity())
				m.SetJogVelocity(m.ConfigState.GetMaxVelocity() / 2.0)
			}
		} else if *rx.Type == pb.ContainerType_MT_EMCSTAT_INCREMENTAL_UPDATE {
			switch rest[0].(string) {
			case "task":
				if m.TaskState != nil {
					proto.Merge(m.TaskState, rx.GetEmcStatusTask())
				}
			case "motion":
				if m.MotionState != nil {
					proto.Merge(m.MotionState, rx.GetEmcStatusMotion())
				}
			case "io":
				if m.IoState != nil {
					proto.Merge(m.IoState, rx.GetEmcStatusIo())
				}
			case "interp":
				if m.InterpState != nil {
					proto.Merge(m.InterpState, rx.GetEmcStatusInterp())
				}
			case "config":
				if m.ConfigState != nil {
					proto.Merge(m.ConfigState, rx.GetEmcStatusConfig())
				}
			}
		} else {
			log.Printf("Got type: %+v", *rx.Type)
		}
	})
	m.status.Start()

	m.preview = pathview.NewPreviewClientBase(0, "preview")
	m.preview.SetPreviewUri(m.Dsn["preview"])
	m.preview.SetPreviewstatusUri(m.Dsn["previewstatus"])
	m.preview.AddPreviewTopic("preview")
	m.preview.AddPreviewstatusTopic("preview")
	m.preview.AddPreviewstatusTopic("previewstatus")
	m.preview.OnPreviewMsgReceived = append(m.preview.OnPreviewMsgReceived, func(rx *pb.Container, rest ...interface{}) {
	})
	m.preview.OnPreviewstatusMsgReceived = append(m.preview.OnPreviewstatusMsgReceived, func(rx *pb.Container, rest ...interface{}) {
	})
	m.preview.Start()

	log.Printf("BUILDING MACHINE %s", m.uuid)

	m.ftp = network.NewFtpConn(m.Dsn["file"][6:])
	m.ftp.AddCb("filesReady", func(files []network.FileEntry) {
		m.fMutex.Lock()
		m.files = network.CopyFileEntries(files)
		m.fMutex.Unlock()
	})

	m.finishBuilding()

	m.Complete = true
}

func (m *Machine) Synced() bool {
	return m.syncedChannels == MotionChannel|IoChannel|ConfigChannel|TaskChannel|InterpChannel
}

func (m *Machine) Increments() []float64 {
	return m.increments
}

func (m *Machine) Running() bool {
	if m.TaskState != nil && m.InterpState != nil {
		taskModeCorrect := (m.TaskState.GetTaskMode() == pb.EmcTaskModeType_EMC_TASK_MODE_AUTO) ||
			(m.TaskState.GetTaskMode() == pb.EmcTaskModeType_EMC_TASK_MODE_MDI)
		return taskModeCorrect && m.InterpState.GetInterpState() != pb.EmcInterpStateType_EMC_TASK_INTERP_IDLE
	}
	return false
}

func (m *Machine) Files() []network.FileEntry {
	m.fMutex.RLock()
	defer m.fMutex.RUnlock()
	return m.files
}

func (m *Machine) finishBuilding() {
	m.XJogAction = NewJogAction(m, 0, m.JogVelocity)
	m.YJogAction = NewJogAction(m, 1, m.JogVelocity)
	m.ZJogAction = NewJogAction(m, 2, m.JogVelocity)
}

func (m *Machine) SetJogDistance(distance float64) {
	m.JogDistance = distance
	m.XJogAction.Distance = m.JogDistance
	m.YJogAction.Distance = m.JogDistance
	m.ZJogAction.Distance = m.JogDistance
}

func (m *Machine) SetJogVelocity(vel float64) {
	m.JogVelocity = vel
	// m.xJogAction.setVelocity(vel)
	// m.yJogAction.setVelocity(vel)
	// m.zJogAction.setVelocity(vel)
}

func (m *Machine) GetTaskStateObject() (*pb.EmcStatusTask, error) {
	if m.TaskState == nil {
		return nil, errors.New("empty")
	}
	return m.TaskState, nil
}

func (m *Machine) GetTaskState() (*pb.EmcTaskStateType, error) {
	if m.TaskState == nil {
		return nil, errors.New("empty")
	}
	return m.TaskState.GetTaskState().Enum(), nil
}

func (m *Machine) GetPosition() []float64 {
	if m.MotionState == nil {
		return []float64{}
	} else {
		return []float64{
			*m.MotionState.ActualPosition.X,
			*m.MotionState.ActualPosition.Y,
			*m.MotionState.ActualPosition.Z,
		}
	}
}

func (m *Machine) GetDtg() []float64 {
	if m.MotionState == nil {
		return []float64{}
	} else {
		return []float64{
			*m.MotionState.Dtg.X,
			*m.MotionState.Dtg.Y,
			*m.MotionState.Dtg.Z,
		}
	}
}

func (m *Machine) GetG92Offset() []float64 {
	if m.MotionState == nil {
		return []float64{}
	} else {
		return []float64{
			*m.MotionState.G92Offset.X,
			*m.MotionState.G92Offset.Y,
			*m.MotionState.G92Offset.Z,
		}
	}
}

func (m *Machine) GetG5XOffset() []float64 {
	if m.MotionState == nil {
		return []float64{}
	} else {
		return []float64{
			*m.MotionState.G5XOffset.X,
			*m.MotionState.G5XOffset.Y,
			*m.MotionState.G5XOffset.Z,
		}
	}
}

func (m *Machine) DownloadRemoteFile(p string) ([]byte, error) {
	m.ftp.Ensure(false)
	buf, err := m.ftp.Retr(p)
	if err == nil {
		return buf, err
	} else {
		log.Printf("ERROR downloading remote file: %+v", err)
		return []byte{}, err
	}
}

var axisIndex map[string]uint32 = map[string]uint32{
	"X": 0,
	"Y": 1,
	"Z": 2,
}

func (m *Machine) HomeAxis(axis string) {
	m.setTaskMode("execute", pb.EmcTaskModeType_EMC_TASK_MODE_MANUAL)
	msg := &pb.Container{
		EmcCommandParams: &pb.EmcCommandParameters{
			Index: util.UI32(axisIndex[axis]),
		},
	}
	m.command.SendEmcAxisHome(msg)
}

func (m *Machine) UnhomeAxis(axis string) {
	m.setTaskMode("execute", pb.EmcTaskModeType_EMC_TASK_MODE_MANUAL)
	msg := &pb.Container{
		EmcCommandParams: &pb.EmcCommandParameters{
			Index: util.UI32(axisIndex[axis]),
		},
	}
	m.command.SendEmcAxisUnhome(msg)
}

func (m *Machine) OverrideLimits() {
	m.setTaskMode("execute", pb.EmcTaskModeType_EMC_TASK_MODE_MANUAL)
	msg := &pb.Container{
		EmcCommandParams: &pb.EmcCommandParameters{},
	}
	m.command.SendEmcAxisOverrideLimits(msg)
}

func (m *Machine) ToggleEstopReset() {
	var newState *pb.EmcTaskStateType
	if *m.TaskState.TaskState == pb.EmcTaskStateType_EMC_TASK_STATE_ESTOP {
		newState = pb.EmcTaskStateType_EMC_TASK_STATE_ESTOP_RESET.Enum()
	} else {
		newState = pb.EmcTaskStateType_EMC_TASK_STATE_ESTOP.Enum()
	}
	msg := &pb.Container{
		InterpName: util.S("execute"),
		EmcCommandParams: &pb.EmcCommandParameters{
			LineNumber: util.I32(0),
			TaskState:  newState,
		},
	}
	m.command.SendEmcTaskSetState(msg)
}
func (m *Machine) TogglePower() {
	var newState *pb.EmcTaskStateType
	if *m.TaskState.TaskState == pb.EmcTaskStateType_EMC_TASK_STATE_ON {
		newState = pb.EmcTaskStateType_EMC_TASK_STATE_OFF.Enum()
	} else {
		newState = pb.EmcTaskStateType_EMC_TASK_STATE_ON.Enum()
	}
	msg := &pb.Container{
		InterpName: util.S("execute"),
		EmcCommandParams: &pb.EmcCommandParameters{
			LineNumber: util.I32(0),
			TaskState:  newState,
		},
	}
	m.command.SendEmcTaskSetState(msg)
}

func (m *Machine) jog(jogType JogType, axis uint32, velocity float64, distance float64) {
	m.setTaskMode("execute", pb.EmcTaskModeType_EMC_TASK_MODE_MANUAL)
	msg := &pb.Container{
		EmcCommandParams: &pb.EmcCommandParameters{
			Index: util.UI32(axis),
		},
	}
	switch jogType {
	case StopJog:
		m.command.SendEmcAxisAbort(msg)
	case ContinuousJog:
		msg.EmcCommandParams.Velocity = util.F64(velocity)
		m.command.SendEmcAxisJog(msg)
	case IncrementJog:
		msg.EmcCommandParams.Velocity = util.F64(velocity)
		msg.EmcCommandParams.Distance = util.F64(distance)
		m.command.SendEmcAxisIncrJog(msg)
	}
}

func (m *Machine) setTaskMode(interp string, taskMode pb.EmcTaskModeType) {
	m.MotionState.GetState()
	if m.TaskState.GetTaskMode() != taskMode {
		msg := &pb.Container{
			EmcCommandParams: &pb.EmcCommandParameters{
				TaskMode: taskMode.Enum(),
			},
			InterpName: util.S(interp),
		}
		m.command.SendEmcTaskSetMode(msg)
	}
}

func (m *Machine) ExecuteMdi(interp string, mdiCommand string) {
	log.Printf("Execute command: '%s'", mdiCommand)
	m.setTaskMode("execute", pb.EmcTaskModeType_EMC_TASK_MODE_MDI)
	msg := &pb.Container{
		EmcCommandParams: &pb.EmcCommandParameters{
			Command: util.S(mdiCommand),
		},
		InterpName: util.S(interp),
	}
	m.command.SendEmcTaskPlanExecute(msg)
}

func (m *Machine) ExecuteProgram(path string) {
	log.Printf("Execute program %s", path)
	m.setTaskMode("execute", pb.EmcTaskModeType_EMC_TASK_MODE_AUTO)
	m.ResetProgram("execute")
	m.OpenProgram("execute", path)
}

func (m *Machine) CloseProgram() {
	if m.CurrentProgram != "" {
		m.ResetProgram("execute")
		m.setTaskMode("execute", pb.EmcTaskModeType_EMC_TASK_MODE_MANUAL)
		m.CurrentProgram = ""
	}
}

func (m *Machine) GetRemotePath() string {
	if m.ConfigState != nil {
		return m.ConfigState.GetRemotePath()
	} else {
		return ""
	}
}
func (m *Machine) OpenProgram(interp string, path string) {
	m.CurrentProgram = path
	msg := &pb.Container{
		EmcCommandParams: &pb.EmcCommandParameters{
			Path: util.S(path),
		},
		InterpName: util.S(interp),
	}
	m.command.SendEmcTaskPlanOpen(msg)
}

func (m *Machine) ResetProgram(interp string) {
	msg := &pb.Container{
		InterpName: util.S(interp),
	}
	m.command.SendEmcTaskPlanInit(msg)
}

func (m *Machine) RunProgram(interp string, startLine int) {
	m.setTaskMode("execute", pb.EmcTaskModeType_EMC_TASK_MODE_AUTO)
	msg := &pb.Container{
		EmcCommandParams: &pb.EmcCommandParameters{
			LineNumber: util.I32(int32(startLine)),
		},
		InterpName: util.S(interp),
	}
	m.command.SendEmcTaskPlanRun(msg)
}

func (m *Machine) Abort(interp string) {
	msg := &pb.Container{
		InterpName: util.S(interp),
	}
	m.command.SendEmcTaskAbort(msg)
}

func (m *Machine) PauseProgram(interp string) {
	msg := &pb.Container{
		InterpName: util.S(interp),
	}
	m.command.SendEmcTaskPlanPause(msg)
}

func (m *Machine) ResumeProgram(interp string) {
	msg := &pb.Container{
		InterpName: util.S(interp),
	}
	m.command.SendEmcTaskPlanResume(msg)
}

func (m *Machine) StepProgram(interp string) {
	msg := &pb.Container{
		InterpName: util.S(interp),
	}
	m.command.SendEmcTaskPlanStep(msg)
}

var lcsNameToIndex = map[string]int{
	"G54":   1,
	"G55":   2,
	"G56":   3,
	"G57":   4,
	"G58":   5,
	"G59":   6,
	"G59.1": 7,
	"G59.2": 8,
	"G59.3": 9,
}

func (m *Machine) SetLcsToCurrent(lcs string) {
	lcsIndex := lcsNameToIndex[lcs]
	m.ExecuteMdi("execute", fmt.Sprintf("G10 L20 P%d X0 Y0 Z0", lcsIndex))
}
