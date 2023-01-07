package message

import (
	"encoding/json"
	"fmt"

	"github.com/biandc/nhole/pkg/tools"
)

const (
	PackageHeadLen = 4

	REGISTER            = "REGISTER"
	CreateForwardConn   = "CREATE_FORWARD_CONN"
	CreateForwardServer = "CREATE_FORWARD_SERVER"
	HEARTBEAT           = "HEARTBEAT"

	ControlConn = "CONTROL"
	ForwardConn = "FORWARD"
)

type Message struct {
	ClientID  string `json:"clientID"`
	ConnType  string `json:"conn_type"`
	Operation string `json:"operation"`
	Error     int    `json:"error"`
	ErrorInfo string `json:"error_info"`
	Data      string `json:"data"`
}

func NewMessage(clientID, connType, operation string, errInt int, errInfo, data string) (m *Message) {
	m = &Message{
		ClientID:  clientID,
		ConnType:  connType,
		Operation: operation,
		Error:     errInt,
		ErrorInfo: errInfo,
		Data:      data,
	}
	return
}

func (m *Message) Validate() (err error) {
	err = tools.ValidateUUID(m.ClientID)
	if err != nil {
		return
	}
	err = ValidateOperation(m.Operation)
	if err != nil {
		return
	}
	err = ValidateConnType(m.ConnType)
	return
}

func (m *Message) String() (msgStr string) {
	msgStr = fmt.Sprintf("{clientID:%s,conn_type:%s,operation:%s,error:%d,error_info:%s,data:%s}", m.ClientID, m.ConnType, m.Operation, m.Error, m.ErrorInfo, m.Data)
	return
}

type CreateConnData struct {
	ServerID  string `json:"forward_server_id"`
	ForwardID string `json:"forward_id"`
}

func NewCreateConnData(serverID, forwardID string) (c *CreateConnData) {
	c = &CreateConnData{
		ServerID:  serverID,
		ForwardID: forwardID,
	}
	return
}

func (c *CreateConnData) Validate() (err error) {
	err = tools.ValidateUUID(c.ServerID)
	if err != nil {
		return
	}
	err = tools.ValidateUUID(c.ForwardID)
	return
}

func UnmarshalCreateConnData(str string) (data *CreateConnData, err error) {
	data = &CreateConnData{}
	err = json.Unmarshal([]byte(str), data)
	if err != nil {
		return
	}
	return
}

func MarshalCreateConnData(serverID, forwardID string) (data string, err error) {
	var bytes []byte
	bytes, err = json.Marshal(NewCreateConnData(serverID, forwardID))
	if err != nil {
		return
	}
	data = string(bytes)
	return
}

func ValidateOperation(operation string) (err error) {
	switch operation {
	case REGISTER:
	case CreateForwardConn:
	case CreateForwardServer:
	case HEARTBEAT:
	default:
		err = fmt.Errorf("%s ValidateOperation error", operation)
	}
	return
}

func ValidateConnType(connType string) (err error) {
	switch connType {
	case ControlConn:
	case ForwardConn:
	default:
		err = fmt.Errorf("%s ValidateConnType error", connType)
	}
	return
}

func UnmarshalMessage(data []byte) (msg *Message, err error) {
	msg = &Message{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		return
	}
	return
}

func MarshalMessage(msg *Message) (data []byte, err error) {
	data, err = json.Marshal(msg)
	return
}
