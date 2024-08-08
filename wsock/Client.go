package wsock

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"sync"
)

type Client struct {
	groupFlag string
	conn      *websocket.Conn
}

func newClient(groupFlag string, conn *websocket.Conn) *Client {
	return &Client{groupFlag: groupFlag, conn: conn}
}

func (c *Client) Conn() *websocket.Conn {
	return c.conn
}

func (c *Client) GroupFlag() string {
	return c.groupFlag
}

func (c *Client) Flag() string {
	return c.conn.RemoteAddr().String()
}

type ClientGroup struct {
	flag    string
	clients sync.Map
}

func NewClientGroup(flag string) *ClientGroup {
	return &ClientGroup{flag: flag, clients: sync.Map{}}
}

func (g *ClientGroup) Flag() string {
	return g.flag
}

func (g *ClientGroup) setFlag(flag string) {
	g.flag = flag
}

func (g *ClientGroup) Append(client *Client) {
	if g.Flag() == "" {
		g.setFlag(client.Flag())
	}
	g.clients.Store(client.Flag(), client)
}

func (g *ClientGroup) Delete(flag string) {
	g.clients.Delete(flag)
}

func (g *ClientGroup) Len() int {
	length := 0
	g.Range(func(clientFlag string, client *Client) bool {
		length++
		return true
	})
	return length
}

func (g *ClientGroup) Range(fn func(clientFlag string, client *Client) bool) {
	g.clients.Range(func(key, value any) bool {
		return fn(key.(string), value.(*Client))
	})
}

func (g *ClientGroup) Response(ctx *gin.Context) *Responder {
	return newResponder(ctx, g)
}

type GroupContainer struct {
	group sync.Map
}

func NewGroupContainer() *GroupContainer {
	return &GroupContainer{group: sync.Map{}}
}

func (this *GroupContainer) Get(key string) (*ClientGroup, bool) {
	group, ok := this.group.Load(key)
	if ok {
		return group.(*ClientGroup), ok
	}
	return nil, ok
}

func (this *GroupContainer) Store(key string, clientGroup *ClientGroup) {
	this.group.Store(key, clientGroup)
}

func (this *GroupContainer) Delete(key string) {
	this.group.Delete(key)
}

func (this *GroupContainer) Range(fn func(key string, clientGroup *ClientGroup) bool) {
	this.group.Range(func(key, value any) bool {
		return fn(key.(string), value.(*ClientGroup))
	})
}

func (this *GroupContainer) Len() int {
	length := 0
	this.Range(func(key string, clientGroup *ClientGroup) bool {
		length++
		return true
	})
	return length
}
