package main

import (
	"os"

	"github.com/adragomir/linuxcncgo/machine"
	"github.com/adragomir/linuxcncgo/ui"
)

func main() {
	if len(os.Args) > 1 {
		mainCli()
	} else {
		mainUi()
	}
}

func mainCli() {
	services := machine.NewServices()
	c := NewCli(services)
	c.Start()
	services.Start()
	c.Wait()
}

func mainUi() {
	services := machine.NewServices()
	ui.StartUi(services)
}
