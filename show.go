package pschecker

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

type Checker struct {
	types      int
	outputPath string
}

func NewChecker(typesString string, outputPath string) (*Checker, error) {
	m := new(Checker)
	var err error
	m.types, err = parseTypes(typesString)
	if err != nil {
		return m, errors.Wrap(err, "cause in NewChecker")
	}
	m.outputPath = outputPath
	return m, nil
}

func clearFile(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrap(err, "cause in clearFile")
	}
	file.Close()
	return nil
}
func (checker *Checker) Show() error {

	targets, err := getProcessInfo(checker.types)
	if err != nil {
		return errors.Wrap(err, "cause in Show")
	}
	if len(targets) != 0 {
		if err := clearFile(checker.outputPath); err != nil {
			return errors.Wrap(err, "cause in Show")
		}
	}
	for _, target := range targets {
		text := "-  "
		if target.Exec != "" {
			if text != "-  " {
				text += "  "
			}
			text += fmt.Sprintf("exec: %s\n", target.Exec)
		}
		if target.Cmd != "" {
			if text != "-  " {
				text += "  "
			}
			text += fmt.Sprintf("cmd: %s\n", target.Cmd)
		}
		if len(target.Open) != 0 {
			if text != "-  " {
				text += "  "
			}
			text += "open: \n"
			for _, file := range target.Open {
				text += fmt.Sprintf("    - %s", file)
			}
		}
		if target.User != "" {
			if text != "-  " {
				text += "  "
			}
			text += fmt.Sprintf("user: %s\n", target.User)
		}
		if target.Pid != 0 {
			if text != "-  " {
				text += "  "
			}
			text += fmt.Sprintf("pid: %s\n", target.Pid)
		}
		if text == "- " {
			continue
		}
		if err := appendFile(checker.outputPath, text); err != nil {
			return errors.Wrap(err, "cause in Show:")
		}
	}

	return nil
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
				fmt.Fprintf(os.Stderr, "Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: exec"))
			} else {
				target.Exec = exec
			}

		}
		if checkTypes&cmdFlag != 0 {
			cmd, err := ps.Cmdline()
			// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
			if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
				fmt.Fprintf(os.Stderr, "Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: cmd"))
			} else {
				target.Cmd = cmd
			}
		}
		if checkTypes&openFlag != 0 {
			files, err := ps.OpenFiles()
			// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
			if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
				fmt.Fprintf(os.Stderr, "Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: open"))
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
				fmt.Fprintf(os.Stderr, "Error: %s\n", errors.Wrap(err, "cause in getProcessInfo: user"))
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
