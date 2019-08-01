// +build !linux

package psmonitor

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
