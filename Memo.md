# ファイルシステム監視による複数プロセス間のノート同期

notebox を別ペインで複数プロセス起動していても、各プロセス間でノートの状態が共有されるようにしたい。

## WIP 引き継ぎ

- 問題:
  - fsnotify による同期シグナルが送信されるたびに並び替えが発生してしまう
  - 新規追加したノートは一番下に追加されて欲しい + 勝手に位置が変わっていやだ
  - ファイル名の先頭に連番し常に順番が保証されるようにする
  - ただ、既存ファイルのマイグレーションがめんどい
  - ツールを起動するたびにマイグレーションが走ると起動速度低下が懸念。

- 対策案
  - `~/.notebox/config.json` に `version` フィールドを追加する
  - インストールする時のバージョンと設定ファイルのバージョンが一致しない場合のみマイグレーション実行
  - インストール完了後に設定ファイルのバージョンを書き換える

  - もしくは、そこまで速度的に大きな影響もなさそうだったら毎度実行でも別にいいかもな

## 背景・目的

NoteBox を複数ターミナルペインで同時に開いている場合、一方でノートを作成・更新・削除しても、もう片方のプロセスには反映されない。`fsnotify` でノートディレクトリを監視し、全プロセスで自動的にリストとプレビューを同期する。

## 設計方針

- `fsnotify` でノートディレクトリを監視し、BubbleTea のメッセージループに統合する
- BubbleTea の標準パターン: Cmd がブロック → Msg を返す → Update が処理して再度 Cmd を発行
- `tea.Tick` によるデバウンス（200ms）で連続イベントをまとめる
- カーソル位置はファイルパスで照合して保持する

## ファイル名形式の変更

新規作成ノートが常にリスト末尾に来るよう、ファイル名先頭に連番を付与する。
`filepath.Walk` の辞書順がそのまま作成順になる。編集してもファイル名は変わらないため位置は安定。

- **旧:** `{title}-{YYYY-MM-DD}.md` (例: `hello-2026-02-11.md`)
- **新:** `{seq:04d}-{title}-{YYYY-MM-DD}.md` (例: `0001-hello-2026-02-11.md`)
- マイグレーション: `notebox migrate` CLI サブコマンドで実行（起動時の自動実行はしない）

## 変更ファイル一覧

| ファイル                               | 変更内容                                                                 |
| -------------------------------------- | ------------------------------------------------------------------------ |
| `go.mod` / `go.sum`                    | `github.com/fsnotify/fsnotify` 追加                                      |
| `internal/tui/file_operations.go`      | `GetNextSeq`, `HasSeqPrefix` をエクスポート、`getTitleFromFilename` 修正 |
| `internal/tui/file_operations_test.go` | 連番関連テスト追加、既存テスト修正                                       |
| `internal/cli/migrate_command.go`      | 新規: `notebox migrate` サブコマンド                                     |
| `internal/cli/migrate_command_test.go` | 新規: マイグレーションテスト                                             |
| `internal/cli/register.go`             | `migrateCmd` 登録追加                                                    |
| `internal/tui/messages.go`             | メッセージ型 4 つ、コマンド 2 つ追加、`createNewNoteCmd` 連番対応        |
| `internal/tui/messages_test.go`        | `TestReloadNotesCmd` 追加                                                |
| `internal/tui/model.go`                | watcher フィールド追加、Init/Update 拡張、Close() 追加                   |
| `internal/tui/listPanel.go`            | `calcSyncNoteList` 純粋関数、`syncNoteList` メソッド追加                 |
| `internal/tui/listPanel_test.go`       | `TestCalcSyncNoteList` 追加                                              |
| `internal/tui/watcher_test.go`         | 新規: `TestWatchNotesCmd`                                                |
| `cmd/notebox/main.go`                  | `defer m.Close()` 追加                                                   |

## エッジケース

| ケース                 | 挙動                                                                       |
| ---------------------- | -------------------------------------------------------------------------- |
| 自プロセスのイベント   | fsnotify で再リロードされるが同じ結果になるだけで問題なし                  |
| 選択中ノートが外部削除 | カーソルが同じインデックスにクランプ。新しい位置のノートのプレビューを表示 |
| 末尾ノートが外部削除   | カーソルが `len-1` にクランプ                                              |
| 全ノート削除           | カーソル 0、ダミーノート表示                                               |
| モーダル表示中の変更   | リストは背後で更新。モーダル操作に影響なし                                 |
| 連番の競合（同時作成） | `os.OpenFile` + `O_EXCL` で重複検出しリトライ                              |

