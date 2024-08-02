package wsock

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

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
	request, err := http.NewRequest(this.route.Method(), this.route.AbsolutePath(), reader)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Length", strconv.Itoa(reader.Len()))
	request.RemoteAddr = this.context.Request.RemoteAddr
	request.URL.RawQuery = this.context.Request.URL.Query().Encode()
	request.Body = io.NopCloser(reader)

	ctx.Request = request
	this.isAborted = false
	for _, handle := range this.route.Handlers() {
		connUpGraderHandlerOk, err := regexp.MatchString(ConnUpGraderPattern, handle.FuncForPcName())
		if err != nil {
			panic(err)
		}
		if connUpGraderHandlerOk {
			continue
		}
		if !this.runHandle(ctx, handle.HandlerFunc()) {
			break
		}
	}
}

func (this *WebsocketConn) runHandle(ctx *gin.Context, handler interface{}) bool {
	if this.isAborted {
		return false
	}
	if handle, ok := handler.(func(*gin.Context)); ok {
		handle(ctx)
	} else if handle, ok := handler.(gin.HandlerFunc); ok {
		handle(ctx)
	} else {
		panic(`Non-supported Handle Type`)
	}
	this.isAborted = ctx.IsAborted()
	return true
}
