package common

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/looplab/fsm"
	zmq "github.com/pebbe/zmq4"

	"sync"

	uuid "github.com/nu7hatch/gouuid"
	"github.com/trivigy/event"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"

	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type RpcService struct {
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

	OnSocketMsgReceived []func(*pb.Container, ...interface{})
	OnStateChanged      []func(string)
	fsm                 *fsm.FSM
}

func NewRpcService(Debuglevel int, Debugname string) *RpcService {
	tmp := &RpcService{}
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
	// more efficient to reuse protobuf messages
	// XXXXX socket socket socket
	tmp.socketRx = &pb.Container{}
	tmp.socketTx = &pb.Container{}

	// callbacks
	tmp.OnSocketMsgReceived = make([]func(*pb.Container, ...interface{}), 0)
	tmp.OnStateChanged = make([]func(string), 0)

	// fsm
	tmp.fsm = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "start", Src: []string{"down"}, Dst: "up"},
			{Name: "ping_received", Src: []string{"up"}, Dst: "up"},
			{Name: "stop", Src: []string{"up"}, Dst: "down"},
		},
		fsm.Callbacks{
			"down":                func(e *fsm.Event) { tmp.OnFsm_down(e) },
			"after_start":         func(e *fsm.Event) { tmp.OnFsm_start(e) },
			"up":                  func(e *fsm.Event) { tmp.OnFsm_up(e) },
			"after_ping_received": func(e *fsm.Event) { tmp.OnFsm_ping_received(e) },
			"after_stop":          func(e *fsm.Event) { tmp.OnFsm_stop(e) },
		},
	)
	return tmp
}

func (self *RpcService) OnFsm_down(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state DOWN", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("down")
	}
}

func (self *RpcService) OnFsm_start(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event START", self.Debugname)
	}
	self.StartSocket()
}

func (self *RpcService) OnFsm_up(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: state UP", self.Debugname)
	}
	for _, cb := range self.OnStateChanged {
		cb("up")
	}
}

func (self *RpcService) OnFsm_ping_received(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event PING RECEIVED", self.Debugname)
	}
	self.SendPingAcknowledge()
}

func (self *RpcService) OnFsm_stop(e *fsm.Event) {
	if self.Debuglevel > 0 {
		log.Printf("[%s]: event STOP", self.Debugname)
	}
	self.StopSocket()
}

func (self *RpcService) ErrorString() string {
	return self.errorString
}

func (self *RpcService) SetErrorString(es string) {
	if self.errorString != "" {
		return
	}
	self.errorString = es
	for _, cb := range self.OnErrorStringChanged {
		cb(es)
	}
}

func (self *RpcService) GetSocketPort() int {
	self.SocketPortEvent.Wait(nil)
	return self.SocketPort
}

func (self *RpcService) GetSocketDsn() string {
	self.SocketDsnEvent.Wait(nil)
	return self.SocketDsn
}

// trigger
func (self *RpcService) Start() {
	if self.fsm.Is("down") {
		self.fsm.Event("start")
	}
}

// trigger
func (self *RpcService) Stop() {
	if self.fsm.Is("up") {
		self.fsm.Event("stop")
	}
}

func (self *RpcService) socketWorker(context *zmq.Context, uri string) {
	poll := zmq.NewPoller()
	socket, _ := self.context.NewSocket(zmq.ROUTER)
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

func (self *RpcService) StartSocket() {
	go self.socketWorker(self.context, self.SocketUri)
}

func (self *RpcService) StopSocket() {
	self.shutdown.Send(" ", 0)   // trigger socket thread shutdown
	self.SocketPortEvent.Clear() // clear sync for port
}

// process all messages received on socket
func (self *RpcService) SocketMsgReceived(socket *zmq.Socket) {
	frames, _ := socket.RecvMessageBytes(0)
	identity := frames[0]
	msg := frames[1]
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
	// INCOMING ping 1 1 0

	// react to ping message
	if *rx.Type == pb.ContainerType_MT_PING {
		if self.fsm.Is("up") {
			self.fsm.Event("ping_received")
		}
		return // ping is uninteresting
		// INCOMING * 0 0 0
	} //BBBBBBBB *

	for _, cb := range self.OnSocketMsgReceived {
		cb(rx, string(identity))
	}
}

func (self *RpcService) SendSocketMsg(identity string, msg_type pb.ContainerType, tx *pb.Container) {
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
func (self *RpcService) SendPingAcknowledge() {
	tx := self.socketTx
	ids := self.SocketTopics
	for receiver, _ := range ids {
		self.SendSocketMsg(receiver, pb.ContainerType_MT_PING_ACKNOWLEDGE, tx)
	}
}
