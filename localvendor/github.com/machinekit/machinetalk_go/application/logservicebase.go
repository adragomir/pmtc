package application

import (
	"fmt"
	"log"
	"sync"

	"github.com/looplab/fsm"

	"github.com/machinekit/machinetalk_go/common"

	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type LogServiceBase struct {
	Debuglevel           int
	Debugname            string
	errorString          string
	OnErrorStringChanged []func(string)
	txLock               sync.Mutex
	// Log socket
	LogChannel *common.Publish
	// more efficient to reuse protobuf messages
	logTx *pb.Container

	OnStateChanged []func(string)
	fsm            *fsm.FSM
}

func NewLogServiceBase(Debuglevel int, Debugname string) *LogServiceBase {
	tmp := &LogServiceBase{}
	tmp.Debuglevel = Debuglevel
	tmp.Debugname = Debugname
	tmp.errorString = ""
	tmp.OnErrorStringChanged = make([]func(string), 0)

	// Log socket
	tmp.LogChannel = common.NewPublish(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "log"))
	tmp.LogChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "log")
	// more efficient to reuse protobuf messages
	tmp.logTx = &pb.Container{}

	// callbacks
	tmp.OnStateChanged = make([]func(string), 0)

	// fsm
	tmp.fsm = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "connect", Src: []string{"down"}, Dst: "up"},
			{Name: "disconnect", Src: []string{"up"}, Dst: "down"},
		},
		fsm.Callbacks{
			"down":             func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_connect":    func(e *fsm.Event) { tmp.OnFsm_connect(e) },
			"up":               func(e *fsm.Event) { tmp.OnFsm_up(e) },
			"after_disconnect": func(e *fsm.Event) { tmp.OnFsm_disconnect(e) },
		},
	)
	return tmp
}

func (self *LogServiceBase) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *LogServiceBase) OnFsm_connect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event CONNECT", self.Debugname)
	}
	self.StartLogChannel()
}

func (self *LogServiceBase) OnFsm_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("up")
	}
}

func (self *LogServiceBase) OnFsm_disconnect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event DISCONNECT", self.Debugname)
	}
	self.StopLogChannel()
}

func (self *LogServiceBase) ErrorString() string {
	return self.errorString
}

func (self *LogServiceBase) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

func (self *LogServiceBase) GetLogUri() string {
	return self.LogChannel.SocketUri
}

// @log_uri.setter
func (self *LogServiceBase) SetLogUri(value string) {
	self.LogChannel.SocketUri = value
}

func (self *LogServiceBase) GetLogPort() int {
	return self.LogChannel.GetSocketPort()
}

func (self *LogServiceBase) GetLogDsn() string {
	return self.LogChannel.GetSocketDsn()
}

// trigger
func (self *LogServiceBase) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("connect")
	}
}

// trigger
func (self *LogServiceBase) Stop() {
	if self.fsm.Is("up") {
		self.fsm.Event("disconnect")
	}
}

func (self *LogServiceBase) AddLogTopic(name string) {
	self.LogChannel.AddSocketTopic(name)
}

func (self *LogServiceBase) RemoveLogTopic(name string) {
	self.LogChannel.RemoveSocketTopic(name)
}

func (self *LogServiceBase) ClearLogTopics() {
	self.LogChannel.ClearSocketTopics()
}

func (self *LogServiceBase) StartLogChannel() {
	self.LogChannel.Start()
}

func (self *LogServiceBase) StopLogChannel() {
	self.LogChannel.Stop()
}

func (self *LogServiceBase) SendLogMsg(identity string, msg_type pb.ContainerType, tx *pb.Container) {
	self.LogChannel.SendSocketMsg(identity, msg_type, tx)
}
func (self *LogServiceBase) SendLogMessage(identity string, tx *pb.Container) {
	ids := map[string]bool{
		identity: true,
	}
	for receiver, _ := range ids {
		self.SendLogMsg(receiver, pb.ContainerType_MT_LOG_MESSAGE, tx)
	}
}
