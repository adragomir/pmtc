package application

import (
	"fmt"
	"log"
	"sync"

	"github.com/looplab/fsm"

	"github.com/machinekit/machinetalk_go/common"
	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type ConfigBase struct {
	Debuglevel           int
	Debugname            string
	errorString          string
	OnErrorStringChanged []func(string)
	txLock               sync.Mutex
	// Config socket
	ConfigChannel *common.RpcClient
	// more efficient to reuse protobuf messages
	configRx *pb.Container
	configTx *pb.Container

	OnConfigMsgReceived []func(*pb.Container, ...interface{})
	OnStateChanged      []func(string)
	fsm                 *fsm.FSM
}

func NewConfigBase(Debuglevel int, Debugname string) *ConfigBase {
	tmp := &ConfigBase{}
	tmp.Debuglevel = Debuglevel
	tmp.Debugname = Debugname
	tmp.errorString = ""
	tmp.OnErrorStringChanged = make([]func(string), 0)

	// Config socket
	tmp.ConfigChannel = common.NewRpcClient(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "config"))
	tmp.ConfigChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "config")
	tmp.ConfigChannel.OnStateChanged = append(tmp.ConfigChannel.OnStateChanged, tmp.ConfigChannel_state_changed)
	tmp.ConfigChannel.OnSocketMsgReceived = append(tmp.ConfigChannel.OnSocketMsgReceived, tmp.ConfigChannelMsgReceived)
	// more efficient to reuse protobuf messages
	// XXXXX config config config
	tmp.configRx = &pb.Container{}
	tmp.configTx = &pb.Container{}

	// callbacks
	tmp.OnConfigMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnStateChanged = make([]func(string), 0)

	// fsm
	tmp.fsm = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "connect", Src: []string{"down"}, Dst: "trying"},
			{Name: "config_up", Src: []string{"trying"}, Dst: "listing"},
			{Name: "disconnect", Src: []string{"listing", "loading", "up", "trying"}, Dst: "down"},
			{Name: "application_retrieved", Src: []string{"listing"}, Dst: "up"},
			{Name: "config_trying", Src: []string{"up", "listing"}, Dst: "trying"},
			{Name: "load_application", Src: []string{"up"}, Dst: "loading"},
			{Name: "application_loaded", Src: []string{"loading"}, Dst: "up"},
			{Name: "config_trying", Src: []string{"loading"}, Dst: "trying"},
		},
		fsm.Callbacks{
			"down":                        func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_connect":               func(e *fsm.Event) { tmp.OnFsm_connect(e) },
			"trying":                      func(e *fsm.Event) { tmp.OnFsm_trying(e) },
			"after_config_up":             func(e *fsm.Event) { tmp.OnFsm_config_up(e) },
			"after_disconnect":            func(e *fsm.Event) { tmp.OnFsm_disconnect(e) },
			"listing":                     func(e *fsm.Event) { tmp.OnFsm_listing(e) },
			"after_application_retrieved": func(e *fsm.Event) { tmp.OnFsm_application_retrieved(e) },
			"after_config_trying":         func(e *fsm.Event) { tmp.OnFsm_config_trying(e) },
			"up":                          func(e *fsm.Event) { tmp.OnFsm_up(e) },
			"after_load_application":      func(e *fsm.Event) { tmp.OnFsm_load_application(e) },
			"leave_up":                    func(e *fsm.Event) { tmp.OnFsm_up_exit(e) },
			"loading":                     func(e *fsm.Event) { tmp.OnFsm_loading(e) },
			"after_application_loaded":    func(e *fsm.Event) { tmp.OnFsm_application_loaded(e) },
		},
	)
	return tmp
}

func (self *ConfigBase) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *ConfigBase) OnFsm_connect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event CONNECT", self.Debugname)
	}
	self.StartConfigChannel()
}

func (self *ConfigBase) OnFsm_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state TRYING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("trying")
	}
}

func (self *ConfigBase) OnFsm_config_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event CONFIG UP", self.Debugname)
	}
	self.SendListApplications()
}

func (self *ConfigBase) OnFsm_disconnect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event DISCONNECT", self.Debugname)
	}
	self.StopConfigChannel()
}

func (self *ConfigBase) OnFsm_listing(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state LISTING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("listing")
	}
}

func (self *ConfigBase) OnFsm_application_retrieved(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event APPLICATION RETRIEVED", self.Debugname)
	}
}

