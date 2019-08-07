package psmonitor

import "errors"

type MessageType string

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
	ErrorMessage   = MessageType("ERROR")
	WarningMessage = MessageType("WARNIG")
	ActiveMessage  = MessageType("ACTIVE")
)

var (
	ErrNotImplementedError = errors.New("not implemented yet")
)
