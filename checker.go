package pschecker

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
	"gopkg.in/yaml.v2"
)

// Target is item struct of whitelist and blacklist
type Target struct {
	Exec   string   `yaml:"exec"`
	Cmd    string   `yaml:"cmd"`
	Open   []string `yaml:"open"`
	User   string   `yaml:"user"`
	Pid    int      `yaml:"pid"`
	Regexp string   `yaml:"regexp"`
}

func getProcessesInfo(checkTypes int) ([]Target, error) {
	targets := make([]Target, 0)
	pses, err := process.Processes()
	if err != nil {
		return targets, errors.Wrap(err, "getProcessesInfo: processes")
	}
	wg := &sync.WaitGroup{}
	ch := make(chan Target, 100)
	for _, ps := range pses {
		wg.Add(1)
		go getProcessInfo(checkTypes, ps, wg, ch)
	}

	go func() {
		wg.Wait()
		close(ch)

	}()

	for target := range ch {
		targets = append(targets, target)
	}

	return targets, nil
}

func getProcessInfo(checkTypes int, ps *process.Process, wgp *sync.WaitGroup, ch chan Target) {
	defer wgp.Done()
	target := Target{}
	wg := &sync.WaitGroup{}
	if checkTypes&execFlag != 0 {
		wg.Add(1)
		go getExe(ps, &target, wg)
	}
	if checkTypes&cmdFlag != 0 {
		wg.Add(1)
		go getCmd(ps, &target, wg)
	}
	if checkTypes&openFlag != 0 {
		wg.Add(1)
		go getOpen(ps, &target, wg)
	}
	if checkTypes&userFlag != 0 {
		wg.Add(1)
		go getUser(ps, &target, wg)
	}
	if checkTypes&cmdFlag != 0 {
		target.Pid = int(ps.Pid)
	}
	wg.Wait()
	ch <- target
}

func getExe(ps *process.Process, target *Target, wgp *sync.WaitGroup) {
	defer wgp.Done()
	exec, err := ps.Exe()
	// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
	if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
		log.Printf("Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: exec"))
	} else {
		target.Exec = exec
	}
}

func getCmd(ps *process.Process, target *Target, wgp *sync.WaitGroup) {
	defer wgp.Done()
	cmd, err := ps.Cmdline()
	// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
	if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
		log.Printf("Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: cmd"))
	} else {
		target.Cmd = cmd
	}
}

func getOpen(ps *process.Process, target *Target, wgp *sync.WaitGroup) {
	defer wgp.Done()
	files, err := ps.OpenFiles()
	// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
	if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
		log.Printf("Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: open"))
	} else if len(files) != 0 {
		for _, file := range files {
			target.Open = append(target.Open, file.String())
		}
	}
}

func getUser(ps *process.Process, target *Target, wgp *sync.WaitGroup) {
	defer wgp.Done()
	user, err := ps.Username()
	// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
	if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
		log.Printf("Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: user"))
	} else {
		target.User = user
	}
}

func clearFile(path string) error {
	if path == "" {
		return nil
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrap(err, "cause in clearFile")
	}
	file.Close()
	return nil
}

func parseConfigYml(configPath string) ([]Target, error) {
	targets := make([]Target, 10000)
	errorWrap := func(err error) error {
		return errors.Wrap(err, "cause in parseConfigYml")
	}

	buf, err := ioutil.ReadFile(configPath)
	if err != nil {
		return targets, errorWrap(err)
	}
	if err := yaml.Unmarshal(buf, &targets); err != nil {
		return targets, errorWrap(err)
	}

	return targets, nil

}
