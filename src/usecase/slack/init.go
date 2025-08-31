package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"fuagfuga-2025-LinkGate/src/model"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SlackHandler struct {
	api        *slack.Client
	collection *mongo.Collection
	ctx        context.Context
}

// NewSlackHandler creates a new Slack handler
func NewSlackHandler(collection *mongo.Collection, ctx context.Context) *SlackHandler {
	token := os.Getenv("SLACK_BOT_TOKEN")
	if token == "" {
		log.Fatal("SLACK_BOT_TOKEN environment variable is required")
	}

	log.Printf("Slack Bot Token: %s...", token[:10]) // 最初の10文字だけ表示

	api := slack.New(token)

	// トークンの有効性をテスト
	_, err := api.AuthTest()
	if err != nil {
		log.Printf("Warning: Slack API auth test failed: %v", err)
		log.Printf("This might cause user info retrieval to fail")
	} else {
		log.Printf("Slack API authentication successful")
	}

	return &SlackHandler{
		api:        api,
		collection: collection,
		ctx:        ctx,
	}
}

// SlackMessage represents a Slack message stored in database
type SlackMessage struct {
	ID          string    `bson:"_id,omitempty" json:"id,omitempty"`
	ChannelID   string    `bson:"channel_id" json:"channel_id"`
	ChannelName string    `bson:"channel_name" json:"channel_name"`
	UserID      string    `bson:"user_id" json:"user_id"`
	UserName    string    `bson:"user_name" json:"user_name"`
	Message     string    `bson:"message" json:"message"`
	Timestamp   string    `bson:"timestamp" json:"timestamp"`
	SlackTS     string    `bson:"slack_ts" json:"slack_ts"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	Source      string    `bson:"source" json:"source"`
}

// HandleSlackEvents handles Slack event subscriptions
func (h *SlackHandler) HandleSlackEvents(c *gin.Context) {
	log.Printf("=== Slack Event Received ===")

	// Slackからのリクエストを検証
	verifier, err := slack.NewSecretsVerifier(c.Request.Header, os.Getenv("SLACK_SIGNING_SECRET"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// リクエストボディを読み取り
	body, err := c.GetRawData()
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// log.Printf("Request body: %s", string(body))

	// 署名を検証
	if _, err := verifier.Write(body); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := verifier.Ensure(); err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	// URL verification challengeの処理
	var challenge struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
	}

	if err := json.Unmarshal(body, &challenge); err == nil && challenge.Type == "url_verification" {
		c.JSON(http.StatusOK, gin.H{"challenge": challenge.Challenge})
		return
	}

	// イベントの処理 - 元のボディを直接パース
	h.handleSlackEvent(body)

	log.Printf("=== Slack Event Processing Complete ===")
	c.Status(http.StatusOK)
}

// handleSlackEvent processes Slack events from raw JSON
func (h *SlackHandler) handleSlackEvent(body []byte) {
	var eventWrapper struct {
		Type  string `json:"type"`
		Event struct {
			Type      string `json:"type"`
			Text      string `json:"text"`
			User      string `json:"user"`
			Channel   string `json:"channel"`
			Timestamp string `json:"ts"`
			BotID     string `json:"bot_id"`
		} `json:"event"`
	}

	if err := json.Unmarshal(body, &eventWrapper); err != nil {
		return
	}

	if eventWrapper.Type == "event_callback" {
		log.Printf("=== Processing Event Callback ===")

		if eventWrapper.Event.Type == "message" {
			// ボット自身のメッセージは無視
			if eventWrapper.Event.BotID != "" {
				return
			}

			// データベースに保存
			h.saveMessageToDatabase(eventWrapper.Event)
		} else {
			log.Printf("Non-message event type: %s", eventWrapper.Event.Type)
		}
	} else {
		log.Printf("Unsupported event wrapper type: %s", eventWrapper.Type)
	}
}

// saveMessageToDatabase saves Slack message to MongoDB
func (h *SlackHandler) saveMessageToDatabase(event struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	User      string `json:"user"`
	Channel   string `json:"channel"`
	Timestamp string `json:"ts"`
	BotID     string `json:"bot_id"`
}) {
	log.Printf("=== Getting User Info ===")
	log.Printf("User ID: %s", event.User)

	// ユーザー情報を取得
	user, err := h.api.GetUserInfo(event.User)

	userName := "unknown"
	iconURL := ""

	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		log.Printf("Error type: %T", err)
	} else if user != nil {
		log.Printf("User found: %+v", user)
		log.Printf("User Profile: %+v", user.Profile)

		// ユーザー名の優先順位: DisplayName > RealName > Name
		if user.Profile.DisplayName != "" {
			userName = user.Profile.DisplayName
		} else if user.Profile.RealName != "" {
			userName = user.Profile.RealName
		} else if user.Name != "" {
			userName = user.Name
		} else {
			userName = event.User // フォールバックとしてユーザーIDを使用
		}

		iconURL = user.Profile.Image72
		log.Printf("Selected username: %s", userName)
		log.Printf("Extracted icon URL: %s", iconURL)
	} else {
		log.Printf("User is nil")
	}

	// model.Message構造体に保存内容を格納
	var message model.Message
	message.ID = primitive.NewObjectID()
	message.User.ID = primitive.NewObjectID()
	message.User.UserID = event.User
	message.User.Platform = model.PlatformSlack
	message.User.Name = userName
	message.User.IconUrl = iconURL
	message.Content.ID = primitive.NewObjectID()
	message.Content.Text = event.Text
	message.CreatedAt = time.Now()

	log.Printf("Final message: %+v", message)

	// データベースに挿入（新しいコンテキストを使用）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := h.collection.InsertOne(ctx, message)
	if err != nil {
		log.Printf("Failed to save message to database: %v", err)
		return
	}

	log.Printf("Message saved to database with ID: %v", result.InsertedID)
}

// GetSlackMessages retrieves all Slack messages from database
func (h *SlackHandler) GetSlackMessages() ([]SlackMessage, error) {
	// 新しいコンテキストを作成（タイムアウトを設定）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := h.collection.Find(ctx, bson.M{"source": "slack"})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []SlackMessage
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// SendMessage sends a message to a Slack channel
func (h *SlackHandler) SendMessage(channelID, message string) error {
	_, _, err := h.api.PostMessage(channelID, slack.MsgOptionText(message, false))
	return err
}

// CreateSlackMessage はMongoDBに新規追加されたメッセージをSlackへ転送します。
func CreateSlackMessage(msg model.Message) {
	// Slack Incoming WebhookのURL
	webhookURL := "https://hooks.slack.com/services/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"

	// メッセージ送信者、内容、送信元プラットフォームを取得
	userName := msg.User.Name
	text := msg.Content.Text
	platform := msg.User.Platform

	// Slack用のメッセージペイロードを作成
	slackPayload := map[string]interface{}{
		"text":       fmt.Sprintf("from: %sさん\n\n%s\n\n(Platform: %s)", userName, text, platform),
		"username":   "LinkGate Bot",
		"icon_emoji": ":link:",
	}

	// JSONにエンコード
	body, err := json.Marshal(slackPayload)
	if err != nil {
		log.Printf("SlackメッセージのJSONエンコードに失敗しました: %v", err)
		return
	}

	// HTTPリクエストを作成
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Slackリクエストの生成に失敗しました: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// HTTPクライアントでリクエストを送信
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Slack Incoming Webhook呼び出しでエラーが発生しました: %v", err)
		return
	}
	defer resp.Body.Close()

	// レスポンスのステータスコードをチェック
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("Slack Incoming Webhookからエラーコード %d が返されました: %s", resp.StatusCode, string(respBody))
		return
	}

	log.Println("Slack送信成功")
}
