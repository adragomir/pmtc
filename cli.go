package main

import (
	"fmt"
	"os"

	"github.com/adragomir/linuxcncgo/machine"
)

type cli struct {
	done     chan bool
	services *machine.Services
}

func NewCli(s *machine.Services) *cli {
	return &cli{
		done:     make(chan bool),
		services: s,
	}
}

func (c *cli) Start() {
	cmd := os.Args[1]
	var neededState string

	switch cmd {
	case "list":
		neededState = "launcher"
	case "start":
		neededState = "launcher"
	case "stop":
		neededState = "launcher"
	}

	c.services.AddCb("launcherUpdate", func(l *machine.Launcher) {
		if neededState == "launcher" || neededState == "machine" {
			switch cmd {
			case "list":
				for uuid, l := range c.services.Launchers {
					fmt.Printf("Launcher %s: %t\n", uuid, l.Running)
				}
				c.done <- true
			}

		}
	})
	c.services.AddCb("machineReady", func(m *machine.Machine) {
	})
}

func (c *cli) Wait() {
	<-c.done
}
