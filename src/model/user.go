package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	// ドキュメントID
	Id primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	// ユーザー名
	Name string `bson:"name,omitempty" json:"name"`
	// プラットフォーム
	Platform string `bson:"platform,omitempty" json:"platform"`
	// 管理者権限の有無
	IsAdmin bool `bson:"isAdmin,omitempty" json:"isAdmin"`
}
