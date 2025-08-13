#!/bin/bash
set -e

echo "🚀 開発環境をセットアップ中..."

# Goモジュール初期化
if [ ! -f "go.mod" ]; then
  echo "📦 Go moduleを初期化中..."
  go mod init fugafuga-2025-LinkGate
fi

# 基本的なGinの依存関係を追加
echo "📝 Go依存関係を追加中..."
go mod tidy
if ! grep -q "github.com/gin-gonic/gin" go.mod 2>/dev/null; then
  go get github.com/gin-gonic/gin
fi

# Swagger用の依存関係を追加
if ! grep -q "github.com/swaggo/gin-swagger" go.mod 2>/dev/null; then
  go get github.com/swaggo/gin-swagger
  go get github.com/swaggo/files
fi

# MongoDB driver追加
if ! grep -q "go.mongodb.org/mongo-driver" go.mod 2>/dev/null; then
  go get go.mongodb.org/mongo-driver/mongo
  go get go.mongodb.org/mongo-driver/bson
fi

# lefthookセットアップ
echo "🔧 lefthookを設定中..."
if [ ! -f ".git/hooks/pre-commit" ]; then
  lefthook install
fi

# go mod tidy to clean up dependencies
go mod tidy

echo "✅ セットアップ完了！"
echo "🔗 アプリケーションを起動するにはターミナルで: task dev"
echo "🌐 ブラウザで http://localhost:8080 にアクセスしてください"