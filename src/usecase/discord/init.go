package discord

// このパッケージは Discord ボットへのメッセージ送信と受信を担当します。
// LINE ボットと同様に、MongoDB に登録された新規メッセージを
// Discord のチャンネルへ転送し、Discord からのメッセージを監視してDBに保存します。

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"fuagfuga-2025-LinkGate/src/model"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Discord API への認証に使用する Bot トークン。
// 環境変数 DISCORD_BOT_TOKEN に設定してください。
var botToken = os.Getenv("DISCORD_BOT_TOKEN")

// メッセージ送信先となる Discord チャンネル ID。
// 環境変数 DISCORD_CHANNEL_ID に設定してください。
var channelID = os.Getenv("DISCORD_CHANNEL_ID")

// Discord セッション
var session *discordgo.Session

var mongoCollection *mongo.Collection

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

func StartDiscordBot(collection *mongo.Collection) error {
	if botToken == "" {
		log.Println("DiscordのBotトークンが設定されていません")
		return fmt.Errorf("DISCORD_BOT_TOKEN is not set")
	}
	if channelID == "" {
		log.Println("DiscordのチャンネルIDが設定されていません")
		return fmt.Errorf("DISCORD_CHANNEL_ID is not set")
	}

	mongoCollection = collection

	var err error
	session, err = discordgo.New("Bot " + botToken)
	if err != nil {
		log.Printf("Discordセッションの作成に失敗しました: %v", err)
		return err
	}

	session.AddHandler(messageCreate)

	session.Identify.Intents = discordgo.IntentsGuildMessages

	err = session.Open()
	if err != nil {
		log.Printf("Discordセッションのオープンに失敗しました: %v", err)
		return err
	}

	log.Println("🔍 Discord bot started and monitoring messages...")
	return nil
}

func CloseDiscordBot() {
	if session != nil {
		session.Close()
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.ChannelID != channelID {
		return
	}

	if m.Type != discordgo.MessageTypeDefault {
		return
	}

	SaveDiscordMessageToMongoDB(s, m)
}

func SaveDiscordMessageToMongoDB(s *discordgo.Session, m *discordgo.MessageCreate) {
	if mongoCollection == nil {
		log.Println("MongoDB collection is not initialized")
		return
	}

	userName := m.Author.Username
	if m.Member != nil && m.Member.Nick != "" {
		userName = m.Member.Nick
	}

	iconURL := m.Author.AvatarURL("64")

	var message model.Message

	// Message構造体に保存内容を格納
	message.ID = primitive.NewObjectID()
	message.User.ID = primitive.NewObjectID()
	message.User.UserID = m.Author.ID
	message.User.Platform = model.PlatformDiscord
	message.User.Name = userName
	message.User.IconUrl = iconURL
	message.Content.ID = primitive.NewObjectID()
	message.Content.Text = strings.TrimSpace(m.Content)
	message.CreatedAt = time.Now()

	// MongoDB にドキュメントを挿入
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := mongoCollection.InsertOne(ctx, message); err != nil {
		log.Printf("Discordメッセージのデータ登録に失敗しました: %v", err)
		return
	}

	log.Printf("Discord message saved: %s from %s", message.Content.Text, userName)
}

func InitializeDiscordBot(collection *mongo.Collection) {
	if err := StartDiscordBot(collection); err != nil {
		log.Printf("Discord ボットの起動に失敗しました: %v", err)
		return
	}

	log.Println("🔍 Discord bot initialized successfully")
	
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	
	<-stop
	log.Println("🛑 Discord bot shutting down gracefully...")
	CloseDiscordBot()
}