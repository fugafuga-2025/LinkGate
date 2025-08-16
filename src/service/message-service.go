package service

import (
	"context"
	"fuagfuga-2025-LinkGate/src/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// 全てのメッセージを取得
func GetMessages(c *gin.Context, collection *mongo.Collection) {
	// タイムアウト付き context を作成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// すべてのドキュメントを取得
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "データ取得に失敗しました", "details": err.Error()})
		return
	}
	defer cur.Close(ctx)
	var messages []model.Message
	if err := cur.All(ctx, &messages); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "データのパースに失敗しました", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}

// 新規メッセージを作成
func CreateMessage(c *gin.Context, collection *mongo.Collection) {
	var message model.Message

	// リクエスト JSON を構造体にバインド
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なリクエストです", "details": err.Error()})
		return
	}
	// ID と作成日時を設定
	message.ID = primitive.NewObjectID()
	message.Content.ID = primitive.NewObjectID()
	message.User.ID = primitive.NewObjectID()
	message.CreatedAt = time.Now()

	// MongoDB にドキュメントを挿入
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := collection.InsertOne(ctx, message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "データ登録に失敗しました", "details": err.Error()})
		return
	}
	// 登録したドキュメントを返却
	c.JSON(http.StatusCreated, message)
}

// 全てのメッセージを削除
func DeleteAllMessage(c *gin.Context, collection *mongo.Collection) {
	// タイムアウト付き context を作成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 全てのドキュメントを削除
	if _, err := collection.DeleteMany(ctx, bson.M{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "データ削除に失敗しました", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "全てのメッセージを削除しました"})
}
