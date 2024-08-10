package wsock

import (
	"fmt"
	"github.com/dunpju/higo-router/router"
	"github.com/dunpju/higo-utils/utils/maputil"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

const (
	LoggerWithConfigPattern         = `\/gin\.LoggerWithConfig\.func1$`
	CustomRecoveryWithWriterPattern = `\/gin\.CustomRecoveryWithWriter\.func1$`
	ConnUpGraderPattern             = `\/wsock\.ConnUpGrader\.func1$`
)

var (
	// WsRecoverHandle Recover处理函数(可自定义替换)
	WsRecoverHandle WsRecoverFunc
	wsRecoverOnce   sync.Once
	Encode          Encrypt
	Decode          Encrypt
	FailLimit       int
)

type WsRecoverFunc func(conn *WebsocketConn, r interface{}) string

type Encrypt func(data []byte) []byte

type WebsocketConn struct {
	flag            string
	groupFlag       string
	context         *gin.Context
	route           *router.Route
	conn            *websocket.Conn
	readChan        chan *WsReadMessage
	writeChan       chan WsWriteMessage
	closeChan       chan byte
	isAborted       bool
	PingFailCounter int
	PongFailCounter int
}

func newWebsocketConn(flag, groupFlag string, ctx *gin.Context, route *router.Route, conn *websocket.Conn) *WebsocketConn {
	return &WebsocketConn{flag: flag, groupFlag: groupFlag, context: ctx, route: route, conn: conn, readChan: make(chan *WsReadMessage),
		writeChan: make(chan WsWriteMessage), closeChan: make(chan byte)}
}

func (this *WebsocketConn) Abort() {
	this.isAborted = true
}

func (this *WebsocketConn) Flag() string {
	return this.flag
}

func (this *WebsocketConn) GroupFlag() string {
	return this.groupFlag
}

func (this *WebsocketConn) Conn() *websocket.Conn {
	return this.conn
}

func (this *WebsocketConn) Close() {
	this.close()
}

func (this *WebsocketConn) close() {
	_ = this.conn.Close()
	WsContainer.Remove(this)
	this.closeChan <- 1
}

func (this *WebsocketConn) ping(wait time.Duration) {
	for WsPingHandle(this, wait) {
	}
	this.close()
}

func (this *WebsocketConn) pong(wait time.Duration) {
	for WsPongHandle(this, wait) {
	}
	this.close()
}

func (this *WebsocketConn) readLoop() {
	for {
		t, msg, err := this.conn.ReadMessage()
		if err != nil {
			this.close()
			break
		}
		msg = Decode(msg)
		this.readChan <- NewReadMessage(t, msg)
	}
}

func (this *WebsocketConn) writeLoop() {
	defer func() {
		if r := recover(); r != nil {
			this.writeLoop()
		}
	}()
loop:
	for {
		select {
		case msg := <-this.writeChan:
			msg.MessageData = Encode(msg.MessageData)
			if WsResperror == msg.MessageType {
				_ = this.conn.WriteMessage(websocket.TextMessage, msg.MessageData)
				break loop
			} else {
				err := this.conn.WriteMessage(websocket.TextMessage, msg.MessageData)
				if err != nil {
					break loop
				}
			}
		case <-this.closeChan:
			return
		}
	}
}

func (this *WebsocketConn) listenLoop() {
	defer this.recover()

	for {
		select {
		case msg := <-this.readChan:
			this.dispatch(msg)
		case <-this.closeChan:
			return
		}
	}
}

func (this *WebsocketConn) Send(msg string) error {
	return this.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (this *WebsocketConn) WriteMessage(message string) {
	go func(msg string) {
		this.writeChan <- WsRespString(msg)
	}(message)
}

func (this *WebsocketConn) WriteMap(message *maputil.ArrayMap) {
	go func(msg *maputil.ArrayMap) {
		this.writeChan <- WsRespMap(msg)
	}(message)
}

func (this *WebsocketConn) WriteStruct(message interface{}) {
	go func(msg interface{}) {
		this.writeChan <- WsRespStruct(msg)
	}(message)
}

func (this *WebsocketConn) WriteError(message string) {
	go func(msg string) {
		this.writeChan <- WsRespError(msg)
	}(message)
}

func (this *WebsocketConn) WriteClose() {
	go func() {
		this.writeChan <- WsRespClose()
	}()
}

func Response(ctx *gin.Context) *WebsocketConn {
	conn := conn(ctx)
	conn.Abort()
	return conn
}

func conn(ctx *gin.Context) *WebsocketConn {
	client, ok := ctx.Get(WsConnIp)
	if !ok {
		panic(fmt.Errorf("conn: websocket conn client non-existent"))
	}
	if conn, ok := WsContainer.clients.Load(client); ok {
		return conn.(*WebsocketConn)
	} else {
		panic(fmt.Errorf("conn: websocket conn non-existent"))
	}
}

func wsPingFunc(websocketConn *WebsocketConn, wait time.Duration) bool {
	defer func(websocketConn *WebsocketConn, wait time.Duration) {
		if r := recover(); r != nil {
			if websocketConn.PingFailCounter < FailLimit {
				wsPingFunc(websocketConn, wait)
			}
		}
	}(websocketConn, wait)
	time.Sleep(wait)
	err := websocketConn.conn.WriteMessage(websocket.PingMessage, []byte(PingFunc()))
	if err != nil {
		websocketConn.PingFailCounter++
		if websocketConn.PingFailCounter >= FailLimit {
			WsContainer.Remove(websocketConn)
			return false
		}
	} else {
		websocketConn.PingFailCounter = 0
	}
	return true
}

func wsPongFunc(websocketConn *WebsocketConn, wait time.Duration) bool {
	defer func(websocketConn *WebsocketConn, wait time.Duration) {
		if r := recover(); r != nil {
			if websocketConn.PongFailCounter < FailLimit {
				wsPingFunc(websocketConn, wait)
			}
		}
	}(websocketConn, wait)
	time.Sleep(wait)
	err := websocketConn.conn.WriteMessage(websocket.PongMessage, []byte(PingFunc()))
	if err != nil {
		websocketConn.PongFailCounter++
		if websocketConn.PongFailCounter >= FailLimit {
			WsContainer.Remove(websocketConn)
			return false
		}
	} else {
		websocketConn.PongFailCounter = 0
	}
	return true
}
