package application

import (
	"fmt"
	"log"
	"sync"

	"github.com/looplab/fsm"
	"github.com/machinekit/machinetalk_go/common"
	"google.golang.org/protobuf/encoding/prototext"

	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type CommandBase struct {
	Debuglevel           int
	Debugname            string
	errorString          string
	OnErrorStringChanged []func(string)
	txLock               sync.Mutex
	// Command socket
	CommandChannel *common.RpcClient
	// more efficient to reuse protobuf messages
	commandRx *pb.Container
	commandTx *pb.Container

	OnCommandMsgReceived []func(*pb.Container, ...interface{})
	OnStateChanged       []func(string)
	fsm                  *fsm.FSM
}

func NewCommandBase(Debuglevel int, Debugname string) *CommandBase {
	tmp := &CommandBase{}
	tmp.Debuglevel = Debuglevel
	tmp.Debugname = Debugname
	tmp.errorString = ""
	tmp.OnErrorStringChanged = make([]func(string), 0)

	// Command socket
	tmp.CommandChannel = common.NewRpcClient(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "command"))
	tmp.CommandChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "command")
	tmp.CommandChannel.OnStateChanged = append(tmp.CommandChannel.OnStateChanged, tmp.CommandChannel_state_changed)
	tmp.CommandChannel.OnSocketMsgReceived = append(tmp.CommandChannel.OnSocketMsgReceived, tmp.CommandChannelMsgReceived)
	// more efficient to reuse protobuf messages
	// XXXXX command command command
	tmp.commandRx = &pb.Container{}
	tmp.commandTx = &pb.Container{}

	// callbacks
	tmp.OnCommandMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnStateChanged = make([]func(string), 0)

	// fsm
	tmp.fsm = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "connect", Src: []string{"down"}, Dst: "trying"},
			{Name: "command_up", Src: []string{"trying"}, Dst: "up"},
			{Name: "disconnect", Src: []string{"trying"}, Dst: "down"},
			{Name: "command_trying", Src: []string{"up"}, Dst: "trying"},
			{Name: "disconnect", Src: []string{"up"}, Dst: "down"},
		},
		fsm.Callbacks{
			"down":                 func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_connect":        func(e *fsm.Event) { tmp.OnFsm_connect(e) },
			"trying":               func(e *fsm.Event) { tmp.OnFsm_trying(e) },
			"after_command_up":     func(e *fsm.Event) { tmp.OnFsm_command_up(e) },
			"after_disconnect":     func(e *fsm.Event) { tmp.OnFsm_disconnect(e) },
			"up":                   func(e *fsm.Event) { tmp.OnFsm_up(e) },
			"after_command_trying": func(e *fsm.Event) { tmp.OnFsm_command_trying(e) },
			"leave_up":             func(e *fsm.Event) { tmp.OnFsm_up_exit(e) },
		},
	)
	return tmp
}

func (self *CommandBase) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *CommandBase) OnFsm_connect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event CONNECT", self.Debugname)
	}
	self.StartCommandChannel()
}

func (self *CommandBase) OnFsm_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state TRYING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("trying")
	}
}

func (self *CommandBase) OnFsm_command_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event COMMAND UP", self.Debugname)
	}
}

func (self *CommandBase) OnFsm_disconnect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event DISCONNECT", self.Debugname)
	}
	self.StopCommandChannel()
	self.ClearConnected()
}

func (self *CommandBase) OnFsm_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP entry", self.Debugname)
	}
	self.SetConnected()
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("up")
	}
}

func (self *CommandBase) OnFsm_command_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event COMMAND TRYING", self.Debugname)
	}
}

func (self *CommandBase) OnFsm_up_exit(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP exit", self.Debugname)
	}
	self.ClearConnected()
}

func (self *CommandBase) ErrorString() string {
	return self.errorString
}

func (self *CommandBase) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

func (self *CommandBase) GetCommandUri() string {
	return self.CommandChannel.SocketUri
}

// @command_uri.setter
func (self *CommandBase) SetCommandUri(value string) {
	self.CommandChannel.SocketUri = value
}

func (self *CommandBase) SetConnected() {
	// log.Printf("WARNING: slot set connected unimplemented")
}

