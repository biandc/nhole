package control

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/biandc/nhole/pkg/config"
	"github.com/biandc/nhole/pkg/core"
	"github.com/biandc/nhole/pkg/log"
	"github.com/biandc/nhole/pkg/message"
)

type ServiceInfo struct {
	ip   string
	port int
}

type ControlClient struct {
	ip       string
	port     int
	clientID string
	net.Conn
	ctx context.Context
	//cancelFunc context.CancelFunc
	logger   *log.Logger
	msgCh    chan *message.Message
	services map[int]ServiceInfo
}

func NewControlClienter(ctx context.Context, ip string, port int) (c *ControlClient, err error) {
	//var conn net.Conn
	//conn, err = core.NewConner(ip, port)
	//if err != nil {
	//	return
	//}
	//reader := bufio.NewReader(conn)
	//coreConn := core.NewCoreConner(core.NewReadWriteCloser(reader, conn, nil), conn)
	//newCtx, cancel := context.WithCancel(ctx)
	newCtx := ctx
	//msgCh := core.Decode2MsgCh(reader)
	cfg := ctx.Value("cfg").(*config.ClientCfg)
	services := make(map[int]ServiceInfo, len(cfg.Services))
	for _, service := range cfg.Services {
		if _, ok := services[service.ForwardPort]; ok {
			err = fmt.Errorf("config forward_port duplication")
			return
		}
		services[service.ForwardPort] = ServiceInfo{
			ip:   service.Ip,
			port: service.Port,
		}
	}
	c = &ControlClient{
		ip:       ip,
		port:     port,
		clientID: "",
		//Conn:     coreConn,
		ctx: newCtx,
		//cancelFunc: cancel,
		logger: log.FromContextSafe(newCtx),
		//msgCh:    msgCh,
		services: services,
	}
	return
}

func (c *ControlClient) Init() (err error) {
	var conn net.Conn
	conn, err = core.NewConner(c.ip, c.port)
	if err != nil {
		return
	}
	reader := bufio.NewReader(conn)
	coreConn := core.NewCoreConner(core.NewReadWriteCloser(reader, conn, nil), conn)
	msgCh := core.Decode2MsgCh(reader)
	c.Conn = coreConn
	c.msgCh = msgCh
	addr := c.RemoteAddr().String()
	c.logger.Info("connect to nhole-server %s ...", addr)
	c.logger.AppendPrefix(addr)
	return
}

func (c *ControlClient) Run() {
	c.register()
	go c.heartbeat()
	c.handleData()
}

func (c *ControlClient) register() {
	var (
		msgBytes []byte
		msg      *message.Message
		err      error
	)
	defer func() {
		if err != nil {
			c.logger.Error(err.Error())
		} else {
			c.logger.Info("register send %s", msg.String())
		}
	}()
	msgBytes, msg, err = core.EncodeOneMsg("", message.ControlConn, message.REGISTER, 0, "", "")
	if err != nil {
		return
	}
	for {
		_, err = c.Write(msgBytes)
		if err == nil || strings.Contains(err.Error(), "use of closed network connection") {
			err = nil
			break
		}
	}
}

func (c *ControlClient) handleRegister(msg *message.Message) {
	c.clientID = msg.ClientID
	c.logger.Info("set clientID %s ...", c.clientID)
	c.logger.AppendPrefix(c.clientID)
	c.createServer()
}

//func (c *ControlClient) createConn() {
//
//}

func (c *ControlClient) handleCreateConn(msg *message.Message) {
	var (
		data        *message.CreateConnData
		forwardPort int
		clienter    *ForwardClient
		err         error
		//errInt      = 0
	)
	defer func() {
		if err != nil {
			log.Error("Error creating forwarding connection %s !!!", err.Error())
		} else {
			log.Info("Successfully created forwarding connection %s .", data.ForwardID)
		}
	}()
	//defer func() {
	//}()
	switch msg.Error {
	case 0:
		data, err = message.UnmarshalCreateConnData(msg.Data)
		if err != nil {
			//errInt = 1
			return
		}
		forwardPort, err = strconv.Atoi(data.ServerID)
		if err != nil {
			//errInt = 2
			return
		}
		if localConnInfo, ok := c.services[forwardPort]; !ok {
			err = fmt.Errorf("no local connection information found %d", forwardPort)
			//errInt = 3
		} else {
			clienter, err = NewForwardClienter(
				localConnInfo.ip,
				localConnInfo.port,
				c.ip,
				c.port,
				data.ServerID,
				data.ForwardID,
			)
			if err != nil {
				//errInt = 4
				return
			}
			clienter.Run()
		}
	default:
		// PASS
	}
}

