package executor

import (
	"bufio"
	"context"
	"fmt"
	"github.com/MR5356/jietan/pkg/common"
	"github.com/MR5356/jietan/pkg/executor/api"
	"github.com/MR5356/jietan/pkg/utils/fileutil"
	"github.com/axgle/mahonia"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type ShellExecutor struct {
	log    []string
	stderr []string
}

func (e *ShellExecutor) Execute(context context.Context, params *api.ExecuteParams) *api.ExecuteResult {
	logrus := logrus.WithField("prefix", "ShellExecutor")

	script := params.GetScript()
	ps := params.GetParams()

	res := &api.ExecuteResult{}
	enc := mahonia.NewDecoder("utf-8")

	scriptName := fmt.Sprintf("/tmp/%s-%d.sh", common.PackageName, time.Now().UnixMilli())
	err := fileutil.WriteFile(scriptName, script)
	if err != nil {
		res.Status = api.Failed
		res.Message = err.Error()
		return res
	}
	defer os.RemoveAll(scriptName)

	logrus.Debugf("cmd: %s", fmt.Sprintf("chmod +x %s; sh %s %s", scriptName, scriptName, ps))
	cmd := exec.Command("sh", "-c", fmt.Sprintf("chmod +x %s && sh %s %s", scriptName, scriptName, ps))

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	wg := sync.WaitGroup{}
	go func() {
		wg.Add(1)
		defer wg.Done()
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
		wg.Add(1)
		defer wg.Done()
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
	err = cmd.Start()
	if err != nil {
		res.Status = api.Failed
		res.Message = err.Error()
		return res
	}
	logrus.Infof("start cmd: %s", script)

	go func() {
		wg.Wait()
		err = cmd.Wait()
		if err != nil {
			logrus.Errorf("run command error: %+v", strings.Join(e.stderr, "\n"))
			logrus.Errorf("[%s] cmd: %s", err.Error(), script)
			res.Status = api.Failed
			res.Message = err.Error()
			finish <- struct{}{}
		} else {
			logrus.Debugf("run command success: %s", script)
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

func (e *ShellExecutor) GetResult(field api.ResultField, params interface{}) interface{} {
	return nil
}
