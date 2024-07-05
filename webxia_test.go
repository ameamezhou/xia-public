package xia

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/xia/xlog"
)

func TestWebService(t *testing.T) {
	x := New()
	x.GET("/xia", func(c *Context) {
		str := "<html><head></head><body><h2>我好喜欢仙剑五前啊！</h2></body></html>"
		c.HTML(http.StatusOK, str)
	})
	x.POST("/xia", testPost)

	g1 := x.Group("/api")
	g1.GET("*", testGroup)
	xlog.Debug("start")
	x.ServerStart()
}

func testPost(c *Context) {
	name := c.PostForm("username")
	fmt.Println(name)
	c.Status(http.StatusOK)
}

func testGroup(c *Context) {
	str := "<html><head></head><body><h2>Group test</h2></body></html>"
	a := []int{1}
	xlog.Debug(a[1])
	c.HTML(http.StatusOK, str)
}

func TestMiddleware(t *testing.T) {
	x := New()
	g1 := x.Group("/")
	g1.Use(TimeLogger)
	g2 := x.Group("/api")
	g2.GET("test/group", testGroup)
	xlog.Debug("start")
	x.ServerStart()
}
