package service

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"simpleGinIm/define"
	"simpleGinIm/helper"
	"log"
	"simpleGinIm/model"
	"time"
)

func GetLoginToken(ctx *gin.Context) {
	userId := ctx.PostForm("user_id");
	currentTime := time.Now().Unix()

	// 获取该userId是否已经注册，如果没注册，入表
	user, err := model.GetUserByUserId(userId)
	if err == mongo.ErrNoDocuments {
		// 没有数据，相当于没注册，往用户表入一条数据
		user = &model.User{
			UserId:userId,
			LoginStatus:define.LOGIN_STATUS_ONLINE,
			LastLoginTime:currentTime,
			LastLoginOutTime:0,
			UserStatus:define.USER_STATUS_OK,
			CreateAt:currentTime,
		}
		err = model.UserInsertOne(user)
		if err != nil {
			log.Printf("[DB ERROR]%v\n", err)
			helper.FailResponse(ctx, "系统错误")
			return
		}
	} else if err != nil {
		log.Printf("[DB ERROR]%v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	} else {
		if user.LoginStatus == 1 {
			helper.FailResponse(ctx, "用户已登录")
			return
		} else {
			// 修改用户登录时间等信息
			_ = model.UpdateUserLoginStatusByUserId(userId, define.LOGIN_STATUS_ONLINE, currentTime)
		}
	}

	token, err := helper.GenerateToken(userId)
	if err != nil {
		log.Printf("生成token失败,%v", err)
		helper.FailResponse(ctx, "获取token失败，请重试")
		return
	}

	data := map[string]string{
		"token":token,
	}

	helper.SucResponse(ctx, "获取token成功", data)
}

func LoginOut(ctx *gin.Context) {
	// userId := ctx.PostForm("user_id");
	userToken := ctx.MustGet("user_token").(*helper.UserToken)
	currentTime := time.Now().Unix()

	err := model.UpdateUserLoginOutStatusByUserId(userToken.UserId, define.LOGIN_STATUS_OFFLINE, currentTime)
	if err != nil {
		log.Printf("[DB ERROR]%v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	}

	// 删除用户连接
	RemoveUserConnect(userToken.UserId)

	helper.SucResponse(ctx, "退出登录成功", make(map[string]interface{}))
}
