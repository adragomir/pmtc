package common

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

type RpcClient struct {
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
	SocketUri    string
	SocketTopics map[string]bool
	// more efficient to reuse protobuf messages
	socketRx *pb.Container
	socketTx *pb.Container

	// Heartbeat timer
	HeartbeatLock          sync.Mutex
	HeartbeatInterval      int
	HeartbeatTimer         *time.Timer
	HeartbeatTimerChan     chan struct{}
	HeartbeatActive        bool
	HeartbeatLiveness      int
	HeartbeatResetLiveness int
	OnSocketMsgReceived    []func(*pb.Container, ...interface{})
	OnStateChanged         []func(string)
	fsm                    *fsm.FSM
}

func NewRpcClient(Debuglevel int, Debugname string) *RpcClient {
	tmp := &RpcClient{}
	tmp.Debuglevel = Debuglevel
	tmp.Debugname = Debugname
	tmp.errorString = ""
	tmp.OnErrorStringChanged = make([]func(string), 1)
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
	// more efficient to reuse protobuf messages
	// XXXXX socket socket socket
	tmp.socketRx = &pb.Container{}
	tmp.socketTx = &pb.Container{}
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
			{Name: "any_msg_received", Src: []string{"trying", "up"}, Dst: "up"},
			{Name: "heartbeat_timeout", Src: []string{"up", "trying"}, Dst: "trying"},
			{Name: "heartbeat_tick", Src: []string{"trying"}, Dst: "trying"},
			{Name: "any_msg_sent", Src: []string{"trying"}, Dst: "trying"},
			{Name: "stop", Src: []string{"up", "trying"}, Dst: "down"},
			{Name: "heartbeat_tick", Src: []string{"up"}, Dst: "up"},
			{Name: "any_msg_sent", Src: []string{"up"}, Dst: "up"},
		},
		fsm.Callbacks{
			"down":                    func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_start":             func(e *fsm.Event) { tmp.OnFsm_start(e) },
			"trying":                  func(e *fsm.Event) { tmp.OnFsm_trying(e) },
			"after_any_msg_received":  func(e *fsm.Event) { tmp.OnFsm_any_msg_received(e) },
			"after_heartbeat_timeout": func(e *fsm.Event) { tmp.OnFsm_heartbeat_timeout(e) },
			"after_heartbeat_tick":    func(e *fsm.Event) { tmp.OnFsm_heartbeat_tick(e) },
			"after_any_msg_sent":      func(e *fsm.Event) { tmp.OnFsm_any_msg_sent(e) },
			"after_stop":              func(e *fsm.Event) { tmp.OnFsm_stop(e) },
			"up":                      func(e *fsm.Event) { tmp.OnFsm_up(e) },
		},
	)
	return tmp
}

func (self *RpcClient) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *RpcClient) OnFsm_start(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event START", self.Debugname)
	}
	self.StartSocket()
	self.ResetHeartbeatLiveness()
	self.SendPing()
	self.StartHeartbeatTimer()
}

func (self *RpcClient) OnFsm_trying(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state TRYING", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("trying")
	}
}

func (self *RpcClient) OnFsm_any_msg_received(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event ANY MSG RECEIVED", self.Debugname)
	}
	self.ResetHeartbeatLiveness()
	self.ResetHeartbeatTimer()
}

func (self *RpcClient) OnFsm_heartbeat_timeout(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event HEARTBEAT TIMEOUT", self.Debugname)
	}
	self.StopSocket()
	self.StartSocket()
	self.ResetHeartbeatLiveness()
	self.SendPing()
}

func (self *RpcClient) OnFsm_heartbeat_tick(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event HEARTBEAT TICK", self.Debugname)
	}
	self.SendPing()
}

func (self *RpcClient) OnFsm_any_msg_sent(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event ANY MSG SENT", self.Debugname)
	}
	self.ResetHeartbeatTimer()
}

func (self *RpcClient) OnFsm_stop(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event STOP", self.Debugname)
	}
	self.StopHeartbeatTimer()
	self.StopSocket()
}

func (self *RpcClient) OnFsm_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("up")
	}
}

func (self *RpcClient) ErrorString() string {
	return self.errorString
}

func (self *RpcClient) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

// trigger
func (self *RpcClient) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("start")
	}
}

// trigger
func (self *RpcClient) Stop() {
	if self.fsm.Is("trying") {
		self.fsm.Event("stop")
	} else if self.fsm.Is("up") {
		self.fsm.Event("stop")
	}
}

func (self *RpcClient) socketWorker(context *zmq.Context, uri string) {
	poll := zmq.NewPoller()
	socket, _ := self.context.NewSocket(zmq.DEALER)
	socket.SetLinger(0)
	socket.Connect(uri)
	poll.Add(socket, zmq.POLLIN)

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
				tmp, _ := pipe.RecvBytes(0)
				socket.SendBytes(tmp, zmq.DONTWAIT)
			case socket:
				self.SocketMsgReceived(socket)
			}
		}
	}
}

func (self *RpcClient) StartSocket() {
	go self.socketWorker(self.context, self.SocketUri)
}

func (self *RpcClient) StopSocket() {
	self.shutdown.Send(" ", 0) // trigger socket thread shutdown
}

func (self *RpcClient) HeartbeatTimerTick() {
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
		} else if self.fsm.Is("trying") {
			self.fsm.Event("heartbeat_timeout")
		}
		return
	}

	if self.fsm.Is("up") {
		self.fsm.Event("heartbeat_tick")
	} else if self.fsm.Is("trying") {
		self.fsm.Event("heartbeat_tick")
	}
}

func (self *RpcClient) ResetHeartbeatLiveness() {
	self.HeartbeatLiveness = self.HeartbeatResetLiveness
}

func (self *RpcClient) ResetHeartbeatTimer() {
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

func (self *RpcClient) StartHeartbeatTimer() {
	self.HeartbeatActive = true
	self.ResetHeartbeatTimer()
}

func (self *RpcClient) StopHeartbeatTimer() {
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
func (self *RpcClient) SocketMsgReceived(socket *zmq.Socket) {
	msg, _ := socket.RecvBytes(0)
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
	// INCOMING * 0 2 0

	// react to any incoming message
	if self.fsm.Is("trying") {
		self.fsm.Event("any_msg_received")
	} else if self.fsm.Is("up") {
		self.fsm.Event("any_msg_received")
	}
	// INCOMING ping acknowledge 1 0 0

	// react to ping acknowledge message
	if *rx.Type == pb.ContainerType_MT_PING_ACKNOWLEDGE {
		return // ping acknowledge is uninteresting
	} //AAAAAAAA

	for _, cb := range self.OnSocketMsgReceived {
		cb(rx)
	}
}

func (self *RpcClient) SendSocketMsg(msg_type pb.ContainerType, tx *pb.Container) {
	self.txLock.Lock()
	defer self.txLock.Unlock()
	tx.Type = msg_type.Enum()
	if self.Debuglevel > 0 {
		log.Printf("[%s] sending message: %s", self.Debugname, msg_type)
		if self.Debuglevel > 1 {
			log.Printf("%s", prototext.Format(tx))
		}
	}
	out, _ := proto.Marshal(tx)

	self.pipe.SendBytes(out, 0)
	tx = &pb.Container{}

	if self.fsm.Is("up") {
		self.fsm.Event("any_msg_sent")
	} else if self.fsm.Is("trying") {
		self.fsm.Event("any_msg_sent")
	}
}
func (self *RpcClient) SendPing() {
	tx := self.socketTx
	self.SendSocketMsg(pb.ContainerType_MT_PING, tx)
}
