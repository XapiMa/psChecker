package psmonitor

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

type OpenFile struct {
	Path string `json:"path"`
	Fd   int    `json:"fd"`
}

type Proc struct {
	exec string
	cmd  string
	open []OpenFile
	user string
	pid  int
}

func getProcFromPs(ps *process.Process, proc chan Proc, pwg *sync.WaitGroup) {
	defer pwg.Done()

	p := Proc{}
	p.pid = int(ps.Pid)

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go p.setExec(ps, wg)
	wg.Add(1)
	go p.setCmd(ps, wg)
	wg.Add(1)
	go p.setOpen(ps, wg)
	wg.Add(1)
	go p.setUser(ps, wg)

	wg.Wait()
	proc <- p
}

func getProcFromPid(pid int) (proc chan Proc, pwg *sync.WaitGroup) {
	defer pwg.Done()

	p := Proc{}
	p.pid = pid

	ps, err := process.NewProcess(int32(p.pid))
	if err != nil {
		fmt.Printf("in getProc: %s", err)
		return
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go p.setExec(ps, wg)
	wg.Add(1)
	go p.setCmd(ps, wg)
	wg.Add(1)
	go p.setOpen(ps, wg)
	wg.Add(1)
	go p.setUser(ps, wg)

	wg.Wait()
	proc <- p
	return
}

func getOpenFromFdPath(path string) (OpenFile, error) {
	o := OpenFile{}
	path = filepath.Clean(path)
	p := strings.Split(path, "/")
	if p[0] != "" || p[1] != "proc" || !isNum(p[2]) || p[3] != "proc" || !isNum(p[4]) {
		return o, fmt.Errorf("%s is not fd", path)
	}
	filepath, err := os.Readlink(path)
	if err != nil {
		return o, fmt.Errorf("%s cant read", path)
	}
	o.Path = filepath
	num, _ := strconv.Atoi(p[4])
	o.Fd = num
	return o, nil
}

func (p *Proc) setExec(ps *process.Process, wgp *sync.WaitGroup) {
	defer wgp.Done()
	exec, err := getExec(ps)
	if err == ErrNotImplementedError {
		return
	} else if err != nil {
		log.Printf("Error: %s\n", errors.Wrap(err, "cause in getExec"))
		return
	}
	p.exec = exec
}

func (p *Proc) setCmd(ps *process.Process, wgp *sync.WaitGroup) {
	defer wgp.Done()
	cmd, err := getCmd(ps)
	if err == ErrNotImplementedError {
		return
	} else if err != nil {
		log.Printf("Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: cmd"))
		return
	}
	p.cmd = cmd
}

func (p *Proc) setOpen(ps *process.Process, wgp *sync.WaitGroup) {
	defer wgp.Done()
	o, err := getOpen(ps)
	if err == ErrNotImplementedError {
		return
	} else if err != nil {
		log.Printf("Error: %s\n", errors.Wrap(err, "in setOpen"))
		return
	}
	p.open = make([]OpenFile, len(o))
	copy(p.open, o)
}

func (p *Proc) setUser(ps *process.Process, wgp *sync.WaitGroup) {
	defer wgp.Done()
	user, err := getUser(ps)
	if err == ErrNotImplementedError {
		return
	} else if err != nil {
		log.Printf("Error: %s\n", errors.Wrap(err, "in setUser"))
		return
	}
	p.user = user
}

func getExec(ps *process.Process) (string, error) {
	exec, err := ps.Exe()
	err = isImplemented(err)
	return exec, err
}

func getCmd(ps *process.Process) (string, error) {
	cmd, err := ps.Cmdline()
	err = isImplemented(err)
	return cmd, err
}

func getOpen(ps *process.Process) ([]OpenFile, error) {
	opens := make([]OpenFile, 0)
	files, err := ps.OpenFiles()
	err = isImplemented(err)
	if err != nil {
		return opens, err
	}
	for _, file := range files {
		o, err := parseOpen(file.Path)
		if err != nil {
			return nil, errors.Wrap(err, "in getOpen")
		}
		opens = append(opens, o)
	}
	sort.Slice(opens, func(i, j int) bool { return opens[i].Path < opens[j].Path })
	return opens, err
}

func getUser(ps *process.Process) (string, error) {
	user, err := ps.Username()
	err = isImplemented(err)
	return user, err
}

func isImplemented(err error) error {
	// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
	if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
		return ErrNotImplementedError
	}
	return err
}

func parseOpen(str string) (OpenFile, error) {
	o := OpenFile{}
	if err := json.Unmarshal([]byte(str), &o); err != nil {
		return o, err
	}
	return o, nil
}
