package control

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/biandc/nhole/pkg/core"
	"github.com/biandc/nhole/pkg/log"
	"github.com/biandc/nhole/pkg/message"
	"github.com/biandc/nhole/pkg/tools"
)

type ClientRecord struct {
	clientMap map[string]net.Conn
	sync.RWMutex
}

func NewClientRecord() (c *ClientRecord) {
	c = &ClientRecord{
		clientMap: make(map[string]net.Conn, 0),
	}
	return
}

func (c *ClientRecord) Get(clientID string) (clienter net.Conn, err error) {
	var ok bool
	c.RLock()
	defer c.RUnlock()
	clienter, ok = c.clientMap[clientID]
	if !ok {
		err = fmt.Errorf("clientRecord not has %s", clientID)
	}
	return
}

func (c *ClientRecord) Add(clientID string, clienter net.Conn) {
	c.Lock()
	defer c.Unlock()
	c.clientMap[clientID] = clienter
}

func (c *ClientRecord) Del(clientID string) {
	c.Lock()
	defer c.Unlock()
	if clienter, ok := c.clientMap[clientID]; ok {
		_ = clienter.Close()
		delete(c.clientMap, clientID)
	}
}

type ControlRecord struct {
	clientServer map[string][]string
	serverMap    map[string]*ForwardServ
	sync.RWMutex
}

func NewControlRecord() (c *ControlRecord) {
	c = &ControlRecord{
		clientServer: make(map[string][]string, 0),
		serverMap:    make(map[string]*ForwardServ, 0),
	}
	return
}

func (c *ControlRecord) GetByServerID(serverID string) (server *ForwardServ, err error) {
	var ok bool
	c.RLock()
	defer c.RUnlock()
	server, ok = c.serverMap[serverID]
	if !ok {
		err = fmt.Errorf("serverID %s not find in ControlRecord", serverID)
	}
	return
}

func (c *ControlRecord) Add(clientID, serverID string, server *ForwardServ) {
	c.Lock()
	defer c.Unlock()
	if c.clientServer == nil || c.serverMap == nil {
		return
	}
	if _, ok := c.clientServer[clientID]; !ok {
		c.clientServer[clientID] = make([]string, 0, 1)
	}
	c.clientServer[clientID] = append(c.clientServer[clientID], serverID)
	c.serverMap[serverID] = server
}

func (c *ControlRecord) Del(clientID string) {
	c.Lock()
	defer c.Unlock()
	if serverIDs, ok := c.clientServer[clientID]; ok {
		for _, serverID := range serverIDs {
			if server, ok := c.serverMap[serverID]; ok {
				go func(server io.Closer) {
					_ = server.Close()
				}(server)
				delete(c.serverMap, serverID)
			}
		}
		delete(c.clientServer, clientID)
	}
	return
}

func (c *ControlRecord) Clear() {
	c.Lock()
	defer c.Unlock()
	for _, value := range c.serverMap {
		value.Release()
	}
	c.clientServer = nil
	c.serverMap = nil
}

type ControlServ struct {
	ip            string
	port          int
	controlRecord *ControlRecord
	clientRecord  *ClientRecord
	net.Listener
	ctx        context.Context
	cancelFunc context.CancelFunc
	logger     *log.Logger
	connCh     chan net.Conn
}

func NewControlServer(ctx context.Context, ip string, port int) (c *ControlServ, err error) {
	var listener net.Listener
	listener, err = core.NewListener(ip, port)
	if err != nil {
		return
	}
	newCtx, cancel := context.WithCancel(ctx)
	c = &ControlServ{
		ip:            ip,
		port:          port,
		controlRecord: NewControlRecord(),
		clientRecord:  NewClientRecord(),
		Listener:      listener,
		ctx:           newCtx,
		cancelFunc:    cancel,
		logger:        log.FromContextSafe(newCtx),
		connCh:        make(chan net.Conn, 100),
	}
	c.logger.AppendPrefix(c.Addr().String())
	return
}

func (c *ControlServ) Run() {
	go c.accept()
	go c.HandleConn()
}

func (c *ControlServ) accept() {
	defer close(c.connCh)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			conn, err := c.Accept()
			if err != nil {
				c.logger.Info(err.Error())
				return
			}
			c.connCh <- conn
		}
	}
}

func (c *ControlServ) handleRegister(conn net.Conn, msg *message.Message) {
	var (
		msgBytes []byte
		msgRes   *message.Message
		err      error
	)
	defer func() {
		if err != nil {
			c.logger.Error(err.Error())
		} else {
			addr := conn.RemoteAddr().String()
			c.logger.Info("register %s %s", addr, msgRes.String())
		}
	}()
	clientID := tools.GenerateUUID()
	msgBytes, msgRes, err = core.EncodeOneMsg(clientID, msg.ConnType, msg.Operation, 0, "", "")
	if err != nil {
		return
	}
	for {
		_, err = conn.Write(msgBytes)
		if err == nil || strings.Contains(err.Error(), "use of closed network connection") {
			err = nil
			break
		}
	}
	if msg.ConnType == message.ControlConn {
		c.clientRecord.Add(clientID, conn)
		if coreConn, ok := conn.(*core.Conn); ok {
			coreConn.SetCloseFunc(func() (err error) {
				c.clientRecord.Del(clientID)
				c.controlRecord.Del(clientID)
				return
			})
		}
	}
}

