package service

import (
	"github.com/smallnest/libp2p/crypto"
	"github.com/smallnest/libp2p/p2p/node"
)

// Message is an interface to represent a simple message structure
type Message interface {
	Sender() node.Node
	Bytes() []byte
}

// Service is an interface that represents a networking service (ideally p2p) that we can use to send messages or listen to incoming messages
type Service interface {
	Start() error
	RegisterProtocol(protocol string) chan Message
	SendMessage(nodeID string, protocol string, payload []byte) error
	SubscribePeerEvents() (new chan crypto.PublicKey, del chan crypto.PublicKey)
	ProcessProtocolMessage(sender node.Node, protocol string, payload Data) error
	Broadcast(protocol string, payload []byte) error
	Shutdown()
}

type Data interface {
	messageData()
	Bytes() []byte
}

type DataBytes struct {
	Payload []byte
}

type DataMsgWrapper struct {
	Req     bool
	MsgType uint32
	ReqID   uint64
	Payload []byte
}

func (m DataBytes) messageData() {}

func (m DataBytes) Bytes() []byte {
	return m.Payload
}

func (m DataMsgWrapper) messageData() {}

func (m DataMsgWrapper) Bytes() []byte {
	return m.Payload
}
