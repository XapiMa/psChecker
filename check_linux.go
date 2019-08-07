package psmonitor

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

func (m *Monitor) check() error {

	err := m.initialProc()
	if err != nil {
		return errors.Wrap(err, "in check")
	}
	fmt.Println("Checking processes infomretions")

	if err := m.firstCheck(); err != nil {
		return errors.Wrap(err, "in check")
	}

	// for {
	// 	select {
	// 	case event := <-m.watcher.Events:
	// 		switch {
	// 		case event.Op&fsnotify.Create == fsnotify.Create:
	// 			if isProcDir(event.Name) {
	// 				go m.addProc(event.Name)
	// 			} else if isfdDir(event.Name) {
	// 				go m.changeProc(openFlag, event.Name)
	// 			}
	// 		case event.Op&fsnotify.Remove == fsnotify.Remove:
	// 			if isProcDir(event.Name) {
	// 				go m.delProc(event.Name)
	// 			} else if isfdDir(event.Name) {
	// 				go m.changeProc(openFlag, event.Name)
	// 			}
	// 		}
	// 	case err := <-m.watcher.Errors:
	// 		return errors.Wrap(err, "in check")
	// 	}
	// }
	return nil
}

func (m *Monitor) initialProc() error {
	proc := "/proc"
	m.watcher.Add(proc)
	fileinfos, err := ioutil.ReadDir(proc)
	if err != nil {
		return errors.Wrap(err, "in addProc")
	}
	wg := &sync.WaitGroup{}
	for _, fi := range fileinfos {
		path := filepath.Join(proc, fi.Name())
		if isFdDir(path) {
			wg.Add(1)
			go m.addFd(path, wg)
		}
	}
	wg.Wait()
	return nil
}

func (m *Monitor) firstCheck() error {
	pses, err := process.Processes()
	if err != nil {
		return err
	}
	proc := make(chan Proc)
	wgGet := &sync.WaitGroup{}
	wgCheck := &sync.WaitGroup{}
	wgCheck.Add(1)
	go func() {
		for p := range proc {
			wgCheck.Add(1)
			go m.addCaseCheckList(p, wgCheck)
		}
		wgCheck.Done()
	}()
	for _, ps := range pses {
		wgGet.Add(1)

		go getProcFromPs(ps, proc, wgGet)
	}
	go func() {
		wgGet.Wait()
		close(proc)
	}()
	wgCheck.Wait()
	return nil
}

func (m *Monitor) addFd(path string, wgp *sync.WaitGroup) {
	defer wgp.Done()
	if isExist(path) {
		m.watcher.Add(path)
	}
}

func (m *Monitor) addCaseCheckList(p Proc, pwg *sync.WaitGroup) {
	defer pwg.Done()
	ts := []*Target{}
	if ok, t := m.inWhite(p); ok {
		if err := m.wcc.add(t); err != nil {
			log.Printf("in addCaseCheckList: %s", err)
		}
		ts = append(ts, t)
	}
	if ok, t := m.inBlack(p); ok {
		if err := m.bcc.add(t); err != nil {
			log.Printf("in addCaseCheckList: %s", err)
		}
		ts = append(ts, t)
	}
	if err := m.cache.add(cacheItem{ts, p}); err != nil {
		log.Printf("in addCaseCheckList: %s", err)
	}
}

func (m *Monitor) inWhite(p Proc) (bool, *Target) {
	for t := range m.wcc {
		if m.equal(p, t) {
			return true, t
		}
	}
	return false, nil
}

func (m *Monitor) inBlack(p Proc) (bool, *Target) {
	for t := range m.bcc {
		if m.equal(p, t) {
			return true, t
		}
	}
	return false, nil
}

func (m *Monitor) equal(p Proc, t *Target) bool {
	if t.Exec != "" && t.Exec != p.exec {
		return false
	}
	if t.Cmd != "" && t.Cmd != p.cmd {
		return false
	}
	if len(t.Open) != 0 {
		if fmt.Sprintf("%v", t.Open) != fmt.Sprintf("%v", p.open) {
			return false
		}
	}
	if t.User != "" && t.User != p.user {
		return false
	}
	if t.Pid != 0 && t.Pid != p.pid {
		return false
	}
	if t.Regexp != "" {
		for x := 0; x < 1; x++ {
			if ok, err := m.matchPattern(p.exec, t.Regexp); err != nil {
				log.Print(errors.Wrap(err, "couse in is"))
			} else if ok {
				return true
			}
			if ok, err := m.matchPattern(p.cmd, t.Regexp); err != nil {
				log.Print(errors.Wrap(err, "couse in is"))
			} else if ok {
				return true
			}
			if ok, err := m.matchPattern(p.user, t.Regexp); err != nil {
				log.Print(errors.Wrap(err, "couse in is"))
			} else if ok {
				return true
			}
			for _, open := range p.open {
				if ok, err := m.matchPattern(open.Path, t.Regexp); err != nil {
					log.Print(errors.Wrap(err, "couse in is"))
				} else if ok {
					return true
				}
			}
		}
		return false
	}
	return true
}
