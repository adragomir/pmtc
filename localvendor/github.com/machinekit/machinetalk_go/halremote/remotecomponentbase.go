package halremote

import (
	"fmt"
	"log"
	"sync"

	"github.com/looplab/fsm"

	"github.com/machinekit/machinetalk_go/common"

	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type RemoteComponentBase struct {
	Debuglevel           int
	Debugname            string
	errorString          string
	OnErrorStringChanged []func(string)
	txLock               sync.Mutex
	// Halrcmd socket
	HalrcmdChannel *common.RpcClient
	// more efficient to reuse protobuf messages
	halrcmdRx *pb.Container
	halrcmdTx *pb.Container
	// Halrcomp socket
	HalrcompChannel *HalrcompSubscribe
	// more efficient to reuse protobuf messages
	halrcompRx *pb.Container

	OnHalrcmdMsgReceived  []func(*pb.Container, ...interface{})
	OnHalrcompMsgReceived []func(*pb.Container, ...interface{})
	OnStateChanged        []func(string)
	fsm                   *fsm.FSM
}

func NewRemoteComponentBase(Debuglevel int, Debugname string) *RemoteComponentBase {
	tmp := &RemoteComponentBase{}
	tmp.Debuglevel = Debuglevel
	tmp.Debugname = Debugname
	tmp.errorString = ""
	tmp.OnErrorStringChanged = make([]func(string), 0)

	// Halrcmd socket
	tmp.HalrcmdChannel = common.NewRpcClient(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "halrcmd"))
	tmp.HalrcmdChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "halrcmd")
	tmp.HalrcmdChannel.OnStateChanged = append(tmp.HalrcmdChannel.OnStateChanged, tmp.HalrcmdChannel_state_changed)
	tmp.HalrcmdChannel.OnSocketMsgReceived = append(tmp.HalrcmdChannel.OnSocketMsgReceived, tmp.HalrcmdChannelMsgReceived)
	// more efficient to reuse protobuf messages
	// XXXXX halrcmd halrcmd halrcmd
	tmp.halrcmdRx = &pb.Container{}
	tmp.halrcmdTx = &pb.Container{}

	// Halrcomp socket
	tmp.HalrcompChannel = NewHalrcompSubscribe(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "halrcomp"))
	tmp.HalrcompChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "halrcomp")
	tmp.HalrcompChannel.OnStateChanged = append(tmp.HalrcompChannel.OnStateChanged, tmp.HalrcompChannel_state_changed)
	tmp.HalrcompChannel.OnSocketMsgReceived = append(tmp.HalrcompChannel.OnSocketMsgReceived, tmp.HalrcompChannelMsgReceived)
	// more efficient to reuse protobuf messages
	// XXXXX halrcomp halrcomp halrcomp
	tmp.halrcompRx = &pb.Container{}

	// callbacks
	tmp.OnHalrcmdMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnHalrcompMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnStateChanged = make([]func(string), 1)

	// fsm
	tmp.fsm = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "connect", Src: []string{"down"}, Dst: "trying"},
			{Name: "halrcmd_up", Src: []string{"trying"}, Dst: "bind"},
			{Name: "halrcomp_bind_msg_sent", Src: []string{"bind"}, Dst: "binding"},
			{Name: "no_bind", Src: []string{"bind"}, Dst: "syncing"},
			{Name: "bind_confirmed", Src: []string{"binding"}, Dst: "syncing"},
			{Name: "bind_rejected", Src: []string{"binding"}, Dst: "error"},
			{Name: "halrcmd_trying", Src: []string{"binding", "syncing", "synced"}, Dst: "trying"},
			{Name: "disconnect", Src: []string{"trying", "binding", "syncing", "synced", "error"}, Dst: "down"},
			{Name: "halrcomp_up", Src: []string{"syncing"}, Dst: "sync"},
			{Name: "sync_failed", Src: []string{"syncing"}, Dst: "error"},
			{Name: "pins_synced", Src: []string{"sync"}, Dst: "synced"},
			{Name: "halrcomp_trying", Src: []string{"synced"}, Dst: "syncing"},
			{Name: "set_rejected", Src: []string{"synced"}, Dst: "error"},
			{Name: "halrcomp_set_msg_sent", Src: []string{"synced"}, Dst: "synced"},
		},
		fsm.Callbacks{
			"down":                         func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_connect":                func(e *fsm.Event) { tmp.OnFsm_connect(e) },
			"leave_down":                   func(e *fsm.Event) { tmp.OnFsm_down_exit(e) },
			"trying":                       func(e *fsm.Event) { tmp.OnFsm_trying(e) },
			"after_halrcmd_up":             func(e *fsm.Event) { tmp.OnFsm_halrcmd_up(e) },
			"after_disconnect":             func(e *fsm.Event) { tmp.OnFsm_disconnect(e) },
			"bind":                         func(e *fsm.Event) { tmp.OnFsm_bind(e) },
			"after_halrcomp_bind_msg_sent": func(e *fsm.Event) { tmp.OnFsm_halrcomp_bind_msg_sent(e) },
			"after_no_bind":                func(e *fsm.Event) { tmp.OnFsm_no_bind(e) },
			"binding":                      func(e *fsm.Event) { tmp.OnFsm_binding(e) },
			"after_bind_confirmed":         func(e *fsm.Event) { tmp.OnFsm_bind_confirmed(e) },
			"after_bind_rejected":          func(e *fsm.Event) { tmp.OnFsm_bind_rejected(e) },
			"after_halrcmd_trying":         func(e *fsm.Event) { tmp.OnFsm_halrcmd_trying(e) },
			"syncing":                      func(e *fsm.Event) { tmp.OnFsm_syncing(e) },
			"after_halrcomp_up":            func(e *fsm.Event) { tmp.OnFsm_halrcomp_up(e) },
			"after_sync_failed":            func(e *fsm.Event) { tmp.OnFsm_sync_failed(e) },
			"sync":                         func(e *fsm.Event) { tmp.OnFsm_sync(e) },
			"after_pins_synced":            func(e *fsm.Event) { tmp.OnFsm_pins_synced(e) },
			"synced":                       func(e *fsm.Event) { tmp.OnFsm_synced(e) },
			"after_halrcomp_trying":        func(e *fsm.Event) { tmp.OnFsm_halrcomp_trying(e) },
			"after_set_rejected":           func(e *fsm.Event) { tmp.OnFsm_set_rejected(e) },
			"after_halrcomp_set_msg_sent":  func(e *fsm.Event) { tmp.OnFsm_halrcomp_set_msg_sent(e) },
			"error":                        func(e *fsm.Event) { tmp.OnFsm_error(e) },
		},
	)
	return tmp
}

