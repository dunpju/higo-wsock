package wsock

import (
	"fmt"
	"github.com/dunpju/higo-router/router"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
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
	serve            string
	UpGrader         websocket.Upgrader
	WsPingHandle     WebsocketFunc
	WsPongHandle     WebsocketFunc
	WsContainer      *WebsocketClient
	WsGroupContainer *GroupContainer
	WsCheckOrigin    WebsocketCheckFunc
	WsPitPatSleep    time.Duration
	PingFunc         PitPatFunc
	PongFunc         PitPatFunc
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
	WsGroupContainer = NewGroupContainer()
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

// ConnUpGrader 连接升级
func ConnUpGrader() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if router.GetRoutes(Serve()).Exist(ctx.Request.Method, ctx.Request.URL.Path) {
			route, err := router.GetRoutes(Serve()).Route(ctx.Request.Method, ctx.Request.URL.Path)
			if err != nil {
				panic(err)
			}
			if route.IsWs() {
				flag, err := Upgrade(ctx)
				if err != nil {
					panic(err)
				}
				ctx.Set(WsConnIp, flag)
				EventConn(ctx, flag)
				ctx.Abort()
				return
			}
		}
		ctx.Next()
	}
}

func handler(handlerFunc gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			// 向调度器传递panic
			if r := recover(); r != nil {
				panic(r)
			}
		}()
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

func Handler(handlerFunc gin.HandlerFunc) gin.HandlerFunc {
	return handler(handlerFunc)
}
