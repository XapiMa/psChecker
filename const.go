// +build !linux

package psmonitor

import "errors"

const (
	execSentence = "exec"
	cmdSentence  = "cmd"
	openSentence = "open"
	userSentence = "user"
	pidSentence  = "pid"
	execFlag     = 1 << iota
	cmdFlag
	openFlag
	userFlag
	pidFlag
)

var (
	ErrNotImplementedError = errors.New("not implemented yet")
)
