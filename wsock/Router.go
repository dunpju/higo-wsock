package wsock

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"reflect"
	"runtime"
)

func Use(handlerFunc gin.HandlerFunc) gin.HandlerFunc {
	fmt.Println(runtime.FuncForPC(reflect.ValueOf(handlerFunc).Pointer()).Name())
	return handlerFunc
}

func Group(relativePath string, handlers ...gin.HandlerFunc) (string, []gin.HandlerFunc) {
	for _, handler := range handlers {
		fmt.Println(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name())
	}
	return relativePath, handlers
}
