package wsock

import (
	"github.com/dunpju/higo-utils/utils/maputil"
	"github.com/gin-gonic/gin"
)

type Responder struct {
	ctx *gin.Context
	g   *ClientGroup
}

func newResponder(ctx *gin.Context, g *ClientGroup) *Responder {
	return &Responder{ctx: ctx, g: g}
}

func (this *Responder) Send(message string) (err error) {
	this.g.Range(func(clientFlag string, client *Client) bool {
		wsConn, ok := WsContainer.Get(clientFlag)
		if ok {
			err = wsConn.Send(message)
			if err != nil {
				return false
			}
			wsConn.Abort()
		}
		return true
	})
	return
}

func (this *Responder) WriteMessage(message string) {
	this.g.Range(func(clientFlag string, client *Client) bool {
		wsConn, ok := WsContainer.Get(clientFlag)
		if ok {
			wsConn.WriteMessage(message)
			wsConn.Abort()
		}
		return true
	})
}

func (this *Responder) WriteMap(message *maputil.ArrayMap) {
	this.g.Range(func(clientFlag string, client *Client) bool {
		wsConn, ok := WsContainer.Get(clientFlag)
		if ok {
			wsConn.WriteMap(message)
			wsConn.Abort()
		}
		return true
	})
}

func (this *Responder) WriteStruct(message interface{}) {
	this.g.Range(func(clientFlag string, client *Client) bool {
		wsConn, ok := WsContainer.Get(clientFlag)
		if ok {
			wsConn.WriteStruct(message)
			wsConn.Abort()
		}
		return true
	})
}

func (this *Responder) WriteError(message string) {
	this.g.Range(func(clientFlag string, client *Client) bool {
		wsConn, ok := WsContainer.Get(clientFlag)
		if ok {
			wsConn.WriteError(message)
			wsConn.Abort()
		}
		return true
	})
}

func (this *Responder) WriteClose() {
	this.g.Range(func(clientFlag string, client *Client) bool {
		wsConn, ok := WsContainer.Get(clientFlag)
		if ok {
			wsConn.WriteClose()
			wsConn.Abort()
		}
		return true
	})
}
