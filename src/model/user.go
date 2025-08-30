package model

import (
	"encoding/json"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	// ID
	ID primitive.ObjectID `bson:"_id" json:"_id"`
	// ユーザーID
	UserID string `bson:"userId" json:"userId"`
	// ユーザー名
	Name string `bson:"name" json:"name"`
	// プラットフォーム
	Platform Platform `bson:"platform" json:"platform"`
	// アイコンURL
	IconUrl string `bson:"iconUrl" json:"iconUrl"`
}

var allowedPlatforms = map[Platform]struct{}{
	PlatformLINE:    {},
	PlatformDiscord: {},
	PlatformSlack:   {},
}

// プラットフォームのバリデーション
func (p *Platform) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	platform := Platform(s)
	if _, ok := allowedPlatforms[platform]; !ok {
		return errors.New("platformの値が無効です: " + s)
	}
	*p = platform
	return nil
}
