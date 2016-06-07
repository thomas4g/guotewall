// vim: noet sts=2 ts=2 sw=2
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/mgo.v2"
)

type Quote struct {
	Text   string
	Author string
	TS     time.Time
}

type NewQuoteForm struct {
	Quote  string `form:"quote" bind:"required"`
	Author string `form:"author" bind:"required"`
}

func MongoConnector() gin.HandlerFunc {
	return func(c *gin.Context) {
		// before request
		session, err := mgo.Dial("localhost")
		if err != nil {
			panic(err)
		}
		defer session.Close()
		conn := session.DB("test")

		c.Set("dbConn", conn)

		// dispatch request
		c.Next()

		// after request
		// - NOTHING YET -
	}
}

func quotesGet(c *gin.Context) {
	conn := c.MustGet("dbConn").(*mgo.Database)

	quotes := []Quote{}
	err := conn.C("quotes").Find(nil).Sort("-time").All(&quotes)
	if err != nil {
		log.Fatal(err)
		c.String(http.StatusNotFound, err.Error())
	}
	tmpl_obj := gin.H{"quotes": quotes}
	c.HTML(http.StatusOK, "home.tmpl", tmpl_obj)
}

func uploadGet(c *gin.Context) {
	c.HTML(http.StatusOK, "upload.tmpl", gin.H{})
}

func uploadPost(c *gin.Context) {
	conn := c.MustGet("dbConn").(*mgo.Database)
	var form NewQuoteForm

	c.BindWith(&form, binding.Form)
	err := conn.C("quotes").Insert(&Quote{form.Quote, form.Author, time.Now()})
	if err != nil {
		log.Fatal(err)
		c.String(http.StatusInternalServerError,
			"Something went wrong when saving to Mongo!")
	}
	c.Redirect(http.StatusMovedPermanently, "/")
}

func main() {
	r := gin.Default()

	r.Use(MongoConnector())
	r.LoadHTMLGlob("templates/*")

	r.GET("/", quotesGet)
	r.GET("/upload", uploadGet)
	r.POST("/upload", uploadPost)

	// Listen and serve on 0.0.0.0:8080
	gin.SetMode(gin.ReleaseMode)
	r.Run("127.0.0.1:8080")
}
