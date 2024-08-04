package wsock

import (
	"gitee.com/dengpju/higo-code/code"
	"github.com/dunpju/higo-logger/logger"
	"github.com/dunpju/higo-throw/exception"
	"github.com/dunpju/higo-utils/utils/maputil"
	"github.com/dunpju/higo-utils/utils/runtimeutil"
)

func init() {
	wsRecoverOnce.Do(func() {
		WsRecoverHandle = func(conn *WebsocketConn, r interface{}) (respMsg string) {
			goid, _ := runtimeutil.GoroutineID()
			logger.LoggerStack(r, goid)
			if msg, ok := r.(*code.CodeMessage); ok {
				respMsg = maputil.Array().
					Put("code", msg.Code).
					Put("message", msg.Message).
					Put("data", nil).
					String()
			} else if arrayMap, ok := r.(maputil.ArrayMap); ok {
				respMsg = arrayMap.String()
			} else if arrayMap, ok := r.(*maputil.ArrayMap); ok {
				respMsg = arrayMap.String()
			} else {
				respMsg = maputil.Array().
					Put("code", 0).
					Put("message", exception.ErrorToString(r)).
					Put("data", nil).
					String()
			}
			return
		}
	})
	Encode = func(data []byte) []byte {
		return data
	}
	Decode = func(data []byte) []byte {
		return data
	}
	FailLimit = 10
}
