package daemon

import (
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheMrViper/gdaemon/pid"
	"github.com/kardianos/osext"
)

var (
	StopedSucceful = errors.New("Daemon stoped")
)

func (d *Daemon) search() (daemon *os.Process, err error) {
	pidFile, err := pid.OpenFile(d.config.PidFileName, os.O_RDONLY, 0640)
	if err != nil {
		return nil, err
	}
	defer pidFile.Close()

	pid, err := pidFile.GetPid()
	if err != nil {
		return nil, err
	}
	return os.FindProcess(pid)
}

var initialized = false

func (d *Daemon) kill() (err error) {
	if initialized == false {
		return nil
	}
	return d.pidFile.Remove()
}

func (d *Daemon) work() (err error) {
	if initialized {
		return os.ErrInvalid
	}
	initialized = true

	decoder := json.NewDecoder(os.Stdin)
	if err = decoder.Decode(d); err != nil {
		return
	}

	if err = syscall.Close(0); err != nil {
		return
	}
	if err = syscall.Dup2(3, 0); err != nil {
		return
	}

	d.pidFile = pid.New(os.NewFile(4, d.config.PidFileName))
	if err = d.pidFile.SetPid(); err != nil {
		return err
	}

	if d.config.Umask != 0 {
		syscall.Umask(int(d.config.Umask))
	}
	if len(d.config.Chroot) > 0 {
		err = syscall.Chroot(d.config.Chroot)
	}

	return
}
func (d *Daemon) toSend() bool {
	for flag := range d.cmds {
		if flag.IsSet() {
			return true
		}
	}
	return false
}
func (d *Daemon) fork() (err error) {
	if err := d.prepareEnv(); err != nil {
		return err
	}

	defer d.closeFiles()
	if err := d.openFiles(); err != nil {
		return err
	}

	attr := &os.ProcAttr{
		Dir:   d.config.WorkDir,
		Env:   d.config.Env,
		Files: d.files(),
		Sys: &syscall.SysProcAttr{
			Setsid: true,
		},
	}

	if _, err = os.StartProcess(d.abspath, d.config.Args, attr); err != nil {
		d.pidFile.Remove()
		return
	}

	d.rpipe.Close()
	encoder := json.NewEncoder(d.wpipe)
	return encoder.Encode(d)
}

func (d *Daemon) files() []*os.File {
	return []*os.File{
		d.rpipe,        // (0) stdin
		d.logFile,      // (1) stdout
		d.logFile,      // (2) stderr
		d.nullFile,     // (3) dup on fd 0 after initialization
		d.pidFile.File, // (4) pid file
	}
}

func (d *Daemon) closeFiles() (err error) {
	cl := func(file **os.File) {
		if *file != nil {
			(*file).Close()
			*file = nil
		}
	}
	cl(&d.rpipe)
	cl(&d.wpipe)
	cl(&d.logFile)
	cl(&d.nullFile)
	if d.pidFile != nil {
		d.pidFile.Close()
		d.pidFile = nil
	}
	return
}

func (d *Daemon) openFiles() (err error) {
	if d.nullFile, err = os.Open(os.DevNull); err != nil {
		return
	}

	if d.pidFile, err = pid.OpenFile(d.config.PidFileName, os.O_RDWR|os.O_CREATE, d.config.PidFilePerm); err != nil {
		return
	}
	if err = d.pidFile.Lock(); err != nil {
		return
	}

	if d.logFile, err = os.OpenFile(d.config.LogFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, d.config.LogFilePerm); err != nil {
		return
	}

	d.rpipe, d.wpipe, err = os.Pipe()
	return
}

func (d *Daemon) prepareEnv() (err error) {
	if d.abspath, err = osext.Executable(); err != nil {
		return
	}
	return
}

func (d *Daemon) recvSignals() (err error) {
	unique := make(map[os.Signal]struct{})

	for k := range d.funcs {
		unique[k] = struct{}{}
	}
	for _, v := range d.restartSignals {
		unique[v] = struct{}{}
	}
	for _, v := range d.shutdownSignals {
		unique[v] = struct{}{}
	}
	for _, v := range d.terminateSignals {
		unique[v] = struct{}{}
	}

	signals := make([]os.Signal, 0, len(unique))

	for k := range unique {
		signals = append(signals, k)
	}

	ch := make(chan os.Signal, len(signals))
	signal.Notify(ch, signals...)
	defer signal.Stop(ch)

	for sig := range ch {
		if err = d.funcs[sig](sig); err != nil {
			return
		}

		for _, v := range d.restartSignals {
			if v == sig {
				for _, s := range d.services {
					if err = s.Shutdown(); err != nil {
						return
					}
					if err = s.Start(); err != nil {
						return
					}
				}
			}
		}
		for _, v := range d.shutdownSignals {
			if v == sig {
				for _, s := range d.services {
					if err = s.Shutdown(); err != nil {
						return
					}
				}
			}
		}
		for _, v := range d.terminateSignals {
			if v == sig {
				for _, s := range d.services {
					if err = s.Terminate(); err != nil {
						return
					}
				}
			}
		}
	}

	return err
}

func (d *Daemon) sendSignals(child *os.Process) (err error) {
	for flag, sig := range d.cmds {
		if flag.IsSet() {
			if err = child.Signal(sig); err != nil {
				return
			}
		}
		continue
	}
	return
}
