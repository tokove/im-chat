package router

import (
	"backend/internal/middleware"
	"backend/pkg/response"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxFileSize = 50 << 20
)

func handleFileUpload(c *gin.Context) {
	senderIDAny, exists := c.Get(middleware.CtxUserID)
	if !exists {
		log.Print("handleFileUpload: senderID not exists, not login")
		response.JSON(c, http.StatusUnauthorized, false, "Unauthorized", nil)
		return
	}
	senderID, ok := senderIDAny.(int64)
	if !ok {
		log.Print("handleFileUpload: senderID type error")
		response.JSON(c, http.StatusUnauthorized, false, "Unauthorized", nil)
		return
	}

	privateIDStr := c.Param("private_id")
	privateID, err := strconv.ParseInt(privateIDStr, 10, 64)
	if err != nil {
		log.Printf("handleFileUpload: parseInt privateID failed, privateIDStr: %v, err: %v", privateIDStr, err)
		response.JSON(c, http.StatusBadRequest, false, "missing private_id", nil)
		return
	}

	if err := c.Request.ParseMultipartForm(maxFileSize); err != nil {
		log.Print("handleFileUpload: file size is too large")
		response.JSON(c, http.StatusInternalServerError, false, "file upload failed", nil)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("handleFileUpload: FormFile err: %v", err)
		if strings.Contains(err.Error(), "http: request body too large") {
			response.JSON(c, http.StatusRequestEntityTooLarge, false, fmt.Sprintf("file size is not allowed to exceed %d", maxFileSize), nil)
		} else {
			response.JSON(c, http.StatusInternalServerError, false, "file upload failed", nil)
		}
		return
	}

	originFileName := filepath.Base(file.Filename)
	ext := strings.ToLower(filepath.Ext(originFileName))

	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".pdf": true, ".doc": true, ".docx": true, ".txt": true,
		".mp4": true, ".mov": true, ".mp3": true,
	}

	if !allowedExts[ext] {
		response.JSON(c, http.StatusBadRequest, false, "file type not allowed", nil)
		return
	}

	uniqueFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dirPath := filepath.Join("files", "chats", fmt.Sprintf("%d", privateID), fmt.Sprintf("%d", senderID))

	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		log.Printf("handleFileUpload: MkdirAll err: %v", err)
		response.JSON(c, http.StatusInternalServerError, false, "failed to create dir", nil)
		return
	}

	filePath := filepath.Join(dirPath, uniqueFileName)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log.Printf("handleFileUpload: SaveUploadedFile err: %v", err)
		response.JSON(c, http.StatusInternalServerError, false, "failed to save file", nil)
		return
	}

	fileUrl := fmt.Sprintf("/files/chats/%d/%d/%s", privateID, senderID, uniqueFileName)
	response.JSON(c, http.StatusOK, true, "success", fileUrl)
}
