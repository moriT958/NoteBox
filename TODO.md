# TODO

- CRUDの実装

  - [x] `note new [title]`: ノートを新規作成
  - [x] `note ls`: メモ一覧を表示
  - [x] `note edit [id]`: メモを編集
  - [x] `note rm [id]`: メモを削除
  - [ ] `NoteStore`オブジェクトの実装

- その他コマンドの実装(MVP)

  - [ ] `note grep <string>`: ノートを文字列で検索
  - [ ] `note config`: NoteBoxアプリの設定を編集。
  - [ ] `note help`: コマンド一覧やコマンドのヘルプを表示
  - [ ] `note view <id>`: メモをプレビュー
  - [ ] `note upload`: GitHubへプッシュ

- 機能・方向性の見直し

  - [ ] Qiita書き始める
  - [ ] 他の類似ツールと比較

- Box機能の追加

  - [ ] `note box ls`: ノートのグループ(ボックス)一覧を表示
  - [ ] `note box status <box-id>`: ボックス内のノート一覧を表示
  - [ ] `note box add <id> <box-id>`: ノートをボックスに追加
  - [ ] `note box out <id> <box-id>`: ノートをボックスから取り出す

- より良くするために
  - [ ] TUIの強化(lipgloss使う)
  - [ ] リファクタリング・アーキテクチャ
  - [ ] SQLite3の検討
  - [ ] パフォーマンス改善
  - [ ] ロガーの作成
