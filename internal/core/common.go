package core

import (
	"io"
	"syscall"
)

var _ io.ReadWriter = &FDCommon{}

type FDCommon struct {
	FD int64
}

func (fd *FDCommon) Read(p []byte) (n int, err error) {
	return syscall.Read(int(fd.FD), p)
}

func (fd *FDCommon) Write(p []byte) (n int, err error) {
	return syscall.Write(int(fd.FD), p)
}
