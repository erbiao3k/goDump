package handler

import (
	"github.com/gin-gonic/gin"
	"goDump/config"
	"net/http"
	"os"
	"path/filepath"
)

func Download(c *gin.Context) {
	// 检查文件是否存在
	filePath := filepath.Join(config.ShareDir, c.Param("filename"))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.Status(http.StatusNotFound)
		c.Writer.Write([]byte("File not found"))
		return
	}
	// 设置文件下载
	c.File(filePath)
}