func (self *RemoteComponentBase) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN entry", self.Debugname)
	}
	self.SetDisconnected()
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *RemoteComponentBase) OnFsm_connect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event CONNECT", self.Debugname)
	}
	self.AddPins()
	self.StartHalrcmdChannel()
}

func (self *RemoteComponentBase) OnFsm_down_exit(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN exit", self.Debugname)
	}
	self.SetConnecting()
}

func (self *RemoteComponentBase) OnFsm_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state TRYING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("trying")
	}
}

func (self *RemoteComponentBase) OnFsm_halrcmd_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event HALRCMD UP", self.Debugname)
	}
	self.BindComponent()
}

func (self *RemoteComponentBase) OnFsm_disconnect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event DISCONNECT", self.Debugname)
	}
	self.StopHalrcmdChannel()
	self.StopHalrcompChannel()
	self.RemovePins()
}

func (self *RemoteComponentBase) OnFsm_bind(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state BIND", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("bind")
	}
}

func (self *RemoteComponentBase) OnFsm_halrcomp_bind_msg_sent(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event HALRCOMP BIND MSG SENT", self.Debugname)
	}
}

func (self *RemoteComponentBase) OnFsm_no_bind(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event NO BIND", self.Debugname)
	}
	self.StartHalrcompChannel()
}

func (self *RemoteComponentBase) OnFsm_binding(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state BINDING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("binding")
	}
}

func (self *RemoteComponentBase) OnFsm_bind_confirmed(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event BIND CONFIRMED", self.Debugname)
	}
	self.StartHalrcompChannel()
}

func (self *RemoteComponentBase) OnFsm_bind_rejected(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event BIND REJECTED", self.Debugname)
	}
	self.StopHalrcmdChannel()
}

func (self *RemoteComponentBase) OnFsm_halrcmd_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event HALRCMD TRYING", self.Debugname)
	}
}

func (self *RemoteComponentBase) OnFsm_syncing(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state SYNCING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("syncing")
	}
}

func (self *RemoteComponentBase) OnFsm_halrcomp_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event HALRCOMP UP", self.Debugname)
	}
}

func (self *RemoteComponentBase) OnFsm_sync_failed(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event SYNC FAILED", self.Debugname)
	}
	self.StopHalrcompChannel()
	self.StopHalrcmdChannel()
}

func (self *RemoteComponentBase) OnFsm_sync(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state SYNC", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("sync")
	}
}

func (self *RemoteComponentBase) OnFsm_pins_synced(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event PINS SYNCED", self.Debugname)
	}
}

