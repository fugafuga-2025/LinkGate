package model

type Attachment struct {
	// 添付ファイル種別
	Type string `bson:"type,omitempty" json:"type"`
	// 添付ファイルURL
	URL string `bson:"url,omitempty" json:"url"`
}
