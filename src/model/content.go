package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Content struct {
	// コンテンツID
	ID primitive.ObjectID `bson:"_id" json:"id"`
	// コンテンツ文章
	Text string `bson:"text" json:"text"`
	// 添付ファイル配列
	Attachments []Attachment `bson:"attachments" json:"attachments"`
}
