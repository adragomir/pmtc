package param

import (
	"fmt"
	"log"
	"sync"

	"github.com/looplab/fsm"
	"github.com/machinekit/machinetalk_go/common"

	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type ParamClient struct {
	Debuglevel           int
	Debugname            string
	errorString          string
	OnErrorStringChanged []func(string)
	txLock               sync.Mutex
	// Paramcmd socket
	ParamcmdChannel *common.RpcClient
	// more efficient to reuse protobuf messages
	paramcmdTx *pb.Container
	// Param socket
	ParamChannel *common.Subscribe
	// more efficient to reuse protobuf messages
	paramRx *pb.Container

	OnParamMsgReceived []func(*pb.Container, ...interface{})
	OnStateChanged     []func(string)
	fsm                *fsm.FSM
}

func NewParamClient(Debuglevel int, Debugname string) *ParamClient {
	tmp := &ParamClient{}
	tmp.Debuglevel = Debuglevel
	tmp.Debugname = Debugname
	tmp.errorString = ""
	tmp.OnErrorStringChanged = make([]func(string), 0)

	// Paramcmd socket
	tmp.ParamcmdChannel = common.NewRpcClient(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "paramcmd"))
	tmp.ParamcmdChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "paramcmd")
	tmp.ParamcmdChannel.OnStateChanged = append(tmp.ParamcmdChannel.OnStateChanged, tmp.ParamcmdChannel_state_changed)
	// more efficient to reuse protobuf messages
	tmp.paramcmdTx = &pb.Container{}

	// Param socket
	tmp.ParamChannel = common.NewSubscribe(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "param"))
	tmp.ParamChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "param")
	tmp.ParamChannel.OnStateChanged = append(tmp.ParamChannel.OnStateChanged, tmp.ParamChannel_state_changed)
	tmp.ParamChannel.OnSocketMsgReceived = append(tmp.ParamChannel.OnSocketMsgReceived, tmp.ParamChannelMsgReceived)
	// more efficient to reuse protobuf messages
	// XXXXX param param param
	tmp.paramRx = &pb.Container{}

	// callbacks
	tmp.OnParamMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnStateChanged = make([]func(string), 0)

	// fsm
	tmp.fsm = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "connect", Src: []string{"down"}, Dst: "connecting"},
			{Name: "paramcmd_up", Src: []string{"connecting"}, Dst: "syncing"},
			{Name: "param_up", Src: []string{"connecting"}, Dst: "trying"},
			{Name: "disconnect", Src: []string{"connecting", "syncing", "trying", "up"}, Dst: "down"},
			{Name: "param_up", Src: []string{"syncing"}, Dst: "up"},
			{Name: "paramcmd_trying", Src: []string{"syncing"}, Dst: "connecting"},
			{Name: "paramcmd_up", Src: []string{"trying"}, Dst: "up"},
			{Name: "param_trying", Src: []string{"trying"}, Dst: "connecting"},
			{Name: "paramcmd_trying", Src: []string{"up"}, Dst: "trying"},
			{Name: "param_trying", Src: []string{"up"}, Dst: "syncing"},
		},
		fsm.Callbacks{
			"down":                  func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_connect":         func(e *fsm.Event) { tmp.OnFsm_connect(e) },
			"connecting":            func(e *fsm.Event) { tmp.OnFsm_connecting(e) },
			"after_paramcmd_up":     func(e *fsm.Event) { tmp.OnFsm_paramcmd_up(e) },
			"after_param_up":        func(e *fsm.Event) { tmp.OnFsm_param_up(e) },
			"after_disconnect":      func(e *fsm.Event) { tmp.OnFsm_disconnect(e) },
			"syncing":               func(e *fsm.Event) { tmp.OnFsm_syncing(e) },
			"after_paramcmd_trying": func(e *fsm.Event) { tmp.OnFsm_paramcmd_trying(e) },
			"trying":                func(e *fsm.Event) { tmp.OnFsm_trying(e) },
			"after_param_trying":    func(e *fsm.Event) { tmp.OnFsm_param_trying(e) },
			"up":                    func(e *fsm.Event) { tmp.OnFsm_up(e) },
			"leave_up":              func(e *fsm.Event) { tmp.OnFsm_up_exit(e) },
		},
	)
	return tmp
}

func (self *ParamClient) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *ParamClient) OnFsm_connect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event CONNECT", self.Debugname)
	}
	self.StartParamcmdChannel()
	self.StartParamChannel()
}

func (self *ParamClient) OnFsm_connecting(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state CONNECTING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("connecting")
	}
}

func (self *ParamClient) OnFsm_paramcmd_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event PARAMCMD UP", self.Debugname)
	}
}

func (self *ParamClient) OnFsm_param_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event PARAM UP", self.Debugname)
	}
}

func (self *ParamClient) OnFsm_disconnect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event DISCONNECT", self.Debugname)
	}
	self.StopParamcmdChannel()
	self.StopParamChannel()
	self.RemoveKeys()
}

