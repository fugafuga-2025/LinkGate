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
var botToken = os.Getenv("DISCORD_BOT_TOKEN")

// メッセージ送信先となる Discord チャンネル ID。
var channelID = os.Getenv("DISCORD_CHANNEL_ID")

// discordMessage は Discord API へ送信する JSON ペイロードの形式です。
// 必要に応じてフィールドを拡張してください。
type discordMessage struct {
    Content string `json:"content"`
}

// CreateDiscordMessage は受け取ったメッセージを Discord の指定チャンネルに送信します。
//
// 例:
// from: 山田太郎さん
//
// こんにちは！
//
// (Platform: Slack)
//
// この関数は MongoDB への新規挿入イベントをフックして呼び出されることを想定しています。
func CreateDiscordMessage(msg model.Message) {
    // 認証情報が設定されていない場合は処理を中断
    if botToken == "" || channelID == "" {
        log.Println("DiscordのBotトークンまたはチャンネルIDが設定されていません")
        return
    }

    // 送信内容を作成
    content := fmt.Sprintf("from: %sさん\n\n%s\n\n(Platform: %s)", msg.User.Name, msg.Content.Text, msg.User.Platform)

    payload := discordMessage{
        Content: content,
    }
    body, err := json.Marshal(payload)
    if err != nil {
        log.Printf("DiscordメッセージのJSONエンコードに失敗しました: %v", err)
        return
    }

    // Discord API エンドポイント
    url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
    if err != nil {
        log.Printf("Discordリクエストの生成に失敗しました: %v", err)
        return
    }
    // 認証トークンをヘッダーに設定
    req.Header.Set("Authorization", "Bot "+botToken)
    req.Header.Set("Content-Type", "application/json")

    // HTTP クライアントを用意
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Discord API 呼び出しでエラーが発生しました: %v", err)
        return
    }
    defer resp.Body.Close()
    // ステータスコードチェック
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        // 失敗時はレスポンスボディを読み取りログ出力
        respBody, _ := io.ReadAll(resp.Body)
        log.Printf("Discord API からエラーコード %d が返されました: %s", resp.StatusCode, string(respBody))
        return
    }
    log.Println("Discord送信成功")
}
