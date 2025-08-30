package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"fuagfuga-2025-LinkGate/src/usecase/line"
)

func LINEController(c *gin.Context, collection *mongo.Collection) {
	// リクエストボディを読み取る
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("読み取りエラー:", err)
		c.Status(http.StatusBadRequest)
		return
	}

	var obj map[string]interface{}
	err = json.Unmarshal(body, &obj)
	if err != nil {
		fmt.Println("ParseError:jsonの解析に失敗しました:", err)
	}

	// LINEbotがキャッチしたイベントの種類を取得する
	eventType := obj["events"].([]interface{})[0].(map[string]interface{})["type"]

	// ユーザーからのメッセージ送信だった場合
	if eventType == "message" {

		// mongoDBへメッセージ内容を保存する
		line.SaveLINEMessageToMongoDB(obj, collection)
	}

	// JSONをそのままコンソールに出力
	fmt.Println("==== Webhook JSON ====")
	fmt.Println(string(body))
	fmt.Println("======================")

	// 200 OK を返す
	c.Status(http.StatusOK)
}
