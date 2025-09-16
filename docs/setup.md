# LinkGate

## 📦 使用技術

- [Go 1.24](https://go.dev/)
- [Dev Containers](https://containers.dev/)
- 拡張機能（自動インストール）:
  - Go
  - GitLens
  - TODO Tree
  - Error Lens
  - Code Spell Checker
  - 日本語言語パック

---

## 🛠️ 前提条件

以下をインストールしておいてください。

- [Visual Studio Code](https://code.visualstudio.com/)
- [Dev Containers 拡張機能](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
- [Docker](https://www.docker.com/)

---

## 🚀 開発環境の立ち上げ手順

### 1. リポジトリをクローン

```bash
git clone https://github.com/fugafuga-2025/LinkGate.git
```

```bash
cd LinkGate
```

### 2. 開発環境の立ち上げ

1. Docker Desktop を立ち上げる

2. 右下に表示される`コンテナーで再度開く`を押すか、
   VS Code 上でコマンドパレットを開き、`Dev Containers: Reopen in Container` を選択
3. `cp -n .env.example .env`で環境変数ファイルをコピー
4. `task dev`を実行

初回起動時は Docker イメージのビルドや拡張機能のインストールに数分かかります。

---

## 📝 Taskfile の使用方法

本プロジェクトでは、タスクランナーとして [Task](https://taskfile.dev/) を使用しています。

### 利用可能なタスク一覧を表示

```bash
task
```

### 主要なタスク

#### 開発関連

- `task dev` - ホットリロード付きで開発サーバーを起動
- `task run -- ファイル` - アプリケーションを実行
- `task build -- ファイル` - アプリケーションをビルド

#### コード品質

- `task fmt` - コードをフォーマット
- `task lint` - リンターを実行
- `task test` - テストを実行
- `task check` - 全てのコード品質チェックを実行

## 💬 使えるコメントの説明

本プロジェクトでは、`TODO Tree` 拡張機能を利用して、コード内のコメントを一覧表示できます。以下のようなコメントをコード中に記述することで、タスクや注意事項を簡単に管理できます。

- `// TODO: やるべきこと`
- `// FIXME: 修正が必要な箇所`
- `// NOTE: 補足情報や注意点`
- `// BUG: バグの報告`

これらのキーワードを含むコメントは、ファイル内でハイライト、自動的に検出され、一覧表示されます。
