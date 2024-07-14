package executor

import (
	"bufio"
	"context"
	"fmt"
	"github.com/MR5356/jietan/pkg/common"
	"github.com/MR5356/jietan/pkg/executor/api"
	"github.com/axgle/mahonia"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io"
	"strings"
	"sync"
	"time"
)

type RemoteShellExecutor struct {
	log       []string
	hostInfos map[string]*HostResult
}

type HostResult struct {
	hostInfo *api.HostInfo
	log      []string
	stderr   []string
	finish   bool
	err      error
}

func (e *RemoteShellExecutor) Execute(context context.Context, params *api.ExecuteParams) *api.ExecuteResult {
	hosts := params.GetParam(api.ExecuteParamHosts).([]*api.HostInfo)
	script := params.GetScript()
	ps := params.GetParams()
	e.hostInfos = make(map[string]*HostResult)
	wg := sync.WaitGroup{}

	for _, host := range hosts {
		wg.Add(1)
		e.hostInfos[host.Id()] = &HostResult{
			hostInfo: host,
			log:      make([]string, 0),
			finish:   false,
			stderr:   make([]string, 0),
		}
		host := host
		go func() {
			defer wg.Done()
			e.hostInfos[host.Id()].err = e.remoteExecute(context, host, script, ps)
			e.hostInfos[host.Id()].finish = true
		}()
	}
	wg.Wait()
	result := &api.ExecuteResult{
		Status:  api.Success,
		Message: "success",
		Data: map[string]interface{}{
			"log":   e.getLog(),
			"error": e.getError(),
		},
	}
	return result
}

func (e *RemoteShellExecutor) GetResult(field api.ResultField, params interface{}) interface{} {
	switch field {
	case api.ResultFieldLog:
		return e.getLog()
	case api.ResultFieldErr:
		return e.getError()
	case api.ResultFieldIncrLog:
		return e.getIncrLog(params.(int))
	}
	return nil
}

func (e *RemoteShellExecutor) getError() map[string]error {
	result := make(map[string]error)
	for _, hostInfo := range e.hostInfos {
		result[hostInfo.hostInfo.String()] = hostInfo.err
	}
	return result

}

func (e *RemoteShellExecutor) getLog() map[string][]string {
	log := make(map[string][]string, len(e.hostInfos))
	for _, hostInfo := range e.hostInfos {
		log[hostInfo.hostInfo.String()] = hostInfo.log
		if hostInfo.err != nil {
			log[hostInfo.hostInfo.String()] = append(log[hostInfo.hostInfo.String()], "[ERROR]"+hostInfo.err.Error())
		}
	}
	return log
}

func (e *RemoteShellExecutor) getIncrLog(start int) *api.IncrementalLog {
	if start > len(e.log) {
		start = len(e.log)
	}
	more := false
	for _, hostInfo := range e.hostInfos {
		if !hostInfo.finish {
			more = true
			break
		}
	}

	return &api.IncrementalLog{
		Start: start,
		End:   len(e.log),
		Log:   e.log[start:],
		More:  more,
	}
}

func (e *RemoteShellExecutor) remoteExecute(context context.Context, hostInfo *api.HostInfo, script string, params ...string) error {
	logrus := logrus.WithField("prefix", e.getLogPrefix(hostInfo))

	clientConfig := &ssh.ClientConfig{
		Timeout:         time.Second * 3,
		User:            hostInfo.User,
		Auth:            hostInfo.GetAuthMethods(),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", hostInfo.Host, hostInfo.Port), clientConfig)
	if err != nil {
		logrus.Errorf("ssh dial error: %v", err)
		return err
	}
	defer client.Close()

	scp, err := sftp.NewClient(client)
	if err != nil {
		logrus.Errorf("sftp new client error: %v", err)
		return err
	}
	defer scp.Close()
	scriptName := fmt.Sprintf("/tmp/%s-%d.sh", common.PackageName, time.Now().UnixMilli())
	target, err := scp.Create(scriptName)
	if err != nil {
		logrus.Errorf("sftp open error: %v", err)
		return err
	}
	_, err = io.Copy(target, strings.NewReader(script))
	if err != nil {
		return err
	}
	logrus.Infof("transfer script to remote file: %s", scriptName)

	err = target.Close()
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		logrus.Errorf("ssh new session error: %v", err)
		return err
	}
	defer session.Close()

	stdout, _ := session.StdoutPipe()
	stderr, _ := session.StderrPipe()

	enc := mahonia.NewDecoder("utf-8")
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
			e.hostInfos[hostInfo.Id()].log = append(e.hostInfos[hostInfo.Id()].log, s)
			e.log = append(e.log, fmt.Sprintf("%s %s", s))
			logrus.Debugf("%s", s)
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
			e.hostInfos[hostInfo.Id()].log = append(e.hostInfos[hostInfo.Id()].log, s)
			e.hostInfos[hostInfo.Id()].stderr = append(e.hostInfos[hostInfo.Id()].stderr, s)
			e.log = append(e.log, fmt.Sprintf("%s %s", s))
			logrus.Debugf("%s", s)
		}
	}()

	logrus.Debugf("cmd: %s", fmt.Sprintf(`chmod +x %s; sh %s %s; rm -f %s`, scriptName, scriptName, strings.Join(params, " "), scriptName))
	err = session.Start(fmt.Sprintf(`chmod +x %s; %s %s; rm -f %s`, scriptName, scriptName, strings.Join(params, " "), scriptName))
	if err != nil {
		logrus.Errorf("ssh start error: %v", err)
	}
	finish := make(chan struct{})
	go func() {
		defer func() {
			logrus.Infof("command execution completed")
			finish <- struct{}{}
		}()
		wg.Wait()
		err = session.Wait()
		if err != nil {
			logrus.Errorf("ssh wait error: %v", err)
		}
	}()

	select {
	case <-context.Done():
		err := session.Signal(ssh.SIGTERM)
		if err != nil {
			logrus.Errorf("session close error: %v", err)
		}

		return err
	case <-finish:
		return err
	}
}

func (e *RemoteShellExecutor) getLogPrefix(hostInfo *api.HostInfo) string {
	return fmt.Sprintf("RemoteShellExecutor-%s@%s:%d", hostInfo.User, hostInfo.Host, hostInfo.Port)
}
