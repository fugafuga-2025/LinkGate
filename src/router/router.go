package router

import (
	"context"
	"fuagfuga-2025-LinkGate/src/controller"
	"fuagfuga-2025-LinkGate/src/service"
	"fuagfuga-2025-LinkGate/src/usecase/slack"
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

	// DELETE /messages: 登録されている全ての投稿を削除します。
	r.DELETE("/messages", func(c *gin.Context) {
		service.DeleteAllMessage(c, collection)
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

	// === LINE API ===
	// === LINE API ===
	// webhookのイベントをキャッチ
	r.Any("/linehook", func(c *gin.Context) {
		controller.LINEController(c, collection)
	})

	// === SLACK API ===
	slackHandler := slack.NewSlackHandler(collection, ctx)
	r.POST("/slack/events", slackHandler.HandleSlackEvents)

	// Slackメッセージを取得するエンドポイント
	r.GET("/slack/messages", func(c *gin.Context) {
		messages, err := slackHandler.GetSlackMessages()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, messages)
	})
}
