package model

type Attachment struct {
	// 添付ファイル種別
	Type string `bson:"type" json:"type"`
	// 添付ファイルURL
	URL string `bson:"url" json:"url"`
}