func (self *CommandBase) ClearConnected() {
	// log.Printf("WARNING: slot clear connected unimplemented")
}

// trigger
func (self *CommandBase) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("connect")
	}
}

// trigger
func (self *CommandBase) Stop() {
	if self.fsm.Is("trying") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("up") {
		self.fsm.Event("disconnect")
	}
}

func (self *CommandBase) StartCommandChannel() {
	self.CommandChannel.Start()
}

func (self *CommandBase) StopCommandChannel() {
	self.CommandChannel.Stop()
}

// process all messages received on command
func (self *CommandBase) CommandChannelMsgReceived(rx *pb.Container, rest ...interface{}) {
	// INCOMING emccmd executed 0 0 1
	log.Printf("RECEIVED COMMAND CHANNEL %s", prototext.Format(rx))

	// react to emccmd executed message
	if *rx.Type == pb.ContainerType_MT_EMCCMD_EXECUTED {
		self.EmccmdExecutedReceived(rx)
		// INCOMING emccmd completed 0 0 1

		// react to emccmd completed message
	} else if *rx.Type == pb.ContainerType_MT_EMCCMD_COMPLETED {
		self.EmccmdCompletedReceived(rx)
		// INCOMING error 0 0 1

		// react to error message
	} else if *rx.Type == pb.ContainerType_MT_ERROR {
		// update error string with note
		self.errorString = ""
		for _, note := range rx.GetNote() {
			self.errorString += note + "\n"
		}
	} //AAAAAAAA

	for _, cb := range self.OnCommandMsgReceived {
		cb(rx)
	}
}

func (self *CommandBase) EmccmdExecutedReceived(rx *pb.Container) {
	// log.Printf("SLOT emccmd executed unimplemented")
}

func (self *CommandBase) EmccmdCompletedReceived(rx *pb.Container) {
	// log.Printf("SLOT emccmd completed unimplemented")
}