---

## 実装フェーズ（各フェーズ = 1 TDD サイクル）

### Phase 1: ファイル名連番対応

**対象ファイル:** `internal/tui/file_operations.go`, `internal/tui/file_operations_test.go`, `internal/tui/messages.go`, `internal/tui/messages_test.go`

**Red（テスト記述）:**

- `TestGetTitleFromFilename` を新形式に対応（`0001-hello-2026-02-11.md` → `hello`）
- `TestHasSeqPrefix` 追加: `0001-xxx` → true、`hello-xxx` → false
- `TestGetNextSeq` 追加: 空ディレクトリ→1、既存ファイルあり→最大+1
- `TestCreateNewNoteCmd` 追加: 連番付きファイル名で作成されること

**Green（実装）:**

- `getTitleFromFilename` 修正: 先頭 1 要素（連番）と末尾 3 要素を除去
- `HasSeqPrefix(filename string) bool` 新規（エクスポート）: ファイル名が `\d{4}-` で始まるか判定
- `GetNextSeq(notesDir string) (int, error)` 新規（エクスポート）: ディレクトリ内の最大連番+1 を返す
- `createNewNoteCmd` 修正: `GetNextSeq` で連番取得、ファイル名に連番付与
- testdata 配下のテストファイルも新形式にリネーム

**テスト実行:**

```bash
go test ./internal/tui/... -v -run "TestGetTitleFromFilename|TestHasSeqPrefix|TestGetNextSeq|TestCreateNewNoteCmd"
```

---

### Phase 2: `notebox migrate` CLI サブコマンド

**対象ファイル:** `internal/cli/migrate_command.go`（新規）, `internal/cli/migrate_command_test.go`（新規）, `internal/cli/register.go`

**Red（テスト記述）:**

- `TestMigrateExecute` 追加:
  - 連番なしファイルが連番付きにリネームされること
  - 既に連番ありのファイルはスキップされること
  - 混在時の挙動（連番ありは維持、なしのみリネーム）
  - 空ディレクトリでもエラーなし

**Green（実装）:**

- `migrateCmd` 構造体: `google/subcommands.Command` インターフェースを実装
  - `Name()` → `"migrate"`
  - `Synopsis()` → `"migrate note files to sequential naming format"`
  - `Execute()`:
    1. `config.GetConfig()` でノートディレクトリ取得
    2. ディレクトリをスキャンし、`HasSeqPrefix` で連番なしファイルを検出
    3. 各ファイルに対して `GetNextSeq` で連番取得、`os.Rename` でリネーム
    4. リネーム結果を標準出力に表示
- `register.go` に `subcommands.Register(&migrateCmd{}, "")` 追加

**テスト実行:**

```bash
go test ./internal/cli/... -v -run "TestMigrateExecute"
```

---

### Phase 3: リスト同期ロジック

**対象ファイル:** `internal/tui/listPanel.go`, `internal/tui/listPanel_test.go`

**Red（テスト記述）:**

- `TestCalcSyncNoteList` 追加（テーブル駆動、8 ケース）:
  - 変更なし: カーソル・オフセット維持
  - ノート追加: カーソル位置維持
  - 選択中ノート削除: 同インデックスにクランプ
  - 末尾ノート削除: `len-1` にクランプ
  - 全削除: cursor=0, offset=0
  - 空→ノート追加: cursor=0
  - offset 付きでカーソル維持
  - パスで照合（リスト順序変更でも追従）

**Green（実装）:**

- `calcSyncNoteList(oldItems, newItems []note, oldCursor, oldOffset, height int) (newCursor, newOffset int)` 新規
  - 選択中ノートのパスで新リストを検索
  - 見つからなければインデックスをクランプ
  - offset をカーソルが表示範囲内に収まるよう調整
- `syncNoteList(newNotes []note) tea.Cmd`（model メソッド）
  - `calcSyncNoteList` を呼び、`listPanel.items` を差し替え
  - `renderPreviewCmd` を返す

