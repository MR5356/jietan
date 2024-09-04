package executor

import (
	"bufio"
	"context"
	"github.com/MR5356/jietan/pkg/executor/api"
	"github.com/axgle/mahonia"
	"github.com/sirupsen/logrus"
	"io"
	"os/exec"
	"strings"
)

type CommandExecutor struct {
	log    []string
	stderr []string
}

func (e *CommandExecutor) Execute(context context.Context, params *api.ExecuteParams) *api.ExecuteResult {
	res := &api.ExecuteResult{}
	enc := mahonia.NewDecoder("utf-8")

	cmd := exec.Command("sh", "-c", params.GetScript())

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	go func() {
		reader := bufio.NewReader(stdout)
		for {
			readStr, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				return
			}
			s := enc.ConvertString(strings.ReplaceAll(readStr, "\n", ""))
			e.log = append(e.log, s)
			logrus.Debugln(s)
		}
	}()

	go func() {
		reader := bufio.NewReader(stderr)
		for {
			readStr, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				return
			}
			s := enc.ConvertString(strings.ReplaceAll(readStr, "\n", ""))
			e.log = append(e.log, s)
			e.stderr = append(e.stderr, s)
			logrus.Debugln(s)
		}
	}()

	finish := make(chan struct{})
	err := cmd.Start()
	if err != nil {
		res.Status = api.Failed
		res.Message = err.Error()
		return res
	}
	logrus.Infof("[%d] cmd: %s", cmd.Process.Pid, cmd)

	go func() {
		err = cmd.Wait()
		if err != nil {
			logrus.Errorf("run command error: %+v", strings.Join(e.stderr, "\n"))
			logrus.Errorf("[%s] cmd: %s", err.Error(), cmd)
			res.Status = api.Failed
			res.Message = err.Error()
			finish <- struct{}{}
		} else {
			logrus.Debugf("[done] cmd: %s", cmd)
			res.Status = api.Success
			res.Message = "success"
			finish <- struct{}{}
		}
	}()

	select {
	case <-context.Done():
		err := cmd.Process.Kill()
		if err != nil {
			logrus.Errorf("kill error: %v", err)
		}

		res.Status = api.Failed
		res.Message = context.Err().Error()
		return res
	case <-finish:
		return res
	}
}

func (e *CommandExecutor) GetResult(field api.ResultField, param interface{}) interface{} {
	return nil
}
