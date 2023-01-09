package core

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

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

func Decode2MsgCh(reader io.Reader) (msgCh chan *message.Message) {
	msgCh = make(chan *message.Message)
	go func() {
		defer close(msgCh)
		for {
			header := tools.GetBuf(PackageHeadLen)
			n, err := io.ReadFull(reader, header)
			if err != nil || n != len(header) {
				tools.PutBuf(header)
				return
			}
			dataLen := tools.Bytes2Uint32(header)
			tools.PutBuf(header)
			data := tools.GetBuf(dataLen)
			n, err = io.ReadFull(reader, data)
			if err != nil || n != len(data) {
				tools.PutBuf(data)
				return
			}
			msg, err := message.UnmarshalMessage(data)
			tools.PutBuf(data)
			if err != nil {
				continue
			}
			msgCh <- msg
		}
	}()
	return
}

// Deprecated
// *Bufio.Reader cannot return io timeout(error).
//func Decode2MsgCh2(rwc io.Reader) (msgCh chan *message.Message) {
//	msgCh = make(chan *message.Message)
//	reader, ok := rwc.(*bufio.Reader)
//	switch ok {
//	// rwc.reader is *bufio.Reader
//	case true:
//		go func() {
//			var buf []byte
//			for {
//				header, err := reader.Peek(PackageHeadLen)
//				if err != nil {
//					if err == io.EOF || strings.Contains(err.Error(), "use of closed network connection") {
//						goto closeMsg
//					}
//					continue
//				}
//				n := tools.Bytes2Uint32(header)
//				packageLen := PackageHeadLen + n
//				if reader.Buffered() < packageLen {
//					goto closeMsg
//				}
//				buf = tools.GetBuf(packageLen)
//				_, err = reader.Read(buf)
//				if err != nil {
//					tools.PutBuf(buf)
//					goto closeMsg
//				}
//				msg, err := message.UnmarshalMessage(buf[PackageHeadLen:])
//				tools.PutBuf(buf)
//				if err != nil {
//					continue
//				}
//				msgCh <- msg
//			}
//		closeMsg:
//			close(msgCh)
//		}()
//	case false:
//		// rwc.reader not is *bufio.Reader
//		close(msgCh)
//	}
//	return
//}

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
	dataBytes = make([]byte, msgBytesLen+PackageHeadLen)
	header := tools.Uint322Bytes(msgBytesLen)
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
	readTimeout time.Duration
	net.Conn
	closeFn func() (err error)
	closed  bool
	sync.RWMutex
}

func WrapConner(conn net.Conn, readTimeout time.Duration, closeFn func() (err error)) (conner *Conn) {
	conner = &Conn{
		readTimeout: readTimeout,
		Conn:        conn,
		closeFn:     closeFn,
		closed:      false,
	}
	return
}

func (c *Conn) Read(b []byte) (n int, err error) {
	if c.readTimeout > 0 {
		err = c.SetReadDeadline(time.Now().Add(c.readTimeout))
		if err != nil {
			return
		}
	}
	return c.Conn.Read(b)
}

func (c *Conn) Close() (err error) {
	c.Lock()
	defer c.Unlock()
	if c.closed {
		return
	}
	if err = c.Conn.Close(); err != nil {
		return
	}
	if c.closeFn != nil {
		if err = c.closeFn(); err != nil {
			return
		}
	}
	c.closed = true
	return
}

func (c *Conn) SetReadTimeout(readTimeout time.Duration) {
	c.readTimeout = readTimeout
}

func (c *Conn) SetCloseFn(closeFn func() (err error)) {
	c.Lock()
	defer c.Unlock()
	c.closeFn = closeFn
}
