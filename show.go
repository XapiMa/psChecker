package pschecker

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

type Monitor struct {
	types      int
	outputPath string
}

func NewMonitor(typesString string, outputPath string) (*Monitor, error) {
	m := new(Monitor)
	var err error
	m.types, err = parseTypes(typesString)
	if err != nil {
		return m, errors.Wrap(err, "cause in NewMonitor")
	}
	m.outputPath = outputPath
	return m, nil
}
func (monitor *Monitor) Show() error {
	pses, err := process.Processes()
	if err != nil {
		return errors.Wrap(err, "cause in Show: processes")
	}
	for _, ps := range pses {
		text := "- "
		if monitor.types&execFlag != 0 {
			exec, err := ps.Exe()
			// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
			if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
				fmt.Fprintf(os.Stderr, "Error: %s\n", errors.Wrap(err, "cause in Show: exec"))
				continue

			} else {

				if text != "- " {
					text += "  "
				}
				text += fmt.Sprintf("exec: %s\n", exec)
			}
		}
		if monitor.types&cmdFlag != 0 {
			cmd, err := ps.Cmdline()
			// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
			if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
				fmt.Fprintf(os.Stderr, "Error: %s\n", errors.Wrap(err, "cause in Show: cmd"))
			} else {
				if text != "- " {
					text += "  "
				}
				text += fmt.Sprintf("cmd: %s\n", cmd)
			}
		}
		if monitor.types&openFlag != 0 {
			files, err := ps.OpenFiles()
			// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
			if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
				fmt.Fprintf(os.Stderr, "Error: %s\n", errors.Wrap(err, "cause in Show: open"))
			} else {
				if len(files) != 0 {
					if text != "- " {
						text += "  "
					}
					text += fmt.Sprintf("open:\n")
					for _, file := range files {
						text += fmt.Sprintf("    - %s\n", file)
					}
				}
			}
		}
		if monitor.types&userFlag != 0 {
			user, err := ps.Username()
			// ErrNotImplementedError cant call because it is defined in ithub.com/shirou/gopsutil/internal/common
			if err != nil && fmt.Sprintf("%s", err) != "not implemented yet" {
				fmt.Fprintf(os.Stderr, "Error: %s\n", errors.Wrap(err, "cause in Show: user"))
			} else {
				if text != "- " {
					text += "  "
				}
				text += fmt.Sprintf("user: %s\n", user)
			}

		}
		if monitor.types&cmdFlag != 0 {
			if text != "- " {
				text += "  "
			}
			pid := ps.Pid
			text += fmt.Sprintf("pid: %d\n", pid)
		}
		if text == "- " {
			continue
		}
		if err := writeNewFile(monitor.outputPath, text); err != nil {
			return errors.Wrap(err, "cause in Show: append")
		}
	}
	return nil
}
