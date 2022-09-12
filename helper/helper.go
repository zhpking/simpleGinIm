package helper

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/ini.v1"
	"log"
	"net/http"
	"simpleGinIm/define"
	"time"
	)

// 加密字符串
var myKey = []byte("im")

type UserToken struct {
	UserId string `json:"user_id"`
	LoginExpire int64 `json:"login_expire"`
	jwt.StandardClaims
}

// 生成token
func GenerateToken(userId string) (string, error) {
	path := define.GetSysConfigPath()
	cfg, err := ini.Load(path)
	if err != nil {
		log.Printf("[SYS CONFIG ERROR] %v\n", err)
		return "", err
	}
	// 获取mongo分区的key
	loginExpire, _ := cfg.Section("login").Key("loginExpire").Int64() // 将结果转为int
	// 配置为0的话，登录不超时
	if loginExpire != 0 {
		loginExpire = time.Now().Unix() + loginExpire
	}

	userToken := &UserToken{
		// Identity:objId,
		UserId:userId,
		LoginExpire:loginExpire,
		StandardClaims:jwt.StandardClaims{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userToken)
	tokenString, err := token.SignedString(myKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// 解析token
func AnalyseToken(tokenString string) (*UserToken, error) {
	userToken := new(UserToken)
	claims, err := jwt.ParseWithClaims(tokenString, userToken, func(token *jwt.Token) (interface{}, error) {
		return myKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !claims.Valid {
		return nil, fmt.Errorf("Analyse Token Error:%v", err)
	}

	return userToken, nil
}

// 获取uuid
func GetUuid() string {
	return fmt.Sprintf("%x", uuid.New())
}

func SucResponse(ctx *gin.Context, msg string, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg": msg,
		"data": data,
	})
}

func FailResponse(ctx *gin.Context, msg string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code": -1,
		"msg": msg,
	})
}
