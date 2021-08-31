package application

import (
	"fmt"
	"log"
	"sync"

	"github.com/looplab/fsm"

	"github.com/machinekit/machinetalk_go/common"
	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type LauncherBase struct {
	Debuglevel           int
	Debugname            string
	errorString          string
	OnErrorStringChanged []func(string)
	txLock               sync.Mutex
	// Launchercmd socket
	LaunchercmdChannel *common.RpcClient
	// more efficient to reuse protobuf messages
	launchercmdRx *pb.Container
	launchercmdTx *pb.Container
	// Launcher socket
	LauncherChannel *LauncherSubscribe
	// more efficient to reuse protobuf messages
	launcherRx *pb.Container

	OnLaunchercmdMsgReceived []func(*pb.Container, ...interface{})
	OnLauncherMsgReceived    []func(*pb.Container, ...interface{})
	OnStateChanged           []func(string)
	fsm                      *fsm.FSM
}

func NewLauncherBase(Debuglevel int, Debugname string) *LauncherBase {
	tmp := &LauncherBase{}
	tmp.Debuglevel = Debuglevel
	tmp.Debugname = Debugname
	tmp.errorString = ""
	tmp.OnErrorStringChanged = make([]func(string), 0)

	// Launchercmd socket
	tmp.LaunchercmdChannel = common.NewRpcClient(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "launchercmd"))
	tmp.LaunchercmdChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "launchercmd")
	tmp.LaunchercmdChannel.OnStateChanged = append(tmp.LaunchercmdChannel.OnStateChanged, tmp.LaunchercmdChannel_state_changed)
	tmp.LaunchercmdChannel.OnSocketMsgReceived = append(tmp.LaunchercmdChannel.OnSocketMsgReceived, tmp.LaunchercmdChannelMsgReceived)
	// more efficient to reuse protobuf messages
	// XXXXX launchercmd launchercmd launchercmd
	tmp.launchercmdRx = &pb.Container{}
	tmp.launchercmdTx = &pb.Container{}

	// Launcher socket
	tmp.LauncherChannel = NewLauncherSubscribe(Debuglevel, fmt.Sprintf("%s - %s", tmp.Debugname, "launcher"))
	tmp.LauncherChannel.Debugname = fmt.Sprintf("%s - %s", tmp.Debugname, "launcher")
	tmp.LauncherChannel.OnStateChanged = append(tmp.LauncherChannel.OnStateChanged, tmp.LauncherChannel_state_changed)
	tmp.LauncherChannel.OnSocketMsgReceived = append(tmp.LauncherChannel.OnSocketMsgReceived, tmp.LauncherChannelMsgReceived)
	// more efficient to reuse protobuf messages
	// XXXXX launcher launcher launcher
	tmp.launcherRx = &pb.Container{}

	// callbacks
	tmp.OnLaunchercmdMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnLauncherMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnStateChanged = make([]func(string), 0)

	// fsm
	tmp.fsm = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "connect", Src: []string{"down"}, Dst: "trying"},
			{Name: "launchercmd_up", Src: []string{"trying"}, Dst: "syncing"},
			{Name: "disconnect", Src: []string{"trying", "syncing", "synced"}, Dst: "down"},
			{Name: "launchercmd_trying", Src: []string{"synced", "syncing"}, Dst: "trying"},
			{Name: "launcher_up", Src: []string{"syncing"}, Dst: "synced"},
			{Name: "launcher_trying", Src: []string{"synced"}, Dst: "syncing"},
		},
		fsm.Callbacks{
			"down":                     func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_connect":            func(e *fsm.Event) { tmp.OnFsm_connect(e) },
			"trying":                   func(e *fsm.Event) { tmp.OnFsm_trying(e) },
			"after_launchercmd_up":     func(e *fsm.Event) { tmp.OnFsm_launchercmd_up(e) },
			"after_disconnect":         func(e *fsm.Event) { tmp.OnFsm_disconnect(e) },
			"syncing":                  func(e *fsm.Event) { tmp.OnFsm_syncing(e) },
			"after_launchercmd_trying": func(e *fsm.Event) { tmp.OnFsm_launchercmd_trying(e) },
			"after_launcher_up":        func(e *fsm.Event) { tmp.OnFsm_launcher_up(e) },
			"synced":                   func(e *fsm.Event) { tmp.OnFsm_synced(e) },
			"after_launcher_trying":    func(e *fsm.Event) { tmp.OnFsm_launcher_trying(e) },
			"leave_synced":             func(e *fsm.Event) { tmp.OnFsm_synced_exit(e) },
		},
	)
	return tmp
}

