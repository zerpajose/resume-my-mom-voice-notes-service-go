package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zerpajose/resume-my-mom-voice-notes-service-go/clients"
	"github.com/zerpajose/resume-my-mom-voice-notes-service-go/config"
	"github.com/zerpajose/resume-my-mom-voice-notes-service-go/controllers"
)

func main() {
	cfg := config.Load()
	r := gin.Default()

	ctx := context.Background()
	firestoreClient := clients.NewFirestoreClient(ctx, cfg.GoogleProjectID)
	storageClient := clients.NewStorageClient(ctx)

	r.POST("/upload-voice-note/:chatId", func(c *gin.Context) {
		chatId := c.Param("chatId")
		fileHeader, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
			return
		}
		err = controllers.UploadVoiceNote(
			c.Request.Context(),
			firestoreClient,
			storageClient,
			cfg.BucketName,
			cfg.CollectionName,
			chatId,
			fileHeader,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "uploaded"})
	})

	r.GET("/finish-thread/:chatId", func(c *gin.Context) {
		chatId := c.Param("chatId")

		result, err := controllers.FinishThread(
			c.Request.Context(),
			firestoreClient,
			storageClient,
			cfg.BucketName,
			cfg.CollectionName,
			cfg.GoogleProjectNumber,
			cfg.GeminiAPIKey,
			chatId,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"result": result})
	})

	r.Run(":" + cfg.Port)
}
