package utils

// ToCmdLine convert strings to [][]byte
func ToCmdLine(cmd ...string) [][]byte {
	args := make([][]byte, len(cmd))
	for i, s := range cmd {
		args[i] = []byte(s)
	}
	return args
}

// ToCmdLineWithName convert command name and args to [][]byte
func ToCmdLineWithName(name string, args ...[]byte) [][]byte {
	cmd := make([][]byte, len(args)+1)
	cmd[0] = []byte(name)
	for i, s := range args {
		cmd[i+1] = s
	}
	return cmd
}