func (self *RemoteComponentBase) OnFsm_synced(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state SYNCED entry", self.Debugname)
	}
	self.SetConnected()
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state SYNCED", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("synced")
	}
}

func (self *RemoteComponentBase) OnFsm_halrcomp_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event HALRCOMP TRYING", self.Debugname)
	}
	self.UnsyncPins()
	self.SetTimeout()
}

func (self *RemoteComponentBase) OnFsm_set_rejected(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event SET REJECTED", self.Debugname)
	}
	self.StopHalrcompChannel()
	self.StopHalrcmdChannel()
}

func (self *RemoteComponentBase) OnFsm_halrcomp_set_msg_sent(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event HALRCOMP SET MSG SENT", self.Debugname)
	}
}

func (self *RemoteComponentBase) OnFsm_error(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state ERROR entry", self.Debugname)
	}
	self.SetError()
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state ERROR", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("error")
	}
}

func (self *RemoteComponentBase) ErrorString() string {
	return self.errorString
}

func (self *RemoteComponentBase) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

func (self *RemoteComponentBase) GetHalrcmdUri() string {
	return self.HalrcmdChannel.SocketUri
}

// @halrcmd_uri.setter
func (self *RemoteComponentBase) SetHalrcmdUri(value string) {
	self.HalrcmdChannel.SocketUri = value
}

func (self *RemoteComponentBase) GetHalrcompUri() string {
	return self.HalrcompChannel.SocketUri
}

// @halrcomp_uri.setter
func (self *RemoteComponentBase) SetHalrcompUri(value string) {
	self.HalrcompChannel.SocketUri = value
}

func (self *RemoteComponentBase) BindComponent() {
	// log.Printf("WARNING: slot bind component unimplemented")
}

func (self *RemoteComponentBase) AddPins() {
	// log.Printf("WARNING: slot add pins unimplemented")
}

func (self *RemoteComponentBase) RemovePins() {
	// log.Printf("WARNING: slot remove pins unimplemented")
}

func (self *RemoteComponentBase) UnsyncPins() {
	// log.Printf("WARNING: slot unsync pins unimplemented")
}

func (self *RemoteComponentBase) SetConnected() {
	// log.Printf("WARNING: slot set connected unimplemented")
}

func (self *RemoteComponentBase) SetError() {
	// log.Printf("WARNING: slot set error unimplemented")
}

func (self *RemoteComponentBase) SetDisconnected() {
	// log.Printf("WARNING: slot set disconnected unimplemented")
}

func (self *RemoteComponentBase) SetConnecting() {
	// log.Printf("WARNING: slot set connecting unimplemented")
}

func (self *RemoteComponentBase) SetTimeout() {
	// log.Printf("WARNING: slot set timeout unimplemented")
}

// trigger
func (self *RemoteComponentBase) NoBind() {
	if self.fsm.Is("bind") {
		self.fsm.Event("no_bind")
	}
}

// trigger
func (self *RemoteComponentBase) PinsSynced() {
	if self.fsm.Is("sync") {
		self.fsm.Event("pins_synced")
	}
}

// trigger
func (self *RemoteComponentBase) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("connect")
	}
}

// trigger
func (self *RemoteComponentBase) Stop() {
	if self.fsm.Is("trying") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("binding") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("syncing") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("synced") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("error") {
		self.fsm.Event("disconnect")
	}
}

func (self *RemoteComponentBase) AddHalrcompTopic(name string) {
	self.HalrcompChannel.AddSocketTopic(name)
}

func (self *RemoteComponentBase) RemoveHalrcompTopic(name string) {
	self.HalrcompChannel.RemoveSocketTopic(name)
}

func (self *RemoteComponentBase) ClearHalrcompTopics() {
	self.HalrcompChannel.ClearSocketTopics()
}

func (self *RemoteComponentBase) StartHalrcmdChannel() {
	self.HalrcmdChannel.Start()
}

func (self *RemoteComponentBase) StopHalrcmdChannel() {
	self.HalrcmdChannel.Stop()
}

func (self *RemoteComponentBase) StartHalrcompChannel() {
	self.HalrcompChannel.Start()
}

func (self *RemoteComponentBase) StopHalrcompChannel() {
	self.HalrcompChannel.Stop()
}

