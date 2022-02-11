package main

import (
	"fmt"
	"github.com/dengpju/higo-wsock/wsock"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
)

func main() {
	r := gin.Default()
	r.Use(wsock.WsConnMiddleWare(r))
	r.Handle("GET", "/conn", func(context *gin.Context) {
		fmt.Println("hhh")
		fmt.Println(context.Writer)
		loginEntity := NewLoginEntity()
		err := context.ShouldBind(loginEntity)
		if err != nil {
			panic(err)
		}
		fmt.Println("Conn", loginEntity)
		wsock.WsConn(context).Conn().WriteMessage(websocket.TextMessage, []byte("ttt"))
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