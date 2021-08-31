package application

import (
	"fmt"
	"log"
	"sync"

	"github.com/looplab/fsm"

	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type ErrorBase struct {
	Debuglevel           int
	Debugname            string
	errorString          string
	OnErrorStringChanged []func(string)
	txLock               sync.Mutex
	// Error socket
	ErrorChannel *ErrorSubscribe
	// more efficient to reuse protobuf messages
	errorRx *pb.Container

	OnErrorMsgReceived []func(*pb.Container, ...interface{})
	OnStateChanged     []func(string)
	fsm                *fsm.FSM
}

func NewErrorBase(Debuglevel int, Debugname string) *ErrorBase {
	tmp := &ErrorBase{}
	tmp.Debuglevel = Debuglevel
	tmp.Debugname = Debugname
	tmp.errorString = ""
	tmp.OnErrorStringChanged = make([]func(string), 0)

	// Error socket
	tmp.ErrorChannel = NewErrorSubscribe(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "error"))
	tmp.ErrorChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "error")
	tmp.ErrorChannel.OnStateChanged = append(tmp.ErrorChannel.OnStateChanged, tmp.ErrorChannel_state_changed)
	tmp.ErrorChannel.OnSocketMsgReceived = append(tmp.ErrorChannel.OnSocketMsgReceived, tmp.ErrorChannelMsgReceived)
	// more efficient to reuse protobuf messages
	// XXXXX error error error
	tmp.errorRx = &pb.Container{}

	// callbacks
	tmp.OnErrorMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnStateChanged = make([]func(string), 0)

	// fsm
	tmp.fsm = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "connect", Src: []string{"down"}, Dst: "trying"},
			{Name: "error_up", Src: []string{"trying"}, Dst: "up"},
			{Name: "disconnect", Src: []string{"trying"}, Dst: "down"},
			{Name: "error_trying", Src: []string{"up"}, Dst: "trying"},
			{Name: "disconnect", Src: []string{"up"}, Dst: "down"},
		},
		fsm.Callbacks{
			"down":               func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_connect":      func(e *fsm.Event) { tmp.OnFsm_connect(e) },
			"trying":             func(e *fsm.Event) { tmp.OnFsm_trying(e) },
			"after_error_up":     func(e *fsm.Event) { tmp.OnFsm_error_up(e) },
			"after_disconnect":   func(e *fsm.Event) { tmp.OnFsm_disconnect(e) },
			"up":                 func(e *fsm.Event) { tmp.OnFsm_up(e) },
			"after_error_trying": func(e *fsm.Event) { tmp.OnFsm_error_trying(e) },
			"leave_up":           func(e *fsm.Event) { tmp.OnFsm_up_exit(e) },
		},
	)
	return tmp
}

func (self *ErrorBase) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *ErrorBase) OnFsm_connect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event CONNECT", self.Debugname)
	}
	self.UpdateTopics()
	self.StartErrorChannel()
}

func (self *ErrorBase) OnFsm_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state TRYING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("trying")
	}
}

func (self *ErrorBase) OnFsm_error_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event ERROR UP", self.Debugname)
	}
}

func (self *ErrorBase) OnFsm_disconnect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event DISCONNECT", self.Debugname)
	}
	self.StopErrorChannel()
}

func (self *ErrorBase) OnFsm_up(e *fsm.Event) {
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

func (self *ErrorBase) OnFsm_error_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event ERROR TRYING", self.Debugname)
	}
}

func (self *ErrorBase) OnFsm_up_exit(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP exit", self.Debugname)
	}
	self.ClearConnected()
}

func (self *ErrorBase) ErrorString() string {
	return self.errorString
}

func (self *ErrorBase) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

func (self *ErrorBase) GetErrorUri() string {
	return self.ErrorChannel.SocketUri
}

// @error_uri.setter
func (self *ErrorBase) SetErrorUri(value string) {
	self.ErrorChannel.SocketUri = value
}

func (self *ErrorBase) UpdateTopics() {
	// log.Printf("WARNING: slot update topics unimplemented")
}

func (self *ErrorBase) SetConnected() {
	// log.Printf("WARNING: slot set connected unimplemented")
}

func (self *ErrorBase) ClearConnected() {
	// log.Printf("WARNING: slot clear connected unimplemented")
}

// trigger
func (self *ErrorBase) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("connect")
	}
}

// trigger
func (self *ErrorBase) Stop() {
	if self.fsm.Is("trying") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("up") {
		self.fsm.Event("disconnect")
	}
}

func (self *ErrorBase) AddErrorTopic(name string) {
	self.ErrorChannel.AddSocketTopic(name)
}

func (self *ErrorBase) RemoveErrorTopic(name string) {
	self.ErrorChannel.RemoveSocketTopic(name)
}

func (self *ErrorBase) ClearErrorTopics() {
	self.ErrorChannel.ClearSocketTopics()
}

func (self *ErrorBase) StartErrorChannel() {
	self.ErrorChannel.Start()
}

func (self *ErrorBase) StopErrorChannel() {
	self.ErrorChannel.Stop()
}

// process all messages received on error
func (self *ErrorBase) ErrorChannelMsgReceived(rx *pb.Container, rest ...interface{}) {
	//parse identity
	identity := rest[0].(string)
	// INCOMING emc nml error 0 0 1

	// react to emc nml error message
	if *rx.Type == pb.ContainerType_MT_EMC_NML_ERROR {
		self.EmcNmlErrorReceived(identity, rx)
		// INCOMING emc nml text 0 0 1

		// react to emc nml text message
	} else if *rx.Type == pb.ContainerType_MT_EMC_NML_TEXT {
		self.EmcNmlTextReceived(identity, rx)
		// INCOMING emc nml display 0 0 1

		// react to emc nml display message
	} else if *rx.Type == pb.ContainerType_MT_EMC_NML_DISPLAY {
		self.EmcNmlDisplayReceived(identity, rx)
		// INCOMING emc operator text 0 0 1

		// react to emc operator text message
	} else if *rx.Type == pb.ContainerType_MT_EMC_OPERATOR_TEXT {
		self.EmcOperatorTextReceived(identity, rx)
		// INCOMING emc operator error 0 0 1

		// react to emc operator error message
	} else if *rx.Type == pb.ContainerType_MT_EMC_OPERATOR_ERROR {
		self.EmcOperatorErrorReceived(identity, rx)
		// INCOMING emc operator display 0 0 1

		// react to emc operator display message
	} else if *rx.Type == pb.ContainerType_MT_EMC_OPERATOR_DISPLAY {
		self.EmcOperatorDisplayReceived(identity, rx)
	} //AAAAAAAA

	for _, cb := range self.OnErrorMsgReceived {
		cb(rx, string(identity))
	}
}

func (self *ErrorBase) EmcNmlErrorReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT emc nml error unimplemented")
}

func (self *ErrorBase) EmcNmlTextReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT emc nml text unimplemented")
}

func (self *ErrorBase) EmcNmlDisplayReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT emc nml display unimplemented")
}

func (self *ErrorBase) EmcOperatorTextReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT emc operator text unimplemented")
}

func (self *ErrorBase) EmcOperatorErrorReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT emc operator error unimplemented")
}

func (self *ErrorBase) EmcOperatorDisplayReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT emc operator display unimplemented")
}
func (self *ErrorBase) ErrorChannel_state_changed(state string) {

	if state == "trying" {
		if self.fsm.Is("up") {
			self.fsm.Event("error_trying")
		}

	} else if state == "up" {
		if self.fsm.Is("trying") {
			self.fsm.Event("error_up")
		}
	}
}
