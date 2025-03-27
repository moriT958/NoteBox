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
  - `notebox new [title]`: ノートを新規作成
  - `notebox ls`: メモ一覧を表示
  - `notebox edit [id]`: メモを編集
  - `notebox rm [id]`: メモを削除
  - `notebox config`: NoteBoxアプリの設定を編集。
- 未完成
  - `notebox view <id>`: メモをプレビュー
  - `notebox grep <string>`: ノートを文字列で検索
  - `notebox help`: コマンド一覧やコマンドのヘルプを表示
  - `notebox upload`: GitHubへプッシュ
  - `notebox shelf ls`: ノートのグループ(ボックス)一覧を表示
  - `notebox shelf status <box-id>`: ボックス内のノート一覧を表示
  - `notebox shelf add <id> <box-id>`: ノートをボックスに追加
  - `notebox shelf out <id> <box-id>`: ノートをボックスから取り出す
