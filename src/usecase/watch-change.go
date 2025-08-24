package usecase

import (
	"context"
	"fmt"
	"fuagfuga-2025-LinkGate/src/model"
	"fuagfuga-2025-LinkGate/src/usecase/line"
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

		if platform != model.PlatformLINE {
			if event.OperationType == "insert" {
				line.CreateLINEMessage(fullDoc)
			}
		}

		// コンソール通知
		fmt.Printf("📢 DBに変更がありました: %+v\n", event)
	}

	if err := stream.Err(); err != nil {
		log.Println("Stream error:", err)
	}
}