func (self *ParamClient) OnFsm_syncing(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state SYNCING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("syncing")
	}
}

func (self *ParamClient) OnFsm_paramcmd_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event PARAMCMD TRYING", self.Debugname)
	}
}

func (self *ParamClient) OnFsm_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state TRYING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("trying")
	}
}

func (self *ParamClient) OnFsm_param_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event PARAM TRYING", self.Debugname)
	}
}

func (self *ParamClient) OnFsm_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP entry", self.Debugname)
	}
	self.SetSynced()
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("up")
	}
}

func (self *ParamClient) OnFsm_up_exit(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP exit", self.Debugname)
	}
	self.ClearSynced()
	self.UnsyncKeys()
}

func (self *ParamClient) ErrorString() string {
	return self.errorString
}

func (self *ParamClient) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

func (self *ParamClient) GetParamcmdUri() string {
	return self.ParamcmdChannel.SocketUri
}

// @paramcmd_uri.setter
func (self *ParamClient) SetParamcmdUri(value string) {
	self.ParamcmdChannel.SocketUri = value
}

func (self *ParamClient) GetParamUri() string {
	return self.ParamChannel.SocketUri
}

// @param_uri.setter
func (self *ParamClient) SetParamUri(value string) {
	self.ParamChannel.SocketUri = value
}

func (self *ParamClient) RemoveKeys() {
	// log.Printf("WARNING: slot remove keys unimplemented")
}

func (self *ParamClient) UnsyncKeys() {
	// log.Printf("WARNING: slot unsync keys unimplemented")
}

func (self *ParamClient) SetSynced() {
	// log.Printf("WARNING: slot set synced unimplemented")
}

func (self *ParamClient) ClearSynced() {
	// log.Printf("WARNING: slot clear synced unimplemented")
}

// trigger
func (self *ParamClient) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("connect")
	}
}

// trigger
func (self *ParamClient) Stop() {
	if self.fsm.Is("connecting") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("syncing") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("trying") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("up") {
		self.fsm.Event("disconnect")
	}
}

func (self *ParamClient) AddParamTopic(name string) {
	self.ParamChannel.AddSocketTopic(name)
}

func (self *ParamClient) RemoveParamTopic(name string) {
	self.ParamChannel.RemoveSocketTopic(name)
}

func (self *ParamClient) ClearParamTopics() {
	self.ParamChannel.ClearSocketTopics()
}

func (self *ParamClient) StartParamcmdChannel() {
	self.ParamcmdChannel.Start()
}

func (self *ParamClient) StopParamcmdChannel() {
	self.ParamcmdChannel.Stop()
}

func (self *ParamClient) StartParamChannel() {
	self.ParamChannel.Start()
}

func (self *ParamClient) StopParamChannel() {
	self.ParamChannel.Stop()
}

// process all messages received on param
func (self *ParamClient) ParamChannelMsgReceived(rx *pb.Container, rest ...interface{}) {
	//parse identity
	identity := rest[0].(string)
	// INCOMING full update 0 0 1

	// react to full update message
	if *rx.Type == pb.ContainerType_MT_FULL_UPDATE {
		self.FullUpdateReceived(identity, rx)
		// INCOMING incremental update 0 0 1

		// react to incremental update message
	} else if *rx.Type == pb.ContainerType_MT_INCREMENTAL_UPDATE {
		self.IncrementalUpdateReceived(identity, rx)
	} //AAAAAAAA

	for _, cb := range self.OnParamMsgReceived {
		cb(rx, string(identity))
	}
}

func (self *ParamClient) FullUpdateReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT full update unimplemented")
}

func (self *ParamClient) IncrementalUpdateReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT incremental update unimplemented")
}

func (self *ParamClient) SendParamcmdMsg(msg_type pb.ContainerType, tx *pb.Container) {
	self.ParamcmdChannel.SendSocketMessage(msg_type, tx)
}
func (self *ParamClient) SendIncrementalUpdate(tx *pb.Container) {
	self.SendParamcmdMsg(pb.ContainerType_MT_INCREMENTAL_UPDATE, tx)
}
func (self *ParamClient) ParamcmdChannel_state_changed(state string) {

	if state == "trying" {
		if self.fsm.Is("syncing") {
			self.fsm.Event("paramcmd_trying")
		} else if self.fsm.Is("up") {
			self.fsm.Event("paramcmd_trying")
		}

	} else if state == "up" {
		if self.fsm.Is("trying") {
			self.fsm.Event("paramcmd_up")
		} else if self.fsm.Is("connecting") {
			self.fsm.Event("paramcmd_up")
		}
	}
}
func (self *ParamClient) ParamChannel_state_changed(state string) {

	if state == "trying" {
		if self.fsm.Is("trying") {
			self.fsm.Event("param_trying")
		} else if self.fsm.Is("up") {
			self.fsm.Event("param_trying")
		}

	} else if state == "up" {
		if self.fsm.Is("syncing") {
			self.fsm.Event("param_up")
		} else if self.fsm.Is("connecting") {
			self.fsm.Event("param_up")
		}
	}
}
