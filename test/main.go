package main

import (
	"fmt"
	"github.com/dengpju/higo-utils/utils/randomutil"
	"github.com/dengpju/higo-utils/utils/stringutil"
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
	g2.Upgrade("GET", "/conn", func(context *gin.Context) {
		fmt.Println("conn1")
		fmt.Println(context.Writer)
		//context.Abort()
		return
		loginEntity := NewLoginEntity()
		err := context.ShouldBind(loginEntity)
		if err != nil {
			panic(err)
		}
		fmt.Println("Conn", loginEntity)
		ran := randomutil.Random().Int(1000)
		loginEntity.Time = loginEntity.Time + stringutil.IntString(ran)
		wsock.Response(context).WriteStruct(loginEntity)
	}, func(context *gin.Context) {
		fmt.Println("conn2")
		wsock.Response(context).WriteMessage("11")
	})
	err := r.Run(":8080")
	if err != nil {
		log.Fatalln(err)
	}
}

type LoginEntity struct {
	UserName    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaCode string `json:"captcha_code" binding:"required"`
	Time        string `json:"time" binding:"required"`
}

func NewLoginEntity() *LoginEntity {
	return &LoginEntity{}
}
