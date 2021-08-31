package application

import (
	"fmt"
	"log"
	"sync"

	"github.com/looplab/fsm"

	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type StatusBase struct {
	Debuglevel           int
	Debugname            string
	errorString          string
	OnErrorStringChanged []func(string)
	txLock               sync.Mutex
	// Status socket
	StatusChannel *StatusSubscribe
	// more efficient to reuse protobuf messages
	statusRx *pb.Container

	OnStatusMsgReceived []func(*pb.Container, ...interface{})
	OnStateChanged      []func(string)
	fsm                 *fsm.FSM
}

func NewStatusBase(Debuglevel int, Debugname string) *StatusBase {
	tmp := &StatusBase{}
	tmp.Debuglevel = Debuglevel
	tmp.Debugname = Debugname
	tmp.errorString = ""
	tmp.OnErrorStringChanged = make([]func(string), 0)

	// Status socket
	tmp.StatusChannel = NewStatusSubscribe(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "status"))
	tmp.StatusChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "status")
	tmp.StatusChannel.OnStateChanged = append(tmp.StatusChannel.OnStateChanged, tmp.StatusChannel_state_changed)
	tmp.StatusChannel.OnSocketMsgReceived = append(tmp.StatusChannel.OnSocketMsgReceived, tmp.StatusChannelMsgReceived)
	// more efficient to reuse protobuf messages
	// XXXXX status status status
	tmp.statusRx = &pb.Container{}

	// callbacks
	tmp.OnStatusMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnStateChanged = make([]func(string), 0)

	// fsm
	tmp.fsm = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "connect", Src: []string{"down"}, Dst: "trying"},
			{Name: "status_up", Src: []string{"trying"}, Dst: "syncing"},
			{Name: "disconnect", Src: []string{"trying", "syncing", "up"}, Dst: "down"},
			{Name: "channels_synced", Src: []string{"syncing"}, Dst: "up"},
			{Name: "status_trying", Src: []string{"syncing", "up"}, Dst: "trying"},
		},
		fsm.Callbacks{
			"down":                  func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_connect":         func(e *fsm.Event) { tmp.OnFsm_connect(e) },
			"trying":                func(e *fsm.Event) { tmp.OnFsm_trying(e) },
			"after_status_up":       func(e *fsm.Event) { tmp.OnFsm_status_up(e) },
			"after_disconnect":      func(e *fsm.Event) { tmp.OnFsm_disconnect(e) },
			"syncing":               func(e *fsm.Event) { tmp.OnFsm_syncing(e) },
			"after_channels_synced": func(e *fsm.Event) { tmp.OnFsm_channels_synced(e) },
			"after_status_trying":   func(e *fsm.Event) { tmp.OnFsm_status_trying(e) },
			"up":                    func(e *fsm.Event) { tmp.OnFsm_up(e) },
			"leave_up":              func(e *fsm.Event) { tmp.OnFsm_up_exit(e) },
		},
	)
	return tmp
}

func (self *StatusBase) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *StatusBase) OnFsm_connect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event CONNECT", self.Debugname)
	}
	self.UpdateTopics()
	self.StartStatusChannel()
}

func (self *StatusBase) OnFsm_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state TRYING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("trying")
	}
}

func (self *StatusBase) OnFsm_status_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event STATUS UP", self.Debugname)
	}
}

func (self *StatusBase) OnFsm_disconnect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event DISCONNECT", self.Debugname)
	}
	self.StopStatusChannel()
}

func (self *StatusBase) OnFsm_syncing(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state SYNCING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("syncing")
	}
}

func (self *StatusBase) OnFsm_channels_synced(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event CHANNELS SYNCED", self.Debugname)
	}
}

func (self *StatusBase) OnFsm_status_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event STATUS TRYING", self.Debugname)
	}
}

func (self *StatusBase) OnFsm_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP entry", self.Debugname)
	}
	self.SyncStatus()
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("up")
	}
}

func (self *StatusBase) OnFsm_up_exit(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP exit", self.Debugname)
	}
	self.UnsyncStatus()
}

func (self *StatusBase) ErrorString() string {
	return self.errorString
}

func (self *StatusBase) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

func (self *StatusBase) GetStatusUri() string {
	return self.StatusChannel.SocketUri
}

// @status_uri.setter
func (self *StatusBase) SetStatusUri(value string) {
	self.StatusChannel.SocketUri = value
}

func (self *StatusBase) SyncStatus() {
	// log.Printf("WARNING: slot sync status unimplemented")
}

func (self *StatusBase) UnsyncStatus() {
	// log.Printf("WARNING: slot unsync status unimplemented")
}

func (self *StatusBase) UpdateTopics() {
	// log.Printf("WARNING: slot update topics unimplemented")
}

// trigger
func (self *StatusBase) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("connect")
	}
}

// trigger
func (self *StatusBase) Stop() {
	if self.fsm.Is("trying") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("up") {
		self.fsm.Event("disconnect")
	}
}

// trigger
func (self *StatusBase) ChannelsSynced() {
	if self.fsm.Is("syncing") {
		self.fsm.Event("channels_synced")
	}
}

func (self *StatusBase) AddStatusTopic(name string) {
	self.StatusChannel.AddSocketTopic(name)
}

func (self *StatusBase) RemoveStatusTopic(name string) {
	self.StatusChannel.RemoveSocketTopic(name)
}

func (self *StatusBase) ClearStatusTopics() {
	self.StatusChannel.ClearSocketTopics()
}

func (self *StatusBase) StartStatusChannel() {
	self.StatusChannel.Start()
}

func (self *StatusBase) StopStatusChannel() {
	self.StatusChannel.Stop()
}

// process all messages received on status
func (self *StatusBase) StatusChannelMsgReceived(rx *pb.Container, rest ...interface{}) {
	//parse identity
	identity := rest[0].(string)
	// INCOMING emcstat full update 0 0 1

	// react to emcstat full update message
	if *rx.Type == pb.ContainerType_MT_EMCSTAT_FULL_UPDATE {
		self.EmcstatFullUpdateReceived(identity, rx)
		// INCOMING emcstat incremental update 0 0 1

		// react to emcstat incremental update message
	} else if *rx.Type == pb.ContainerType_MT_EMCSTAT_INCREMENTAL_UPDATE {
		self.EmcstatIncrementalUpdateReceived(identity, rx)
	} //AAAAAAAA

	for _, cb := range self.OnStatusMsgReceived {
		cb(rx, string(identity))
	}
}

func (self *StatusBase) EmcstatFullUpdateReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT emcstat full update unimplemented")
}

func (self *StatusBase) EmcstatIncrementalUpdateReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT emcstat incremental update unimplemented")
}
func (self *StatusBase) StatusChannel_state_changed(state string) {

	if state == "trying" {
		if self.fsm.Is("up") {
			self.fsm.Event("status_trying")
		}

	} else if state == "trying" {
		if self.fsm.Is("syncing") {
			self.fsm.Event("status_trying")
		}

	} else if state == "up" {
		if self.fsm.Is("trying") {
			self.fsm.Event("status_up")
		}
	}
}
