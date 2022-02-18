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
	"sync"
	"time"
)

var (
	//Recover处理函数(可自定义替换)
	WsRecoverHandle WsRecoverFunc
	wsRecoverOnce   sync.Once
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
}

type WsRecoverFunc func(conn *WebsocketConn, r interface{}) string

type WebsocketConn struct {
	lock      sync.RWMutex
	ctx       *gin.Context
	route     *router.Route
	conn      *websocket.Conn
	readChan  chan *WsReadMessage
	writeChan chan WsWriteMessage
	closeChan chan byte
}

func NewWebsocketConn(ctx *gin.Context, route *router.Route, conn *websocket.Conn) *WebsocketConn {
	return &WebsocketConn{ctx: ctx, route: route, conn: conn, readChan: make(chan *WsReadMessage),
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
	for {
		WsPingHandle(this, waittime)
	}
}

func (this *WebsocketConn) pong(waittime time.Duration) {
	for {
		WsPongHandle(this, waittime)
	}
}

func (this *WebsocketConn) readLoop() {
	for {
		t, message, err := this.conn.ReadMessage()
		if err != nil {
			this.close()
			break
		}
		this.readChan <- NewReadMessage(t, message)
	}
}

func (this *WebsocketConn) writeLoop() {
loop:
	for {
		select {
		case msg := <-this.writeChan:
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
			// 调度处理信息
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
	handle := this.route.Handle()
	ctx := this.ctx
	reader := bytes.NewReader(msg.MessageData)
	request, err := http.NewRequest(router.POST, this.route.AbsolutePath(), reader)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")
	ctx.Request = request
	handle.(gin.HandlerFunc)(ctx)
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
	return conn(ctx)
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

//webSocket请求连接
func upgraderConnFunc(ctx *gin.Context) string {
	//升级get请求为webSocket协议
	client, err := Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		panic(err)
	}

	route := router.GetRoutes(WebsocketServe).Route(ctx.Request.Method, ctx.Request.URL.Path).SetHeader(ctx.Request.Header)

	WsContainer.Store(ctx, route, client)
	return client.RemoteAddr().String()
}

func wsPingFunc(websocketConn *WebsocketConn, waittime time.Duration) {
	time.Sleep(waittime)
	err := websocketConn.conn.WriteMessage(websocket.TextMessage, []byte("ping"))
	if err != nil {
		WsContainer.Remove(websocketConn.conn)
		return
	}
}

func wsPongFunc(websocketConn *WebsocketConn, waittime time.Duration) {
	time.Sleep(waittime)
	err := websocketConn.conn.WriteMessage(websocket.TextMessage, []byte("pong"))
	if err != nil {
		WsContainer.Remove(websocketConn.conn)
		return
	}
}
