package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Content struct {
	// コンテンツID
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	// コンテンツ文章
	Text string `bson:"text,omitempty" json:"text"`
	// 添付ファイル配列
	Attachments []Attachment `bson:"attachments,omitempty" json:"attachments"`
}