// process all messages received on halrcmd
func (self *RemoteComponentBase) HalrcmdChannelMsgReceived(rx *pb.Container, rest ...interface{}) {
	// INCOMING halrcomp bind confirm 0 1 0

	// react to halrcomp bind confirm message
	if *rx.Type == pb.ContainerType_MT_HALRCOMP_BIND_CONFIRM {
		if self.fsm.Is("binding") {
			self.fsm.Event("bind_confirmed")
		}
		// INCOMING halrcomp bind reject 0 1 0

		// react to halrcomp bind reject message
	} else if *rx.Type == pb.ContainerType_MT_HALRCOMP_BIND_REJECT {
		// update error string with note
		self.errorString = ""
		for _, note := range rx.GetNote() {
			self.errorString += note + "\n"
		}
		if self.fsm.Is("binding") {
			self.fsm.Event("bind_rejected")
		}
		// INCOMING halrcomp set reject 0 1 0

		// react to halrcomp set reject message
	} else if *rx.Type == pb.ContainerType_MT_HALRCOMP_SET_REJECT {
		// update error string with note
		self.errorString = ""
		for _, note := range rx.GetNote() {
			self.errorString += note + "\n"
		}
		if self.fsm.Is("synced") {
			self.fsm.Event("set_rejected")
		}
	} //AAAAAAAA

	for _, cb := range self.OnHalrcmdMsgReceived {
		cb(rx)
	}
}

// process all messages received on halrcomp
func (self *RemoteComponentBase) HalrcompChannelMsgReceived(rx *pb.Container, rest ...interface{}) {
	//parse identity
	identity := rest[0].(string)
	// INCOMING halrcomp full update 0 0 1

	// react to halrcomp full update message
	if *rx.Type == pb.ContainerType_MT_HALRCOMP_FULL_UPDATE {
		self.HalrcompFullUpdateReceived(identity, rx)
		// INCOMING halrcomp incremental update 0 0 1

		// react to halrcomp incremental update message
	} else if *rx.Type == pb.ContainerType_MT_HALRCOMP_INCREMENTAL_UPDATE {
		self.HalrcompIncrementalUpdateReceived(identity, rx)
		// INCOMING halrcomp error 0 1 1

		// react to halrcomp error message
	} else if *rx.Type == pb.ContainerType_MT_HALRCOMP_ERROR {
		// update error string with note
		self.errorString = ""
		for _, note := range rx.GetNote() {
			self.errorString += note + "\n"
		}
		if self.fsm.Is("syncing") {
			self.fsm.Event("sync_failed")
		}
		self.HalrcompErrorReceived(identity, rx)
	} //AAAAAAAA

	for _, cb := range self.OnHalrcompMsgReceived {
		cb(rx, string(identity))
	}
}

func (self *RemoteComponentBase) HalrcompFullUpdateReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT halrcomp full update unimplemented")
}

func (self *RemoteComponentBase) HalrcompIncrementalUpdateReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT halrcomp incremental update unimplemented")
}

func (self *RemoteComponentBase) HalrcompErrorReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT halrcomp error unimplemented")
}

func (self *RemoteComponentBase) SendHalrcmdMsg(msg_type pb.ContainerType, tx *pb.Container) {
	self.HalrcmdChannel.SendSocketMessage(msg_type, tx)
	if msg_type == pb.MT_HALRCOMP_BIND {
		if self.fsm.Is("bind") {
			self.fsm.Event("halrcomp_bind_msg_sent")
		}
	} else if msg_type == pb.MT_HALRCOMP_SET {
		if self.fsm.Is("synced") {
			self.fsm.Event("halrcomp_set_msg_sent")
		}
	} // A
}
func (self *RemoteComponentBase) SendHalrcompBind(tx *pb.Container) {
	self.SendHalrcmdMsg(pb.ContainerType_MT_HALRCOMP_BIND, tx)
}
func (self *RemoteComponentBase) SendHalrcompSet(tx *pb.Container) {
	self.SendHalrcmdMsg(pb.ContainerType_MT_HALRCOMP_SET, tx)
}
func (self *RemoteComponentBase) HalrcmdChannel_state_changed(state string) {

	if state == "trying" {
		if self.fsm.Is("syncing") {
			self.fsm.Event("halrcmd_trying")
		} else if self.fsm.Is("synced") {
			self.fsm.Event("halrcmd_trying")
		} else if self.fsm.Is("binding") {
			self.fsm.Event("halrcmd_trying")
		}

	} else if state == "up" {
		if self.fsm.Is("trying") {
			self.fsm.Event("halrcmd_up")
		}
	}
}
func (self *RemoteComponentBase) HalrcompChannel_state_changed(state string) {

	if state == "trying" {
		if self.fsm.Is("synced") {
			self.fsm.Event("halrcomp_trying")
		}

	} else if state == "up" {
		if self.fsm.Is("syncing") {
			self.fsm.Event("halrcomp_up")
		}
	}
}
