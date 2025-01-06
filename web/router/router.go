package router

import (
	"github.com/gin-gonic/gin"
	"goDump/web/handler"
	"log"
)

func Router() {
	r := gin.Default()

	r.MaxMultipartMemory = 30 << 30

	r.POST("/upload", handler.Upload)
	r.GET("/download/:filename", handler.Download)
	r.GET("/", handler.ViewGet)
	r.POST("/", handler.ViewPost)
	r.GET("/files", handler.ViewFiles)
	r.GET("/iteminfo", handler.ItemInfo)

	log.Println("Serving on 0.0.0.0:12399...")
	if err := r.Run("0.0.0.0:12399"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
