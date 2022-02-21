package wsock

import (
	"github.com/dengpju/higo-router/router"
	"github.com/gin-gonic/gin"
	"reflect"
	"regexp"
	"runtime"
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

func (this *RouterGroup) Upgrade(httpMethod, relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	groupHandlers := make([]interface{}, 0)
	for _, handler := range this.group.Handlers {
		handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
		b0, err := regexp.MatchString(`^github\.com\/gin\-gonic\/gin\.LoggerWithConfig\.`, handlerName)
		if err != nil {
			panic(err)
		}
		b1, err := regexp.MatchString(`github\.com\/gin\-gonic\/gin\.CustomRecoveryWithWriter\.`, handlerName)
		if err != nil {
			panic(err)
		}
		b2, err := regexp.MatchString(`github\.com\/dengpju\/higo\-wsock\/wsock\.ConnUpgrader\.`, handlerName)
		if err != nil {
			panic(err)
		}
		if !b0 && !b1 && !b2 {
			//fmt.Println(runtime.FuncForPC(reflect.ValueOf(Handler).Pointer()).Name())
			groupHandlers = append(groupHandlers, handler)
		}
	}
	var lastHandler interface{}
	if len(handlers) > 0 {
		for _, handler := range handlers[:len(handlers)-1] {
			groupHandlers = append(groupHandlers, handler)
		}
		lastHandler = handle(handlers[len(handlers)-1])
		handlers[len(handlers)-1] = lastHandler.(gin.HandlerFunc)
	}
	path := this.group.BasePath() + relativePath
	router.AddRoute(httpMethod, path, lastHandler, router.Flag(router.Unique(httpMethod, path)),
		router.IsWs(true), router.Middleware(groupHandlers...))
	this.group.Handle(httpMethod, relativePath, handlers...)
	return this
}
