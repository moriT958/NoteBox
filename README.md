# NoteBox

ローカルでのメモ管理をしやすくするCLI。

## 誰におすすめか

- なるべくターミナル上で作業を完結させたい
- Markdownで記録したメモやノートのファイルがいろんなディレクトリに分散してしまっている
- ローカルのPCで編集したメモをスマホでも確認したい

## 特徴

- メモ管理をCLIで完結
- Markdown形式でメモ・ノートテイキング
- VimやVSCodeなど好きなエディタで編集
- メモのグルーピング (未完成)
- メモの文字列検索機能 (未完成)
- Markdownのプレビュー (未完成)
- 自由な設定
- GitHubへのノートの自動プッシュ (未完成)
- MarkdownからPukiWiki形式への変換 (未完成)

## 使用方法(完成状況)

- 完成
  - `note new [title]`: ノートを新規作成
  - `note ls`: メモ一覧を表示
  - `note edit [id]`: メモを編集
  - `note rm [id]`: メモを削除
  - `note config`: NoteBoxアプリの設定を編集。
- 未完成
  - `note view <id>`: メモをプレビュー
  - `note grep <string>`: ノートを文字列で検索
  - `note help`: コマンド一覧やコマンドのヘルプを表示
  - `note upload`: GitHubへプッシュ
  - `note shelf ls`: ノートのグループ(ボックス)一覧を表示
  - `note shelf status <box-id>`: ボックス内のノート一覧を表示
  - `note shelf add <id> <box-id>`: ノートをボックスに追加
  - `note shelf out <id> <box-id>`: ノートをボックスから取り出す
