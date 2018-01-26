package huker

import (
	"bytes"
	"fmt"
	"github.com/qiniu/log"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	DATA_DIR      = "data"
	LOG_DIR       = "log"
	PKG_DIR       = "pkg"
	CONF_DIR      = "conf"
	LIBRARY_DIR   = ".packages"
	STDOUT_DIR    = "stdout"
	HOOKS_DIR     = "hooks"
	StatusRunning = "Running"
	StatusStopped = "Stopped"
)

func progDirs() []string {
	return []string{DATA_DIR, LOG_DIR, CONF_DIR, STDOUT_DIR}
}

type Program struct {
	Name       string            `json:"name"`
	Job        string            `json:"job"`
	TaskId     int               `json:"task_id"`
	Bin        string            `json:"bin"`
	Args       []string          `json:"args"`
	Configs    map[string]string `json:"configs"`
	PkgAddress string            `json:"pkg_address"`
	PkgName    string            `json:"pkg_name"`
	PkgMD5Sum  string            `json:"pkg_md5sum"`
	PID        int               `json:"pid"`
	Status     string            `json:"status"`
	RootDir    string            `json:"root_dir"`
	Hooks      map[string]string `json:"hooks"`
}

// <agent-root-dir>/<cluster-name>/<job-name>.<task-id>
func (p *Program) getJobRootDir(agentRootDir string) string {
	return path.Join(agentRootDir, p.Name, fmt.Sprintf("%s.%d", p.Job, p.TaskId))
}

// Update <job-root-dir>/pkg link.
func (p *Program) updatePackage(agentRootDir string) error {
	libsDir := path.Join(agentRootDir, LIBRARY_DIR)
	tmpPackageDir := path.Join(libsDir, fmt.Sprintf("%s.tmp", p.PkgMD5Sum))
	md5sumPackageDir := path.Join(libsDir, p.PkgMD5Sum)
	// Step.0 Create <agent-root-dir>/packages/<md5sum>.tmp directory if not exists.
	if _, err := os.Stat(md5sumPackageDir); os.IsNotExist(err) {
		if _, err := os.Stat(tmpPackageDir); err == nil {
			if err := os.RemoveAll(tmpPackageDir); err != nil {
				return err
			}
		}
		if err := os.MkdirAll(tmpPackageDir, 0755); err != nil {
			return err
		}

		// step.1 Download the package
		pkgFilePath := path.Join(tmpPackageDir, p.PkgName)
		resp, err := http.Get(p.PkgAddress)
		if err != nil {
			log.Errorf("Downloading package failed. package : %s, err: %s", p.PkgAddress, err.Error())
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			log.Errorf("Downloading package failed. package : %s, err: %s", p.PkgAddress, resp.Status)
			data, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("%s", string(data))
		}
		out, err := os.Create(pkgFilePath)
		if err != nil {
			log.Errorf("Create package file error: %v", err)
			return err
		}
		defer out.Close()
		io.Copy(out, resp.Body)

		// step.2 Verify md5 checksum
		// TODO reuse those codes with pkgsrv.go
		md5sum, md5Err := calcFileMD5Sum(pkgFilePath)
		if md5Err != nil {
			log.Errorf("Calculate the md5 checksum of file %s failed, cause: %v", pkgFilePath, md5Err)
			return md5Err
		}
		if md5sum != p.PkgMD5Sum {
			return fmt.Errorf("md5sum mismatch, %s != %s, package: %s", md5sum, p.PkgMD5Sum, p.PkgName)
		}

		// step.3 Extract package
		tarCmd := []string{"tar", "xzvf", pkgFilePath, "-C", tmpPackageDir}
		cmd := exec.Command(tarCmd[0], tarCmd[1:]...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); err != nil {
			log.Errorf("exec cmd failed. [cmd: %s], [stdout: %s], [stderr: %s]",
				strings.Join(tarCmd, " "), stdout.String(), stderr.String())
			return err
		}
		if err := os.Rename(tmpPackageDir, md5sumPackageDir); err != nil {
			return err
		}
	}

	// Step.4 Link <job-root-dir>/pkg to <agent-root-dir>/packages/<md5sum>
	files, errs := ioutil.ReadDir(md5sumPackageDir)
	if errs != nil {
		return errs
	}
	for _, f := range files {
		if f.IsDir() {
			linkPkgDir := path.Join(p.getJobRootDir(agentRootDir), PKG_DIR)
			pkgDir := path.Join(md5sumPackageDir, f.Name())
			if _, err := os.Stat(linkPkgDir); err == nil {
				os.RemoveAll(linkPkgDir)
			}
			return os.Symlink(pkgDir, linkPkgDir)
		}
	}
	return fmt.Errorf("Sub-directory under %s does not exist.", md5sumPackageDir)
}

func (p *Program) updateConfigFiles(agentRootDir string) error {
	for fname, content := range p.Configs {
		// When fname is /tmp/huker/agent01/myid case, we should write directly.
		cfgPath := fname
		if !strings.Contains(fname, "/") {
			cfgPath = path.Join(p.getJobRootDir(agentRootDir), CONF_DIR, fname)
		}
		out, err := os.Create(cfgPath)
		if err != nil {
			log.Errorf("save configuration file error: %v", err)
			return err
		}
		defer out.Close()
		io.Copy(out, bytes.NewBufferString(content))
	}
	return nil
}

