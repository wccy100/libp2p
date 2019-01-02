package connectionpool

import (
	"errors"
	"fmt"
	"github.com/smallnest/libp2p/crypto"
	"github.com/smallnest/libp2p/p2p/net"
	"github.com/smallnest/libp2p/p2p/node"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func generatePublicKey() crypto.PublicKey {
	_, pubKey, _ := crypto.GenerateKeyPair()
	return pubKey
}

func generateIpAddress() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Int31n(255), rand.Int31n(255), rand.Int31n(255), rand.Int31n(255))
}

func TestGetConnectionWithNoConnection(t *testing.T) {
	net := net.NewNetworkMock()
	net.SetDialDelayMs(50)
	net.SetDialResult(nil)
	cPool := NewConnectionPool(net, generatePublicKey())
	remotePub := generatePublicKey()
	addr := "1.1.1.1"
	conn, err := cPool.GetConnection(addr, remotePub)
	assert.Nil(t, err)
	assert.Equal(t, remotePub.String(), conn.RemotePublicKey().String())
	assert.Equal(t, int32(1), net.DialCount())
}

func TestGetConnectionWithConnection(t *testing.T) {
	net := net.NewNetworkMock()
	net.SetDialDelayMs(50)
	net.SetDialResult(nil)
	cPool := NewConnectionPool(net, generatePublicKey())
	remotePub := generatePublicKey()
	addr := "1.1.1.1"
	cPool.GetConnection(addr, remotePub)
	conn, err := cPool.GetConnection(addr, remotePub)
	assert.Nil(t, err)
	assert.Equal(t, remotePub.String(), conn.RemotePublicKey().String())
	assert.Equal(t, int32(1), net.DialCount())
}

func TestGetConnectionWithError(t *testing.T) {
	net := net.NewNetworkMock()
	net.SetDialDelayMs(50)
	eErr := errors.New("err")
	net.SetDialResult(eErr)
	cPool := NewConnectionPool(net, generatePublicKey())
	remotePub := generatePublicKey()
	addr := "1.1.1.1"
	conn, aErr := cPool.GetConnection(addr, remotePub)
	assert.Equal(t, eErr, aErr)
	assert.Nil(t, conn)
	assert.Equal(t, int32(1), net.DialCount())
}

func TestGetConnectionDuringDial(t *testing.T) {
	n := net.NewNetworkMock()
	remotePub := generatePublicKey()
	addr := "1.1.1.1"
	n.SetDialDelayMs(100)
	n.SetDialResult(nil)

	cPool := NewConnectionPool(n, generatePublicKey())
	waitCh := make(chan net.Connection)
	// dispatch 2 GetConnection calls
	dispatchF := func(ch chan net.Connection) {
		conn, _ := cPool.GetConnection(addr, remotePub)
		assert.Equal(t, remotePub.String(), conn.RemotePublicKey().String())
		ch <- conn
	}
	go dispatchF(waitCh)
	go dispatchF(waitCh)
	cnt := 0
	var prevId string
Loop:
	for {
		select {
		case c := <-waitCh:
			if prevId == "" {
				prevId = c.ID()
			} else {
				assert.Equal(t, prevId, c.ID())
			}
			cnt++
			if cnt == 2 {
				break Loop
			}
		case <-time.After(120 * time.Millisecond):
			fmt.Println("timeout!")
			assert.True(t, false)
			break Loop
		}
	}
	assert.Equal(t, int32(1), n.DialCount())
}

func TestRemoteConnectionWithNoConnection(t *testing.T) {
	n := net.NewNetworkMock()
	remotePub := generatePublicKey()
	addr := "1.1.1.1"
	n.SetDialDelayMs(50)
	n.SetDialResult(nil)

	cPool := NewConnectionPool(n, generatePublicKey())
	rConn := net.NewConnectionMock(remotePub)
	rConn.SetSession(net.NewSessionMock([]byte("aaa")))
	cPool.newRemoteConn <- net.NewConnectionEvent{rConn, node.EmptyNode}
	time.Sleep(50 * time.Millisecond)
	conn, err := cPool.GetConnection(addr, remotePub)
	assert.Equal(t, remotePub.String(), conn.RemotePublicKey().String())
	assert.Equal(t, rConn.ID(), conn.ID())
	assert.Nil(t, err)
	assert.Equal(t, int32(0), n.DialCount())
}

func TestRemoteConnectionWithExistingConnection(t *testing.T) {
	n := net.NewNetworkMock()
	addr := "1.1.1.1"
	cPool := NewConnectionPool(n, generatePublicKey())

	// local connection has session ID < remote's session ID
	remotePub := generatePublicKey()
	localSessionID := []byte("110")
	n.SetNextDialSessionID(localSessionID)
	lConn, _ := cPool.GetConnection(addr, remotePub)
	rConn := net.NewConnectionMock(remotePub)
	rConn.SetSession(net.NewSessionMock([]byte("111")))
	cPool.newRemoteConn <- net.NewConnectionEvent{rConn, node.EmptyNode}
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, remotePub.String(), lConn.RemotePublicKey().String())
	assert.Equal(t, int32(1), n.DialCount())
	assert.False(t, rConn.Closed())
	assert.True(t, lConn.Closed())

	// local connection has session ID > remote's session ID
	remotePub = generatePublicKey()
	localSessionID = []byte("111")
	n.SetNextDialSessionID(localSessionID)
	lConn, _ = cPool.GetConnection(addr, remotePub)
	rConn = net.NewConnectionMock(remotePub)
	rConn.SetSession(net.NewSessionMock([]byte("110")))
	cPool.newRemoteConn <- net.NewConnectionEvent{rConn, node.EmptyNode}
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, remotePub.String(), lConn.RemotePublicKey().String())
	assert.Equal(t, int32(2), n.DialCount())
	assert.True(t, rConn.Closed())
	assert.False(t, lConn.Closed())
}

