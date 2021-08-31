package application

import (
	"fmt"
	"log"
	"sync"

	"github.com/looplab/fsm"

	"github.com/machinekit/machinetalk_go/common"

	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type LogBase struct {
	Debuglevel           int
	Debugname            string
	errorString          string
	OnErrorStringChanged []func(string)
	txLock               sync.Mutex
	// Log socket
	LogChannel *common.SimpleSubscribe
	// more efficient to reuse protobuf messages
	logRx *pb.Container

	OnLogMsgReceived []func(*pb.Container, ...interface{})
	OnStateChanged   []func(string)
	fsm              *fsm.FSM
}

func NewLogBase(Debuglevel int, Debugname string) *LogBase {
	tmp := &LogBase{}
	tmp.Debuglevel = Debuglevel
	tmp.Debugname = Debugname
	tmp.errorString = ""
	tmp.OnErrorStringChanged = make([]func(string), 0)

	// Log socket
	tmp.LogChannel = common.NewSimpleSubscribe(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "log"))
	tmp.LogChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "log")
	tmp.LogChannel.OnStateChanged = append(tmp.LogChannel.OnStateChanged, tmp.LogChannel_state_changed)
	tmp.LogChannel.OnSocketMsgReceived = append(tmp.LogChannel.OnSocketMsgReceived, tmp.LogChannelMsgReceived)
	// more efficient to reuse protobuf messages
	// XXXXX log log log
	tmp.logRx = &pb.Container{}

	// callbacks
	tmp.OnLogMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnStateChanged = make([]func(string), 0)

	// fsm
	tmp.fsm = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "connect", Src: []string{"down"}, Dst: "trying"},
			{Name: "log_up", Src: []string{"trying"}, Dst: "up"},
			{Name: "disconnect", Src: []string{"trying", "up"}, Dst: "down"},
		},
		fsm.Callbacks{
			"down":             func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_connect":    func(e *fsm.Event) { tmp.OnFsm_connect(e) },
			"trying":           func(e *fsm.Event) { tmp.OnFsm_trying(e) },
			"after_log_up":     func(e *fsm.Event) { tmp.OnFsm_log_up(e) },
			"after_disconnect": func(e *fsm.Event) { tmp.OnFsm_disconnect(e) },
			"up":               func(e *fsm.Event) { tmp.OnFsm_up(e) },
			"leave_up":         func(e *fsm.Event) { tmp.OnFsm_up_exit(e) },
		},
	)
	return tmp
}

func (self *LogBase) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *LogBase) OnFsm_connect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event CONNECT", self.Debugname)
	}
	self.UpdateTopics()
	self.StartLogChannel()
}

func (self *LogBase) OnFsm_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state TRYING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("trying")
	}
}

func (self *LogBase) OnFsm_log_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event LOG UP", self.Debugname)
	}
}

func (self *LogBase) OnFsm_disconnect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event DISCONNECT", self.Debugname)
	}
	self.StopLogChannel()
}

func (self *LogBase) OnFsm_up(e *fsm.Event) {
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

func (self *LogBase) OnFsm_up_exit(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP exit", self.Debugname)
	}
	self.ClearConnected()
}

func (self *LogBase) ErrorString() string {
	return self.errorString
}

func (self *LogBase) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

func (self *LogBase) GetLogUri() string {
	return self.LogChannel.SocketUri
}

// @log_uri.setter
func (self *LogBase) SetLogUri(value string) {
	self.LogChannel.SocketUri = value
}

func (self *LogBase) UpdateTopics() {
	// log.Printf("WARNING: slot update topics unimplemented")
}

func (self *LogBase) SetConnected() {
	// log.Printf("WARNING: slot set connected unimplemented")
}

func (self *LogBase) ClearConnected() {
	// log.Printf("WARNING: slot clear connected unimplemented")
}

// trigger
func (self *LogBase) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("connect")
	}
}

// trigger
func (self *LogBase) Stop() {
	if self.fsm.Is("trying") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("up") {
		self.fsm.Event("disconnect")
	}
}

func (self *LogBase) AddLogTopic(name string) {
	self.LogChannel.AddSocketTopic(name)
}

func (self *LogBase) RemoveLogTopic(name string) {
	self.LogChannel.RemoveSocketTopic(name)
}

func (self *LogBase) ClearLogTopics() {
	self.LogChannel.ClearSocketTopics()
}

func (self *LogBase) StartLogChannel() {
	self.LogChannel.Start()
}

func (self *LogBase) StopLogChannel() {
	self.LogChannel.Stop()
}

// process all messages received on log
func (self *LogBase) LogChannelMsgReceived(rx *pb.Container, rest ...interface{}) {
	//parse identity
	identity := rest[0].(string)
	// INCOMING log message 0 0 1

	// react to log message message
	if *rx.Type == pb.ContainerType_MT_LOG_MESSAGE {
		self.LogMessageReceived(identity, rx)
	} //AAAAAAAA

	for _, cb := range self.OnLogMsgReceived {
		cb(rx, string(identity))
	}
}

func (self *LogBase) LogMessageReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT log message unimplemented")
}
func (self *LogBase) LogChannel_state_changed(state string) {

	if state == "up" {
		if self.fsm.Is("trying") {
			self.fsm.Event("log_up")
		}
	}
}
