package pathview

import (
	"fmt"
	"log"
	"sync"

	"github.com/looplab/fsm"

	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type PreviewClientBase struct {
	Debuglevel           int
	Debugname            string
	errorString          string
	OnErrorStringChanged []func(string)
	txLock               sync.Mutex
	// Preview socket
	PreviewChannel *PreviewSubscribe
	// more efficient to reuse protobuf messages
	previewRx *pb.Container
	// Previewstatus socket
	PreviewstatusChannel *PreviewSubscribe
	// more efficient to reuse protobuf messages
	previewstatusRx *pb.Container

	OnPreviewMsgReceived       []func(*pb.Container, ...interface{})
	OnPreviewstatusMsgReceived []func(*pb.Container, ...interface{})
	OnStateChanged             []func(string)
	fsm                        *fsm.FSM
}

func NewPreviewClientBase(Debuglevel int, Debugname string) *PreviewClientBase {
	tmp := &PreviewClientBase{}
	tmp.Debuglevel = Debuglevel
	tmp.Debugname = Debugname
	tmp.errorString = ""
	tmp.OnErrorStringChanged = make([]func(string), 0)

	// Preview socket
	tmp.PreviewChannel = NewPreviewSubscribe(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "preview"))
	tmp.PreviewChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "preview")
	tmp.PreviewChannel.OnStateChanged = append(tmp.PreviewChannel.OnStateChanged, tmp.PreviewChannel_state_changed)
	tmp.PreviewChannel.OnSocketMsgReceived = append(tmp.PreviewChannel.OnSocketMsgReceived, tmp.PreviewChannelMsgReceived)
	// more efficient to reuse protobuf messages
	// XXXXX preview preview preview
	tmp.previewRx = &pb.Container{}

	// Previewstatus socket
	tmp.PreviewstatusChannel = NewPreviewSubscribe(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "previewstatus"))
	tmp.PreviewstatusChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "previewstatus")
	tmp.PreviewstatusChannel.OnStateChanged = append(tmp.PreviewstatusChannel.OnStateChanged, tmp.PreviewstatusChannel_state_changed)
	tmp.PreviewstatusChannel.OnSocketMsgReceived = append(tmp.PreviewstatusChannel.OnSocketMsgReceived, tmp.PreviewstatusChannelMsgReceived)
	// more efficient to reuse protobuf messages
	// XXXXX previewstatus previewstatus previewstatus
	tmp.previewstatusRx = &pb.Container{}

	// callbacks
	tmp.OnPreviewMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnPreviewstatusMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnStateChanged = make([]func(string), 0)

	// fsm
	tmp.fsm = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "connect", Src: []string{"down"}, Dst: "trying"},
			{Name: "status_up", Src: []string{"trying"}, Dst: "previewtrying"},
			{Name: "preview_up", Src: []string{"trying"}, Dst: "statustrying"},
			{Name: "disconnect", Src: []string{"trying", "previewtrying", "statustrying", "up"}, Dst: "down"},
			{Name: "preview_up", Src: []string{"previewtrying"}, Dst: "up"},
			{Name: "status_trying", Src: []string{"previewtrying"}, Dst: "trying"},
			{Name: "status_up", Src: []string{"statustrying"}, Dst: "up"},
			{Name: "preview_trying", Src: []string{"statustrying"}, Dst: "trying"},
			{Name: "preview_trying", Src: []string{"up"}, Dst: "previewtrying"},
			{Name: "status_trying", Src: []string{"up"}, Dst: "statustrying"},
		},
		fsm.Callbacks{
			"down":                 func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_connect":        func(e *fsm.Event) { tmp.OnFsm_connect(e) },
			"trying":               func(e *fsm.Event) { tmp.OnFsm_trying(e) },
			"after_status_up":      func(e *fsm.Event) { tmp.OnFsm_status_up(e) },
			"after_preview_up":     func(e *fsm.Event) { tmp.OnFsm_preview_up(e) },
			"after_disconnect":     func(e *fsm.Event) { tmp.OnFsm_disconnect(e) },
			"previewtrying":        func(e *fsm.Event) { tmp.OnFsm_previewtrying(e) },
			"after_status_trying":  func(e *fsm.Event) { tmp.OnFsm_status_trying(e) },
			"statustrying":         func(e *fsm.Event) { tmp.OnFsm_statustrying(e) },
			"after_preview_trying": func(e *fsm.Event) { tmp.OnFsm_preview_trying(e) },
			"up":                   func(e *fsm.Event) { tmp.OnFsm_up(e) },
			"leave_up":             func(e *fsm.Event) { tmp.OnFsm_up_exit(e) },
		},
	)
	return tmp
}

func (self *PreviewClientBase) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *PreviewClientBase) OnFsm_connect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event CONNECT", self.Debugname)
	}
	self.StartPreviewChannel()
	self.StartPreviewstatusChannel()
}

func (self *PreviewClientBase) OnFsm_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state TRYING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("trying")
	}
}

func (self *PreviewClientBase) OnFsm_status_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event STATUS UP", self.Debugname)
	}
}

func (self *PreviewClientBase) OnFsm_preview_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event PREVIEW UP", self.Debugname)
	}
}

func (self *PreviewClientBase) OnFsm_disconnect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event DISCONNECT", self.Debugname)
	}
	self.StopPreviewChannel()
	self.StopPreviewstatusChannel()
}

