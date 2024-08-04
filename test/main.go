package main

import (
	"fmt"
	"github.com/dunpju/higo-router/router"
	"github.com/dunpju/higo-wsock/wsock"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"log"
	"time"
)

// jmeter 压测 https://www.bbsmax.com/A/8Bz8jog15x/
// http://wstool.js.org/
func main() {
	router.AddServe(wsock.Serve())
	r := wsock.Default()
	r.Gin().Static("/index", "./dist")
	r.Use(wsock.ConnUpGrader())
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
		//context.Abort()
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
		fmt.Println("conn")
		wsock.Response(context).WriteMessage("11")
	})
	// {"username":"3","password":"2","captcha_code":"731","time":"5350"}
	r.Upgrade("/conn1", func(context *gin.Context) {
		fmt.Println("conn2")
		fmt.Println(context.Writer)
		loginEntity := NewLoginEntity()
		err := context.ShouldBindBodyWith(loginEntity, binding.JSON)
		if err != nil {
			panic(err)
		}
		fmt.Println("Conn", loginEntity)
		conn, _ := wsock.Conn("1")
		conn.WriteStruct(loginEntity) // 测试自定义flag
		//wsock.Response(context).WriteStruct(loginEntity)
	}, router.IsAuth(true))
	router.AddServe(wsock.Serve()).ForEach(func(route *router.Route) {
		fmt.Println(*route)
	})
	wsock.Encode = func(data []byte) []byte {
		//return []byte(encodeutil.Base64Encode(data))
		return data
	}
	wsock.Decode = func(data []byte) []byte {
		fmt.Println("wsock.Decode")
		fmt.Println(string(data))
		return data
	}
	wsock.FailLimit = 3
	wsock.WsPitPatSleep = time.Second * 5
	// 自定义flag
	wsock.ClientFlag = func() *wsock.ClientGroup {
		return wsock.NewClientGroup("1")
	}
	go func() {
		for {
			time.Sleep(6 * time.Second)
			fmt.Println("WsContainer", wsock.WsContainer.Len())
			fmt.Println("WsGroupContainer", wsock.WsGroupContainer.Len())
		}
	}()
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