func TestShutdown(t *testing.T) {
	n := net.NewNetworkMock()
	n.SetDialDelayMs(100)
	n.SetDialResult(nil)
	remotePub := generatePublicKey()
	addr := "1.1.1.1"

	cPool := NewConnectionPool(n, generatePublicKey())
	newConns := make(chan net.Connection)
	go func() {
		conn, _ := cPool.GetConnection(addr, remotePub)
		newConns <- conn
	}()
	time.Sleep(20 * time.Millisecond)
	cPool.Shutdown()
	conn := <-newConns
	cMock := conn.(*net.ConnectionMock)
	assert.True(t, cMock.Closed())
}

func TestGetConnectionAfterShutdown(t *testing.T) {
	n := net.NewNetworkMock()
	n.SetDialDelayMs(100)
	n.SetDialResult(nil)
	remotePub := generatePublicKey()
	addr := "1.1.1.1"

	cPool := NewConnectionPool(n, generatePublicKey())
	cPool.Shutdown()
	conn, err := cPool.GetConnection(addr, remotePub)
	assert.NotNil(t, err)
	assert.Nil(t, conn)
}

func TestShutdownWithMultipleDials(t *testing.T) {
	n := net.NewNetworkMock()
	n.SetDialDelayMs(100)
	n.SetDialResult(nil)

	cPool := NewConnectionPool(n, generatePublicKey())
	newConns := make(chan net.Connection)
	iterCnt := 20
	for i := 0; i < iterCnt; i++ {
		go func() {
			addr := generateIpAddress()
			key := generatePublicKey()
			conn, err := cPool.GetConnection(addr, key)
			if err == nil {
				newConns <- conn
			}
		}()
	}
	time.Sleep(20 * time.Millisecond)
	cPool.Shutdown()
	var cnt int
	for conn := range newConns {
		cMock := conn.(*net.ConnectionMock)
		assert.True(t, cMock.Closed(), "connection %s is still open", cMock.ID())
		cnt++
		if cnt == iterCnt {
			break
		}
	}
}

func TestClosedConnection(t *testing.T) {
	nMock := net.NewNetworkMock()
	nMock.SetDialDelayMs(50)
	nMock.SetDialResult(nil)
	cPool := NewConnectionPool(nMock, generatePublicKey())
	remotePub := generatePublicKey()
	addr := "1.1.1.1"

	// create connection
	conn, _ := cPool.GetConnection(addr, remotePub)

	// report that the connection was closed
	nMock.PublishClosingConnection(conn)
	time.Sleep(20 * time.Millisecond)

	// query same connection and assert that it's a new instance
	conn2, _ := cPool.GetConnection(addr, remotePub)
	assert.NotEqual(t, conn.ID(), conn2.ID())
	assert.Equal(t, int32(2), nMock.DialCount())
}

func TestRandom(t *testing.T) {
	type Peer struct {
		key  crypto.PublicKey
		addr string
	}

	peerCnt := 30
	peers := make([]Peer, 0)
	for i := 0; i < peerCnt; i++ {
		peers = append(peers, Peer{generatePublicKey(), generateIpAddress()})
	}

	nMock := net.NewNetworkMock()
	nMock.SetDialDelayMs(50)
	nMock.SetDialResult(nil)
	cPool := NewConnectionPool(nMock, generatePublicKey())
	rand.Seed(time.Now().UnixNano())
	for {
		r := rand.Int31n(3)
		if r == 0 {
			go func() {
				peer := peers[rand.Int31n(int32(peerCnt))]
				rConn := net.NewConnectionMock(peer.key)
				sID := make([]byte, 4)
				rand.Read(sID)
				rConn.SetSession(net.NewSessionMock(sID))
				cPool.newRemoteConn <- net.NewConnectionEvent{rConn, node.EmptyNode}
			}()
		} else if r == 1 {
			go func() {
				peer := peers[rand.Int31n(int32(peerCnt))]
				conn, err := cPool.GetConnection(peer.addr, peer.key)
				assert.Nil(t, err)
				nMock.PublishClosingConnection(conn)
			}()
		} else {
			go func() {
				peer := peers[rand.Int31n(int32(peerCnt))]
				_, err := cPool.GetConnection(peer.addr, peer.key)
				assert.Nil(t, err)
			}()
		}
		time.Sleep(10 * time.Millisecond)

		if rand.Int31n(100) == 0 {
			cPool.Shutdown()
			break
		}
	}
}
