package psmonitor

import "github.com/pkg/errors"

// Shower is struct for show command
type Shower struct {
	types      int
	outputPath string
}

// NewShower create new shower object
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

// Show shows prosesses information
func (shower *Shower) Show() error {

	return nil
}
