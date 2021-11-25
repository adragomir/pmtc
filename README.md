# MachineKit UI

An UI for a network connected [MachineKit](https://www.machinekit.io/) controlled CNC.

Features:

* Jogging
* Homing
* Gcode execution (MDI)
* Setting LCS coordinates
* Loading and executing Gcode files, with 3d preview

# Usage

* Tested only on Mac OS X
  * Probably only works on Mac OS X, due to usage of [tmm1/dnssd](https://github.com/tmm1/dnssd) - because this was the only way I found to get reliable zeroconf discovery in go, on a mac

# Kudos

* Gcode parser / simulator heavily inspired (read "copied") from (webgcode](https://github.com/nraynaud/webgcode)
* Modified [looplab/fsm](https://github.com/looplab/fsm) to work w/ concurrent access
* Used [machinetalk-gsl](https://github.com/machinekoder/machinetalk-gsl) to build the go bindings, but gave up at a certain point and started modifying by hand. 

