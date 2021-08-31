package network

import (
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/tmm1/dnssd"
)

type BrowseCallback struct {
	op          *dnssd.BrowseOp
	add         bool
	iface       int
	name        string
	serviceType string
	domain      string
}

type ResolveCallback struct {
	op   *dnssd.ResolveOp
	name string
	host string
	port int
	txt  map[string]string
}

type StateItem struct {
	resolveOp *dnssd.ResolveOp
	exists    bool
	resolved  bool
	host      string
	port      int
	txt       map[string]string
}

type Resolver struct {
	serviceType string
	browseOp    *dnssd.BrowseOp
	browseChan  chan BrowseCallback
	resolveChan chan ResolveCallback
	stopChan    chan int
	state       sync.Map
	OnItemAdded []func(name string, host string, port int, txt map[string]string)
}

func NewResolver(serviceType string) *Resolver {
	tmp := &Resolver{
		serviceType: serviceType,
		browseChan:  make(chan BrowseCallback),
		resolveChan: make(chan ResolveCallback),
		stopChan:    make(chan int),
	}
	tmp.browseOp = dnssd.NewBrowseOp(tmp.serviceType, func(op *dnssd.BrowseOp, err error, add bool, interfaceIndex int, name string, serviceType string, domain string) {
		if err != nil {
			log.Printf("Browse operation failed: %s", err)
			return
		}
		// log.Printf("Browse operation %s %s service “%s” in %s on interface %d", add, serviceType, name, domain, interfaceIndex)
		if rawVal, ok := tmp.state.Load(name); ok {
			// the name already exists in the state
			val := rawVal.(*StateItem)
			if val.exists != add {
				if !add {
					// log.Printf("Cleaning up state for %s", name)
					val.resolveOp.Stop()
					tmp.state.Delete(name)
				} else {
					// log.Printf("Adding new state item %s", name)
					tmp.browseChan <- BrowseCallback{
						op:          op,
						add:         add,
						iface:       interfaceIndex,
						name:        name,
						serviceType: serviceType,
						domain:      domain,
					}
				}
			}
		} else {
			// the name does not exist, add it
			// log.Printf("Adding new name %s", name)
			tmp.browseChan <- BrowseCallback{
				op:          op,
				add:         add,
				iface:       interfaceIndex,
				name:        name,
				serviceType: serviceType,
				domain:      domain,
			}
		}
	})
	return tmp
}

func (r *Resolver) Start() {
	r.browseOp.Start()
	go func(bc chan BrowseCallback, rc chan ResolveCallback, sc chan int) {
		for {
			select {
			case bopr := <-bc:
				shouldAdd := false
				shouldDelete := false
				// we got a new browse op result
				if rawVal, ok := r.state.Load(bopr.name); ok {
					// the name already exists in the state
					val := rawVal.(*StateItem)
					if val.exists != bopr.add {
						if !bopr.add {
							shouldDelete = true
						} else {
							shouldAdd = true
						}
					}
				} else {
					if bopr.add {
						shouldAdd = true
					}
				}
				// log.Printf("Browse result for %s: %s %s", bopr.name, shouldAdd, shouldDelete)
				if shouldDelete {
					if rawVal, ok := r.state.Load(bopr.name); ok {
						val := rawVal.(*StateItem)
						if val.resolveOp != nil {
							val.resolveOp.Stop()
						}
						r.state.Delete(bopr.name)
					}
				}
				if shouldAdd {
					stateItem := &StateItem{
						resolveOp: dnssd.NewResolveOp(bopr.iface, bopr.name, bopr.serviceType, bopr.domain, func(op *dnssd.ResolveOp, err error, host string, port int, txt map[string]string) {
							if err != nil {
								log.Printf("Resolve operation failed: %s", err)
								return
							}
							// log.Printf("Resolve operation for '%s': %s %d %+v", bopr.name, host, port, txt)
							rc <- ResolveCallback{
								op:   op,
								name: op.Name(),
								host: host,
								port: port,
								txt:  txt,
							}
						}),
						exists:   true,
						resolved: false,
					}
					r.state.Store(bopr.name, stateItem)
					stateItem.resolveOp.Start()
				}
			case ropr := <-rc:
				// log.Printf("Resolve result for %s: %s %d %+v", ropr.name, ropr.host, ropr.port, ropr.txt)
				if rawVal, ok := r.state.Load(ropr.name); ok {
					val := rawVal.(*StateItem)
					if !val.resolved {
						val.host = ropr.host
						val.port = ropr.port
						val.txt = ropr.txt
						val.resolved = true
						for _, cb := range r.OnItemAdded {
							cb(ropr.name, ropr.host, ropr.port, ropr.txt)
						}
					} else {
						// have to check for differences
						if val.host != ropr.host || val.port != ropr.port || !reflect.DeepEqual(val.txt, ropr.txt) {
							val.host = ropr.host
							val.port = ropr.port
							val.txt = ropr.txt
							val.resolved = true
							for _, cb := range r.OnItemAdded {
								cb(ropr.name, ropr.host, ropr.port, ropr.txt)
							}
						}
					}
				} else {
					log.Printf("Should not get here !!!!!!!!!!!")
				}

			case <-sc:
				fmt.Println("quit")
				return
			}
		}
	}(r.browseChan, r.resolveChan, r.stopChan)
}

func (r *Resolver) Stop() {
	r.browseOp.Stop()
	r.stopChan <- 0
}
