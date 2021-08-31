package common

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/looplab/fsm"
	zmq "github.com/pebbe/zmq4"

	"sync"

	uuid "github.com/nu7hatch/gouuid"
	"github.com/trivigy/event"
	"google.golang.org/protobuf/proto"

	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type Publish struct {
	Debuglevel           int
	Debugname            string
	errorString          string
	OnErrorStringChanged []func(string)
	txLock               sync.Mutex
	context              *zmq.Context
	shutdown             *zmq.Socket
	shutdownUri          string
	pipe                 *zmq.Socket
	pipeUri              string
	// Socket socket
	SocketUri       string
	SocketPort      int
	SocketPortEvent *event.Event
	SocketDsn       string
	SocketDsnEvent  *event.Event
	SocketTopics    map[string]bool
	// more efficient to reuse protobuf messages
	socketRx *pb.Container
	socketTx *pb.Container

	// Heartbeat timer
	HeartbeatLock       sync.Mutex
	HeartbeatInterval   int
	HeartbeatTimer      *time.Timer
	HeartbeatActive     bool
	OnSocketMsgReceived []func(*pb.Container, ...interface{})
	OnStateChanged      []func(string)
	fsm                 *fsm.FSM
}

func NewPublish(Debuglevel int, Debugname string) *Publish {
	tmp := &Publish{}
	tmp.Debuglevel = Debuglevel
	tmp.Debugname = Debugname
	tmp.errorString = ""
	tmp.OnErrorStringChanged = make([]func(string), 0)
	// ZeroMQ
	context, _ := zmq.NewContext()
	tmp.context = context
	// pipe to signalize a shutdown
	tmp.shutdown, _ = context.NewSocket(zmq.PUSH)
	u4s, _ := uuid.NewV4()
	tmp.shutdownUri = fmt.Sprintf("inproc://shutdown-%s", u4s.String())
	tmp.shutdown.Bind(tmp.shutdownUri)
	// pipe for outgoing messages
	tmp.pipe, _ = context.NewSocket(zmq.PUSH)
	u4p, _ := uuid.NewV4()
	tmp.pipeUri = fmt.Sprintf("inproc://pipe-%s", u4p.String())
	tmp.pipe.Bind(tmp.pipeUri)
	//tmp._thread = None  // socket worker tread
	tmp.txLock = sync.Mutex{} // lock for outgoing messages

	// Socket socket
	tmp.SocketUri = ""
	tmp.SocketPort = 0
	tmp.SocketPortEvent = event.New() // sync event for port
	tmp.SocketDsn = ""
	tmp.SocketDsnEvent = event.New() // sync event for dsn
	tmp.SocketTopics = make(map[string]bool)
	// more efficient to reuse protobuf messages
	// XXXXX socket socket socket
	tmp.socketRx = &pb.Container{}
	tmp.socketTx = &pb.Container{}
	// Heartbeat timer
	tmp.HeartbeatLock = sync.Mutex{}
	tmp.HeartbeatInterval = 2500
	tmp.HeartbeatTimer = nil
	tmp.HeartbeatActive = false

	// callbacks
	tmp.OnSocketMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnStateChanged = make([]func(string), 0)

	// fsm
	tmp.fsm = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "start", Src: []string{"down"}, Dst: "up"},
			{Name: "stop", Src: []string{"up"}, Dst: "down"},
			{Name: "heartbeat_tick", Src: []string{"up"}, Dst: "up"},
		},
		fsm.Callbacks{
			"down":                 func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_start":          func(e *fsm.Event) { tmp.OnFsm_start(e) },
			"up":                   func(e *fsm.Event) { tmp.OnFsm_up(e) },
			"after_stop":           func(e *fsm.Event) { tmp.OnFsm_stop(e) },
			"after_heartbeat_tick": func(e *fsm.Event) { tmp.OnFsm_heartbeat_tick(e) },
		},
	)
	return tmp
}

func (self *Publish) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *Publish) OnFsm_start(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event START", self.Debugname)
	}
	self.StartSocket()
	self.StartHeartbeatTimer()
}

func (self *Publish) OnFsm_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("up")
	}
}

func (self *Publish) OnFsm_stop(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event STOP", self.Debugname)
	}
	self.StopHeartbeatTimer()
	self.StopSocket()
}

func (self *Publish) OnFsm_heartbeat_tick(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event HEARTBEAT TICK", self.Debugname)
	}
	self.SendPing()
	self.ResetHeartbeatTimer()
}

func (self *Publish) ErrorString() string {
	return self.errorString
}

func (self *Publish) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

func (self *Publish) GetSocketPort() int {
	self.SocketPortEvent.Wait(nil)
	return self.SocketPort
}

func (self *Publish) GetSocketDsn() string {
	self.SocketDsnEvent.Wait(nil)
	return self.SocketDsn
}

// trigger
func (self *Publish) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("start")
	}
}

// trigger
func (self *Publish) Stop() {
	if self.fsm.Is("up") {
		self.fsm.Event("stop")
	}
}

func (self *Publish) AddSocketTopic(name string) {
	self.SocketTopics[name] = true
}

func (self *Publish) RemoveSocketTopic(name string) {
	delete(self.SocketTopics, name)
}

func (self *Publish) ClearSocketTopics() {
	self.SocketTopics = make(map[string]bool)
}

