# Merkle Tree 自作教材 設計書

## 目的

Go で Merkle Tree を step ごとに自作し、構造・Proof・安全性・応用までを理解する教材を作る。
学習者は次のどちらの進め方も選べる。

- **自作モード**: ガイドの概念説明とヒントだけを読み、雛形を埋めてテストを通す
- **写経モード**: ガイド後半の模範実装を書き写し、テストを通して動作を確かめる

参考: https://github.com/hackmylife/like-a-ethereum （README と手順書 `.plan/milestone-*.md` を分ける構成、標準ライブラリのみ、Makefile・`.tool-versions` を持つ）

## 前提

- 言語: Go 1.26.2（`.tool-versions` に `golang 1.26.2`）
- 外部ライブラリは使わない（標準ライブラリのみ。テストも `testing` のみ）
- 文書はすべて日本語
- ハッシュは SHA-256

## ディレクトリ構成

```text
my_merkle_tree/
├── README.md                  # 何を作るか / セットアップ / 進め方 / コマンド / 本物との違い
├── guide/
│   ├── 00-overview.md         # Merkle Tree とは、全体の流れ、2つの進め方
│   ├── 01-hash.md
│   ├── 02-build-tree.md
│   ├── 03-any-size.md
│   ├── 04-generate-proof.md
│   ├── 05-verify-proof.md
│   ├── 06-second-preimage.md
│   └── 07-file-integrity-cli.md
├── skeleton/                  # 雛形（コミット対象・学習者は編集しない）
│   ├── merkle/
│   │   ├── hash.go            # Step 1
│   │   ├── tree.go            # Step 2, 3
│   │   ├── proof.go           # Step 4, 5
│   │   ├── attack.go          # Step 6（NaiveRoot は実装済みで提供）
│   │   └── chunk.go           # Step 7
│   └── cmd/merklecli/main.go  # Step 7
├── tests/                     # テスト（コミット対象）
│   ├── step1_hash_test.go
│   ├── ...
│   └── step7_chunk_test.go
├── workspace/                 # 学習者の作業場所。.gitignore 対象
├── docs/specs/                # この設計書
├── Makefile
├── go.mod                     # module my_merkle_tree
├── .tool-versions
└── .gitignore                 # workspace/
```

### workspace と tests の関係

- `make init` で `skeleton/` を `workspace/` にコピーする。`workspace/` が既にあれば何もしない（学習者のコードを上書きしない）。
- `tests/` は `package tests`（テストファイルのみのパッケージ）として `my_merkle_tree/workspace/merkle` を import する。テストは学習者が触らない場所にあるため、教材側でテストを直しても学習者のコードと衝突しない。
- `make init` 前は `workspace/` が無いので `tests/` はコンパイルできない。README とエラー時の案内で `make init` を最初に実行するよう明記する。
- 雛形の未実装関数は `panic("TODO: Step N")` とし、全 step のテストが常にコンパイルできるようにする。step を絞ったテスト実行では、未着手 step の panic に巻き込まれない。

## Makefile

| ターゲット | 内容 |
|---|---|
| `make init` | `skeleton/` → `workspace/` コピー（既存なら中止してメッセージ表示） |
| `make test STEP=N` | `go test ./tests -run '^TestStepN_' -v` |
| `make test-all` | `go test ./tests` |
| `make vet` | `go vet ./workspace/... ./tests/...` |
| `make reset` | 確認メッセージ後に `workspace/` を削除して `init` し直す（`CONFIRM=yes` 必須） |

テスト関数名は `TestStep{N}_{内容}` で統一し、`-run '^TestStepN_'` で step 単位に絞れるようにする（末尾の `_` で `TestStep1` が `TestStep10` に当たらないようにする）。

## API（workspace/merkle パッケージ）

```go
// Step 1
func HashLeaf(data []byte) []byte              // sha256(0x00 || data)
func HashNode(left, right []byte) []byte       // sha256(0x01 || left || right)

// Step 2, 3
type Tree struct {
    Levels [][][]byte // Levels[0] = リーフハッシュ列, Levels[len-1] = [root]。リーフ 0 個なら空
}
func NewTree(leaves [][]byte) *Tree
func (t *Tree) Root() []byte

// Step 4
type Sibling struct {
    Hash   []byte
    IsLeft bool // 兄弟が左側にあるなら true
}
type Proof struct {
    Index    int
    Siblings []Sibling // リーフに近い順
}
func (t *Tree) Proof(index int) (Proof, error)

// Step 5
func Verify(root, leafData []byte, proof Proof) bool

// Step 6
func NaiveRoot(leaves [][]byte) []byte          // 雛形で提供（プレフィックスなし sha256、奇数は持ち上げ）
func Forge(leaves [][]byte) [][]byte            // 学習者が実装

// Step 7
func Chunk(data []byte, size int) [][]byte
```

### 仕様の決めごと