**テスト実行:**

```bash
go test ./internal/tui/... -v -run "TestCalcSyncNoteList"
```

---

### Phase 4: ファイル監視コマンド

**対象ファイル:** `internal/tui/messages.go`, `internal/tui/messages_test.go`, `internal/tui/watcher_test.go`（新規）

**依存:** `go get github.com/fsnotify/fsnotify`

**Red（テスト記述）:**

- `TestWatchNotesCmd` 追加（`watcher_test.go` 新規）:
  - 一時ディレクトリにファイル作成 → `fileChangedMsg` が返ること
  - watcher close 後 → nil が返ること
- `TestReloadNotesCmd` 追加（`messages_test.go`）:
  - 有効なディレクトリ → `notesReloadedMsg` が返ること
  - 存在しないディレクトリ → `errMsg` が返ること

**Green（実装）:**

- メッセージ型追加:
  - `fileChangedMsg{op fsnotify.Op, name string}` — FS イベント通知
  - `fileWatchErrMsg error` — watcher エラー
  - `debounceTickMsg time.Time` — デバウンス用 tick
  - `notesReloadedMsg []note` — リロード完了通知
- コマンド追加:
  - `watchNotesCmd(watcher *fsnotify.Watcher) tea.Cmd` — watcher チャネルをブロッキングで待つ
  - `reloadNotesCmd(notesDir string) tea.Cmd` — `loadNoteFiles` を非同期実行

**テスト実行:**

```bash
go test ./internal/tui/... -v -run "TestWatchNotesCmd|TestReloadNotesCmd"
```

---

### Phase 5: model 統合 + main.go

**対象ファイル:** `internal/tui/model.go`, `cmd/notebox/main.go`

**Red（テスト記述）:**

- このフェーズは統合テスト。自動テストは全テストスイート実行で確認
- 手動テストが主な検証手段

**Green（実装）:**

- `model` にフィールド追加: `watcher *fsnotify.Watcher`, `pendingFSChange bool`, `lastFSEventTime time.Time`
- `NewModel()` 修正:
  - watcher 生成 + `watcher.Add(cfg.NotesDir)`
- `Init()` 修正: `watchNotesCmd(m.watcher)` をバッチに追加
- `Update()` に case 追加:
  - `fileChangedMsg` → `.md` のみ反応、`pendingFSChange=true`, `lastFSEventTime=now`、`tea.Tick(200ms)` + `watchNotesCmd` を返す
  - `debounceTickMsg` → `time.Since(lastFSEventTime) >= 200ms` なら `reloadNotesCmd` 発行
  - `notesReloadedMsg` → `syncNoteList()` 呼び出し
  - `fileWatchErrMsg` → ログ出力 + `watchNotesCmd` で再購読
- `Close() error` メソッド追加: `m.watcher.Close()`
- `cmd/notebox/main.go`: `NewModel()` 後に `defer m.Close()` 追加

**テスト実行:**

```bash
go test ./... -v
```

---

## 検証手順（全フェーズ完了後）

### 自動テスト

```bash
go test ./... -v
```

### 手動テスト

1. **マイグレーション確認**: 既存ノート（連番なし）がある状態で `notebox migrate` → ファイルがリネームされることを確認。リネーム結果が標準出力に表示されること
2. ターミナルを 2 ペインに分割し、両方で `task run` を実行
3. **作成**: ペイン 1 で `n` → ノート作成 → ペイン 2 のリストに反映。リスト末尾に追加されていること
4. **削除**: ペイン 1 で `d` → Enter → ペイン 2 から消えること
5. **編集**: ペイン 1 で `e` → 編集・保存 → ペイン 2 でプレビュー更新。リスト位置は不変
6. **外部変更**: `touch ~/.notebox/notes/9999-test-2026-02-11.md` → 両インスタンスに反映
7. **デバウンス**: `for i in $(seq 1 10); do touch ~/.notebox/notes/999$i-rapid-2026-01-01.md; done` → リスト一括更新
8. **カーソル保持**: ペイン 1 でノート B 選択中 → ペイン 2 からノート A を削除 → ペイン 1 のカーソルがノート B に留まる