func (c *ControlServ) createConn(clientID, fserverID, forwardID string) {
	var (
		data     string
		clienter net.Conn
		msgBytes []byte
		err      error
	)
	defer func() {
		if err != nil {
			c.logger.Error(err.Error())
		}
	}()
	data, err = message.MarshalCreateConnData(fserverID, forwardID)
	if err != nil {
		return
	}
	clienter, err = c.clientRecord.Get(clientID)
	if err != nil {
		return
	}
	msgBytes, _, err = core.EncodeOneMsg(clientID, message.ControlConn, message.CreateForwardConn, 0, "", data)
	if err != nil {
		return
	}
	for {
		_, err = clienter.Write(msgBytes)
		if err == nil || strings.Contains(err.Error(), "use of closed network connection") {
			err = nil
			break
		}
	}
}

func (c *ControlServ) handleCreateConn(conn net.Conn, msg *message.Message) {
	if msg.ConnType != message.ForwardConn {
		log.Error("handleCreateConn msg.ConnType not is %s", message.ForwardConn)
		return
	}
	var (
		data    *message.CreateConnData
		fserver *ForwardServ
		fclient net.Conn
		err     error
	)
	defer func() {
		if err != nil {
			c.logger.Error(err.Error())
		} else {
			addr := conn.RemoteAddr().String()
			addr2 := fclient.RemoteAddr().String()
			c.logger.Info("forwarding %s %s ...", addr, addr2)
		}
	}()
	data, err = message.UnmarshalCreateConnData(msg.Data)
	if err != nil {
		return
	}
	fserver, err = c.controlRecord.GetByServerID(data.ServerID)
	if err != nil {
		return
	}
	fclient, err = fserver.Get(data.ForwardID)
	if err != nil {
		return
	}
	core.Forward(fclient, conn)
}

func (c *ControlServ) handleCreateServer(conn net.Conn, msg *message.Message) {
	if msg.ConnType != message.ControlConn {
		log.Error("handleCreateServer msg.ConnType not is %s", message.ControlConn)
		return
	}
	var (
		port     int
		fserver  *ForwardServ
		msgBytes []byte
		errInt   = 0
		err      error
	)
	defer func() {
		if err != nil {
			c.logger.Error(err.Error())
		} else {
			c.logger.Info("create forward server %s:%d %s", c.ip, port, fserver.serverID)
		}
	}()
	defer func() {
		errInfo := ""
		if err != nil {
			errInfo = err.Error()
		}
		msgBytes, _, _ = core.EncodeOneMsg(
			msg.ClientID,
			message.ControlConn,
			message.CreateForwardServer,
			errInt,
			errInfo,
			msg.Data,
		)
		for {
			_, writeErr := conn.Write(msgBytes)
			if writeErr == nil || strings.Contains(writeErr.Error(), "use of closed network connection") {
				break
			}
		}
	}()
	port, err = strconv.Atoi(msg.Data)
	if err != nil {
		errInt = 1
		return
	}
	err = tools.ValidatePort(port)
	if err != nil {
		errInt = 2
		return
	}
	fserver, err = NewForwardServer(c.ctx, c.ip, port, msg.ClientID, tools.GenerateUUID(), c.createConn)
	if err != nil {
		errInt = 3
	} else {
		c.controlRecord.Add(msg.ClientID, msg.Data, fserver)
		//c.controlRecord.Add(msg.ClientID, fserver.serverID, fserver)
		fserver.Run()
	}
}

func (c *ControlServ) handleHeartbeat(_ interface{}, msg *message.Message) {
	// PASS
}

func (c *ControlServ) handleConn(conn net.Conn) {
	c.logger.Info("Connection from %s", conn.RemoteAddr().String())
	reader := bufio.NewReader(conn)
	coreConn := core.NewCoreConner(core.NewReadWriteCloser(reader, conn, nil), conn)
	for {
		msg, err := core.DecodeOneMsg(coreConn)
		if err != nil {
			return
		}
		if msg.Operation != message.REGISTER ||
			(msg.ConnType != message.ControlConn && msg.ConnType != message.ForwardConn) {
			continue
		}
		go c.handleRegister(coreConn, msg)
		switch msg.ConnType {
		case message.ControlConn:
			goto controlConn
		case message.ForwardConn:
			goto forwardConn
		}
	}
forwardConn:
	for {
		msg, err := core.DecodeOneMsg(coreConn)
		if err != nil {
			return
		}
		if msg.Operation == message.CreateForwardConn && msg.ConnType == message.ForwardConn {
			go c.handleCreateConn(coreConn, msg)
			return
		}
	}
controlConn:
	defer func() {
		err := coreConn.Close()
		if err != nil {
			c.logger.Error(err.Error())
		} else {
			c.logger.Info("%s Close.", coreConn.RemoteAddr().String())
		}
	}()
	msgCh := core.Decode2MsgCh(reader)
	for msg := range msgCh {
		if msg.ConnType != message.ControlConn {
			c.logger.Error("connType error message from %s %s", conn.RemoteAddr().String(), msg.String())
			continue
		}
		if msg.Operation != message.HEARTBEAT {
			c.logger.Info("message from %s %s", conn.RemoteAddr().String(), msg.String())
		}
		switch msg.Operation {
		case message.REGISTER:
			// register
			go c.handleRegister(coreConn, msg)
		case message.CreateForwardConn:
			// create forward conn
			go c.handleCreateConn(coreConn, msg)
		case message.CreateForwardServer:
			// create forward server
			go c.handleCreateServer(coreConn, msg)
		case message.HEARTBEAT:
			// heartbeat
			go c.handleHeartbeat(coreConn, msg)
		default:
			// error
			c.logger.Warn("error message from %s %s", conn.RemoteAddr().String(), msg.String())
		}
	}
}

func (c *ControlServ) HandleConn() {
	for conn := range c.connCh {
		go c.handleConn(conn)
	}
}

func (c *ControlServ) Release() {
	err := c.Close()
	if err != nil {
		c.logger.Warn(err.Error())
	}
	c.cancelFunc()
	c.controlRecord.Clear()
	return
}