func (c *ControlClient) createServer() {
	for forwardPort := range c.services {
		msgBytes, msg, err := core.EncodeOneMsg(
			c.clientID,
			message.ControlConn,
			message.CreateForwardServer,
			0,
			"",
			strconv.Itoa(forwardPort),
		)
		if err != nil {
			c.logger.Error(err.Error())
			continue
		}
		for {
			_, err = c.Write(msgBytes)
			if err == nil || strings.Contains(err.Error(), "use of closed network connection") {
				err = nil
				c.logger.Info("createServer send %s", msg.String())
				break
			}
		}
	}
}

func (c *ControlClient) handleCreateServer(msg *message.Message) {
	switch msg.Error {
	case 0:
		c.logger.Info("Successfully created forwarding server %s.", msg.Data)
	default:
		c.logger.Error("Failed to create forwarding server %s %s !!!", msg.Data, msg.ErrorInfo)
		// retry
		time.Sleep(30 * time.Second)
		msgBytes, msg, err := core.EncodeOneMsg(
			c.clientID,
			message.ControlConn,
			message.CreateForwardServer,
			0,
			"",
			msg.Data,
		)
		if err != nil {
			c.logger.Error(err.Error())
			return
		}
		for {
			_, err = c.Write(msgBytes)
			if err == nil || strings.Contains(err.Error(), "use of closed network connection") {
				err = nil
				c.logger.Info("createServer send %s", msg.String())
				break
			}
		}
	}
}

func (c *ControlClient) heartbeat() {
	var (
		msgBytes []byte
		msg      *message.Message
		err      error
	)
	defer func() {
		if err != nil {
			c.logger.Error(err.Error())
		} else {
			c.logger.Debug("heartbeat send %s", msg.String())
		}
	}()
	msgBytes, msg, err = core.EncodeOneMsg(c.clientID, message.ControlConn, message.HEARTBEAT, 0, "", "")
	if err != nil {
		return
	}
	for {
		_, err = c.Write(msgBytes)
		if err == nil || strings.Contains(err.Error(), "use of closed network connection") {
			err = nil
			break
		}
	}
}

func (c *ControlClient) handleHeartbeat(_ interface{}) {
	time.Sleep(10 * time.Second)
	c.heartbeat()
}

func (c *ControlClient) handleData() {
	defer func() {
		c.logger.Info("Close.")
		c.clear()
	}()
	for msg := range c.msgCh {
		if (msg.ConnType != message.ControlConn) || // && msg.ConnType != message.ForwardConn) ||
			(c.clientID != "" && msg.ClientID != c.clientID) {
			c.logger.Error("connType error message %s", msg.String())
			continue
		}
		if msg.Operation != message.HEARTBEAT {
			c.logger.Info("message %s", msg.String())
		}
		switch msg.Operation {
		case message.REGISTER:
			// register
			go c.handleRegister(msg)
		case message.CreateForwardConn:
			// create forward conn
			go c.handleCreateConn(msg)
		case message.CreateForwardServer:
			// create forward server
			go c.handleCreateServer(msg)
		case message.HEARTBEAT:
			// heartbeat
			go c.handleHeartbeat(msg)
		default:
			// error
			c.logger.Warn("error message %s", msg.String())
		}
	}
}

func (c *ControlClient) clear() {
	c.clientID = ""
	c.msgCh = nil
	c.Conn = nil
	c.logger.ResetPrefixes()
}

func (c *ControlClient) Release() {
	if c.Conn != nil {
		_ = c.Close()
	}
}
