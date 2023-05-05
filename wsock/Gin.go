package wsock

import (
	"github.com/dunpju/higo-router/router"
	"github.com/gin-gonic/gin"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

func Default() *Engine {
	eng := gin.Default()
	engine := &Engine{gin: eng}
	engine.group = &eng.RouterGroup
	return engine
}

func Gin(eng *gin.Engine) *Engine {
	engine := &Engine{gin: eng}
	engine.group = &eng.RouterGroup
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
	//fmt.Println(runtime.FuncForPC(reflect.ValueOf(handlerFunc).Pointer()).Name())
	this.gin.Use(handlerFunc)
	return this
}

type RouterGroup struct {
	group *gin.RouterGroup
}

func (this *RouterGroup) Gin() *gin.RouterGroup {
	return this.group
}

func (this *RouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	return &RouterGroup{group: this.group.Group(relativePath, handlers...)}
}

func (this *RouterGroup) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	this.group.Handle(httpMethod, relativePath, handlers...)
	return this
}

func (this *RouterGroup) Upgrade(relativePath string, handle gin.HandlerFunc, attributes ...*router.RouteAttribute) *RouterGroup {
	groupHandlers := make([]interface{}, 0)
	for _, handler := range this.group.Handlers {
		handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
		b0, err := regexp.MatchString(`\/gin\-gonic\/gin\.LoggerWithConfig\.func1$`, handlerName)
		if err != nil {
			panic(err)
		}
		b1, err := regexp.MatchString(`\/gin\-gonic\/gin\.CustomRecoveryWithWriter\.func1$`, handlerName)
		if err != nil {
			panic(err)
		}
		b2, err := regexp.MatchString(`\/higo\-wsock\/wsock\.ConnUpgrader\.func1$`, handlerName)
		if err != nil {
			panic(err)
		}
		if !b0 && !b1 && !b2 {
			groupHandlers = append(groupHandlers, handler)
		}
	}
	path := this.group.BasePath() + relativePath
	path = "/" + strings.TrimLeft(path, "/")
	abs := make([]*router.RouteAttribute, 0)
	abs = append(abs, router.Flag(router.Unique(router.GET, path)), router.IsWs(true))
	abs = append(abs, attributes...)
	if len(groupHandlers) > 0 {
		abs = append(abs, router.Middleware(groupHandlers...))
	}
	router.AddRoute(router.GET, path, handler(handle), abs...)
	this.group.Handle(router.GET, relativePath, handler(handle))
	return this
}

func (this *RouterGroup) WSock(relativePath string, handle gin.HandlerFunc, attributes ...*router.RouteAttribute) *RouterGroup {
	return this.Upgrade(relativePath, handle, attributes...)
}