func (self *LauncherBase) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *LauncherBase) OnFsm_connect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event CONNECT", self.Debugname)
	}
	self.StartLaunchercmdChannel()
}

func (self *LauncherBase) OnFsm_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state TRYING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("trying")
	}
}

func (self *LauncherBase) OnFsm_launchercmd_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event LAUNCHERCMD UP", self.Debugname)
	}
	self.StartLauncherChannel()
}

func (self *LauncherBase) OnFsm_disconnect(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event DISCONNECT", self.Debugname)
	}
	self.StopLaunchercmdChannel()
	self.StopLauncherChannel()
}

func (self *LauncherBase) OnFsm_syncing(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state SYNCING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("syncing")
	}
}

func (self *LauncherBase) OnFsm_launchercmd_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event LAUNCHERCMD TRYING", self.Debugname)
	}
	self.StopLauncherChannel()
}

func (self *LauncherBase) OnFsm_launcher_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event LAUNCHER UP", self.Debugname)
	}
}

func (self *LauncherBase) OnFsm_synced(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state SYNCED entry", self.Debugname)
	}
	self.SyncStatus()
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state SYNCED", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("synced")
	}
}

func (self *LauncherBase) OnFsm_launcher_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event LAUNCHER TRYING", self.Debugname)
	}
}

func (self *LauncherBase) OnFsm_synced_exit(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state SYNCED exit", self.Debugname)
	}
	self.UnsyncStatus()
}

func (self *LauncherBase) ErrorString() string {
	return self.errorString
}

func (self *LauncherBase) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

func (self *LauncherBase) GetLaunchercmdUri() string {
	return self.LaunchercmdChannel.SocketUri
}

// @launchercmd_uri.setter
func (self *LauncherBase) SetLaunchercmdUri(value string) {
	self.LaunchercmdChannel.SocketUri = value
}

func (self *LauncherBase) GetLauncherUri() string {
	return self.LauncherChannel.SocketUri
}

// @launcher_uri.setter
func (self *LauncherBase) SetLauncherUri(value string) {
	self.LauncherChannel.SocketUri = value
}

func (self *LauncherBase) SyncStatus() {
	// log.Printf("[%s] WARNING: slot sync status unimplemented", self.Debugname)
}

func (self *LauncherBase) UnsyncStatus() {
	// log.Printf("[%s] WARNING: slot unsync status unimplemented", self.Debugname)
}

// trigger
func (self *LauncherBase) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("connect")
	}
}

// trigger
func (self *LauncherBase) Stop() {
	if self.fsm.Is("trying") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("syncing") {
		self.fsm.Event("disconnect")
	} else if self.fsm.Is("synced") {
		self.fsm.Event("disconnect")
	}
}

func (self *LauncherBase) AddLauncherTopic(name string) {
	self.LauncherChannel.AddSocketTopic(name)
}

func (self *LauncherBase) RemoveLauncherTopic(name string) {
	self.LauncherChannel.RemoveSocketTopic(name)
}

func (self *LauncherBase) ClearLauncherTopics() {
	self.LauncherChannel.ClearSocketTopics()
}

func (self *LauncherBase) StartLaunchercmdChannel() {
	self.LaunchercmdChannel.Start()
}

func (self *LauncherBase) StopLaunchercmdChannel() {
	self.LaunchercmdChannel.Stop()
}

func (self *LauncherBase) StartLauncherChannel() {
	self.LauncherChannel.Start()
}

func (self *LauncherBase) StopLauncherChannel() {
	self.LauncherChannel.Stop()
}

