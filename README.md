# 📓 NoteBox 📓

ターミナルで動作するノート管理アプリ。

![](/assets/overview.png)

## 特徴

- [BubbleTea](https://github.com/charmbracelet/bubbletea)を使用した綺麗なTUI
- メモ管理をCLIで完結できる
- Vimライクなキーバインディング
- Markdown形式でメモ・ノートテイキング
- VimやVSCodeなど好きなエディタで編集
- Markdownで作成したノートのプレビュー
- Boxによるノートのグルーピング (未完成)
- メモの文字列検索機能 (未完成)
- レイアウトやキーバインドの自由な設定(未完成)
- レスポンシブデザイン(未完成)
- Box単位でノートをまとめてGit管理 (未完成)

## 誰におすすめか

- ターミナル上で作業を完結させたい方
- よくMarkdownでメモやノートを取る方
- ローカルのPCで編集したノートをスマホや別のPCでも確認したい

## 使用方法

`note`コマンドで起動

### キーバインド

- `j/k`または`↓/↑`: カーソルの上下移動
- `ctrl+l/ctrl+h`: プレビューワーへのフォーカス、ノートリストパネルへのフォーカスの切り替え
- `n`: ノートの新規作成
- `e`: ノートの編集(エディター起動)
- `d`: ノートの削除

### サブコマンド

- `note config`: 設定ファイルを編集
- `note help`: ヘルプの表示
- `note version`: バージョンを表示
- `note pl`(未完成): パス一覧(設定ファイルやマークダウンファイルなどを保存している場所)を表示

## インストール

### ソースからビルド(未完成)

1. `git clone https://github.com/moriT958/NoteBox`
2. `cd NoteBox`
3. `make build`

### brew(未完成)

`brew install notebox`(イメージ)

## 使用イメージ

### ノート作成時の画面

![](/assets/create-view.png)

### ノート削除時の画面

![](/assets/delete-view.png)

