package wsock

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"reflect"
	"runtime"
)

func Default() *Engine {
	return &Engine{gin: gin.Default()}
}

func Gin(engine *gin.Engine) *Engine {
	return &Engine{gin: engine}
}

type Engine struct {
	gin *gin.Engine
}

func (this *Engine) Gin() *gin.Engine {
	return this.gin
}

func (this *Engine) Run(addr ...string) (err error) {
	return this.gin.Run(addr...)
}

func (this *Engine) Use(handlerFunc gin.HandlerFunc) *Engine {
	fmt.Println(runtime.FuncForPC(reflect.ValueOf(handlerFunc).Pointer()).Name())
	this.gin.Use(handlerFunc)
	return this
}

func (this *Engine) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	for _, handler := range handlers {
		fmt.Println(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name())
	}
	return &RouterGroup{group: this.gin.Group(relativePath, handlers...)}
}

func (this *Engine) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) *Engine {
	this.gin.Handle(httpMethod, relativePath, handlers...)
	return this
}

type RouterGroup struct {
	group *gin.RouterGroup
}

func (this *RouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	return &RouterGroup{group: this.group.Group(relativePath, handlers...)}
}

func (this *RouterGroup) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	this.group.Handle(httpMethod, relativePath, handlers...)
	return this
}
