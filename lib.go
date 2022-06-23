package goyang

import (
	"errors"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type Protocols uint8

const (
	ProtocolHTTP_BODY_V1 Protocols = iota
	ProtocolWEBSOCKET
	ProtocolWEBTRANSPORT
	ProtocolIFRAME
	ProtocolPOLL
)

type Yang struct {
	protoNegoData []byte
	WSUpgrade     websocket.Upgrader

	connIDCounter uint64
}

var connPool = sync.Pool{
	New: func() interface{} {
		return &Conn{}
	},
}

type Conn struct {
	connID   uint64
	connType Protocols
	conn     *websocket.Conn
	w        http.ResponseWriter
	r        io.ReadCloser
}

func AcquireConn() *Conn {
	return connPool.Get().(*Conn)
}

func ReleaseConn(c *Conn) {
	resetConn(c)
	connPool.Put(c)
}

func resetConn(c *Conn) {
	c.connID = 0
	c.connType = 0
	c.conn = nil
	c.w = nil
	c.r = nil
}

var ErrInvalidProtocol = errors.New("invalid protocol")

func (y *Yang) Upgrade(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	if r.Method == "POST" {
		switch r.URL.Query().Get("y_req") {
		case "1":
			// WebSocket
			c, err := y.WSUpgrade.Upgrade(w, r, nil)
			if err != nil {
				return nil, err
			}
			conn := AcquireConn()
			conn.connID = atomic.AddUint64(&y.connIDCounter, 1)
			conn.connType = ProtocolWEBSOCKET
			conn.conn = c
			conn.w = w
			conn.r = r.Body
			return conn, nil
		case "0":
			// HTTP Body V1
			conn := AcquireConn()
			conn.connID = atomic.AddUint64(&y.connIDCounter, 1)
			conn.connType = ProtocolHTTP_BODY_V1
			conn.w = w
			conn.r = r.Body
			return conn, nil
		}
	}

	return nil, ErrInvalidProtocol
}

// Protocols
//
