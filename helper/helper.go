package helper

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"simpleGinIm/define"
	"strings"
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

// 获取本机ip
func GetLocalIP() []string {
	var ipStr []string
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("net.Interfaces error:", err.Error())
		return ipStr
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()
			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					//获取IPv6
					/*if ipnet.IP.To16() != nil {
					    fmt.Println(ipnet.IP.String())
					    ipStr = append(ipStr, ipnet.IP.String())

					}*/
					//获取IPv4
					if ipnet.IP.To4() != nil {
						fmt.Println(ipnet.IP.String())
						ipStr = append(ipStr, ipnet.IP.String())

					}
				}
			}
		}
	}
	return ipStr

}

// 发送post请求
func SendPost(address string, body string) error {
	// res, err := http.Post(address, "application/x-www-form-urlencoded;charset=utf-8", bytes.NewBuffer(sendData))
	res, err := http.Post(address, "application/x-www-form-urlencoded;charset=utf-8", strings.NewReader(body))
	if err != nil {
		return err
	}

	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	log.Println(string(content))
	return nil
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