func (self *ConfigBase) OnFsm_config_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event CONFIG TRYING", self.Debugname)
	}
}

func (self *ConfigBase) OnFsm_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP entry", self.Debugname)
	}
	self.SyncConfig()
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("up")
	}
}

func (self *ConfigBase) OnFsm_load_application(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event LOAD APPLICATION", self.Debugname)
	}
}

func (self *ConfigBase) OnFsm_up_exit(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP exit", self.Debugname)
	}
	self.UnsyncConfig()
}

func (self *ConfigBase) OnFsm_loading(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state LOADING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("loading")
	}
}

func (self *ConfigBase) OnFsm_application_loaded(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event APPLICATION LOADED", self.Debugname)
	}
}

func (self *ConfigBase) ErrorString() string {
	return self.errorString
}

func (self *ConfigBase) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

func (self *ConfigBase) GetConfigUri() string {
	return self.ConfigChannel.SocketUri
}

// @config_uri.setter
func (self *ConfigBase) SetConfigUri(value string) {
	self.ConfigChannel.SocketUri = value
}

func (self *ConfigBase) SyncConfig() {
	// log.Printf("WARNING: slot sync config unimplemented")
}

func (self *ConfigBase) UnsyncConfig() {
	// log.Printf("WARNING: slot unsync config unimplemented")
}

// trigger
func (self *ConfigBase) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("connect")
	}
}

// trigger
func (self *ConfigBase) Stop() {
	if self.fsm.Is("trying") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("listing") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("up") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("loading") {
		self.fsm.Event("disconnect")
	}
}

func (self *ConfigBase) StartConfigChannel() {
	self.ConfigChannel.Start()
}

func (self *ConfigBase) StopConfigChannel() {
	self.ConfigChannel.Stop()
}

// process all messages received on config
func (self *ConfigBase) ConfigChannelMsgReceived(rx *pb.Container, rest ...interface{}) {
	// INCOMING describe application 0 1 1

	// react to describe application message
	if *rx.Type == pb.ContainerType_MT_DESCRIBE_APPLICATION {
		if self.fsm.Is("listing") {
			self.fsm.Event("application_retrieved")
		}
		self.DescribeApplicationReceived(rx)
		// INCOMING application detail 0 1 1

		// react to application detail message
	} else if *rx.Type == pb.ContainerType_MT_APPLICATION_DETAIL {
		if self.fsm.Is("loading") {
			self.fsm.Event("application_loaded")
		}
		self.ApplicationDetailReceived(rx)
		// INCOMING error 0 0 1

		// react to error message
	} else if *rx.Type == pb.ContainerType_MT_ERROR {
		// update error string with note
		self.errorString = ""
		for _, note := range rx.GetNote() {
			self.errorString += note + "\n"
		}
	} //AAAAAAAA

	for _, cb := range self.OnConfigMsgReceived {
		cb(rx)
	}
}

func (self *ConfigBase) DescribeApplicationReceived(rx *pb.Container) {
	// log.Printf("SLOT describe application unimplemented")
}

func (self *ConfigBase) ApplicationDetailReceived(rx *pb.Container) {
	// log.Printf("SLOT application detail unimplemented")
}

func (self *ConfigBase) SendConfigMsg(msg_type pb.ContainerType, tx *pb.Container) {
	self.ConfigChannel.SendSocketMsg(msg_type, tx)
	if msg_type == pb.ContainerType_MT_RETRIEVE_APPLICATION {
		if self.fsm.Is("up") {
			self.fsm.Event("load_application")
		}
	} // A
}
func (self *ConfigBase) SendListApplications() {
	tx := self.configTx
	self.SendConfigMsg(pb.ContainerType_MT_LIST_APPLICATIONS, tx)
}
func (self *ConfigBase) SendRetrieveApplication(tx *pb.Container) {
	self.SendConfigMsg(pb.ContainerType_MT_RETRIEVE_APPLICATION, tx)
}
func (self *ConfigBase) ConfigChannel_state_changed(state string) {

	if state == "trying" {
		if self.fsm.Is("listing") {
			self.fsm.Event("config_trying")
		} else if self.fsm.Is("up") {
			self.fsm.Event("config_trying")
		} else if self.fsm.Is("loading") {
			self.fsm.Event("config_trying")
		}

	} else if state == "up" {
		if self.fsm.Is("trying") {
			self.fsm.Event("config_up")
		}
	}
}
