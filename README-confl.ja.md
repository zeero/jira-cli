# confl - Confluence CLI

Confluence REST API v2 を利用してページを取得・表示する CLI ツール。
jira-cli プロジェクトの拡張コマンドとして、同じ設定ファイル・認証情報を共有する。

## 📑 目次

1. [概要](#概要)
1. [インストール](#インストール)
1. [前提条件](#前提条件)
1. [コマンド仕様](#コマンド仕様)
   - [confl page view](#confl-page-view)
   - [グローバルフラグ](#グローバルフラグ)
1. [設定](#設定)
   - [設定ファイル](#設定ファイル)
   - [認証](#認証)
1. [アーキテクチャ](#アーキテクチャ)
   - [レイヤー構成](#レイヤー構成)
   - [ファイル一覧](#ファイル一覧)
   - [設計上のポイント](#設計上のポイント)
1. [開発](#開発)
   - [ビルド](#ビルド)
   - [テスト](#テスト)

---

## 概要

`confl` は Atlassian Confluence のページコンテンツを CLI から取得するためのツール。
jira-cli の設計パターンを踏襲し、同じ設定ファイル（`jira init` で生成）を共有する。

現時点では `page view` コマンドのみをサポートし、指定したページの Storage Format (XHTML) を標準出力に表示する。

## インストール

```bash
# リポジトリルートで実行
make install
```

`$GOPATH/bin` に `confl` バイナリがインストールされる。

## 前提条件

- `jira init` で設定ファイルが生成済みであること
- Jira API トークンが利用可能であること（Confluence も同じ Atlassian アカウントで認証する）
- Confluence が Jira と同じ Atlassian インスタンス上にあること（`server` 設定値を共有）

## コマンド仕様

### confl page view

Confluence ページを Storage Format (XHTML) で表示する。

```
confl page view <PAGE-ID>
```

| 項目 | 内容 |
|---|---|
| 引数 | `PAGE-ID` — Confluence ページ ID（数値） |
| API エンドポイント | `GET /wiki/api/v2/pages/{id}?body-format=storage` |
| 出力形式 | Storage Format (XHTML) をそのまま標準出力に出力 |
| 終了コード | 成功時 `0`、エラー時 `1` |

#### 使用例

```bash
# ページを表示
confl page view 12345

# デバッグモードで表示（HTTP リクエスト/レスポンスの詳細をダンプ）
confl --debug page view 12345

# 設定ファイルを指定
confl --config /path/to/config.yml page view 12345
```

#### エラーケース

| 状況 | 出力 |
|---|---|
| 存在しないページ ID | `confluence: unexpected response '404 Not Found'` |
| 認証失敗 | `confluence: unexpected response '401 Unauthorized'` |
| ページの body が空 | `confluence: page body is empty` |

### グローバルフラグ

| フラグ | 短縮形 | 説明 |
|---|---|---|
| `--config` | `-c` | 設定ファイルのパス（デフォルト: `~/.config/.jira/.config.yml`） |
| `--debug` | — | デバッグ出力を有効化（リクエスト/レスポンスのダンプ） |

## 設定

### 設定ファイル

`confl` は `jira` コマンドと同じ設定ファイルを使用する。
設定ファイルの生成は `jira init` で行う。

設定ファイルの探索順序:

1. `--config` フラグで指定されたパス
2. `JIRA_CONFIG_FILE` 環境変数
3. `~/.config/.jira/.config.yml`（デフォルト）

### 認証

以下の優先順位で API トークンを解決する（jira コマンドと同一）:

1. `JIRA_API_TOKEN` 環境変数
2. `.netrc` ファイル
3. OS キーリング（`jira-cli` サービス名）
4. 設定ファイルの `api_token` フィールド

対応する認証方式:

| 方式 | `auth_type` | 説明 |
|---|---|---|
| Basic 認証 | `basic` | メールアドレス + API トークン（Cloud のデフォルト） |
| Bearer (PAT) | `bearer` | Personal Access Token |
| mTLS | `mtls` | クライアント証明書認証 |

## アーキテクチャ

### レイヤー構成

```
cmd/confl/main.go                 エントリーポイント
    |
internal/cmd/confl/root/root.go   ルートコマンド（設定読み込み、認証チェック）
    |
internal/cmd/confl/page/          page サブコマンド群
    |
api/client.go                     ConfluenceClient ファクトリ（config 解決）
    |
pkg/confluence/                   Confluence REST API クライアント
```

### ファイル一覧

| パス | 役割 |
|---|---|
| [`cmd/confl/main.go`](cmd/confl/main.go) | エントリーポイント |
| [`internal/cmd/confl/root/root.go`](internal/cmd/confl/root/root.go) | ルートコマンド定義、設定初期化、認証チェック |
| [`internal/cmd/confl/page/page.go`](internal/cmd/confl/page/page.go) | `page` 親コマンド |
| [`internal/cmd/confl/page/view/view.go`](internal/cmd/confl/page/view/view.go) | `page view` サブコマンド |
| [`pkg/confluence/client.go`](pkg/confluence/client.go) | Confluence HTTP クライアント（認証・TLS 対応） |
| [`pkg/confluence/types.go`](pkg/confluence/types.go) | API レスポンス型定義 |
| [`pkg/confluence/page.go`](pkg/confluence/page.go) | `GetPage()` API メソッド |
| [`api/client.go`](api/client.go) | `ConfluenceClient()` ファクトリ + `resolveConfig()` 共通ヘルパー（変更） |

### 設計上のポイント

#### jira-cli との共有

`confl` は独立したバイナリだが、以下を jira-cli と共有する:

- **設定ファイル**: `jira init` で生成される `~/.config/.jira/.config.yml`
- **認証情報解決ロジック**: `api/client.go` の `resolveConfig()` ヘルパー
- **型定義**: `jira.Config`, `jira.Header`, `jira.AuthType`
- **ユーティリティ**: `internal/cmdutil`（スピナー、エラーハンドリング等）

#### Confluence クライアントの独立性

`pkg/confluence/` パッケージは `pkg/jira/` とは独立した HTTP クライアントを持つ。
これは `jira.Client.request()` メソッドが unexported であり外部パッケージから呼べないため。
ただし、transport 構築パターン（TLS 1.2+, ProxyFromEnvironment）と認証ヘッダー設定ロジックは同一パターンを踏襲している。

#### URL 構築

`server` 設定値（例: `https://example.atlassian.net`）に `/wiki/api/v2` を付与して Confluence API のベース URL とする。
Atlassian Cloud では Jira と Confluence が同一ドメイン上に存在するため、URL 変換処理は不要。

## 開発

### ビルド

```bash
make build      # jira と confl の両方をビルド
make install    # $GOPATH/bin にインストール
```

### テスト

```bash
make lint       # golangci-lint を実行
make test       # レースディテクタ付きテスト
make ci         # lint + test
```
