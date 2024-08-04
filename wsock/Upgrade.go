package wsock

import (
	"github.com/dunpju/higo-router/router"
	"github.com/gin-gonic/gin"
)

var ClientFlag = func() *ClientGroup {
	return NewClientGroup("")
}

// Upgrade 升级
func Upgrade(ctx *gin.Context) (string, error) {
	route, err := router.GetRoutes(Serve()).Route(ctx.Request.Method, ctx.Request.URL.Path)
	if err != nil {
		return "", err
	}
	route.SetHeader(ctx.Request.Header)

	clientGroup := ClientFlag()

	conn, err := UpGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return "", err
	}

	cg, ok := WsGroupContainer.Get(clientGroup.Flag())
	if ok {
		cg.Append(newClient(clientGroup.Flag(), conn))
		WsContainer.Store(ctx, route, cg)
	} else {
		clientGroup.Append(newClient(clientGroup.Flag(), conn))
		WsContainer.Store(ctx, route, clientGroup)
	}

	return clientGroup.Flag(), nil
}
