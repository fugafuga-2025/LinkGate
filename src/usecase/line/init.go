package line

import (
	"fmt"
	"fuagfuga-2025-LinkGate/src/model"
	"log"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"
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
