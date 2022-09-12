package middleware

import (
	"github.com/gin-gonic/gin"
	"simpleGinIm/helper"
	"time"
)

func AuthCheck() gin.HandlerFunc {
	return func (ctx *gin.Context) {
		token := ctx.GetHeader("token")
		userToken, err := helper.AnalyseToken(token)
		if err != nil {
			// Abort用法 https://zhuanlan.zhihu.com/p/479526793
			ctx.Abort()
			helper.FailResponse(ctx, "用户认证不通过")
			return
		}

		// 校验过期时间
		if userToken.LoginExpire != 0 && userToken.LoginExpire < time.Now().Unix() {
			ctx.Abort()
			helper.FailResponse(ctx, "token已失效")
			return
		}

		ctx.Set("user_token", userToken)
		ctx.Next()
	}
}
