package module

import (
	"fmt"
	"github.com/LuisZhou/lpge/chanrpc"
	"github.com/LuisZhou/lpge/conf"
	"github.com/LuisZhou/lpge/log"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// todo: start a tcp server, to listen cluster.
// two kind of agent: module agent,

type Module interface {
	OnInit()
	OnDestroy()
	Run(closeSig chan bool)
	AsynCall(server *chanrpc.Server, id interface{}, args ...interface{})
	SynCall(server *chanrpc.Server, id interface{}, args ...interface{})
	RegisterChanRPC(id interface{}, f interface{})
	GetChanrpcServer() *chanrpc.Server
}

type module struct {
	Module
	closeSig chan bool
	wg       sync.WaitGroup
}

var (
	mods  []*module
	names map[string]*module = make(map[string]*module)
	addr  uint               = 0
	mutex sync.Mutex
)

func Register(mi Module, name string) (err error) {
	namesz := len(name)
	if namesz > 0 {
		mutex.Lock()
		if _, ok := names[name]; ok {
			err = fmt.Errorf("dulplicate name of module")
		}
		mutex.Unlock()

		if err != nil {
			return
		}
	}

	m := new(module)
	m.Module = mi
	m.closeSig = make(chan bool, 1)

	// goroutine safe
	mutex.Lock()
	addr++
	if namesz > 0 {
		names[name] = m
	} else {
		names[strconv.Itoa(int(addr))] = m
	}
	mutex.Unlock()

	// todo
	// search module for cluster.
	// explore API for cluster call.

	m.Module.OnInit()
	m.wg.Add(1)
	go run(m)

	mods = append(mods, m)

	return nil
}

func Destroy() {
	for i := len(mods) - 1; i >= 0; i-- {
		m := mods[i]
		m.closeSig <- true
		// wait the run to return. Reason is that, if the destory release the
		// resource which the goroutine in run is using, the program may panic.
		m.wg.Wait()
		destroy(m)
	}
	// release map
	for k := range names {
		delete(names, k)
	}
}

func run(m *module) {
	m.Module.Run(m.closeSig)
	m.wg.Done()
}

func destroy(m *module) {
	// onDestory is defined by outside, means some kinds of danger of panic,
	// so recover if panic in OnDestroy.
	defer func() {
		if r := recover(); r != nil {
			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				log.Error("%v: %s", r, buf[:l])
			} else {
				log.Error("%v", r)
			}
		}
	}()

	m.Module.OnDestroy()
}

func Search(name string) (m *module, is_remote bool, err error) {
	// todo: to support remote server.
	arr := strings.Split(name, ":")
	len_of_arr := len(arr)
	if len_of_arr == 2 {
		return nil, true, fmt.Errorf("Not support remote name now: %s", name)
	} else if len_of_arr == 1 {
		if m, ok := names[name]; ok {
			return m, false, nil
		} else {
			return nil, false, fmt.Errorf("Not found for name: %s", name)
		}
	} else {
		return nil, false, fmt.Errorf("Unsupport format: %s", name)
	}
}

// Send do a async call to module with name 'name'.
func (m *module) Send(name string, cmd uint16, data interface{}) error {
	m1, is_remote, err := Search(name)
	if err != nil {
		return err
	}

	if !is_remote {
		m.Module.AsynCall(m1.Module.GetChanrpcServer(), cmd, data, func(interface{}, error) {})
	} else {
	}

	return nil
}

// Call
func (m *module) Call(name string, cmd uint16, data interface{}) error {
	m1, is_remote, err := Search(name)
	if err != nil {
		return err
	}

	if !is_remote {
		m.Module.SynCall(m1.Module.GetChanrpcServer(), cmd, data)
	} else {
	}

	return nil
}