// Render the agent root directory for config files and arguments.
func (p *Program) renderVars(agentRootDir string) {
	newConfigMap := make(map[string]string)
	for fname, content := range p.Configs {
		content = strings.Replace(content, "$AgentRootDir", agentRootDir, -1)
		content = strings.Replace(content, "$TaskId", strconv.Itoa(p.TaskId), -1)
		fname = strings.Replace(fname, "$AgentRootDir", agentRootDir, -1)
		fname = strings.Replace(fname, "$TaskId", strconv.Itoa(p.TaskId), -1)
		newConfigMap[fname] = content
	}
	p.Configs = newConfigMap

	for idx, arg := range p.Args {
		arg = strings.Replace(arg, "$AgentRootDir", agentRootDir, -1)
		arg = strings.Replace(arg, "$TaskId", strconv.Itoa(p.TaskId), -1)
		p.Args[idx] = arg
	}
	p.RootDir = p.getJobRootDir(agentRootDir)
}

func (p *Program) Install(agentRootDir string) error {
	jobRootDir := p.getJobRootDir(agentRootDir)

	// step.0 Prev-check
	if relDir, err := filepath.Rel(agentRootDir, jobRootDir); err != nil {
		return err
	} else if strings.Contains(relDir, "..") || agentRootDir == jobRootDir {
		return fmt.Errorf("Permission denied, mkdir %s", jobRootDir)
	}
	if _, err := os.Stat(jobRootDir); err == nil {
		return fmt.Errorf("%s already exists, cleanup it first please.", jobRootDir)
	}

	// step.1 Create directories recursively
	for _, sub := range progDirs() {
		if err := os.MkdirAll(path.Join(jobRootDir, sub), 0755); err != nil {
			return err
		}
	}

	// step.2 Download package and link pkg to library.
	if err := p.updatePackage(agentRootDir); err != nil {
		return err
	}

	// step.3 Dump configuration files
	return p.updateConfigFiles(agentRootDir)
}

// TODO pipe stdout & stderr into pkg_root_dir/stdout directories.
func (p *Program) Start(s *Supervisor) error {
	if isProcessOK(p.PID) {
		return fmt.Errorf("Process %d is already running.", p.PID)
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(p.Bin, p.Args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
		Pgid:   0,
	}
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	go func() {
		if err := cmd.Start(); err != nil {
			log.Errorf("Start job failed. [cmd: %s %s], [stdout: %s], [stderr: %s], err: %v",
				p.Bin, strings.Join(p.Args, " "), stdout.String(), stderr.String(), err)
		}
		if err := cmd.Wait(); err != nil {
			log.Errorf("Wait job failed. [cmd: %s %s], [stdout: %s], [stderr: %s], err: %v",
				p.Bin, strings.Join(p.Args, " "), stdout.String(), stderr.String(), err)
		}
	}()
	time.Sleep(time.Second * 1)

	if cmd.Process != nil && isProcessOK(cmd.Process.Pid) {
		log.Infof("Start process success. [%s %s]", p.Bin, strings.Join(p.Args, " "))
		p.Status = StatusRunning
		p.PID = cmd.Process.Pid
		return nil
	} else {
		return fmt.Errorf("Start job failed.")
	}
}

func (p *Program) Stop(s *Supervisor) error {
	process, err := os.FindProcess(p.PID)
	if err != nil {
		return err
	}
	err = process.Kill()
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	// check the pid in the final
	if isProcessOK(p.PID) {
		return fmt.Errorf("Failed to stop the process %d, still running.", p.PID)
	}
	p.Status = StatusStopped
	return nil
}

func (p *Program) Restart(s *Supervisor) error {
	p.Stop(s)
	if isProcessOK(p.PID) {
		// TODO check process status
		return fmt.Errorf("Failed to stop the process %d, still running.", p.PID)
	}
	return p.Start(s)
}

func (p *Program) hookEnv() []string {
	var env []string
	env = append(env, "PROGRAM_BIN="+p.Bin)
	env = append(env, "PROGRAM_ARGS="+strings.Join(p.Args, " "))
	env = append(env, "PROGRAM_DIR="+p.RootDir)
	env = append(env, os.Environ()...)
	return env
}

func (p *Program) ExecHooks(hook string) error {
	if _, ok := p.Hooks[hook]; !ok {
		return nil
	}
	hooksDir := path.Join(p.RootDir, HOOKS_DIR)
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		if err := os.MkdirAll(hooksDir, 0755); err != nil {
			return err
		}
	} else {
		return err
	}
	hookFile := path.Join(hooksDir, hook)
	if err := ioutil.WriteFile(hookFile, []byte(p.Hooks[hook]), 0744); err != nil {
		return err
	}
	// Execute the hooked bash script.
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("/bin/bash", hookFile)
	cmd.Env = p.hookEnv()
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	log.Infof("Hook %s, Environment variables:\n%s", hook, strings.Join(cmd.Env, "\n"))
	if err := cmd.Run(); err != nil {
		log.Errorf("Execute hook failed. [cmd: /bin/bash %s], [stdout: %s], [stderr: %s]",
			hookFile, stdout.String(), stderr.String())
		return err
	}
	return nil
}