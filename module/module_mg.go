package module

import (
	"fmt"
	"github.com/LuisZhou/lpge/conf"
	"github.com/LuisZhou/lpge/log"
	"runtime"
	"strings"
	"sync"
)

type Module interface {
	OnInit()
	OnDestroy()
	Run(closeSig chan bool)
}

type module struct {
	mi       Module
	closeSig chan bool
	wg       sync.WaitGroup
	address  uint
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
	m.mi = mi
	m.closeSig = make(chan bool, 1)

	// goroutine safe
	mutex.Lock()
	m.address = addr
	addr++
	if namesz > 0 {
		names[name] = m
	}
	mutex.Unlock()

	// todo
	// search module for cluster.
	// explore API for cluster call.

	m.mi.OnInit()
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
	m.mi.Run(m.closeSig)
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

	m.mi.OnDestroy()
}

// todo: to support remote server.
func Search(name string) (m *module, err error) {
	arr := strings.Split(name, ":")
	len_of_arr := len(arr)
	if len_of_arr == 2 {
		return nil, fmt.Errorf("Not support remote name now: ", name)
	} else if len_of_arr == 1 {
		if m, ok := names[name]; ok {
			return m, nil
		} else {
			return nil, fmt.Errorf("Not found for name: ", name)
		}
	} else {
		return nil, fmt.Errorf("Unsupport format: ", name)
	}
}
