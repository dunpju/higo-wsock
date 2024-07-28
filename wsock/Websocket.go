package wsock

import (
	"fmt"
	"github.com/dunpju/higo-router/router"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

const (
	WebsocketServe = "websocket"
	WsConnIp       = "ws_conn_ip"
	WsRequest      = "ws_request"
	WsRespstring   = "string"
	WsRespmap      = "map"
	WsRespstruct   = "struct"
	WsResperror    = "error"
	WsRespclose    = "close"
)

var (
	serve         string
	UpGrader      websocket.Upgrader
	WsPingHandle  WebsocketFunc
	WsPongHandle  WebsocketFunc
	WsContainer   *WebsocketClient
	WsCheckOrigin WebsocketCheckFunc
	WsPitPatSleep time.Duration
	PingFunc      PitPatFunc
	PongFunc      PitPatFunc
)

func init() {
	WsCheckOrigin = func(r *http.Request) bool {
		return true
	}
	UpGrader = websocket.Upgrader{
		CheckOrigin: WsCheckOrigin,
	}
	WsPingHandle = wsPingFunc
	WsPongHandle = wsPongFunc
	WsContainer = NewWebsocketClient()
	WsPitPatSleep = time.Second * 30
	PingFunc = func() string {
		return "ping"
	}
	PongFunc = func() string {
		return "pong"
	}
}

func SetServe(ser string) {
	serve = ser
}

func Serve() string {
	if serve != "" {
		return serve
	}
	return WebsocketServe
}

type WebsocketCheckFunc func(r *http.Request) bool

type WebsocketFunc func(websocketConn *WebsocketConn, wait time.Duration) bool

type PitPatFunc func() string

type WebsocketClient struct {
	clients sync.Map
}

func NewWebsocketClient() *WebsocketClient {
	return &WebsocketClient{}
}

func (this *WebsocketClient) Store(ctx *gin.Context, route *router.Route, conn *websocket.Conn) {
	wsConn := NewWebsocketConn(ctx, route, conn)
	this.clients.Store(conn.RemoteAddr().String(), wsConn)
	go wsConn.ping(WsPitPatSleep) //心跳
	go wsConn.writeLoop()         //写循环
	go wsConn.readLoop()          //读循环
	go wsConn.handlerLoop()       //处理控制循环
}

func (this *WebsocketClient) SendAll(msg string) {
	this.clients.Range(func(key, client interface{}) bool {
		connect := client.(*WebsocketConn).conn
		err := connect.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			this.Remove(connect)
		}
		return true
	})
}

func (this *WebsocketClient) Remove(conn *websocket.Conn) {
	this.clients.Delete(conn.RemoteAddr().String())
}

func (this *WebsocketClient) Get(key string) (*WebsocketConn, bool) {
	val, ok := this.clients.Load(key)
	if ok {
		return val.(*WebsocketConn), ok
	}
	return nil, ok
}

// ConnUpGrader 连接升级
func ConnUpGrader() gin.HandlerFunc {
	router.AddServe(Serve())
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				panic(r)
			}
		}()
		if router.GetRoutes(Serve()).Exist(ctx.Request.Method, ctx.Request.URL.Path) {
			route, err := router.GetRoutes(Serve()).Route(ctx.Request.Method, ctx.Request.URL.Path)
			if err != nil {
				panic(err)
			}
			if route.IsWs() {
				conn := upGrader(ctx)
				ctx.Set(WsConnIp, conn)
				return
			}
		}
		ctx.Next()
	}
}

func handler(handlerFunc gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, ok := ctx.Get(WsRequest)
		if !ok {
			ctx.Abort()
		} else {
			if _, ok = ctx.Get(WsConnIp); !ok {
				panic(fmt.Errorf("handler: websocket conn client non-existent"))
			}
			handlerFunc(ctx)
		}
	}
}
