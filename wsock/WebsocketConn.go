package wsock

import (
	"bytes"
	"fmt"
	"gitee.com/dengpju/higo-code/code"
	"github.com/dengpju/higo-logger/logger"
	"github.com/dengpju/higo-router/router"
	"github.com/dengpju/higo-throw/exception"
	"github.com/dengpju/higo-utils/utils/maputil"
	"github.com/dengpju/higo-utils/utils/runtimeutil"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	//Recover处理函数(可自定义替换)
	WsRecoverHandle WsRecoverFunc
	wsRecoverOnce   sync.Once
	Encode          Encrypt
	Decode          Encrypt
	PingFailLimit   int
)

func init() {
	wsRecoverOnce.Do(func() {
		WsRecoverHandle = func(conn *WebsocketConn, r interface{}) (respMsg string) {
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
	PingFailLimit = 10
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

func (this *WebsocketConn) ping(waittime time.Duration) {
	for WsPingHandle(this, waittime) {
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
			}
			err := this.conn.WriteMessage(websocket.TextMessage, msg.MessageData)
			if err != nil {
				this.close()
				break loop
			}
		}
	}
}

func (this *WebsocketConn) handlerLoop() {
	defer func() {
		if r := recover(); r != nil {
			logger.LoggerStack(r, runtimeutil.GoroutineID())
			this.writeChan <- WsRespError(WsRecoverHandle(this, r))
		}
	}()
loop:
	for {
		select {
		case msg := <-this.readChan:
			this.dispatch(msg)
		case <-this.closeChan:
			break loop
		}
	}
}

func (this *WebsocketConn) dispatch(msg *WsReadMessage) {
	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
	}()
	conn, ok := this.context.Get(WsConnIp)
	if !ok {
		panic(fmt.Errorf("websocket conn client non-existent"))
	}
	ctx := &gin.Context{Request: &http.Request{PostForm: make(url.Values)}}
	ctx.Writer = this.context.Writer
	ctx.Set(WsConnIp, conn)
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
		if handle, ok := handler.(gin.HandlerFunc); ok {
			handle(ctx)
			if this.isAborted {
				break
			}
			this.isAborted = ctx.IsAborted()
		}
	}
}

func (this *WebsocketConn) WriteMessage(message string) {
	go func(msg string) {
		this.writeChan <- WsRespString(msg)
	}(message)
}

func (this *WebsocketConn) WriteMap(message maputil.ArrayMap) {
	go func(msg maputil.ArrayMap) {
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
		panic(fmt.Errorf("websocket conn client non-existent"))
	}
	if conn, ok := WsContainer.clients.Load(client); ok {
		return conn.(*WebsocketConn)
	} else {
		panic(fmt.Errorf("websocket conn non-existent"))
	}
}

// 升级
func upgrader(ctx *gin.Context) string {
	client, err := Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		panic(err)
	}
	route := router.GetRoutes(Serve()).Route(ctx.Request.Method, ctx.Request.URL.Path).SetHeader(ctx.Request.Header)
	WsContainer.Store(ctx, route, client)
	return client.RemoteAddr().String()
}

func wsPingFunc(websocketConn *WebsocketConn, waittime time.Duration) bool {
	time.Sleep(waittime)
	err := websocketConn.conn.WriteMessage(websocket.PingMessage, []byte("ping"))
	if err != nil {
		websocketConn.PingFailCounter++
		if websocketConn.PingFailCounter >= PingFailLimit {
			WsContainer.Remove(websocketConn.conn)
			return false
		}
	} else {
		websocketConn.PingFailCounter = 0
	}
	return true
}
