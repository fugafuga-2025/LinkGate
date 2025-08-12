package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	// ドキュメントID
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	// 投稿内容
	Contents string `bson:"contents" json:"contents"`
	// 作成日時
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	// 投稿者情報
	User User `bson:"user" json:"user"`
}
