package control

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/biandc/nhole/pkg/core"
	"github.com/biandc/nhole/pkg/log"
	"github.com/biandc/nhole/pkg/message"
)

type ForwardServ struct {
	ip   string
	port int

	clientID string
	serverID string

	net.Listener

	ctx    context.Context
	logger *log.Logger

	connCh     chan net.Conn
	createConn func(clientID, fserverID, forwardID string)

	record map[string]net.Conn
	sync.RWMutex
}

func NewForwardServer(
	ctx context.Context,
	ip string,
	port int,
	clientID, serverID string,
	createConn func(clientID, fserverID, forwardID string),
) (f *ForwardServ, err error) {
	var listener net.Listener
	listener, err = core.NewListener(ip, port)
	if err != nil {
		return
	}
	newCtx := ctx
	f = &ForwardServ{
		ip:   ip,
		port: port,

		clientID: clientID,
		serverID: serverID,

		Listener: listener,

		ctx:    newCtx,
		logger: log.FromContextSafe(ctx),

		connCh:     make(chan net.Conn, 100),
		createConn: createConn,

		record: make(map[string]net.Conn, 0),
	}
	f.logger.AppendPrefix(f.Addr().String())
	return
}

func (f *ForwardServ) Run() {
	go f.accept()
	go f.HandleConn()
}

func (f *ForwardServ) accept() {
	defer close(f.connCh)
	for {
		conn, err := f.Accept()
		if err != nil {
			f.logger.Warn(err.Error())
			return
		}
		f.connCh <- conn
	}
}

func (f *ForwardServ) handleConn(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	f.logger.Info("Connection from %s", addr)
	f.Add(addr, conn)
	f.createConn(f.clientID, strconv.Itoa(f.port), addr)
}

func (f *ForwardServ) HandleConn() {
	defer func() {
		err := f.Close()
		if err != nil {
			f.logger.Warn(err.Error())
		} else {
			f.logger.Info("Close.")
		}
	}()
	for conn := range f.connCh {
		go f.handleConn(conn)
	}
}

func (f *ForwardServ) Get(fID string) (fclient net.Conn, err error) {
	var ok bool
	f.RLock()
	defer f.RUnlock()
	fclient, ok = f.record[fID]
	if !ok {
		err = fmt.Errorf("forwardServer %s not has %s", f.Addr().String(), fID)
	}
	return
}

func (f *ForwardServ) Add(fID string, fclient net.Conn) {
	f.Lock()
	defer f.Unlock()
	f.record[fID] = fclient
}

func (f *ForwardServ) clear() {
	f.Lock()
	defer f.Unlock()
	for _, conn := range f.record {
		_ = conn.Close()
	}
	f.record = nil
}

func (f *ForwardServ) Close() (err error) {
	f.clear()
	err = f.Listener.Close()
	if err != nil {
		return
	}
	return
}

type ForwardClient struct {
	clientID  string
	serverID  string
	forwardID string

	localIp     string
	localPort   int
	controlIp   string
	controlPort int

	localConn   net.Conn
	controlConn net.Conn
}

func NewForwardClienter(
	localIp string,
	localPort int,
	cIp string,
	cPort int,
	serverID, forwardID string,
) (f *ForwardClient, err error) {
	var (
		localConn   net.Conn
		controlConn net.Conn
	)
	defer func() {
		if err != nil {
			if localConn != nil {
				_ = localConn.Close()
			}
			if controlConn != nil {
				_ = controlConn.Close()
			}
		}
	}()
	localConn, err = core.NewConner(localIp, localPort)
	if err != nil {
		return
	}
	controlConn, err = core.NewConner(cIp, cPort)
	if err != nil {
		return
	}

	f = &ForwardClient{
		clientID:  "",
		serverID:  serverID,
		forwardID: forwardID,

		localIp:     localIp,
		localPort:   localPort,
		controlIp:   cIp,
		controlPort: cPort,

		localConn:   localConn,
		controlConn: core.WrapConner(controlConn, 0, nil),
	}
	err = f.register()
	if err != nil {
		return
	}
	err = f.handleRegister()
	if err != nil {
		return
	}
	return
}

func (f *ForwardClient) Run() {
	f.forward()
}

func (f *ForwardClient) register() (err error) {
	var (
		msgBytes []byte
	)
	msgBytes, _, err = core.EncodeOneMsg("", message.ForwardConn, message.REGISTER, 0, "", "")
	if err != nil {
		return
	}
	_, err = f.controlConn.Write(msgBytes)
	if err != nil {
		return
	}
	return
}

func (f *ForwardClient) handleRegister() (err error) {
	var (
		msg *message.Message
	)
	for {
		msg, err = core.DecodeOneMsg(f.controlConn)
		if err != nil {
			return
		}
		if msg.Operation == message.REGISTER {
			break
		}
	}
	f.clientID = msg.ClientID
	err = f.sendCreateConn()
	return
}

func (f *ForwardClient) sendCreateConn() (err error) {
	var (
		data     string
		msgBytes []byte
	)
	data, err = message.MarshalCreateConnData(f.serverID, f.forwardID)
	if err != nil {
		return
	}
	msgBytes, _, err = core.EncodeOneMsg(f.clientID, message.ForwardConn, message.CreateForwardConn, 0, "", data)
	if err != nil {
		return
	}
	_, err = f.controlConn.Write(msgBytes)
	return
}

func (f *ForwardClient) forward() {
	core.Forward(f.localConn, f.controlConn)
}
