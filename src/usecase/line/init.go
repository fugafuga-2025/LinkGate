package line

import (
	"context"
	"fmt"
	"fuagfuga-2025-LinkGate/src/model"
	"log"
	"os"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// LINEbotの作成時に必要となる各種トークンの取得
var channelSecret = os.Getenv("LINE_CHANNEL_SECRET")
var channelToken = os.Getenv("LINE_CHANNEL_TOKEN")

// メッセージ送信先となるLINEグループのID
var groupID = os.Getenv("LINE_GROUP_ID")

var bot *linebot.Client

func createBot() {
	var err error

	bot, err = linebot.New(channelSecret, channelToken)
	if err != nil {
		fmt.Printf("LINE botの生成に失敗しました:%s", err)
	}
}

func CreateLINEMessage(msg model.Message) {
	createBot()

	// メッセージ送信者、内容、送信元プラットフォームを取得
	userName := msg.User.Name
	text := msg.Content.Text
	platform := msg.User.Platform

	lineText := linebot.NewTextMessage(fmt.Sprintf("from: %sさん\n\n%s\n\n(Platform: %s)", userName, text, platform))
	if _, err := bot.PushMessage(groupID, lineText).Do(); err != nil {
		if apiErr, ok := err.(*linebot.APIError); ok {
			fmt.Printf("LINE API Error: code=%d message=%s\n", apiErr.Code, apiErr.Response.Message)

			for i, detail := range apiErr.Response.Details {
				fmt.Printf("Detail %d: property=%s message=%s\n", i, detail.Property, detail.Message)
			}
		} else {
			fmt.Println("Other error:", err)
		}
	} else {
		log.Println("LINE送信成功")
	}
}

func SaveLINEMessageToMongoDB(obj map[string]interface{}, collection *mongo.Collection) {
	createBot()

	event := obj["events"].([]interface{})[0].(map[string]interface{})
	userId, _ := GetUserID(event)
	groupId, _ := GetGroupID(event)
	text, _ := GetMessageText(event)

	userName, err := GetUserName(bot, groupId, userId)
	if err != nil {
		println("ユーザー名の取得に失敗しました:", err)
	}

	var message model.Message

	// Message構造体に保存内容を格納
	message.ID = primitive.NewObjectID()
	message.User.ID = primitive.NewObjectID()
	message.User.UserID = userId
	message.User.Platform = "LINE"
	message.User.Name = userName
	message.Content.ID = primitive.NewObjectID()
	message.Content.Text = text
	message.CreatedAt = time.Now()

	// MongoDB にドキュメントを挿入
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := collection.InsertOne(ctx, message); err != nil {
		fmt.Println("データ登録に失敗しました:", err)
	}

	fmt.Printf("%s\n%s\n%s\n%s\n", userId, groupId, text, userName)

	fmt.Println(obj)

}

// イベントから userId を取得
func GetUserID(event map[string]interface{}) (string, bool) {
	sourceIface, ok := event["source"]
	if !ok {
		return "", false
	}

	source, ok := sourceIface.(map[string]interface{})
	if !ok {
		return "", false
	}

	userID, ok := source["userId"].(string)
	return userID, ok
}

// イベントから groupId を取得
func GetGroupID(event map[string]interface{}) (string, bool) {
	sourceIface, ok := event["source"]
	if !ok {
		return "", false
	}

	source, ok := sourceIface.(map[string]interface{})
	if !ok {
		return "", false
	}

	groupID, ok := source["groupId"].(string)
	return groupID, ok
}

// イベントから message.text を取得
func GetMessageText(event map[string]interface{}) (string, bool) {
	messageIface, ok := event["message"]
	if !ok {
		return "", false
	}

	message, ok := messageIface.(map[string]interface{})
	if !ok {
		return "", false
	}

	text, ok := message["text"].(string)
	return text, ok
}

// bot は事前に初期化されている *linebot.Client
func GetUserName(bot *linebot.Client, groupID, userID string) (string, error) {
	profile, err := bot.GetGroupMemberProfile(groupID, userID).Do()
	if err != nil {
		// 取得失敗時
		log.Printf("ユーザー情報取得失敗: %v", err)
		return "", err
	}
	// 名前を返す
	return profile.DisplayName, nil
}
