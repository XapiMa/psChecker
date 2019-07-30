package pschecker

import (
	"testing"
)

func TestParseConfigYml(t *testing.T) {

	targets, err := parseConfigYml("./config.yml")
	if err != nil {
		t.Error(err)
	}
	t.Errorf("%v\n", targets)

}
