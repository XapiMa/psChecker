// +build !linux

package pschecker

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
