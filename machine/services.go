package machine

import (
	"github.com/adragomir/linuxcncgo/network"
	"github.com/adragomir/linuxcncgo/util"
	"github.com/machinekit/machinetalk_go/application"

	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type Services struct {
	*util.Callbacker
	resolver     *network.Resolver
	tempResolved map[string]map[string]string

	Launchers     map[string]*Launcher
	Machines      map[string]*Machine
	ActiveMachine *Machine

	callbacks map[string][]interface{}
}

func NewServices() *Services {
	tmp := &Services{
		Callbacker:   util.NewCallbacker(),
		resolver:     network.NewResolver("_machinekit._tcp"),
		tempResolved: make(map[string]map[string]string),

		Launchers: make(map[string]*Launcher),
		Machines:  make(map[string]*Machine),

		callbacks: make(map[string][]interface{}, 0),
	}

	tmp.resolver.OnItemAdded = append(tmp.resolver.OnItemAdded, func(name string, host string, port int, txt map[string]string) {
		service := txt["service"]
		uuid := txt["uuid"]
		if _, ok := tmp.tempResolved[uuid]; !ok {
			tmp.tempResolved[uuid] = make(map[string]string)
		}
		tmp.tempResolved[uuid][service] = txt["dsn"]
		tmp.TryToBuild()

	})
	return tmp
}

func (s *Services) Start() {
	s.resolver.Start()
}

func (s *Services) TryToBuild() {
	for uuid, _ := range s.tempResolved {
		ldsn, okl := s.tempResolved[uuid]["launcher"]
		lcmddsn, oklcmd := s.tempResolved[uuid]["launchercmd"]
		if okl && oklcmd {
			// got a complete launcher
			l := &Launcher{
				launcherDsn:    ldsn,
				launcherCmdDsn: lcmddsn,
				uuid:           uuid,
				service:        nil,
				index:          -1,

				Name:        "",
				Complete:    false,
				Running:     false,
				Terminating: false,
			}
			s.Launchers[uuid] = l

			l.Complete = true
			l.service = application.NewLauncherBase(0, "launcher")
			l.service.AddLauncherTopic("launcher")
			l.service.OnLaunchercmdMsgReceived = append(l.service.OnLaunchercmdMsgReceived, func(rx *pb.Container, rest ...interface{}) {
			})
			l.service.OnLauncherMsgReceived = append(l.service.OnLauncherMsgReceived, func(rx *pb.Container, rest ...interface{}) {
				if *rx.Type == pb.ContainerType_MT_LAUNCHER_FULL_UPDATE {
					for _, launch := range rx.Launcher {
						l.index = *launch.Index
						l.Name = *launch.Name
						l.Running = *launch.Running
						l.Terminating = *launch.Terminating
						s.RunCbs("launcherUpdate", l)
					}
				}
			})
			l.service.SetLauncherUri(l.launcherDsn)
			l.service.SetLaunchercmdUri(l.launcherCmdDsn)
			l.service.Start()

			delete(s.tempResolved[uuid], "launcher")
			delete(s.tempResolved[uuid], "launchercmd")
		}
		if _, ok := s.Launchers[uuid]; ok {
			// if we already have a launcher - which means we deleted the things
			if len(s.tempResolved[uuid]) == 11 {
				// we also have a complete machine
				m := &Machine{
					Callbacker: util.NewCallbacker(),
					uuid:       uuid,
					Complete:   false,
					Dsn: map[string]string{
						"command":       s.tempResolved[uuid]["command"],
						"error":         s.tempResolved[uuid]["error"],
						"file":          s.tempResolved[uuid]["file"],
						"halgroup":      s.tempResolved[uuid]["halgroup"],
						"halrcmd":       s.tempResolved[uuid]["halrcmd"],
						"halrcomp":      s.tempResolved[uuid]["halrcomp"],
						"log":           s.tempResolved[uuid]["log"],
						"config":        s.tempResolved[uuid]["config"],
						"preview":       s.tempResolved[uuid]["preview"],
						"previewstatus": s.tempResolved[uuid]["previewstatus"],
						"status":        s.tempResolved[uuid]["status"],
					},
					increments: make([]float64, 0),
				}
				s.Machines[uuid] = m
				m.TryBuilding(s)
				s.activateMachine(uuid)
			}
		}
	}
}

func (s *Services) activateMachine(uuid string) {
	s.Launchers[uuid].Running = true
	s.RunCbs("machineReady", s.Machines[uuid])
}
