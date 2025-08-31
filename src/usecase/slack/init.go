package slack

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"go.mongodb.org/mongo-driver/bson"
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

	return &SlackHandler{
		api:        slack.New(token),
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
	log.Printf("Method: %s", c.Request.Method)
	log.Printf("URL: %s", c.Request.URL.String())

	// Slackからのリクエストを検証
	verifier, err := slack.NewSecretsVerifier(c.Request.Header, os.Getenv("SLACK_SIGNING_SECRET"))
	if err != nil {
		log.Printf("Failed to create secrets verifier: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}

	// リクエストボディを読み取り
	body, err := c.GetRawData()
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}

	log.Printf("Request body: %s", string(body))

	// 署名を検証
	if _, err := verifier.Write(body); err != nil {
		log.Printf("Failed to write body to verifier: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}

	if err := verifier.Ensure(); err != nil {
		log.Printf("Invalid signature: %v", err)
		c.Status(http.StatusUnauthorized)
		return
	}

	// URL verification challengeの処理
	var challenge struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
	}

	if err := json.Unmarshal(body, &challenge); err == nil && challenge.Type == "url_verification" {
		log.Printf("URL verification challenge received: %s", challenge.Challenge)
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
		log.Printf("Failed to parse event wrapper: %v", err)
		return
	}

	log.Printf("Event wrapper type: %s", eventWrapper.Type)

	if eventWrapper.Type == "event_callback" {
		log.Printf("=== Processing Event Callback ===")
		log.Printf("Inner event type: %s", eventWrapper.Event.Type)

		if eventWrapper.Event.Type == "message" {
			// ボット自身のメッセージは無視
			if eventWrapper.Event.BotID != "" {
				log.Printf("Ignoring bot message (BotID: %s)", eventWrapper.Event.BotID)
				return
			}

			log.Printf("=== Slack Message Received ===")
			log.Printf("Channel: %s", eventWrapper.Event.Channel)
			log.Printf("User: %s", eventWrapper.Event.User)
			log.Printf("Message: %s", eventWrapper.Event.Text)
			log.Printf("Timestamp: %s", eventWrapper.Event.Timestamp)
			log.Printf("=============================")

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
	// 新しいコンテキストを作成（タイムアウトを設定）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ユーザー情報を取得
	user, err := h.api.GetUserInfo(event.User)

	userName := "unknown"
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
	} else if user != nil {
		userName = user.Name
	}

	log.Printf("=== Slack Message Received ===")
	// log.Printf("Channel: %s", eventWrapper.Event.Channel)
	log.Printf("User: %s", userName)
	log.Printf("Message: %s", event.Text)
	log.Printf("Timestamp: %s", event.Timestamp)
	log.Printf("=============================")

	// SlackMessage構造体を作成
	slackMessage := SlackMessage{
		UserID:    event.User,
		UserName:  userName,
		Message:   event.Text,
		Timestamp: time.Now().Format(time.RFC3339),
		SlackTS:   event.Timestamp,
		CreatedAt: time.Now(),
		Source:    "slack",
	}

	// データベースに挿入（新しいコンテキストを使用）
	result, err := h.collection.InsertOne(ctx, slackMessage)
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

// handleMessageEvent processes message events
func (h *SlackHandler) handleMessageEvent(event *slack.MessageEvent) {
	// ボット自身のメッセージは無視
	if event.BotID != "" {
		return
	}

	// メッセージの詳細を取得
	channel, err := h.api.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID: event.Channel,
	})
	if err != nil {
		log.Printf("Failed to get channel info: %v", err)
		return
	}

	user, err := h.api.GetUserInfo(event.User)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		return
	}

	// メッセージ情報をログ出力
	log.Printf("=== Slack Message Received ===")
	log.Printf("Channel: %s (%s)", channel.Name, channel.ID)
	log.Printf("User: %s (%s)", user.Name, user.ID)
	log.Printf("Message: %s", event.Text)
	log.Printf("Timestamp: %s", event.Timestamp)
	log.Printf("=============================")

	// ここでメッセージの処理ロジックを追加
	// 例: データベースに保存、他のサービスに転送など
}

// SendMessage sends a message to a Slack channel
func (h *SlackHandler) SendMessage(channelID, message string) error {
	_, _, err := h.api.PostMessage(channelID, slack.MsgOptionText(message, false))
	return err
}
