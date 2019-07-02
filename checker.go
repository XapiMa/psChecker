package pschecker

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
	"gopkg.in/yaml.v2"
)

type Target struct {
	Exec string   `yaml:"exec"`
	Cmd  string   `yaml:"cmd"`
	Open []string `yaml:"open"`
	User string   `yaml:"user"`
	Pid  int      `yaml:"pid"`
}

func getProcessInfo(checkTypes int) ([]Target, error) {
	targets := make([]Target, 0)
	pses, err := process.Processes()
	if err != nil {
		return targets, errors.Wrap(err, "getProcessInfo: processes")
	}

	for _, ps := range pses {
		target := Target{}
		if checkTypes&execFlag != 0 {
			exec, err := ps.Exe()
			// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
			if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
				log.Printf("Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: exec"))
			} else {
				target.Exec = exec
			}

		}
		if checkTypes&cmdFlag != 0 {
			cmd, err := ps.Cmdline()
			// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
			if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
				log.Printf("Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: cmd"))
			} else {
				target.Cmd = cmd
			}
		}
		if checkTypes&openFlag != 0 {
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
		if checkTypes&userFlag != 0 {
			user, err := ps.Username()
			// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
			if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
				log.Printf("Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: user"))
			} else {
				target.User = user
			}
		}
		if checkTypes&cmdFlag != 0 {
			target.Pid = int(ps.Pid)
		}
		targets = append(targets, target)
	}
	return targets, nil
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
