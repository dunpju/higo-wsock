package main

import (
	"fmt"
	"github.com/dengpju/higo-router/router"
	"github.com/dengpju/higo-utils/utils/encodeutil"
	"github.com/dengpju/higo-wsock/wsock"
	"github.com/gin-gonic/gin"
	"log"
)

//jmeter 压测 https://www.bbsmax.com/A/8Bz8jog15x/
func main() {
	r := wsock.Default()
	r.Gin().Static("/index", "./dist")
	r.Use(wsock.ConnUpgrader())
	r.Use(func(context *gin.Context) {
		fmt.Println("use1")
		fmt.Println(context.Request.URL.Query().Get("token"))
		context.Next()
	})
	g1 := r.Group("/g1", func(context *gin.Context) {
		fmt.Println("g1")
		context.Next()
	})
	g2 := g1.Group("/g2", func(context *gin.Context) {
		fmt.Println("g2-1")
		context.Next()
	}, func(context *gin.Context) {
		fmt.Println("g2-2")
		context.Next()
	})
	g3 := g2.Group("/g3", func(context *gin.Context) {
		fmt.Println("g3")
		context.Next()
	})
	g3.Handle("GET", "/test", func(context *gin.Context) {
		fmt.Println("test1")
		context.Next()
	}, func(context *gin.Context) {
		fmt.Println("test2")
	})
	g2.Upgrade("/conn", func(context *gin.Context) {
		fmt.Println("conn1")
		wsock.Response(context).WriteMessage("11")
	})
	r.Upgrade("/conn1", func(context *gin.Context) {
		fmt.Println("conn2")
		fmt.Println(context.Writer)
		loginEntity := NewLoginEntity()
		err := context.ShouldBind(loginEntity)
		if err != nil {
			panic(err)
		}
		fmt.Println("Conn", loginEntity)
		wsock.Response(context).WriteStruct(loginEntity)
	}, router.IsAuth(true))
	router.AddServe(wsock.Serve()).ForEach(func(index int, route *router.Route) {
		fmt.Println(route)
	})
	wsock.Encode = func(data []byte) []byte {
		return []byte(encodeutil.Base64Encode(data))
	}
	wsock.Decode = func(data []byte) []byte {
		fmt.Println(string(data))
		return data
	}
	err := r.Run(":8080")
	if err != nil {
		log.Fatalln(err)
	}
}

type LoginEntity struct {
	UserName string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func NewLoginEntity() *LoginEntity {
	return &LoginEntity{}
}
