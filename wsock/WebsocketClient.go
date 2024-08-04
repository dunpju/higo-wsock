package wsock

import (
	"github.com/dunpju/higo-router/router"
	"github.com/gin-gonic/gin"
	"sync"
)

type WebsocketClient struct {
	clients sync.Map
}

func NewWebsocketClient() *WebsocketClient {
	return &WebsocketClient{}
}

func (this *WebsocketClient) Store(ctx *gin.Context, route *router.Route, clientGroup *ClientGroup) {
	WsGroupContainer.Store(clientGroup.Flag(), clientGroup)
	clientGroup.Range(func(clientFlag string, client *Client) bool {
		_, ok := this.clients.Load(clientFlag)
		if !ok {
			wsConn := newWebsocketConn(clientFlag, client.GroupFlag(), ctx, route, client.Conn())
			this.clients.Store(wsConn.Flag(), wsConn)
			go wsConn.ping(WsPitPatSleep) //心跳
			go wsConn.writeLoop()         //写循环
			go wsConn.readLoop()          //读循环
			go wsConn.listenLoop()        //监听循环调度消息
		}
		return true
	})
}

func (this *WebsocketClient) SendAll(msg string) {
	this.Range(func(key string, connect *WebsocketConn) bool {
		err := connect.Send(msg)
		if err != nil {
			this.Remove(connect)
		}
		return true
	})
}

func (this *WebsocketClient) Range(fn func(key string, client *WebsocketConn) bool) {
	this.clients.Range(func(key, client interface{}) bool {
		return fn(key.(string), client.(*WebsocketConn))
	})
}

func (this *WebsocketClient) Remove(wsConn *WebsocketConn) {
	this.clients.Delete(wsConn.Flag())
	clientGroup, ok := WsGroupContainer.Get(wsConn.GroupFlag())
	if ok {
		clientGroup.Delete(wsConn.Flag())
		if clientGroup.Len() == 0 {
			WsGroupContainer.Delete(wsConn.GroupFlag())
		}
	}
}

func (this *WebsocketClient) Get(key string) (*WebsocketConn, bool) {
	conn, ok := this.clients.Load(key)
	if ok {
		return conn.(*WebsocketConn), ok
	}
	return nil, ok
}

func Conn(flag string) (*ClientGroup, bool) {
	return WsGroupContainer.Get(flag)
}
