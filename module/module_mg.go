// Pacakge module define interface and implement one skeleton of Module, provider a global wide module manager.
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

// Module interface defines what API a module should implement.
type Module interface {
	OnInit()                                                                                  // callback when init.
	OnDestroy()                                                                               // callback when destory.
	Run(closeSig chan bool)                                                                   // start run.
	AsynCall(server *chanrpc.Server, id interface{}, args ...interface{}) error               // async call to rpc server.
	SynCall(server *chanrpc.Server, id interface{}, args ...interface{}) (interface{}, error) // sync call to rpc server.
	RegisterChanRPC(id interface{}, f func([]interface{}) (interface{}, error))               // register rpc handler.
	GetChanrpcServer() *chanrpc.Server                                                        // get its rpc server.
}

// module is a the unit, the module manager can manage.
type module struct {
	Module                  // Module interface.
	closeSig chan bool      // closeSig is used to destory Module.
	wg       sync.WaitGroup // wg is used to wait Module stop running completely for the manager.
}

var (
	names map[string]*module = make(map[string]*module) // map between name and module.
	addr  uint               = 0                        // addr counter for assign name to module when there is no name explictly.
	mutex sync.Mutex                                    // mutex for protect register
)

// Register a Module to this manager.
func Register(mi Module, name string) (err error) {
	mutex.Lock()
	defer func() {
		mutex.Unlock()
	}()

	namesz := len(name)
	if namesz > 0 {
		if _, ok := names[name]; ok {
			return fmt.Errorf("dulplicate name of module")
		}
	} else {
		addr++
		name = strconv.Itoa(int(addr))
	}

	m := new(module)
	m.Module = mi
	m.closeSig = make(chan bool, 1)
	names[name] = m

	m.Module.OnInit()
	m.wg.Add(1)
	go run(m)

	return nil
}

// run start running of module.
func run(m *module) {
	m.Module.Run(m.closeSig)
	m.wg.Done()
}

// destory do destory module internal.
func destroy(m *module) {
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

// Destory all module.
func Destroy() {
	for k, _ := range names {
		DestroyOne(k)
	}
}

// Destory one module.
func DestroyOne(name string) {
	m, _ := names[name]
	m.closeSig <- true
	m.wg.Wait()
	destroy(m)
	delete(names, name)
}

// Search module according its name.
func Search(name string) (m *module, is_remote bool, err error) {
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

// AsynCall do a async call to module.
func (m *module) AsynCall(name string, cmd uint16, data interface{}) error {
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

// SynCall do a synCall call to module.
func (m *module) SynCall(name string, cmd uint16, data interface{}) error {
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
