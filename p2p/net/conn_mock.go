package net

import (
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/smallnest/libp2p/crypto"
)

type ConnectionMock struct {
	id        string
	remotePub crypto.PublicKey
	session   NetworkSession
	source    ConnectionSource

	addr string

	sendDelayMs int
	sendRes     error
	sendCnt     int32

	closed bool
}

func NewConnectionMock(key crypto.PublicKey) *ConnectionMock {
	return &ConnectionMock{
		id:        crypto.UUIDString(),
		remotePub: key,
		closed:    false,
	}
}

func (cm ConnectionMock) ID() string {
	return cm.id
}

func (cm ConnectionMock) RemotePublicKey() crypto.PublicKey {
	return cm.remotePub
}

func (cm *ConnectionMock) SetRemotePublicKey(key crypto.PublicKey) {
	cm.remotePub = key
}

func (cm *ConnectionMock) RemoteAddr() net.Addr {

	if cm.addr == "" {
		return nil
	}

	addr, port, _ := net.SplitHostPort(cm.addr)
	portstr, _ := strconv.Atoi(port)

	return &net.TCPAddr{net.ParseIP(addr), portstr, ""}
}

func (cm *ConnectionMock) SetSession(session NetworkSession) {
	cm.session = session
}

func (cm ConnectionMock) Session() NetworkSession {
	return cm.session
}

func (cm ConnectionMock) IncomingChannel() chan []byte {
	return nil
}

func (cm *ConnectionMock) SetSendDelay(delayMs int) {
	cm.sendDelayMs = delayMs
}

func (cm *ConnectionMock) SetSendResult(err error) {
	cm.sendRes = err
}

func (cm ConnectionMock) SendCount() int32 {
	return atomic.LoadInt32(&cm.sendCnt)
}

func (cm *ConnectionMock) Send(m []byte) error {
	atomic.AddInt32(&cm.sendCnt, int32(1))
	time.Sleep(time.Duration(cm.sendDelayMs) * time.Millisecond)
	return cm.sendRes
}

func (cm ConnectionMock) Closed() bool {
	return cm.closed
}

func (cm *ConnectionMock) Close() {
	cm.closed = true
}

func (cm *ConnectionMock) beginEventProcessing() {

}

func (cm ConnectionMock) String() string {
	return cm.id
}
