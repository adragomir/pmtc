package param

import (
	"fmt"
	"log"
	"sync"

	"github.com/looplab/fsm"

	"github.com/machinekit/machinetalk_go/common"

	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type ParamServer struct {
	Debuglevel           int
	Debugname            string
	errorString          string
	OnErrorStringChanged []func(string)
	txLock               sync.Mutex
	// Paramcmd socket
	ParamcmdChannel *common.RpcService
	// more efficient to reuse protobuf messages
	paramcmdRx *pb.Container
	// Param socket
	ParamChannel *common.Publish
	// more efficient to reuse protobuf messages
	paramTx *pb.Container

	OnParamcmdMsgReceived []func(*pb.Container, ...interface{})
	OnStateChanged        []func(string)
	fsm                   *fsm.FSM
}

func NewParamServer(Debuglevel int, Debugname string) *ParamServer {
	tmp := &ParamServer{}
	tmp.Debuglevel = Debuglevel
	tmp.Debugname = Debugname
	tmp.errorString = ""
	tmp.OnErrorStringChanged = make([]func(string), 0)

	// Paramcmd socket
	tmp.ParamcmdChannel = common.NewRpcService(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "paramcmd"))
	tmp.ParamcmdChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "paramcmd")
	tmp.ParamcmdChannel.OnSocketMsgReceived = append(tmp.ParamcmdChannel.OnSocketMsgReceived, tmp.ParamcmdChannelMsgReceived)
	// more efficient to reuse protobuf messages
	// XXXXX paramcmd paramcmd paramcmd
	tmp.paramcmdRx = &pb.Container{}

	// Param socket
	tmp.ParamChannel = common.NewPublish(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "param"))
	tmp.ParamChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "param")
	// more efficient to reuse protobuf messages
	tmp.paramTx = &pb.Container{}

	// callbacks
	tmp.OnParamcmdMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
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

func (self *ParamServer) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *ParamServer) OnFsm_connect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event CONNECT", self.Debugname)
	}
	self.StartParamcmdChannel()
	self.StartParamChannel()
}

func (self *ParamServer) OnFsm_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("up")
	}
}

func (self *ParamServer) OnFsm_disconnect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event DISCONNECT", self.Debugname)
	}
	self.StopParamcmdChannel()
	self.StopParamChannel()
}

func (self *ParamServer) ErrorString() string {
	return self.errorString
}

func (self *ParamServer) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

func (self *ParamServer) GetParamcmdUri() string {
	return self.ParamcmdChannel.SocketUri
}

// @paramcmd_uri.setter
func (self *ParamServer) SetParamcmdUri(value string) {
	self.ParamcmdChannel.SocketUri = value
}

func (self *ParamServer) GetParamcmdPort() int {
	return self.ParamcmdChannel.GetSocketPort()
}

func (self *ParamServer) GetParamcmdDsn() string {
	return self.ParamcmdChannel.GetSocketDsn()
}

func (self *ParamServer) GetParamUri() string {
	return self.ParamChannel.SocketUri
}

// @param_uri.setter
func (self *ParamServer) SetParamUri(value string) {
	self.ParamChannel.SocketUri = value
}

func (self *ParamServer) GetParamPort() int {
	return self.ParamChannel.GetSocketPort()
}

func (self *ParamServer) GetParamDsn() string {
	return self.ParamChannel.GetSocketDsn()
}

func (self *ParamServer) AddParamTopic(name string) {
	self.ParamChannel.AddSocketTopic(name)
}

func (self *ParamServer) RemoveParamTopic(name string) {
	self.ParamChannel.RemoveSocketTopic(name)
}

func (self *ParamServer) ClearParamTopics() {
	self.ParamChannel.ClearSocketTopics()
}

func (self *ParamServer) StartParamcmdChannel() {
	self.ParamcmdChannel.Start()
}

func (self *ParamServer) StopParamcmdChannel() {
	self.ParamcmdChannel.Stop()
}

func (self *ParamServer) StartParamChannel() {
	self.ParamChannel.Start()
}

func (self *ParamServer) StopParamChannel() {
	self.ParamChannel.Stop()
}

// process all messages received on paramcmd
func (self *ParamServer) ParamcmdChannelMsgReceived(rx *pb.Container, rest ...interface{}) {
	//parse identity
	identity := rest[0].(string)
	// INCOMING incremental update 0 0 1

	// react to incremental update message
	if *rx.Type == pb.ContainerType_MT_INCREMENTAL_UPDATE {
		self.IncrementalUpdateReceived(identity, rx)
	} //AAAAAAAA

	for _, cb := range self.OnParamcmdMsgReceived {
		cb(rx, string(identity))
	}
}

func (self *ParamServer) IncrementalUpdateReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT incremental update unimplemented")
}

func (self *ParamServer) SendParamMsg(identity string, msg_type pb.ContainerType, tx *pb.Container) {
	self.ParamChannel.SendSocketMessage(identity, msg_type, tx)
}
func (self *ParamServer) SendFullUpdate(identity string, tx *pb.Container) {
	ids := map[string]bool{
		identity: true,
	}
	for receiver, _ := range ids {
		self.SendParamMsg(receiver, pb.ContainerType_MT_FULL_UPDATE, tx)
	}
}
func (self *ParamServer) SendIncrementalUpdate(identity string, tx *pb.Container) {
	ids := map[string]bool{
		identity: true,
	}
	for receiver, _ := range ids {
		self.SendParamMsg(receiver, pb.ContainerType_MT_INCREMENTAL_UPDATE, tx)
	}
}
