package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const launchAgentsDir = "/Users/wilson1/Library/LaunchAgents"

// CmdResult 命令执行结果（标准输出 + 标准错误）
type CmdResult struct {
	Stdout string
	Stderr string
}

// LaunchCtl 执行 launchctl 命令
type LaunchCtl struct {
	agentsDir string
}

// NewLaunchCtl 创建 LaunchCtl 服务
func NewLaunchCtl() *LaunchCtl {
	return &LaunchCtl{agentsDir: launchAgentsDir}
}

func (l *LaunchCtl) run(cmd *exec.Cmd) (CmdResult, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return CmdResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}, err
}

// Bootstrap 执行 launchctl bootstrap gui/$(id -u) .../name.plist，返回终端输出
func (l *LaunchCtl) Bootstrap(name string) (CmdResult, error) {
	plistPath := filepath.Join(l.agentsDir, name+".plist")
	domain := fmt.Sprintf("gui/%d", os.Getuid())
	cmd := exec.Command("launchctl", "bootstrap", domain, plistPath)
	return l.run(cmd)
}

// Bootout 执行 launchctl bootout gui/$(id -u) .../name.plist，返回终端输出
func (l *LaunchCtl) Bootout(name string) (CmdResult, error) {
	plistPath := filepath.Join(l.agentsDir, name+".plist")
	domain := fmt.Sprintf("gui/%d", os.Getuid())
	cmd := exec.Command("launchctl", "bootout", domain, plistPath)
	return l.run(cmd)
}

// List 执行 launchctl list 并过滤包含 name 的行（等价于 launchctl list | grep name），返回过滤后的输出
func (l *LaunchCtl) List(name string) (CmdResult, error) {
	cmd := exec.Command("launchctl", "list")
	res, err := l.run(cmd)
	if err != nil {
		return res, err
	}
	var out []string
	for _, line := range strings.Split(res.Stdout, "\n") {
		if strings.Contains(line, name) {
			out = append(out, line)
		}
	}
	res.Stdout = strings.Join(out, "\n")
	return res, nil
}
