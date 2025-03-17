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
- ブラウザでのプレビュー (未完成)
- 自由な設定が可能 (未完成)
- GitHubを利用したオンライン保存/同期 (未完成)
- MarkdownからHTML、PukiWiki形式への変換 (未完成)

## 使用方法(完成状況)

- `box new [title]`: ノートを新規作成
- `box ls`: メモ一覧を表示
- `box edit [id]`: メモを編集
- `box rm [id]`: メモを削除
- 未完成
  - `box view <id>`: メモをプレビュー
  - `box grep <string>`: ノートを文字列で検索
  - `box config`: NoteBoxアプリの設定を編集。
  - `box help`: コマンド一覧やコマンドのヘルプを表示
  - `box upload`: GitHubへプッシュ
  - `box shelf ls`: ノートのグループ(ボックス)一覧を表示
  - `box shelf status <box-id>`: ボックス内のノート一覧を表示
  - `box shelf add <id> <box-id>`: ノートをボックスに追加
  - `box shelf out <id> <box-id>`: ノートをボックスから取り出す