func (self *CommandBase) SendCommandMsg(msg_type pb.ContainerType, tx *pb.Container) {
	self.CommandChannel.SendSocketMsg(msg_type, tx)
}
func (self *CommandBase) SendEmcTaskAbort(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TASK_ABORT, tx)
}
func (self *CommandBase) SendEmcTaskPlanRun(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TASK_PLAN_RUN, tx)
}
func (self *CommandBase) SendEmcTaskPlanPause(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TASK_PLAN_PAUSE, tx)
}
func (self *CommandBase) SendEmcTaskPlanStep(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TASK_PLAN_STEP, tx)
}
func (self *CommandBase) SendEmcTaskPlanResume(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TASK_PLAN_RESUME, tx)
}
func (self *CommandBase) SendEmcSetDebug(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_SET_DEBUG, tx)
}
func (self *CommandBase) SendEmcCoolantFloodOn(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_COOLANT_FLOOD_ON, tx)
}
func (self *CommandBase) SendEmcCoolantFloodOff(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_COOLANT_FLOOD_OFF, tx)
}
func (self *CommandBase) SendEmcAxisHome(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_AXIS_HOME, tx)
}
func (self *CommandBase) SendEmcAxisJog(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_AXIS_JOG, tx)
}
func (self *CommandBase) SendEmcAxisAbort(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_AXIS_ABORT, tx)
}
func (self *CommandBase) SendEmcAxisIncrJog(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_AXIS_INCR_JOG, tx)
}
func (self *CommandBase) SendEmcToolLoadToolTable(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TOOL_LOAD_TOOL_TABLE, tx)
}
func (self *CommandBase) SendEmcToolUpdateToolTable(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TOOL_UPDATE_TOOL_TABLE, tx)
}
func (self *CommandBase) SendEmcTaskPlanExecute(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TASK_PLAN_EXECUTE, tx)
}
func (self *CommandBase) SendEmcCoolantMistOn(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_COOLANT_MIST_ON, tx)
}
func (self *CommandBase) SendEmcCoolantMistOff(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_COOLANT_MIST_OFF, tx)
}
func (self *CommandBase) SendEmcTaskPlanInit(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TASK_PLAN_INIT, tx)
}
func (self *CommandBase) SendEmcTaskPlanOpen(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TASK_PLAN_OPEN, tx)
}
func (self *CommandBase) SendEmcTaskPlanSetOptionalStop(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TASK_PLAN_SET_OPTIONAL_STOP, tx)
}
func (self *CommandBase) SendEmcTaskPlanSetBlockDelete(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TASK_PLAN_SET_BLOCK_DELETE, tx)
}
func (self *CommandBase) SendEmcTaskSetMode(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TASK_SET_MODE, tx)
}
func (self *CommandBase) SendEmcTaskSetState(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TASK_SET_STATE, tx)
}
func (self *CommandBase) SendEmcTrajSetSoEnable(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TRAJ_SET_SO_ENABLE, tx)
}
func (self *CommandBase) SendEmcTrajSetFhEnable(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TRAJ_SET_FH_ENABLE, tx)
}
func (self *CommandBase) SendEmcTrajSetFoEnable(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TRAJ_SET_FO_ENABLE, tx)
}
func (self *CommandBase) SendEmcTrajSetMaxVelocity(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TRAJ_SET_MAX_VELOCITY, tx)
}
func (self *CommandBase) SendEmcTrajSetMode(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TRAJ_SET_MODE, tx)
}
func (self *CommandBase) SendEmcTrajSetScale(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TRAJ_SET_SCALE, tx)
}
func (self *CommandBase) SendEmcTrajSetRapidScale(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TRAJ_SET_RAPID_SCALE, tx)
}
func (self *CommandBase) SendEmcTrajSetSpindleScale(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TRAJ_SET_SPINDLE_SCALE, tx)
}
func (self *CommandBase) SendEmcTrajSetTeleopEnable(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TRAJ_SET_TELEOP_ENABLE, tx)
}
func (self *CommandBase) SendEmcTrajSetTeleopVector(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TRAJ_SET_TELEOP_VECTOR, tx)
}
func (self *CommandBase) SendEmcToolSetOffset(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_TOOL_SET_OFFSET, tx)
}
func (self *CommandBase) SendEmcAxisOverrideLimits(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_AXIS_OVERRIDE_LIMITS, tx)
}
func (self *CommandBase) SendEmcSpindleConstant(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_SPINDLE_CONSTANT, tx)
}
func (self *CommandBase) SendEmcSpindleDecrease(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_SPINDLE_DECREASE, tx)
}
func (self *CommandBase) SendEmcSpindleIncrease(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_SPINDLE_INCREASE, tx)
}
func (self *CommandBase) SendEmcSpindleOff(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_SPINDLE_OFF, tx)
}
func (self *CommandBase) SendEmcSpindleOn(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_SPINDLE_ON, tx)
}
func (self *CommandBase) SendEmcSpindleBrakeEngage(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_SPINDLE_BRAKE_ENGAGE, tx)
}
func (self *CommandBase) SendEmcSpindleBrakeRelease(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_SPINDLE_BRAKE_RELEASE, tx)
}
func (self *CommandBase) SendEmcMotionSetAout(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_MOTION_SET_AOUT, tx)
}
func (self *CommandBase) SendEmcMotionSetDout(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_MOTION_SET_DOUT, tx)
}
func (self *CommandBase) SendEmcMotionAdaptive(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_MOTION_ADAPTIVE, tx)
}
func (self *CommandBase) SendEmcAxisSetMaxPositionLimit(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_AXIS_SET_MAX_POSITION_LIMIT, tx)
}
func (self *CommandBase) SendEmcAxisSetMinPositionLimit(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_AXIS_SET_MIN_POSITION_LIMIT, tx)
}
func (self *CommandBase) SendEmcAxisUnhome(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_EMC_AXIS_UNHOME, tx)
}
func (self *CommandBase) SendShutdown(tx *pb.Container) {
	self.SendCommandMsg(pb.ContainerType_MT_SHUTDOWN, tx)
}
func (self *CommandBase) CommandChannel_state_changed(state string) {

	if state == "trying" {
		if self.fsm.Is("up") {
			self.fsm.Event("command_trying")
		}

	} else if state == "up" {
		if self.fsm.Is("trying") {
			self.fsm.Event("command_up")
		}
	}
}
