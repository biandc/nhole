package tools

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/biandc/nhole/pkg/log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/google/uuid"
)

const (
	UuidLen = 36
)

var (
	bufPool16k sync.Pool
	bufPool5k  sync.Pool
	bufPool2k  sync.Pool
	bufPool1k  sync.Pool
	bufPool    sync.Pool
)

func GetBuf(size int) []byte {
	var x interface{}
	if size >= 16*1024 {
		x = bufPool16k.Get()
	} else if size >= 5*1024 {
		x = bufPool5k.Get()
	} else if size >= 2*1024 {
		x = bufPool2k.Get()
	} else if size >= 1*1024 {
		x = bufPool1k.Get()
	} else {
		x = bufPool.Get()
	}
	if x == nil {
		return make([]byte, size)
	}
	buf := *x.(*[]byte)
	if cap(buf) < size {
		return make([]byte, size)
	}
	return buf[:size]
}

func PutBuf(buf []byte) {
	size := cap(buf)
	if size >= 16*1024 {
		bufPool16k.Put(&buf)
	} else if size >= 5*1024 {
		bufPool5k.Put(&buf)
	} else if size >= 2*1024 {
		bufPool2k.Put(&buf)
	} else if size >= 1*1024 {
		bufPool1k.Put(&buf)
	} else {
		bufPool.Put(&buf)
	}
}

func ValidateIp(ip string) (err error) {
	ret := net.ParseIP(ip)
	if ret == nil {
		err = fmt.Errorf("%s ValidateIp error", ip)
	}
	return
}

func ValidatePort(port int) (err error) {
	if port < 0 || port > 65535 {
		err = fmt.Errorf("%d ValidatePort error", port)
	}
	return
}

func ValidateUUID(uuidStr string) (err error) {
	if len(uuidStr) != UuidLen {
		err = fmt.Errorf("%s ValidateUUID error", uuidStr)
	}
	return
}

func GenerateUUID() (uuidStr string) {
	uuidStr = uuid.NewString()
	return
}

func Uint322Bytes(n int) (bs []byte) {
	data := uint32(n)
	buf := bytes.NewBuffer([]byte{})
	_ = binary.Write(buf, binary.BigEndian, data)
	bs = buf.Bytes()
	return
}

func Bytes2Uint32(bs []byte) (n int) {
	buf := bytes.NewBuffer(bs)
	var data uint32
	_ = binary.Read(buf, binary.BigEndian, &data)
	n = int(data)
	return
}

type Releaser interface {
	Release()
}

func ExitClear(r Releaser, exitInfo string) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(
		signalCh,
		os.Interrupt,
		//syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGKILL,
		syscall.SIGTERM,
	)
	_ = <-signalCh
	log.Info(exitInfo)
	r.Release()
}
