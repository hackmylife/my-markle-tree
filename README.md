# my_merkle_tree

Go で Merkle Tree を step ごとに自作して学ぶ教材です。

ハッシュ関数から始めて、ツリーの構築、Merkle Proof の生成と検証、第二原像攻撃とその対策を順に実装し、
最後にファイル改ざん検知 CLI を作ります。各 step にはテストが用意されていて、`go test` が通れば次へ進めます。

## 目的

この教材で学ぶこと:

- リーフハッシュと内部ノードハッシュ（ドメイン分離）
- ボトムアップのツリー構築
- 奇数個のリーフの扱い（持ち上げ方式と複製方式、CVE-2012-2459）
- Merkle Proof の生成と検証
- 第二原像攻撃と、プレフィックスによる対策
- チャンク分割したファイルの改ざん検知

## 進め方

ガイドは [guide/](guide/) にあります。[guide/00-overview.md](guide/00-overview.md) から始めてください。

| Step | ガイド | 実装するもの |
|---|---|---|
| 1 | [ハッシュ](guide/01-hash.md) | `HashLeaf`, `HashNode` |
| 2 | [ツリーを組み立てる](guide/02-build-tree.md) | `NewTree`, `Root`（2 の累乗個） |
| 3 | [任意の個数に対応する](guide/03-any-size.md) | `NewTree`, `Root`（任意の個数） |
| 4 | [Proof を生成する](guide/04-generate-proof.md) | `Tree.Proof` |
| 5 | [Proof を検証する](guide/05-verify-proof.md) | `Verify` |
| 6 | [第二原像攻撃](guide/06-second-preimage.md) | `Forge` |
| 7 | [ファイル改ざん検知 CLI](guide/07-file-integrity-cli.md) | `Chunk`, `merklecli` |

各ガイドは前半が概念とヒント、後半が模範実装（折りたたみ）です。

- **自作モード**: 前半だけを読んで実装し、テストが通ってから模範実装と見比べる
- **写経モード**: 模範実装を書き写してテストを通し、解説を読む

## 必要環境

- Go 1.26 以上（`.tool-versions` は `golang 1.26.2`）
- make
- 外部ライブラリは使いません（標準ライブラリのみ）

## セットアップ

```bash
git clone <this-repo>
cd my_merkle_tree
make init
```

`make init` で `skeleton/` が `workspace/` にコピーされます。実装はすべて `workspace/` で行います。
`workspace/` は `.gitignore` に入っているので、Git にはコミットされません。

## Makefile

| ターゲット | 内容 |
|---|---|
| `make init` | `skeleton/` を `workspace/` にコピーする（`workspace/` があれば何もしない） |
| `make test STEP=N` | Step N のテストを実行する |
| `make test-all` | すべてのテストを実行する |
| `make vet` | `go vet` を実行する |
| `make reset CONFIRM=yes` | `workspace/` を雛形に戻す（実装はすべて消える） |

## ディレクトリ構成

```text
my_merkle_tree/
├── guide/            # 各 step の手順書（ヒントと模範実装）
├── skeleton/         # 雛形。make init で workspace/ にコピーされる（編集しない）
│   ├── merkle/
│   └── cmd/merklecli/
├── tests/            # 各 step のテスト（編集しない）
├── workspace/        # 学習者の作業場所（Git 管理外）
├── docs/specs/       # 教材の設計書
└── Makefile
```

`tests/` は `my_merkle_tree/workspace/merkle` を import してテストします。
`make init` の前は `workspace/` が無いため、`go test ./...` はコンパイルエラーになります。

## 完成すると

```bash
$ go run ./workspace/cmd/merklecli root -file sample.bin
e775bffc73a1cdd672a3d4592e71552e420ef4e4d51adcb44db44a666e88f383

$ go run ./workspace/cmd/merklecli proof -file sample.bin -index 2 > proof.json

$ go run ./workspace/cmd/merklecli verify -file tampered.bin -root e775bffc... -proof proof.json
NG: チャンク 2 は root に含まれていません（改ざんの可能性）
```

## 実システムとの違い

| 項目 | 実システムの例 | この教材 |
|---|---|---|
| ハッシュ | SHA-256（Bitcoin は二重 SHA-256、Ethereum は Keccak-256） | SHA-256 |
| ドメイン分離 | RFC 6962 は `0x00` / `0x01`、Bitcoin はなし | `0x00` / `0x01`（RFC 6962 と同じ） |
| 奇数個 | Bitcoin は複製、RFC 6962 は分割 | 持ち上げ（RFC 6962 と同じ形） |
| 空のツリー | RFC 6962 は `sha256("")` | `sha256("")` |
| Proof の左右 | RFC 6962 は index とツリーサイズから計算 | `IsLeft` を Proof に含める |
| ツリーの種類 | Ethereum は Merkle Patricia Trie | 二分 Merkle Tree |
| 永続化・差分更新 | あり | なし（毎回すべて作り直す） |

## 参考

- [RFC 6962: Certificate Transparency](https://www.rfc-editor.org/rfc/rfc6962)（2.1 節 Merkle Hash Trees）
- [CVE-2012-2459](https://nvd.nist.gov/vuln/detail/CVE-2012-2459)
- [hackmylife/like-a-ethereum](https://github.com/hackmylife/like-a-ethereum)（教材構成の参考）
