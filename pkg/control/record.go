package control

import (
	"fmt"
	"net"
	"sync"
)

type clientRecord struct {
	clientMap map[string]net.Conn
	sync.RWMutex
}

func NewClientRecord() (c *clientRecord) {
	c = &clientRecord{
		clientMap: make(map[string]net.Conn, 0),
	}
	return
}

func (c *clientRecord) Get(clientID string) (clienter net.Conn, err error) {
	var ok bool
	c.RLock()
	defer c.RUnlock()
	clienter, ok = c.clientMap[clientID]
	if !ok {
		err = fmt.Errorf("clientRecord not has %s", clientID)
	}
	return
}

func (c *clientRecord) GetAll() (conns []net.Conn) {
	c.RLock()
	defer c.RUnlock()
	conns = make([]net.Conn, 0, len(c.clientMap))
	for _, value := range c.clientMap {
		conns = append(conns, value)
	}
	return
}

func (c *clientRecord) Add(clientID string, clienter net.Conn) {
	c.Lock()
	defer c.Unlock()
	c.clientMap[clientID] = clienter
}

func (c *clientRecord) Del(clientID string) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.clientMap[clientID]; ok {
		delete(c.clientMap, clientID)
	}
}

func (c *clientRecord) Clear() {
	conns := c.GetAll()
	for _, conner := range conns {
		_ = conner.Close()
	}
}

type controlRecord struct {
	clientServer map[string][]string
	serverMap    map[string]*ForwardServ
	sync.RWMutex
}

func NewControlRecord() (c *controlRecord) {
	c = &controlRecord{
		clientServer: make(map[string][]string, 0),
		serverMap:    make(map[string]*ForwardServ, 0),
	}
	return
}

func (c *controlRecord) GetByServerID(serverID string) (server *ForwardServ, err error) {
	var ok bool
	c.RLock()
	defer c.RUnlock()
	server, ok = c.serverMap[serverID]
	if !ok {
		err = fmt.Errorf("serverID %s not find in ControlRecord", serverID)
	}
	return
}

func (c *controlRecord) Add(clientID, serverID string, server *ForwardServ) {
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

func (c *controlRecord) Del(clientID string) {
	c.Lock()
	defer c.Unlock()
	if serverIDs, ok := c.clientServer[clientID]; ok {
		for _, serverID := range serverIDs {
			if server, ok := c.serverMap[serverID]; ok {
				_ = server.Close()
				delete(c.serverMap, serverID)
			}
		}
		delete(c.clientServer, clientID)
	}
}
