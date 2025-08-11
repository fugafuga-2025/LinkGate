package main

import (
	"context"
	"fuagfuga-2025-LinkGate/src/router"
	"log"
	"time"

	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// MongoDB 接続 URI を環境変数から取得します。設定されていない場合はデフォルトを使用します。
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		// docker-compose.yml で指定されている linkgate データベースを使用
		mongoURI = "mongodb://mongodb:27017/linkgate"
	}

	// MongoDB に接続するための context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// MongoDB クライアントの作成
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	// アプリ終了時にクライアントを必ず切断する
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	// 使用するデータベースとコレクションを選択
	db := client.Database("linkgate")
	collection := db.Collection("posts")

	// Gin エンジンを初期化
	r := gin.Default()

	// ルーティングのセットアップ
	router.SetupRoutes(r, collection, ctx, client)

}
