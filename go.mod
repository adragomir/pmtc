module github.com/adragomir/linuxcncgo

go 1.16

require (
	gioui.org v0.0.0-20210819082505-f47508729638 // indirect
	gioui.org/example v0.0.0-20210804164344-256666b4c0fa // indirect
	gioui.org/x v0.0.0-20210805142029-e0acbf68bc23 // indirect
	github.com/andrewtj/dnssd v0.0.0-20161222030342-242ed8d297c8 // indirect
	github.com/go-gl/gl v0.0.0-20210315015930-ae072cafe09d
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20210311203641-62640a716d48
	github.com/go-gl/mathgl v1.0.0
	github.com/grandcat/zeroconf v1.0.0 // indirect
	github.com/inkyblackness/imgui-go/v4 v4.3.0
	github.com/jlaffaye/ftp v0.0.0-20210307004419-5d4190119067
	github.com/looplab/fsm v0.2.0 // indirect
	github.com/machinekit/machinetalk_go v0.0.0-00010101000000-000000000000
	github.com/machinekit/machinetalk_protobuf_go v0.0.0-00010101000000-000000000000
	github.com/miekg/dns v1.1.43 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/pebbe/zmq4 v1.2.7 // indirect
	github.com/tmm1/dnssd v0.0.0-20180920054121-2469354f9317
	github.com/trivigy/event v1.1.0 // indirect
	golang.org/x/exp v0.0.0-20210722180016-6781d3edade3 // indirect
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d // indirect
	golang.org/x/net v0.0.0-20210726213435-c6fcb2dbf985 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/protobuf v1.27.1
)

replace github.com/machinekit/machinetalk_go => ./localvendor/github.com/machinekit/machinetalk_go/

replace github.com/machinekit/machinetalk_protobuf_go => ./localvendor/github.com/machinekit/machinetalk_protobuf_go/

replace github.com/looplab/fsm => ./localvendor/github.com/looplab/fsm/

replace github.com/go-gl/mathgl => ./localvendor/github.com/go-gl/mathgl/
