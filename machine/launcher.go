package machine

import (
	"github.com/machinekit/machinetalk_go/application"
	pb "github.com/machinekit/machinetalk_protobuf_go"
)

type Launcher struct {
	launcherDsn    string
	launcherCmdDsn string
	uuid           string
	service        *application.LauncherBase
	index          int32

	Name        string
	Complete    bool
	Running     bool
	Terminating bool
}

func (l *Launcher) Start() {
	msg := &pb.Container{
		Index: &l.index,
	}
	l.service.SendLauncherStart(msg)
}
