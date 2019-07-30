package pschecker

import (
	"fmt"
	"log"
	"sync"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

func getProc(pid int) (Proc, error) {
	p := Proc{}
	p.Pid = pid

	ps, err := process.NewProcess(int32(p.Pid))
	if err != nil {
		return p, errors.Wrap(err, "in getProc")
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go p.setExe(ps, wg)
	wg.Add(1)
	go p.setCmd(ps, wg)
	wg.Add(1)
	go p.setOpen(ps, wg)
	wg.Add(1)
	go p.setUser(ps, wg)

	wg.Wait()
	return p, nil
}

func (p *Proc) setExe(ps *process.Process, wgp *sync.WaitGroup) {
	defer wgp.Done()
	exec, err := ps.Exe()
	// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
	if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
		log.Printf("Error: %s\n", errors.Wrap(err, "cause in getExec"))
	} else {
		p.Exec = exec
	}
}

func (p *Proc) setCmd(ps *process.Process, wgp *sync.WaitGroup) {
	defer wgp.Done()
	cmd, err := ps.Cmdline()
	// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
	if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
		log.Printf("Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: cmd"))
	} else {
		p.Cmd = cmd
	}
}

func (p *Proc) setOpen(ps *process.Process, wgp *sync.WaitGroup) {
	defer wgp.Done()

	files, err := ps.OpenFiles()
	// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
	if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
		log.Printf("Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: open"))
	} else if len(files) != 0 {
		for _, file := range files {
			p.Open = append(p.Open, file.String())
		}
	}
}

func (p *Proc) setUser(ps *process.Process, wgp *sync.WaitGroup) {
	defer wgp.Done()
	user, err := ps.Username()
	// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
	if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
		log.Printf("Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: user"))
	} else {
		p.User = user
	}
}
