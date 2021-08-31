package application

import (
	"fmt"
	"log"
	"time"

	"github.com/looplab/fsm"
	zmq "github.com/pebbe/zmq4"

	"sync"

	uuid "github.com/nu7hatch/gouuid"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"

	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type ErrorSubscribe struct {
	Debuglevel           int
	Debugname            string
	errorString          string
	OnErrorStringChanged []func(string)
	txLock               sync.Mutex
	context              *zmq.Context
	shutdown             *zmq.Socket
	shutdownUri          string
	// Socket socket
	SocketUri    string
	SocketTopics map[string]bool
	// more efficient to reuse protobuf messages
	socketRx *pb.Container

	// Heartbeat timer
	HeartbeatLock          sync.Mutex
	HeartbeatInterval      int
	HeartbeatTimer         *time.Timer
	HeartbeatActive        bool
	HeartbeatLiveness      int
	HeartbeatResetLiveness int
	OnSocketMsgReceived    []func(*pb.Container, ...interface{})
	OnStateChanged         []func(string)
	fsm                    *fsm.FSM
}

func NewErrorSubscribe(Debuglevel int, Debugname string) *ErrorSubscribe {
	tmp := &ErrorSubscribe{}
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
	//tmp._thread = None  // socket worker tread
	tmp.txLock = sync.Mutex{} // lock for outgoing messages

	// Socket socket
	tmp.SocketUri = ""
	tmp.SocketTopics = make(map[string]bool)
	// more efficient to reuse protobuf messages
	// XXXXX socket socket socket
	tmp.socketRx = &pb.Container{}
	// Heartbeat timer
	tmp.HeartbeatLock = sync.Mutex{}
	tmp.HeartbeatInterval = 2500
	tmp.HeartbeatTimer = nil
	tmp.HeartbeatActive = false
	tmp.HeartbeatLiveness = 0
	tmp.HeartbeatResetLiveness = 5

	// callbacks
	tmp.OnSocketMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnStateChanged = make([]func(string), 0)

	// fsm
	tmp.fsm = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "start", Src: []string{"down"}, Dst: "trying"},
			{Name: "ping_received", Src: []string{"trying"}, Dst: "up"},
			{Name: "stop", Src: []string{"up", "trying"}, Dst: "down"},
			{Name: "heartbeat_timeout", Src: []string{"up"}, Dst: "trying"},
			{Name: "heartbeat_tick", Src: []string{"up"}, Dst: "up"},
			{Name: "any_msg_received", Src: []string{"up"}, Dst: "up"},
		},
		fsm.Callbacks{
			"down":                    func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_start":             func(e *fsm.Event) { tmp.OnFsm_start(e) },
			"trying":                  func(e *fsm.Event) { tmp.OnFsm_trying(e) },
			"after_ping_received":     func(e *fsm.Event) { tmp.OnFsm_ping_received(e) },
			"after_stop":              func(e *fsm.Event) { tmp.OnFsm_stop(e) },
			"up":                      func(e *fsm.Event) { tmp.OnFsm_up(e) },
			"after_heartbeat_timeout": func(e *fsm.Event) { tmp.OnFsm_heartbeat_timeout(e) },
			"after_heartbeat_tick":    func(e *fsm.Event) { tmp.OnFsm_heartbeat_tick(e) },
			"after_any_msg_received":  func(e *fsm.Event) { tmp.OnFsm_any_msg_received(e) },
		},
	)
	return tmp
}

func (self *ErrorSubscribe) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *ErrorSubscribe) OnFsm_start(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event START", self.Debugname)
	}
	self.StartSocket()
}

func (self *ErrorSubscribe) OnFsm_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state TRYING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("trying")
	}
}

func (self *ErrorSubscribe) OnFsm_ping_received(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event PING RECEIVED", self.Debugname)
	}
	self.ResetHeartbeatLiveness()
	self.StartHeartbeatTimer()
}

func (self *ErrorSubscribe) OnFsm_stop(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event STOP", self.Debugname)
	}
	self.StopHeartbeatTimer()
	self.StopSocket()
}

func (self *ErrorSubscribe) OnFsm_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("up")
	}
}

func (self *ErrorSubscribe) OnFsm_heartbeat_timeout(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event HEARTBEAT TIMEOUT", self.Debugname)
	}
	self.StopHeartbeatTimer()
	self.StopSocket()
	self.StartSocket()
}

func (self *ErrorSubscribe) OnFsm_heartbeat_tick(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event HEARTBEAT TICK", self.Debugname)
	}
	self.ResetHeartbeatTimer()
}

