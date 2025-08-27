package discord

// このパッケージは Discord ボットへのメッセージ送信を担当します。
// LINE ボットと同様に、MongoDB に登録された新規メッセージを
// Discord のチャンネルへ転送する用途を想定しています。

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"fuagfuga-2025-LinkGate/src/model"
)

// Discord API への認証に使用する Bot トークン。
// 環境変数 DISCORD_BOT_TOKEN に設定してください。
var botToken = os.Getenv("DISCORD_BOT_TOKEN")

// メッセージ送信先となる Discord チャンネル ID。
// 環境変数 DISCORD_CHANNEL_ID に設定してください。
var channelID = os.Getenv("DISCORD_CHANNEL_ID")

// CreateDiscordMessage はMongoDBに新規追加されたメッセージをDiscordへ転送します。
func CreateDiscordMessage(msg model.Message) {
	if botToken == "" {
		log.Println("DiscordのBotトークンが設定されていません")
		return
	}
	if channelID == "" {
		log.Println("DiscordのチャンネルIDが設定されていません")
		return
	}

	// プラットフォームごとにEmbedカラーを設定
	var colorInt int
	switch msg.User.Platform {
	case model.PlatformDiscord:
		colorInt = 0x5865F2 // Discordブランドカラー（紫）
	case model.PlatformLINE:
		colorInt = 0x00C300 // LINEブランドカラー（ライトグリーン）
	case model.PlatformSlack:
		colorInt = 0xFFFFFF // Slackは白
	default:
		colorInt = 0xCCCCCC // その他はグレー
	}

	// Embedペイロードを作成
	embedPayload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"color": colorInt,
				"author": map[string]interface{}{
					"name":     msg.User.Name,
					"icon_url": msg.User.IconUrl,
				},
				"description": msg.Content.Text,
			},
		},
	}

	body, err := json.Marshal(embedPayload)
	if err != nil {
		log.Printf("DiscordメッセージのJSONエンコードに失敗しました: %v", err)
		return
	}

	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Discordリクエストの生成に失敗しました: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bot "+botToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Discord API 呼び出しでエラーが発生しました: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("Discord API からエラーコード %d が返されました: %s", resp.StatusCode, string(respBody))
		return
	}
	log.Println("Discord送信成功 (Embed形式)")
}
