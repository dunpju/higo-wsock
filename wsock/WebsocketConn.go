package wsock

import (
	"bytes"
	"fmt"
	"gitee.com/dengpju/higo-code/code"
	"github.com/dunpju/higo-logger/logger"
	"github.com/dunpju/higo-router/router"
	"github.com/dunpju/higo-throw/exception"
	"github.com/dunpju/higo-utils/utils/maputil"
	"github.com/dunpju/higo-utils/utils/runtimeutil"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	// WsRecoverHandle Recover处理函数(可自定义替换)
	WsRecoverHandle WsRecoverFunc
	wsRecoverOnce   sync.Once
	Encode          Encrypt
	Decode          Encrypt
	FailLimit       int
)

func init() {
	wsRecoverOnce.Do(func() {
		WsRecoverHandle = func(conn *WebsocketConn, r interface{}) (respMsg string) {
			goid, _ := runtimeutil.GoroutineID()
			logger.LoggerStack(r, goid)
			if msg, ok := r.(*code.CodeMessage); ok {
				respMsg = maputil.Array().
					Put("code", msg.Code).
					Put("message", msg.Message).
					Put("data", nil).
					String()
			} else if arrayMap, ok := r.(maputil.ArrayMap); ok {
				respMsg = arrayMap.String()
			} else {
				respMsg = maputil.Array().
					Put("code", 0).
					Put("message", exception.ErrorToString(r)).
					Put("data", nil).
					String()
			}
			return
		}
	})
	Encode = func(data []byte) []byte {
		return data
	}
	Decode = func(data []byte) []byte {
		return data
	}
	FailLimit = 10
}

type WsRecoverFunc func(conn *WebsocketConn, r interface{}) string

type Encrypt func(data []byte) []byte

type WebsocketConn struct {
	lock            sync.RWMutex
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

func NewWebsocketConn(ctx *gin.Context, route *router.Route, conn *websocket.Conn) *WebsocketConn {
	return &WebsocketConn{context: ctx, route: route, conn: conn, readChan: make(chan *WsReadMessage),
		writeChan: make(chan WsWriteMessage), closeChan: make(chan byte)}
}

func (this *WebsocketConn) Conn() *websocket.Conn {
	return this.conn
}

func (this *WebsocketConn) Close() {
	this.close()
}

func (this *WebsocketConn) close() {
	_ = this.conn.Close()
	WsContainer.Remove(this.conn)
	this.closeChan <- 1
}

func (this *WebsocketConn) ping(wait time.Duration) {
	for WsPingHandle(this, wait) {
	}
}

func (this *WebsocketConn) pong(wait time.Duration) {
	for WsPongHandle(this, wait) {
	}
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
loop:
	for {
		select {
		case msg := <-this.writeChan:
			msg.MessageData = Encode(msg.MessageData)
			if WsResperror == msg.MessageType {
				_ = this.conn.WriteMessage(websocket.TextMessage, msg.MessageData)
				this.close()
				break loop
			} else {
				err := this.conn.WriteMessage(websocket.TextMessage, msg.MessageData)
				if err != nil {
					this.close()
					break loop
				}
			}
		}
	}
}

func (this *WebsocketConn) listenLoop() {
	defer this.recover()
loop:
	for {
		select {
		case msg := <-this.readChan:
			this.dispatch(msg)
		case <-this.closeChan:
			goto loop
		}
	}
}

func (this *WebsocketConn) recover() {
	if r := recover(); r != nil {
		this.writeChan <- WsRespString(WsRecoverHandle(this, r))
	}
	// 再次拉起监听循环调度
	this.listenLoop()
}

func (this *WebsocketConn) dispatch(msg *WsReadMessage) {
	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
	}()
	if this.conn == nil {
		panic(fmt.Errorf("dispatch: websocket conn client non-existent"))
	}
	ctx := &gin.Context{Request: &http.Request{PostForm: make(url.Values)}}
	ctx.Writer = this.context.Writer
	ctx.Set(WsConnIp, this.conn.RemoteAddr().String())
	ctx.Set(WsRequest, WsRequest)
	reader := bytes.NewReader(msg.MessageData)
	request, err := http.NewRequest(router.POST, this.route.AbsolutePath(), reader)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.RemoteAddr = this.context.Request.RemoteAddr
	request.URL.RawQuery = this.context.Request.URL.Query().Encode()
	ctx.Request = request
	handlers := this.route.Middleware()
	handlers = append(handlers.([]interface{}), this.route.Handle())
	this.isAborted = false
	for _, handler := range handlers.([]interface{}) {
		if !this.runHandle(ctx, handler) {
			break
		}
	}
}

func (this *WebsocketConn) runHandle(ctx *gin.Context, handler interface{}) bool {
	if handle, ok := handler.(func(*gin.Context)); ok {
		handle(ctx)
	} else if handle, ok := handler.(gin.HandlerFunc); ok {
		handle(ctx)
	} else {
		panic(`Non-supported Handle Type`)
	}

	if this.isAborted {
		return false
	}
	this.isAborted = ctx.IsAborted()
	return true
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
	conn.isAborted = true
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

// 升级
func upGrader(ctx *gin.Context) string {
	client, err := UpGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		panic(err)
	}
	route, err := router.GetRoutes(Serve()).Route(ctx.Request.Method, ctx.Request.URL.Path)
	if err != nil {
		panic(err)
	}
	route.SetHeader(ctx.Request.Header)
	WsContainer.Store(ctx, route, client)
	return client.RemoteAddr().String()
}

func wsPingFunc(websocketConn *WebsocketConn, waittime time.Duration) bool {
	time.Sleep(waittime)
	err := websocketConn.conn.WriteMessage(websocket.PingMessage, []byte(PingFunc()))
	if err != nil {
		websocketConn.PingFailCounter++
		if websocketConn.PingFailCounter >= FailLimit {
			WsContainer.Remove(websocketConn.conn)
			return false
		}
	} else {
		websocketConn.PingFailCounter = 0
	}
	return true
}

func wsPongFunc(websocketConn *WebsocketConn, wait time.Duration) bool {
	time.Sleep(wait)
	err := websocketConn.conn.WriteMessage(websocket.PongMessage, []byte(PongFunc()))
	if err != nil {
		websocketConn.PongFailCounter++
		if websocketConn.PongFailCounter >= FailLimit {
			WsContainer.Remove(websocketConn.conn)
			return false
		}
	} else {
		websocketConn.PongFailCounter = 0
	}
	return true
}