func (self *ErrorSubscribe) OnFsm_any_msg_received(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event ANY MSG RECEIVED", self.Debugname)
	}
	self.ResetHeartbeatLiveness()
	self.ResetHeartbeatTimer()
}

func (self *ErrorSubscribe) ErrorString() string {
	return self.errorString
}

func (self *ErrorSubscribe) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

// trigger
func (self *ErrorSubscribe) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("start")
	}
}

// trigger
func (self *ErrorSubscribe) Stop() {
	if self.fsm.Is("trying") {
		self.fsm.Event("stop")
	} else if self.fsm.Is("up") {
		self.fsm.Event("stop")
	}
}

func (self *ErrorSubscribe) AddSocketTopic(name string) {
	self.SocketTopics[name] = true
}

func (self *ErrorSubscribe) RemoveSocketTopic(name string) {
	delete(self.SocketTopics, name)
}

func (self *ErrorSubscribe) ClearSocketTopics() {
	self.SocketTopics = make(map[string]bool)
}

func (self *ErrorSubscribe) socketWorker(context *zmq.Context, uri string) {
	poll := zmq.NewPoller()
	socket, _ := self.context.NewSocket(zmq.SUB)
	socket.SetLinger(0)
	socket.Connect(uri)
	poll.Add(socket, zmq.POLLIN)
	// subscribe is always connected to socket creation
	for topic, _ := range self.SocketTopics {
		socket.SetSubscribe(topic)
	}

	shutdown, _ := self.context.NewSocket(zmq.PULL)
	shutdown.Connect(self.shutdownUri)
	poll.Add(shutdown, zmq.POLLIN)

	for {
		ss, _ := poll.Poll(-1)
		for _, psocket := range ss {
			switch s := psocket.Socket; s {
			case shutdown:
				shutdown.Recv(0)
				return // shutdown signal
			case socket:
				self.SocketMsgReceived(socket)
			}
		}
	}
}

func (self *ErrorSubscribe) StartSocket() {
	go self.socketWorker(self.context, self.SocketUri)
}

func (self *ErrorSubscribe) StopSocket() {
	self.shutdown.Send(" ", 0) // trigger socket thread shutdown
}

func (self *ErrorSubscribe) HeartbeatTimerTick() {
	self.HeartbeatLock.Lock()
	self.HeartbeatTimer = nil // timer is dead on tick
	self.HeartbeatLock.Unlock()

	if self.Debuglevel > 0 {
		log.Printf("[%s] heartbeat timer tick", self.Debugname)
	}

	self.HeartbeatLiveness -= 1
	if self.HeartbeatLiveness == 0 {
		if self.fsm.Is("up") {
			self.fsm.Event("heartbeat_timeout")
		}
		return
	}

	if self.fsm.Is("up") {
		self.fsm.Event("heartbeat_tick")
	}
}

func (self *ErrorSubscribe) ResetHeartbeatLiveness() {
	self.HeartbeatLiveness = self.HeartbeatResetLiveness
}

func (self *ErrorSubscribe) ResetHeartbeatTimer() {
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

func (self *ErrorSubscribe) StartHeartbeatTimer() {
	self.HeartbeatActive = true
	self.ResetHeartbeatTimer()
}

func (self *ErrorSubscribe) StopHeartbeatTimer() {
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
func (self *ErrorSubscribe) SocketMsgReceived(socket *zmq.Socket) {
	tmp, _ := socket.RecvMessageBytes(0) // identity is topic
	// ADR: needed ?
	identity := tmp[0]
	msg := tmp[1]
	if err := proto.Unmarshal(msg, self.socketRx); err != nil {
		log.Printf("Protobuf Decode Error:", err)
		return
	}

	if self.Debuglevel > 0 {
		log.Printf("[%s] received message", self.Debugname)
		if self.Debuglevel > 1 {
			log.Printf("[%s] %s", prototext.Format(self.socketRx))
		}
	}
	rx := self.socketRx
	// INCOMING * 0 1 0

	// react to any incoming message
	if self.fsm.Is("up") {
		self.fsm.Event("any_msg_received")
	}
	// INCOMING ping 1 1 0

	// react to ping message
	if *rx.Type == pb.ContainerType_MT_PING {
		if rx.Pparams != nil {
			interval := int(*rx.Pparams.KeepaliveTimer)
			self.HeartbeatInterval = interval
		}
		if self.fsm.Is("trying") {
			self.fsm.Event("ping_received")
		}
		return // ping is uninteresting
	} //AAAAAAAA

	for _, cb := range self.OnSocketMsgReceived {
		cb(rx, string(identity))
	}
}
