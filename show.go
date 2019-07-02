package pschecker

import (
	"fmt"

	"github.com/pkg/errors"
)

type Shower struct {
	types      int
	outputPath string
}

func NewShower(typesString string, outputPath string) (*Shower, error) {
	s := new(Shower)
	var err error
	s.types, err = parseTypes(typesString)
	if err != nil {
		return s, errors.Wrap(err, "cause in NewShower")
	}
	s.outputPath = outputPath
	return s, nil
}

func (shower *Shower) Show() error {

	targets, err := getProcessInfo(shower.types)
	if err != nil {
		return errors.Wrap(err, "cause in Show")
	}
	if len(targets) != 0 {
		if err := clearFile(shower.outputPath); err != nil {
			return errors.Wrap(err, "cause in Show")
		}
	}
	for _, target := range targets {
		text := "- "
		if target.Exec != "" {
			if text != "- " {
				text += "  "
			}
			text += fmt.Sprintf("exec: %s\n", target.Exec)
		}
		if target.Cmd != "" {
			if text != "- " {
				text += "  "
			}
			text += fmt.Sprintf("cmd: %s\n", target.Cmd)
		}
		if len(target.Open) != 0 {
			if text != "- " {
				text += "  "
			}
			text += "open: \n"
			for _, file := range target.Open {
				text += fmt.Sprintf("    - %s", file)
			}
		}
		if target.User != "" {
			if text != "- " {
				text += "  "
			}
			text += fmt.Sprintf("user: %s\n", target.User)
		}
		if target.Pid != 0 {
			if text != "- " {
				text += "  "
			}
			text += fmt.Sprintf("pid: %d\n", target.Pid)
		}
		if text == "- " {
			continue
		}
		if err := appendFile(shower.outputPath, text); err != nil {
			return errors.Wrap(err, "cause in Show:")
		}
	}

	return nil
}
