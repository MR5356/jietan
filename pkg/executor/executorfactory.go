package executor

//var executorFactory map[string]Executor

func GetExecutor(name string) Executor {
	switch name {
	case "command":
		return &CommandExecutor{}
	case "shell":
		return &ShellExecutor{}
	case "remote":
		return &RemoteShellExecutor{}
	default:
		return nil
	}
}

//func init() {
//	executorFactory = map[string]Executor{
//		"command": &CommandExecutor{},
//		"shell":   &ShellExecutor{},
//		"remote":  &RemoteShellExecutor{},
//	}
//}
