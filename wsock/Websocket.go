package wsock

import (
	"github.com/dengpju/higo-router/router"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"regexp"
	"sync"
	"time"
)

const (
	WebsocketServe = "websocket"
	WsConnIp       = "ws_conn_ip"
	WsRespstring   = "string"
	WsRespmap      = "map"
	WsRespstruct   = "struct"
	WsResperror    = "error"
	WsRespclose    = "close"
)

var (
	isCollect     bool
	Upgrader      websocket.Upgrader
	WsPingHandle  WebsocketPingFunc
	WsPongHandle  WebsocketPingFunc
	WsContainer   *WebsocketClient
	WsCheckOrigin WebsocketCheckFunc
	WsPitpatSleep time.Duration
)

func init() {
	WsCheckOrigin = func(r *http.Request) bool {
		return true
	}
	Upgrader = websocket.Upgrader{
		CheckOrigin: WsCheckOrigin,
	}
	WsPingHandle = wsPingFunc
	WsPongHandle = wsPongFunc
	WsContainer = NewWebsocketClient()
	WsPitpatSleep = time.Second * 5
}

type WebsocketCheckFunc func(r *http.Request) bool

type WebsocketPingFunc func(websocketConn *WebsocketConn, waittime time.Duration)

type WebsocketClient struct {
	clients sync.Map
}

func NewWebsocketClient() *WebsocketClient {
	return &WebsocketClient{}
}

func (this *WebsocketClient) Store(ctx *gin.Context, route *router.Route, conn *websocket.Conn) {
	wsConn := NewWebsocketConn(ctx, route, conn)
	this.clients.Store(conn.RemoteAddr().String(), wsConn)
	go wsConn.ping(WsPitpatSleep) //心跳
	go wsConn.writeLoop()         //写循环
	go wsConn.readLoop()          //读循环
	go wsConn.handlerLoop()       //处理控制循环
}

func (this *WebsocketClient) SendAll(msg string) {
	this.clients.Range(func(key, client interface{}) bool {
		conn := client.(*WebsocketConn).conn
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			this.Remove(conn)
		}
		return true
	})
}

func (this *WebsocketClient) Remove(conn *websocket.Conn) {
	this.clients.Delete(conn.RemoteAddr().String())
}

//ws连接中间件
func WsConnMiddleWare(engine *gin.Engine) gin.HandlerFunc {
	router.AddServe(WebsocketServe)
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				panic(r)
			}
		}()
		if !isCollect {
			reg , err := regexp.Compile(`.createStaticHandler.`)
			if err != nil {
				panic(err)
			}
			for _, route := range engine.Routes() {
				if !reg.MatchString(route.Handler) {
					if !router.GetRoutes(WebsocketServe).Exist(route.Method, route.Path) {
						router.AddRoute(route.Method, route.Path, route.HandlerFunc, router.Flag(route.Handler),
							router.IsWs(true))
					}
				}
			}
			isCollect = true
		}

		route := router.GetRoutes(WebsocketServe).Route(ctx.Request.Method, ctx.Request.URL.Path)

		if route.IsWs() {
			conn := websocketConnFunc(ctx)
			// 设置变量到Context的key中，可以通过Get()取
			ctx.Set(WsConnIp, conn)
			// 终止执行
			ctx.Abort()
		}
		ctx.Next()
	}
}

// 连接升级协议handle
func WsUpgraderHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, ok := ctx.Get(WsConnIp)
		if !ok {
			panic("websocket conn ip non-existent")
		}
	}
}