func (self *Publish) socketWorker(context *zmq.Context, uri string) {
	poll := zmq.NewPoller()
	socket, _ := self.context.NewSocket(zmq.XPUB)
	socket.SetLinger(0)
	if strings.Contains(uri, "ipc://") || strings.Contains(uri, "inproc://") {
		socket.Bind(uri)
	} else {
		socket.Bind(fmt.Sprintf("%s:*", uri))
		tmp, _ := socket.GetLastEndpoint()
		tmp2 := strings.Split(tmp, ":")[1]
		self.SocketPort, _ = strconv.Atoi(tmp2)
		self.SocketPortEvent.Set() // set sync for port
	}
	le, _ := socket.GetLastEndpoint()
	self.SocketDsn = le
	self.SocketDsnEvent.Set()
	poll.Add(socket, zmq.POLLIN)
	socket.SetXpubVerbose(1) // enable verbose subscription messages

	shutdown, _ := self.context.NewSocket(zmq.PULL)
	shutdown.Connect(self.shutdownUri)
	poll.Add(shutdown, zmq.POLLIN)
	pipe, _ := self.context.NewSocket(zmq.PULL)
	pipe.Connect(self.pipeUri)
	poll.Add(pipe, zmq.POLLIN)

	for {
		ss, _ := poll.Poll(-1)
		for _, psocket := range ss {
			switch s := psocket.Socket; s {
			case shutdown:
				shutdown.Recv(0)
				return // shutdown signal
			case pipe:
				tmpmsg, _ := pipe.RecvMessageBytes(0)
				socket.SendMessageDontwait(tmpmsg)
			case socket:
				self.SocketMsgReceived(socket)
			}
		}
	}
}

func (self *Publish) StartSocket() {
	go self.socketWorker(self.context, self.SocketUri)
}

func (self *Publish) StopSocket() {
	self.shutdown.Send(" ", 0)   // trigger socket thread shutdown
	self.SocketPortEvent.Clear() // clear sync for port
}

func (self *Publish) HeartbeatTimerTick() {
	self.HeartbeatLock.Lock()
	self.HeartbeatTimer = nil // timer is dead on tick
	self.HeartbeatLock.Unlock()

	if self.Debuglevel > 0 {
		log.Printf("[%s] heartbeat timer tick", self.Debugname)
	}

	if self.fsm.Is("up") {
		self.fsm.Event("heartbeat_tick")
	}
}

func (self *Publish) ResetHeartbeatTimer() {
	if !self.HeartbeatActive {
		return
	}

	self.HeartbeatLock.Lock()
	defer self.HeartbeatLock.Unlock()
	if self.HeartbeatTimer != nil {
		if !self.HeartbeatTimer.Stop() {
			<-self.HeartbeatTimer.C
		}
		self.HeartbeatTimer = nil
	}

	if self.HeartbeatInterval > 0 {
		self.HeartbeatTimer = time.AfterFunc(time.Duration(self.HeartbeatInterval/1000.0)*time.Second, func() {
			self.HeartbeatTimerTick()
		})
	}
	if self.Debuglevel > 0 {
		log.Printf("[%s] heartbeat timer reset", self.Debugname)
	}
}

func (self *Publish) StartHeartbeatTimer() {
	self.HeartbeatActive = true
	self.ResetHeartbeatTimer()
}

func (self *Publish) StopHeartbeatTimer() {
	self.HeartbeatActive = false
	self.HeartbeatLock.Lock()
	if self.HeartbeatTimer != nil {
		if !self.HeartbeatTimer.Stop() {
			<-self.HeartbeatTimer.C
		}
	}
	self.HeartbeatLock.Unlock()
}

// process all messages received on socket
func (self *Publish) SocketMsgReceived(socket *zmq.Socket) {
	msg, _ := socket.RecvBytes(0)
	var rx = &pb.Container{}
	if err := proto.Unmarshal(msg, rx); err != nil {
		log.Printf("Protobuf Decode Error: %+v", err)
		return
	}
	// INCOMING * 0 0 0

	for _, cb := range self.OnSocketMsgReceived {
		cb(rx) // FIXME
	}
}

func (self *Publish) SendSocketMsg(identity string, msg_type pb.ContainerType, tx *pb.Container) {
	self.txLock.Lock()
	defer self.txLock.Unlock()
	tx.Type = msg_type.Enum()
	if self.Debuglevel > 0 {
		log.Printf("[%s] sending message: %s", self.Debugname, msg_type)
		if self.Debuglevel > 1 {
			log.Printf("%+d", tx)
		}
	}
	out, _ := proto.Marshal(tx)

	self.pipe.SendMessage([][]byte{[]byte(identity), out})
	tx = &pb.Container{}
}
func (self *Publish) SendPing() {
	tx := self.socketTx
	ids := self.SocketTopics
	for receiver, _ := range ids {
		self.SendSocketMsg(receiver, pb.ContainerType_MT_PING, tx)
	}
}
func (self *Publish) SendFullUpdate(identity string, tx *pb.Container) {
	ids := map[string]bool{
		identity: true,
	}
	pparams := tx.Pparams
	tmp := int32(self.HeartbeatInterval)
	pparams.KeepaliveTimer = &tmp
	for receiver, _ := range ids {
		self.SendSocketMsg(receiver, pb.ContainerType_MT_FULL_UPDATE, tx)
	}
}
func (self *Publish) SendIncrementalUpdate(identity string, tx *pb.Container) {
	ids := map[string]bool{
		identity: true,
	}
	for receiver, _ := range ids {
		self.SendSocketMsg(receiver, pb.ContainerType_MT_INCREMENTAL_UPDATE, tx)
	}
}
