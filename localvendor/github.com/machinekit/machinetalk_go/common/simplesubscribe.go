package common

import (
	"fmt"
	"log"

	"github.com/looplab/fsm"
	zmq "github.com/pebbe/zmq4"

	"sync"

	uuid "github.com/nu7hatch/gouuid"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"

	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type SimpleSubscribe struct {
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

	OnSocketMsgReceived []func(*pb.Container, ...interface{})
	OnStateChanged      []func(string)
	fsm                 *fsm.FSM
}

func NewSimpleSubscribe(Debuglevel int, Debugname string) *SimpleSubscribe {
	tmp := &SimpleSubscribe{}
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

	// callbacks
	tmp.OnSocketMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnStateChanged = make([]func(string), 0)

	// fsm
	tmp.fsm = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "start", Src: []string{"down"}, Dst: "up"},
			{Name: "any_msg_received", Src: []string{"up"}, Dst: "up"},
			{Name: "stop", Src: []string{"up"}, Dst: "down"},
		},
		fsm.Callbacks{
			"down":                   func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_start":            func(e *fsm.Event) { tmp.OnFsm_start(e) },
			"up":                     func(e *fsm.Event) { tmp.OnFsm_up(e) },
			"after_any_msg_received": func(e *fsm.Event) { tmp.OnFsm_any_msg_received(e) },
			"after_stop":             func(e *fsm.Event) { tmp.OnFsm_stop(e) },
		},
	)
	return tmp
}

func (self *SimpleSubscribe) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *SimpleSubscribe) OnFsm_start(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event START", self.Debugname)
	}
	self.StartSocket()
}

func (self *SimpleSubscribe) OnFsm_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("up")
	}
}

func (self *SimpleSubscribe) OnFsm_any_msg_received(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event ANY MSG RECEIVED", self.Debugname)
	}
}

func (self *SimpleSubscribe) OnFsm_stop(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event STOP", self.Debugname)
	}
	self.StopSocket()
}

func (self *SimpleSubscribe) ErrorString() string {
	return self.errorString
}

func (self *SimpleSubscribe) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

// trigger
func (self *SimpleSubscribe) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("start")
	}
}

// trigger
func (self *SimpleSubscribe) Stop() {
	if self.fsm.Is("up") {
		self.fsm.Event("stop")
	}
}

func (self *SimpleSubscribe) AddSocketTopic(name string) {
	self.SocketTopics[name] = true
}

func (self *SimpleSubscribe) RemoveSocketTopic(name string) {
	delete(self.SocketTopics, name)
}

func (self *SimpleSubscribe) ClearSocketTopics() {
	self.SocketTopics = make(map[string]bool)
}

func (self *SimpleSubscribe) socketWorker(context *zmq.Context, uri string) {
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

func (self *SimpleSubscribe) StartSocket() {
	go self.socketWorker(self.context, self.SocketUri)
}

func (self *SimpleSubscribe) StopSocket() {
	self.shutdown.Send(" ", 0) // trigger socket thread shutdown
}

// process all messages received on socket
func (self *SimpleSubscribe) SocketMsgReceived(socket *zmq.Socket) {
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

	for _, cb := range self.OnSocketMsgReceived {
		cb(rx, string(identity))
	}
}
