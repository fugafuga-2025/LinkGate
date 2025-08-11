package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// このファイルは LinkGate API のエントリーポイントです。
// MongoDB への接続を行い、簡単な REST API を提供します。
// データモデルは以下の構造体を参照してください。
//
// type User struct {
//     Name     string `bson:"name" json:"name"`
//     Platform string `bson:"platform" json:"platform"`
//     IsAdmin  bool   `bson:"isAdmin" json:"isAdmin"`
// }
//
// type Post struct {
//     ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
//     Contents  string             `bson:"contents" json:"contents"`
//     CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
//     User      User               `bson:"user" json:"user"`
// }

// User は投稿したユーザー情報を表します。
type User struct {
	Name     string `bson:"name" json:"name"`         // ユーザー名
	Platform string `bson:"platform" json:"platform"` // プラットフォーム名
	IsAdmin  bool   `bson:"isAdmin" json:"isAdmin"`   // 管理者権限の有無
}

// Post は投稿データを表します。
type Post struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`    // ドキュメントID
	Contents  string             `bson:"contents" json:"contents"`   // 投稿内容
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"` // 作成日時
	User      User               `bson:"user" json:"user"`           // 投稿者情報
}

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

	// ルート: API が動作しているか確認するための簡易メッセージ
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "LinkGate API is running!!",
		})
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

	// GET /posts: 登録されている全ての投稿を取得します。
	r.GET("/posts", func(c *gin.Context) {
		// タイムアウト付き context を作成
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// すべてのドキュメントを取得
		cur, err := collection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "データ取得に失敗しました"})
			return
		}
		defer cur.Close(ctx)
		var posts []Post
		if err := cur.All(ctx, &posts); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "データのパースに失敗しました"})
			return
		}
		c.JSON(http.StatusOK, posts)
	})

	// POST /posts: 新規投稿を作成します。
	r.POST("/posts", func(c *gin.Context) {
		var post Post
		// リクエスト JSON を構造体にバインド
		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "無効なリクエストです"})
			return
		}
		// ID と作成日時を設定
		post.ID = primitive.NewObjectID()
		post.CreatedAt = time.Now()
		// MongoDB にドキュメントを挿入
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := collection.InsertOne(ctx, post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "データ登録に失敗しました"})
			return
		}
		// 登録したドキュメントを返却
		c.JSON(http.StatusCreated, post)
	})

	// サーバーを起動
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