func (self *PreviewClientBase) OnFsm_previewtrying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state PREVIEWTRYING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("previewtrying")
	}
}

func (self *PreviewClientBase) OnFsm_status_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event STATUS TRYING", self.Debugname)
	}
}

func (self *PreviewClientBase) OnFsm_statustrying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state STATUSTRYING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("statustrying")
	}
}

func (self *PreviewClientBase) OnFsm_preview_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event PREVIEW TRYING", self.Debugname)
	}
}

func (self *PreviewClientBase) OnFsm_up(e *fsm.Event) {
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

func (self *PreviewClientBase) OnFsm_up_exit(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP exit", self.Debugname)
	}
	self.ClearConnected()
}

func (self *PreviewClientBase) ErrorString() string {
	return self.errorString
}

func (self *PreviewClientBase) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

func (self *PreviewClientBase) GetPreviewUri() string {
	return self.PreviewChannel.SocketUri
}

// @preview_uri.setter
func (self *PreviewClientBase) SetPreviewUri(value string) {
	self.PreviewChannel.SocketUri = value
}

func (self *PreviewClientBase) GetPreviewstatusUri() string {
	return self.PreviewstatusChannel.SocketUri
}

// @previewstatus_uri.setter
func (self *PreviewClientBase) SetPreviewstatusUri(value string) {
	self.PreviewstatusChannel.SocketUri = value
}

func (self *PreviewClientBase) SetConnected() {
	// log.Printf("WARNING: slot set connected unimplemented")
}

func (self *PreviewClientBase) ClearConnected() {
	// log.Printf("WARNING: slot clear connected unimplemented")
}

// trigger
func (self *PreviewClientBase) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("connect")
	}
}

// trigger
func (self *PreviewClientBase) Stop() {
	if self.fsm.Is("trying") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("up") {
		self.fsm.Event("disconnect")
	}
}

func (self *PreviewClientBase) AddPreviewTopic(name string) {
	self.PreviewChannel.AddSocketTopic(name)
}

func (self *PreviewClientBase) RemovePreviewTopic(name string) {
	self.PreviewChannel.RemoveSocketTopic(name)
}

func (self *PreviewClientBase) ClearPreviewTopics() {
	self.PreviewChannel.ClearSocketTopics()
}

func (self *PreviewClientBase) AddPreviewstatusTopic(name string) {
	self.PreviewstatusChannel.AddSocketTopic(name)
}

func (self *PreviewClientBase) RemovePreviewstatusTopic(name string) {
	self.PreviewstatusChannel.RemoveSocketTopic(name)
}

func (self *PreviewClientBase) ClearPreviewstatusTopics() {
	self.PreviewstatusChannel.ClearSocketTopics()
}

func (self *PreviewClientBase) StartPreviewChannel() {
	self.PreviewChannel.Start()
}

func (self *PreviewClientBase) StopPreviewChannel() {
	self.PreviewChannel.Stop()
}

func (self *PreviewClientBase) StartPreviewstatusChannel() {
	self.PreviewstatusChannel.Start()
}

func (self *PreviewClientBase) StopPreviewstatusChannel() {
	self.PreviewstatusChannel.Stop()
}

// process all messages received on preview
func (self *PreviewClientBase) PreviewChannelMsgReceived(rx *pb.Container, rest ...interface{}) {
	//parse identity
	identity := rest[0].(string)
	// INCOMING preview 0 0 1

	// react to preview message
	if *rx.Type == pb.ContainerType_MT_PREVIEW {
		self.PreviewReceived(identity, rx)
	} //AAAAAAAA

	for _, cb := range self.OnPreviewMsgReceived {
		cb(rx, string(identity))
	}
}

func (self *PreviewClientBase) PreviewReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT preview unimplemented")
}

// process all messages received on previewstatus
func (self *PreviewClientBase) PreviewstatusChannelMsgReceived(rx *pb.Container, rest ...interface{}) {
	//parse identity
	identity := rest[0].(string)
	// INCOMING interp stat 0 0 1

	// react to interp stat message
	if *rx.Type == pb.ContainerType_MT_INTERP_STAT {
		self.InterpStatReceived(identity, rx)
	} //AAAAAAAA

	for _, cb := range self.OnPreviewstatusMsgReceived {
		cb(rx, string(identity))
	}
}

func (self *PreviewClientBase) InterpStatReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT interp stat unimplemented")
}
func (self *PreviewClientBase) PreviewChannel_state_changed(state string) {

	if state == "trying" {
		if self.fsm.Is("up") {
			self.fsm.Event("preview_trying")
		} else if self.fsm.Is("statustrying") {
			self.fsm.Event("preview_trying")
		}

	} else if state == "up" {
		if self.fsm.Is("trying") {
			self.fsm.Event("preview_up")
		} else if self.fsm.Is("previewtrying") {
			self.fsm.Event("preview_up")
		}
	}
}
func (self *PreviewClientBase) PreviewstatusChannel_state_changed(state string) {

	if state == "trying" {
		if self.fsm.Is("up") {
			self.fsm.Event("status_trying")
		} else if self.fsm.Is("previewtrying") {
			self.fsm.Event("status_trying")
		}

	} else if state == "up" {
		if self.fsm.Is("trying") {
			self.fsm.Event("status_up")
		} else if self.fsm.Is("statustrying") {
			self.fsm.Event("status_up")
		}
	}
}
