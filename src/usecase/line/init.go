package line

import (
	"fmt"
	"fuagfuga-2025-LinkGate/src/model"
	"log"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

var channelSecret = "a37672fceb5a6e39d7ed4aee215bfc9d"
var channelToken = "1cLQiY/aumnpIGCWg2Eme3r3to6UNhdFhFtYveh0d8P43ZcB10B+GxGfQrbV5Bqck7PKD6kwAKjoQNkdPFriu2hgjTMoTy727Ib+lIHVpu2gU+y8v7cCL4JX4a56zvaEAWX99jlfBS/l47kF+ZA8WwdB04t89/1O/w1cDnyilFU="
var testGroupID = "C56e7053ee8c56dbd99e94964216ce69c"
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
	text := msg.Content.Text
	platform := msg.User.Platform
	userName := msg.User.Name

	sendText := fmt.Sprintf("from: %sさん\n\n%s\n\n(Platform: %s)", userName, text, platform)
	log.Print(sendText)

	lineText := linebot.NewTextMessage(sendText)
	if _, err := bot.PushMessage(testGroupID, lineText).Do(); err != nil {
		if apiErr, ok := err.(*linebot.APIError); ok {
			fmt.Printf("LINE API Error: code=%d message=%s\n", apiErr.Code, apiErr.Response.Message)

			for i, detail := range apiErr.Response.Details {
				fmt.Printf("Detail %d: property=%s message=%s\n", i, detail.Property, detail.Message)
			}
		} else {
			fmt.Println("Other error:", err)
		}
	} else {
		log.Printf("LINE送信成功: %s", lineText.Text)
	}
}
