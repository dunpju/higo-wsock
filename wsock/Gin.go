package wsock

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"reflect"
	"runtime"
)

type Engine struct {
	Gin *gin.Engine
}

func Default() *Engine {
	return &Engine{Gin: gin.Default()}
}

func Gin(engine *gin.Engine) *Engine {
	return &Engine{Gin: engine}
}

func (this *Engine) Run(addr ...string) (err error) {
	return this.Gin.Run(addr...)
}

func (this *Engine) Use(handlerFunc gin.HandlerFunc) *Engine {
	fmt.Println(runtime.FuncForPC(reflect.ValueOf(handlerFunc).Pointer()).Name())
	this.Gin.Use(handlerFunc)
	return this
}

func (this *Engine) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	for _, handler := range handlers {
		fmt.Println(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name())
	}
	return &RouterGroup{GinGroup: this.Gin.Group(relativePath, handlers...)}
}

func (this *Engine) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) *Engine {
	this.Gin.Handle(httpMethod, relativePath, handlers...)
	return this
}

type RouterGroup struct {
	GinGroup *gin.RouterGroup
}

func (this *RouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	this.GinGroup.Group(relativePath, handlers...)
	return this
}

func (this *RouterGroup) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	this.GinGroup.Handle(httpMethod, relativePath, handlers...)
	return this
}
