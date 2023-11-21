package api

import (
	"crypto/md5"
	"fmt"
)

type HostInfo struct {
	Host   string `json:"host" yaml:"host"`
	Port   int    `json:"port" yaml:"port"`
	User   string `json:"user" yaml:"user"`
	Passwd string `json:"passwd" yaml:"passwd"`
}

func (i *HostInfo) String() string {
	return fmt.Sprintf("%s@%s:%d", i.User, i.Host, i.Port)
}

func (i *HostInfo) Id() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(i.String())))
}
