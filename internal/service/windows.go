//go:build windows

package service

import (
	"fmt"
	"os/exec"
	"strings"
)

type Windows struct{}

const svcName = "cftunnel"

func (w *Windows) Install(binPath, token string) error {
	binArg := fmt.Sprintf(`%s tunnel run --token %s`, binPath, token)
	if err := exec.Command("sc", "create", svcName, "binPath=", binArg, "start=", "auto").Run(); err != nil {
		return fmt.Errorf("创建服务失败: %w", err)
	}
	return exec.Command("sc", "start", svcName).Run()
}

func (w *Windows) Uninstall() error {
	exec.Command("sc", "stop", svcName).Run()
	return exec.Command("sc", "delete", svcName).Run()
}

func (w *Windows) Running() bool {
	out, err := exec.Command("sc", "query", svcName).Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "RUNNING")
}

func New() Service {
	return &Windows{}
}