| 項目 | 決定 | 理由 |
|---|---|---|
| ハッシュ | リーフ `0x00`、内部ノード `0x01` のプレフィックス付き SHA-256（RFC 6962 方式） | 第二原像攻撃を防ぐため。Step 1 では「Step 6 で理由を体験する」と予告する |
| 奇数個 | 余った最後のノードを複製せず、そのまま次の階層へ持ち上げる | 複製方式（Bitcoin）だと `[a,b,c]` と `[a,b,c,c]` が同じ root になる（CVE-2012-2459）。ガイドで比較する |
| 空のツリー | `NewTree(nil)` の `Levels` は空、`Root()` は `sha256("")`（RFC 6962 の空ツリーハッシュ） | panic やエラーにせず、定義済みの値を返す |
| リーフ 1 個 | root は `HashLeaf(leaf)` | 持ち上げ規則から自然に導かれる |
| Proof の範囲外 index | `Proof` はエラーを返す | |
| 持ち上げられたノードの Proof | その階層では兄弟が無いため Sibling を追加しない | Proof の長さは index により変わりうる |
| Verify の index | `proof.Siblings` の `IsLeft` だけで結合順を決める。`Index` は表示・デバッグ用 | |
| Forge の出力 | `len(leaves) >= 2` が前提。隣り合うペアのリーフハッシュを連結した 64 バイト `sha256(a)‖sha256(b)` を新しいリーフにする。奇数個で余った最後のリーフは元のまま末尾に残す。例: `[a,b,c]` → `[sha256(a)‖sha256(b), c]` | `NaiveRoot` はこの新リーフを `sha256(sha256(a)‖sha256(b))` とハッシュし、元の内部ノードと同じ値になるため、元と同じ root が得られる |
| Chunk | 最後のチャンクは `size` 未満でもよい。空データは空スライス。`size <= 0` は panic | |

### CLI（Step 7）

```text
merklecli root   -file <path> [-chunk 1024]
merklecli proof  -file <path> -index <i> [-chunk 1024]     # JSON で Proof を出力
merklecli verify -file <path> -root <hex> -proof <proof.json> [-chunk 1024]   # proof.index 番目のチャンクを検証。NG なら終了コード 1
```

`go run ./workspace/cmd/merklecli ...` で実行する。雛形では main・JSON 変換・`loadChunks` を提供し、学習者は `runRoot` / `runProof` / `runVerify` を実装する。CLI はガイドの手動確認手順で検証し、自動テストは `Chunk` と、「Chunk → NewTree → Proof → Verify」の結合テストで行う。

## テスト方針

- 期待値はリテラルの 16 進数で固定せず、テスト内で `crypto/sha256` から組み立てて比較する（例: `HashNode(HashLeaf(a), HashLeaf(b))` を手計算した sha256 と比較）。
- 各 step の完了条件をテストで表す。

| Step | 主なテスト |
|---|---|
| 1 | プレフィックス付きハッシュが一致する。`HashLeaf(x) != sha256(x)`。`HashNode(l,r) != HashNode(r,l)`。`HashLeaf(l‖r) != HashNode(l,r)` |
| 2 | 1, 2, 4, 8 個で root が手計算と一致する。階層数と各階層の長さ。リーフを 1 つ変えると root が変わる |
| 3 | 0, 3, 5 個の形と値、1〜17 個の root。`[a,b,c]` と `[a,b,c,c]` で root が異なる |
| 4 | 全 index で Proof の Sibling 数と各 Hash・IsLeft が期待どおりになる。範囲外はエラー |
| 5 | 全サイズ（1〜9）・全 index で Verify が true。改ざんデータ、別 root、Sibling 改ざん、IsLeft 反転では false |
| 6 | 2〜9 個で `NaiveRoot(Forge(x)) == NaiveRoot(x)`、かつ `Forge(x)` は `x` より短い。同じ偽造リーフ列でも `NewTree(...).Root()` は元と一致しない |
| 7 | Chunk の境界（ちょうど割り切れる・端数・空）。ファイル相当データで結合テスト、1 バイト改ざんで Verify が false |

## ガイドの構成（各 step 共通）

1. **この step で学ぶこと**: 概念と図（ASCII）。具体的な入力と出力の例を必ず載せる
2. **実装する関数**: シグネチャと仕様（雛形の doc コメントと同じ内容）
3. **ヒント**: 段階的に開示する。`<details><summary>ヒント1</summary>` のように折りたたむ
4. **テストを実行する**: `make test STEP=N` と、よくある失敗メッセージの読み方
5. **模範実装**: 完全なコード。`<details><summary>模範実装を見る</summary>` で折りたたむ。写経モードの人は最初にここを開いて書き写す
6. **解説**: 模範実装のポイントと、他の実装方法との比較
7. **完了条件と次の step**

`00-overview.md` で 2 つのモードを説明する。写経モードでも、写経後に「解説」を読み、テストを 1 つ壊して失敗することを確かめるよう案内する。

## 教材の品質保証（作成時）

- 模範実装はガイドから抽出したコードを一時ディレクトリの `workspace/` に配置し、`make test-all` が全件通ることを確認する。
- 雛形のままでは、各 step のテストがすべて失敗することを確認する（各テストが実装を本当に検査していることの確認）。
- 模範実装の一部を壊し（プレフィックス除去、奇数時の複製、IsLeft 反転など）、対応するテストが失敗することを確認する。
- 模範解答のコードはリポジトリの `.go` ファイルとしては置かない（ガイド内のみ）。

## 対象外

- Sparse Merkle Tree、Merkle Patricia Trie（README の「発展」で紹介するにとどめる）
- ストリーミング構築、永続化、並列化
