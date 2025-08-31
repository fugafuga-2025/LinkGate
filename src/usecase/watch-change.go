package usecase

import (
	"context"
	"fmt"
	"fuagfuga-2025-LinkGate/src/model"
	"fuagfuga-2025-LinkGate/src/usecase/discord"
	"fuagfuga-2025-LinkGate/src/usecase/line"
	"fuagfuga-2025-LinkGate/src/usecase/slack"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

// dbの変更を監視するための関数
func WatchChanges(coll *mongo.Collection) {
	ctx := context.Background()

	// 監視パイプライン（ここでは全部の変更を監視）
	pipeline := mongo.Pipeline{}

	stream, err := coll.Watch(ctx, pipeline)
	if err != nil {
		log.Fatal("Change Stream エラー:", err)
	}
	defer stream.Close(ctx)

	fmt.Println("🔍 Change stream 開始...")

	for stream.Next(ctx) {
		var event struct {
			OperationType string        `bson:"operationType"`
			FullDocument  model.Message `bson:"fullDocument"`
		}
		if err := stream.Decode(&event); err != nil {
			log.Println("Decode error:", err)
			continue
		}

		fullDoc := event.FullDocument
		platform := fullDoc.User.Platform

		// 新規メッセージ挿入時に各プラットフォームへ転送します。
		if event.OperationType == "insert" {
			// 元プラットフォームがLINEでなければLINEへ送信
			if platform != model.PlatformLINE {
				line.CreateLINEMessage(fullDoc)
			}
			// 元プラットフォームがDiscordでなければDiscordへ送信
			if platform != model.PlatformDiscord {
				discord.CreateDiscordMessage(fullDoc)
			}
			// 元プラットフォームがSlackでなければSlackへ送信
			if platform != model.PlatformSlack {
				slack.CreateSlackMessage(fullDoc)
			}
		}

		// コンソール通知
		fmt.Printf("📢 DBに変更がありました: %+v\n", event)
	}

	if err := stream.Err(); err != nil {
		log.Println("Stream error:", err)
	}
}
