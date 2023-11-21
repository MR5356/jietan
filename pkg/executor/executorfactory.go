package executor

var executorFactory map[string]Executor

func GetExecutor(name string) Executor {
	if executor, ok := executorFactory[name]; ok {
		return executor
	}
	return nil
}

func init() {
	executorFactory = map[string]Executor{
		"command": &CommandExecutor{},
		"shell":   &ShellExecutor{},
		"remote":  &RemoteShellExecutor{},
	}
}
