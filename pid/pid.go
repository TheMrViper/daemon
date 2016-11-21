package pid

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

var (
	ErrWouldBlock = errors.New("Pid file blocked")
)

type File struct {
	*os.File
}

func New(f *os.File) *File {
	return &File{f}
}

func OpenFile(name string, flag int, perm os.FileMode) (lock *File, err error) {
	file, err := os.OpenFile(name, flag, perm)

	return &File{file}, err
}

func (f *File) Lock() (err error) {
	return lockFile(f.Fd())
}

func (f *File) Unlock() (err error) {
	return unlockFile(f.Fd())
}

func (f *File) SetPid() (err error) {
	if _, err = f.Seek(0, os.SEEK_SET); err != nil {
		return
	}
	var fileLen int
	if fileLen, err = fmt.Fprint(f, os.Getpid()); err != nil {
		return
	}
	if err = f.Truncate(int64(fileLen)); err != nil {
		return
	}
	return f.Sync()
}

func (f *File) GetPid() (pid int, err error) {
	if _, err = f.Seek(0, os.SEEK_SET); err != nil {
		return
	}
	_, err = fmt.Fscan(f, &pid)
	return
}

func (f *File) Remove() (err error) {
	defer f.Close()

	if err := f.Unlock(); err != nil {
		return err
	}

	// TODO(yar): keep filename?
	var name string
	if name, err = getFdName(f.Fd()); err != nil {
		return err
	}

	return syscall.Unlink(name)
}
