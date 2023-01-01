package core

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/biandc/nhole/pkg/core/tcp"
	"github.com/biandc/nhole/pkg/message"
	"github.com/biandc/nhole/pkg/tools"
)

const (
	PackageHeadLen = message.PackageHeadLen
)

func NewListener(ip string, port int) (listener net.Listener, err error) {
	return tcp.NewTcpListener(ip, port)
}

func NewConner(ip string, port int) (conn net.Conn, err error) {
	return tcp.NewConner(ip, port)
}

type ReadWriteClose struct {
	reader    io.Reader
	writer    io.Writer
	closeFunc func() (err error)
	closed    bool
	sync.RWMutex
}

func NewReadWriteCloser(reader io.Reader, writer io.Writer, closeFunc func() (err error)) (rwc *ReadWriteClose) {
	rwc = &ReadWriteClose{
		reader:    reader,
		writer:    writer,
		closeFunc: closeFunc,
		closed:    false,
	}
	return
}

func (rwc *ReadWriteClose) Read(b []byte) (n int, err error) {
	return rwc.reader.Read(b)
}

func (rwc *ReadWriteClose) Write(b []byte) (n int, err error) {
	return rwc.writer.Write(b)
}

func (rwc *ReadWriteClose) Close() (errRet error) {
	rwc.Lock()
	if rwc.closed {
		rwc.Unlock()
		return
	}
	rwc.closed = true
	closeFunc := rwc.closeFunc
	rwc.Unlock()

	var err error
	if rc, ok := rwc.reader.(io.Closer); ok {
		err = rc.Close()
		if err != nil {
			errRet = err
		}
	}

	if wc, ok := rwc.writer.(io.Closer); ok {
		err = wc.Close()
		if err != nil {
			errRet = err
		}
	}

	if closeFunc != nil {
		err = closeFunc()
		if err != nil {
			errRet = err
		}
	}
	return
}

func (rwc *ReadWriteClose) SetCloseFunc(closeFunc func() (err error)) {
	rwc.RLock()
	defer rwc.RUnlock()
	if !rwc.closed {
		rwc.closeFunc = closeFunc
	}
	return
}

func DecodeOneMsg(reader io.Reader) (msg *message.Message, err error) {
	var n int
	header := tools.GetBuf(PackageHeadLen)
	defer tools.PutBuf(header)
	n, err = io.ReadFull(reader, header)
	if err != nil {
		return
	}
	if n != len(header) {
		err = fmt.Errorf("bad header %s", header)
		return
	}
	dataLen := tools.Bytes2Uint32(header)
	data := tools.GetBuf(dataLen)
	defer tools.PutBuf(data)
	n, err = io.ReadFull(reader, data)
	if err != nil {
		return
	}
	if n != len(data) {
		err = fmt.Errorf("bad data body %s", data)
		return
	}
	msg, err = message.UnmarshalMessage(data)
	if err != nil {
		return
	}
	return

}

func Decode2MsgCh(rwc io.Reader) (msgCh chan *message.Message) {
	msgCh = make(chan *message.Message)
	reader, ok := rwc.(*bufio.Reader)
	switch ok {
	// rwc.reader is *bufio.Reader
	case true:
		go func() {
			var buf []byte
			for {
				header, err := reader.Peek(PackageHeadLen)
				if err != nil {
					if err == io.EOF || strings.Contains(err.Error(), "use of closed network connection") {
						goto closeMsg
					}
					continue
				}
				n := tools.Bytes2Uint32(header)
				packageLen := PackageHeadLen + n
				if reader.Buffered() < packageLen {
					goto closeMsg
				}
				buf = tools.GetBuf(packageLen)
				_, err = reader.Read(buf)
				if err != nil {
					tools.PutBuf(buf)
					goto closeMsg
				}
				msg, err := message.UnmarshalMessage(buf[PackageHeadLen:])
				tools.PutBuf(buf)
				if err != nil {
					continue
				}
				msgCh <- msg
			}
		closeMsg:
			close(msgCh)
		}()
	case false:
		// TODO: rwc.reader not is *bufio.Reader
		close(msgCh)
	}
	return
}

func EncodeOneMsg(
	uuid, connType, operation string,
	errInt int,
	errInfo, data string,
) (dataBytes []byte, msg *message.Message, err error) {
	var msgBytes []byte
	msg = message.NewMessage(uuid, connType, operation, errInt, errInfo, data)
	msgBytes, err = message.MarshalMessage(msg)
	if err != nil {
		return
	}
	msgBytesLen := len(msgBytes)
	//buf := tools.GetBuf(msgBytesLen + PackageHeadLen)
	//defer tools.PutBuf(buf)
	dataBytes = make([]byte, msgBytesLen+PackageHeadLen)
	header := tools.Uint322Bytes(msgBytesLen)
	//copy(buf[:PackageHeadLen], header)
	//copy(buf[PackageHeadLen:], msgBytes)
	copy(dataBytes[:PackageHeadLen], header)
	copy(dataBytes[PackageHeadLen:], msgBytes)
	return
}

func Forward(clienter1, clienter2 io.ReadWriteCloser) {
	var forward = func(from, to io.ReadWriteCloser) {
		buf := tools.GetBuf(16 * 1024)
		defer tools.PutBuf(buf)
		defer func() {
			_ = from.Close()
			_ = to.Close()
		}()
		_, _ = io.CopyBuffer(to, from, buf)
	}
	go forward(clienter1, clienter2)
	go forward(clienter2, clienter1)
}

type Conn struct {
	*ReadWriteClose
	net.Conn
}

func NewCoreConner(rwc *ReadWriteClose, conn net.Conn) (c *Conn) {
	c = &Conn{
		ReadWriteClose: rwc,
		Conn:           conn,
	}
	return
}

func (c *Conn) Read(b []byte) (n int, err error) {
	if c.ReadWriteClose == nil {
		n, err = c.Conn.Read(b)
	} else {
		n, err = c.ReadWriteClose.Read(b)
	}
	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	if c.ReadWriteClose == nil {
		n, err = c.Conn.Write(b)
	} else {
		n, err = c.ReadWriteClose.Write(b)
	}
	return
}

func (c *Conn) Close() (err error) {
	if c.ReadWriteClose == nil {
		err = c.Conn.Close()
	} else {
		err = c.ReadWriteClose.Close()
	}
	return
}
