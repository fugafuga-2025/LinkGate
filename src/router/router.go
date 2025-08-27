package router

import (
	"context"
	"encoding/json"
	"fmt"
	"fuagfuga-2025-LinkGate/src/service"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"fuagfuga-2025-LinkGate/src/usecase/line"
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
	// webhookのイベントをキャッチ
	r.POST("/linehook", func(c *gin.Context) {
		// リクエストボディを読み取る
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Println("読み取りエラー:", err)
			c.Status(http.StatusBadRequest)
			return
		}

		var obj map[string]interface{}
		err = json.Unmarshal(body, &obj)
		if err != nil {
			fmt.Println("ParseError:jsonの解析に失敗しました:", err)
		}

		// LINEbotがキャッチしたイベントの種類を取得する
		eventType := obj["events"].([]interface{})[0].(map[string]interface{})["type"]

		// ユーザーからのメッセージ送信だった場合
		if eventType == "message" {

			// mongoDBへメッセージ内容を保存する
			line.SaveLINEMessageToMongoDB(obj, collection)
		}

		// JSONをそのままコンソールに出力
		fmt.Println("==== Webhook JSON ====")
		fmt.Println(string(body))
		fmt.Println("======================")

		// 200 OK を返す
		c.Status(http.StatusOK)
	})
}
