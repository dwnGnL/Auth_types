package main

import(
	// "net/http"
	"jwtchat/routs"
	"jwtchat/db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
	"io"
	"jwtchat/models"
	"log"
	"libruary/utils"
)

func main()  {
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