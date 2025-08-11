package router

import (
	"context"
	"fuagfuga-2025-LinkGate/src/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(r *gin.Engine, collection *mongo.Collection, ctx context.Context, client *mongo.Client) {
	// ルート: API が動作しているか確認するための簡易メッセージ
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "LinkGate APIは起動中です🚀",
		})
	})

	// GET /messages: 登録されている全ての投稿を取得します。
	r.GET("/messages", func(c *gin.Context) {
		service.GetMessages(c, collection)
	})

	// POST /post: 新規投稿を作成します。
	r.POST("/post", func(c *gin.Context) {
		service.CreateMessage(c, collection)
	})

	// ヘルスチェック: サーバーと DB の状態をチェック
	r.GET("/health", func(c *gin.Context) {
		if err := client.Ping(ctx, nil); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "unhealthy",
				"database": "disconnected",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"database": "connected",
		})
	})

	// サーバーを起動
	if err := r.Run(":8080"); err != nil {
		log.Fatal("サーバーの起動に失敗🥺:", err)
	}
}
