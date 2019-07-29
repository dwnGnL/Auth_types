package main

import(
	// "net/http"
	"chat/routs"
	"chat/db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	// "github.com/gorilla/sessions"
	// "encoding/gob"
	"os"
	"io"
	"chat/models"
	"log"
	"libruary/utils"
)


func main()  {
	// gob.Register(sesKey(0))
	config := utils.ReadConfig()
	f, _ := os.OpenFile(config.LogName+".log", os.O_RDWR|os.O_CREATE, 0666)
	log.SetOutput(f)
	gin.DefaultWriter = io.MultiWriter(f)
	logger := logrus.New()
	logger.Level = logrus.TraceLevel
	logger.SetOutput(gin.DefaultWriter)
	db.Open(config.DbURI, logger)
 	defer db.GetDB().Close()
	r:=gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/auth", routs.Auth)
	r.POST("/checkLog", routs.CheckLog)
	r.GET("/chat",routs.Chat)
	r.GET("/ws", func(c *gin.Context) {
		routs.Wshandler(c.Writer, c.Request)
	})
		go routs.HandleMessages()
	r.Run(":"+config.Port)
	var history models.History
	db := db.GetDB()
	defer db.Delete(&history)
}