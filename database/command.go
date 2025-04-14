package database

import "strings"

// cmdTable is a map that associates command names (as strings) with their corresponding command structures
var cmdTable = make(map[string]*command)

type command struct {
	exec  ExecFunc // function to execute the command
	arity int      // number of arguments required for the command
}

// RegisterCommand registers a command with the command table
func RegisterCommand(name string, exec ExecFunc, arity int) {
	name = strings.ToLower(name)
	cmdTable[name] = &command{
		exec:  exec,
		arity: arity,
	}
}
