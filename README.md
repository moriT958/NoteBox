# NoteBox

ローカルでのメモ管理をしやすくするCLI。

## 誰におすすめか

- なるべくターミナル上で作業を完結させたい
- 日頃メモやノートをMarkdownで記録しており、ファイルがいろんな場所に分散してしまっている
- ローカルのPCで編集したメモをスマホでも確認したい
- Markdownで編集したメモをそのままWebページにしたい

## 特徴

- GitHubを利用したオンライン保存/同期
- メモ管理をCLIで完結
- メモのグルーピング
- メモの文字列検索機能
- ブラウザでのプレビュー
- Markdown形式でメモが取れる
- MarkdownからHTML、PukiWiki形式への変換
- 自由な設定が可能

## 使用方法(完成状況)

- `note new [title]`: ノートを新規作成
- `note ls`: メモ一覧を表示
- `note edit [id]`: メモを編集
- `note view <id>`: メモをプレビュー
- `note rm [id]`: メモを削除
- `note grep <string>`: ノートを文字列で検索
- `note config`: NoteBoxアプリの設定を編集。
- `note help`: コマンド一覧やコマンドのヘルプを表示
- `note upload`: GitHubへプッシュ

- `note box ls`: ノートのグループ(ボックス)一覧を表示
- `note box status <box-id>`: ボックス内のノート一覧を表示
- `note box add <id> <box-id>`: ノートをボックスに追加
- `note box out <id> <box-id>`: ノートをボックスから取り出す
