package api

import (
	"crypto/md5"
	"fmt"
	"golang.org/x/crypto/ssh"
)

type HostInfo struct {
	Host   string `json:"host" yaml:"host"`
	Port   int    `json:"port" yaml:"port"`
	User   string `json:"user" yaml:"user"`
	Passwd string `json:"passwd" yaml:"passwd"`

	PrivateKey string `json:"privateKey"`
	Passphrase string `json:"passphrase"`
}

func (i *HostInfo) String() string {
	return fmt.Sprintf("%s@%s:%d", i.User, i.Host, i.Port)
}

func (i *HostInfo) Id() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(i.String())))
}

func (h *HostInfo) GetAuthMethods() []ssh.AuthMethod {
	authMethods := make([]ssh.AuthMethod, 0)

	if len(h.PrivateKey) > 0 {
		if len(h.Passphrase) > 0 {
			signer, err := ssh.ParsePrivateKeyWithPassphrase([]byte(h.PrivateKey), []byte(h.Passphrase))
			if err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		} else {
			signer, err := ssh.ParsePrivateKey([]byte(h.PrivateKey))
			if err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}

	if len(h.Passwd) > 0 {
		authMethods = append(authMethods, ssh.Password(h.Passwd))
	}
	return authMethods
}
