package daemon

import (
	"os"
)

const (
	MARK_NAME  = "_GO_DAEMON"
	MARK_VALUE = "_GO_ENABLED"
	MARK       = MARK_NAME + "=" + MARK_VALUE

	DEFAULT_FILE_PERM     = os.FileMode(0640)
	DEFAULT_PID_FILE_NAME = "pid"
	DEFAULT_LOG_FILE_NAME = "log"
)

type Config struct {
	Env  []string
	Args []string

	Umask   int
	Chroot  string
	WorkDir string

	PidFileName string
	PidFilePerm os.FileMode

	LogFileName string
	LogFilePerm os.FileMode
}

func (config Config) Create() *Daemon {

	if len(config.Args) == 0 {
		config.Args = os.Args
	}

	if len(config.Env) == 0 {
		config.Env = os.Environ()
	}
	config.Env = append(config.Env, MARK)

	if config.PidFilePerm == 0 {
		config.PidFilePerm = DEFAULT_FILE_PERM
	}
	if config.LogFilePerm == 0 {
		config.LogFilePerm = DEFAULT_FILE_PERM
	}

	if len(config.PidFileName) == 0 {
		config.PidFileName = DEFAULT_PID_FILE_NAME
	}
	if len(config.LogFileName) == 0 {
		config.LogFileName = DEFAULT_LOG_FILE_NAME
	}

	return &Daemon{
		config: config,

		services: make([]Service, 0),

		cmds:  make(map[Flag]os.Signal),
		funcs: make(map[os.Signal]HandlerFunc),

		restartSignals:   make([]os.Signal, 0),
		shutdownSignals:  make([]os.Signal, 0),
		terminateSignals: make([]os.Signal, 0),
	}
}

func New(config Config) *Daemon {
	return config.Create()
}
