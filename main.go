package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var Rds *redis.Client

func main() {
	//gin.SetMode(gin.ReleaseMode)
	//r := gin.Default()
	//r.GET("/index", func(c *gin.Context) {
	//	c.JSON(http.StatusOK, gin.H{
	//		"code": http.StatusOK,
	//		"data": "success",
	//	})
	//})
	//r.Run()
	InitRedisClient()

	r := gin.Default()
	r.Any("/*proxyPath", proxyHandle)
	err := r.Run("0.0.0.0:8080")
	if err != nil {
		panic(err)
	}
}

func proxyHandle(c *gin.Context) {
	remote, err := url.Parse("http://x.x.x.x:8081")
	if err != nil {
	}
	k := c.Param("proxyPath") + c.Query("keyword")
	//fmt.Println(k)
	ctx := context.Background()
	ret, err1 := Rds.Get(ctx, k).Bytes()
	if err1 != nil {
		proxy := httputil.NewSingleHostReverseProxy(remote)
		// 说明没有缓存
		fmt.Println("没有缓存，请求接口:", err1)
		proxy.Director = func(req *http.Request) {
			req.Header = c.Request.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			req.URL.Path = c.Param("proxyPath")
		}
		proxy.ModifyResponse = func(response *http.Response) error {

			b, _ := io.ReadAll(response.Body)
			fmt.Println("设置相应数据：")
			m := JsonToMap(b)
			setErr := Rds.Set(ctx, k, b, 20*time.Second).Err()
			if setErr != nil {
				fmt.Println("rds set error:", setErr)
			}
			c.JSON(200, &m)
			return nil

		}
		proxy.ServeHTTP(c.Writer, c.Request)
	} else {
		fmt.Println("有缓存，直接返回")
		m := JsonToMap(ret)
		c.JSON(200, &m)
	}

}

func InitRedisClient() {
	Rds = redis.NewClient(&redis.Options{
		Addr:     "x.x.x.x:6379",
		Password: "rds#2023",
		DB:       0,
	})
}

func MapToJson(m map[string]interface{}) string {
	d, _ := json.Marshal(m)
	return string(d)
}

func JsonToMap(bytes []byte) map[string]interface{} {
	tmpMap := map[string]interface{}{}
	err := json.Unmarshal(bytes, &tmpMap)
	if err != nil {
		return nil
	}
	return tmpMap
}