// process all messages received on launchercmd
func (self *LauncherBase) LaunchercmdChannelMsgReceived(rx *pb.Container, rest ...interface{}) {
	// INCOMING error 0 0 1

	// react to error message
	if *rx.Type == pb.ContainerType_MT_ERROR {
		// update error string with note
		self.errorString = ""
		for _, note := range rx.GetNote() {
			self.errorString += note + "\n"
		}
	} //AAAAAAAA

	for _, cb := range self.OnLaunchercmdMsgReceived {
		cb(rx)
	}
}

// process all messages received on launcher
func (self *LauncherBase) LauncherChannelMsgReceived(rx *pb.Container, rest ...interface{}) {
	//parse identity
	identity := rest[0].(string)

	// INCOMING launcher full update 0 0 1

	// react to launcher full update message
	if *rx.Type == pb.ContainerType_MT_LAUNCHER_FULL_UPDATE {
		self.LauncherFullUpdateReceived(identity, rx)
		// INCOMING launcher incremental update 0 0 1

		// react to launcher incremental update message
	} else if *rx.Type == pb.ContainerType_MT_LAUNCHER_INCREMENTAL_UPDATE {
		self.LauncherIncrementalUpdateReceived(identity, rx)
	} //AAAAAAAA

	for _, cb := range self.OnLauncherMsgReceived {
		cb(rx, string(identity))
	}
}

func (self *LauncherBase) LauncherFullUpdateReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT launcher full update unimplemented")
}

func (self *LauncherBase) LauncherIncrementalUpdateReceived(identity string, rx *pb.Container) {
	// log.Printf("SLOT launcher incremental update unimplemented")
}

func (self *LauncherBase) SendLaunchercmdMsg(msg_type pb.ContainerType, tx *pb.Container) {
	self.LaunchercmdChannel.SendSocketMsg(msg_type, tx)
}
func (self *LauncherBase) SendLauncherStart(tx *pb.Container) {
	self.SendLaunchercmdMsg(pb.ContainerType_MT_LAUNCHER_START, tx)
}
func (self *LauncherBase) SendLauncherKill(tx *pb.Container) {
	self.SendLaunchercmdMsg(pb.ContainerType_MT_LAUNCHER_KILL, tx)
}
func (self *LauncherBase) SendLauncherTerminate(tx *pb.Container) {
	self.SendLaunchercmdMsg(pb.ContainerType_MT_LAUNCHER_TERMINATE, tx)
}
func (self *LauncherBase) SendLauncherWriteStdin(tx *pb.Container) {
	self.SendLaunchercmdMsg(pb.ContainerType_MT_LAUNCHER_WRITE_STDIN, tx)
}
func (self *LauncherBase) SendLauncherCall(tx *pb.Container) {
	self.SendLaunchercmdMsg(pb.ContainerType_MT_LAUNCHER_CALL, tx)
}
func (self *LauncherBase) SendLauncherShutdown(tx *pb.Container) {
	self.SendLaunchercmdMsg(pb.ContainerType_MT_LAUNCHER_SHUTDOWN, tx)
}
func (self *LauncherBase) SendLauncherSet(tx *pb.Container) {
	self.SendLaunchercmdMsg(pb.ContainerType_MT_LAUNCHER_SET, tx)
}
func (self *LauncherBase) LaunchercmdChannel_state_changed(state string) {
	if state == "trying" {
		if self.fsm.Is("syncing") {
			self.fsm.Event("launchercmd_trying")
		} else if self.fsm.Is("synced") {
			self.fsm.Event("launchercmd_trying")
		}

	} else if state == "up" {
		if self.fsm.Is("trying") {
			self.fsm.Event("launchercmd_up")
		}
	}
}
func (self *LauncherBase) LauncherChannel_state_changed(state string) {

	if state == "trying" {
		if self.fsm.Is("synced") {
			self.fsm.Event("launcher_trying")
		}

	} else if state == "up" {
		if self.fsm.Is("syncing") {
			self.fsm.Event("launcher_up")
		}
	}
}
