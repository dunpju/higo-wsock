package wsock

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"reflect"
	"runtime"
)

func Default() *Engine {
	g := gin.Default()
	engine := &Engine{gin: g}
	engine.group = g.RouterGroup
	return engine
}

func Gin(eng *gin.Engine) *Engine {
	engine := &Engine{gin: eng}
	engine.group = eng.RouterGroup
	return engine
}

type Engine struct {
	RouterGroup
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

type RouterGroup struct {
	group gin.RouterGroup
}

func (this *RouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	return &RouterGroup{group: *this.group.Group(relativePath, handlers...)}
}

func (this *RouterGroup) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	this.group.Handle(httpMethod, relativePath, handlers...)
	return this
}
