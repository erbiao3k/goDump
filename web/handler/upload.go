package handler

import (
	"github.com/gin-gonic/gin"
	"goDump/config"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Upload 使用方法：curl -X POST -F "upload[]=@/path/to/your/file" http://localhost:12399/upload
func Upload(c *gin.Context) {
	fileHeader, err := c.FormFile("upload[]")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"receive file error": err.Error()})
		return
	}

	// 创建文件
	dst := filepath.Join(config.ShareDir, fileHeader.Filename)
	out, err := os.Create(dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"create file error": err.Error()})
		return
	}
	defer out.Close()

	// 打开上传的文件
	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"open file error": err.Error()})
		return
	}
	defer src.Close()

	// 使用bufio优化内存分配和减少系统调用
	writer := out
	reader := src
	buffer := make([]byte, 1024*1024*3) // 3MB buffer

	// 流式写入文件
	_, err = io.CopyBuffer(writer, reader, buffer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"write file error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "file uploaded successfully"})
}
