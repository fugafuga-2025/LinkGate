package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	// ドキュメントID
	ID primitive.ObjectID `bson:"_id" json:"id"`
	// 投稿者情報
	User User `bson:"user" json:"user"`
	// 投稿内容
	Content Content `bson:"content" json:"content"`
	// 作成日時
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
