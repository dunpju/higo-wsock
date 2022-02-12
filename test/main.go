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
		ran := randomutil.Random().Int(1000)
		loginEntity.Time = loginEntity.Time + stringutil.IntString(ran)
		wsock.Response(context).WriteStruct(loginEntity)
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
