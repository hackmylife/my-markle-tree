# Step 0: はじめに

## Merkle Tree とは

Merkle Tree は、たくさんのデータを **1 つのハッシュ値（root）** にまとめる木構造です。

```text
                 root
              /        \
         H(ab)          H(cd)
         /   \          /   \
      H(a)  H(b)     H(c)  H(d)
       |     |        |     |
       a     b        c     d      ← リーフ（元のデータ）
```

- 一番下の **リーフ** は元のデータのハッシュ
- その上の **内部ノード** は、左右の子のハッシュを連結してハッシュしたもの
- 一番上の **root** は、すべてのデータに依存する 1 つのハッシュ

例えば `a, b, c, d` の root は `33376a3b...` になります（Step 2 で自分で計算します）。
`d` を 1 バイトでも変えると root はまったく別の値になります。

### 何がうれしいのか

root だけを信頼できる場所に置いておけば、次の 2 つができます。

1. **改ざん検知**: データ全体から root を計算し直して比べれば、どこかが変わったことがわかる
2. **部分的な証明（Merkle Proof）**: 「`c` が root に含まれている」ことを、全データではなく
   `H(d)` と `H(ab)` の **2 つのハッシュだけ** で示せる。リーフが 100 万個でも約 20 個のハッシュで済む

この性質は、Bitcoin のブロック内トランザクション、Ethereum の状態ルート、Git、
Certificate Transparency、IPFS、BitTorrent などで使われています。

## この教材で作るもの

`workspace/merkle` パッケージに次の機能を順に実装します。

| Step | ガイド | 実装するもの | 学ぶこと |
|---|---|---|---|
| 1 | [01-hash.md](01-hash.md) | `HashLeaf`, `HashNode` | リーフと内部ノードのハッシュ |
| 2 | [02-build-tree.md](02-build-tree.md) | `NewTree`, `Root`（2 の累乗個） | ボトムアップの木構築 |
| 3 | [03-any-size.md](03-any-size.md) | `NewTree`, `Root`（任意の個数） | 奇数個・空の扱いと、複製方式の落とし穴 |
| 4 | [04-generate-proof.md](04-generate-proof.md) | `Tree.Proof` | Merkle Proof の生成 |
| 5 | [05-verify-proof.md](05-verify-proof.md) | `Verify` | Merkle Proof の検証 |
| 6 | [06-second-preimage.md](06-second-preimage.md) | `Forge` | 第二原像攻撃と、プレフィックスによる対策 |
| 7 | [07-file-integrity-cli.md](07-file-integrity-cli.md) | `Chunk`, `merklecli` | ファイル改ざん検知 CLI |

## 2 つの進め方

各ガイドは次の順で書かれています。

1. この step で学ぶこと（概念と図）
2. 実装する関数（シグネチャと仕様）
3. ヒント（折りたたみ）
4. テストを実行する
5. **模範実装**（折りたたみ）
6. 解説
7. 完了条件

どちらの進め方でも、テストが通れば次の step に進めます。

### 自作モード

1〜4 を読み、`workspace/merkle/*.go` の `panic("TODO: ...")` を自分の実装に置き換えます。
詰まったらヒントを 1 つずつ開いてください。テストが通ったら、模範実装と解説を読んで自分のコードと比べます。

### 写経モード

1〜2 を読んだら、模範実装を開いて **ファイル全体を書き写し** ます（コピー＆ペーストではなく、手で打つのがおすすめです）。
テストが通ったら解説を読み、最後に **自分のコードを 1 か所わざと壊して、テストが失敗することを確かめて** ください。
例えば Step 1 なら、プレフィックスの `0x00` を `0x02` に変えて `make test STEP=1` を実行します。
どのテストが、どんなメッセージで落ちるかを見ると、そのコードが何を守っているのかがわかります。

## 準備

```bash
make init          # skeleton/ を workspace/ にコピーする
make test STEP=1   # Step 1 のテストを実行する（最初はすべて失敗する）
```

`workspace/` は `.gitignore` に入っているので、Git にはコミットされません。
最初からやり直すときは `make reset CONFIRM=yes` で雛形に戻せます（実装はすべて消えます）。

### テストの読み方

未実装の関数は `panic("TODO: Step N ...")` になっているので、最初は次のように表示されます。

```text
--- FAIL: TestStep1_HashLeaf (0.00s)
panic: TODO: Step 1 HashLeaf [recovered, repanicked]
```

`panic` が出ている間は、その関数がまだ実装されていません。
実装が間違っているときは、期待値と実際の値（先頭 8 バイトの 16 進）が表示されます。

```text
--- FAIL: TestStep1_HashLeaf (0.00s)
    step1_hash_test.go:18: HashLeaf("hello") は sha256(0x00 || "hello") と一致するはず
          got:  2cf24dba5fb0a30e...
          want: 8a2a5c9b768827de...
```

テストは `tests/` にあります。`tests/` は編集しないでください。
ただし、何を確かめているのかを知りたいときに読むのは大歓迎です。

次は [Step 1: ハッシュ](01-hash.md) です。
