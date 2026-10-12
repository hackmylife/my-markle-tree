# Step 1: ハッシュ

## この step で学ぶこと

Merkle Tree で使うハッシュは 2 種類あります。

| 種類 | 入力 | 計算 |
|---|---|---|
| リーフハッシュ | 元のデータ `data` | `sha256(0x00 ‖ data)` |
| 内部ノードハッシュ | 左右の子のハッシュ `left`, `right` | `sha256(0x01 ‖ left ‖ right)` |

`‖` はバイト列の連結です。先頭に 1 バイトの **プレフィックス** を付けて、リーフ用と内部ノード用を区別します。

### 具体例

```text
入力: "hello"（5 バイト: 68 65 6c 6c 6f）

sha256("hello")                = 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
HashLeaf("hello")
  = sha256(00 68 65 6c 6c 6f)  = 8a2a5c9b768827de5a9552c38a044c66959c68f6d2f21b5260af54d2f87db827
```

```text
HashLeaf("a") = 022a6979e6dab7aa...
HashLeaf("b") = 57eb35615d47f34e...

HashNode(HashLeaf("a"), HashLeaf("b"))
  = sha256(01 ‖ 022a6979...（32 バイト） ‖ 57eb3561...（32 バイト）)   ← 合計 65 バイトの入力
  = b137985ff484fb600db93107c77b0365c80d78f5b429ded0fd97361d077999eb
```

### なぜプレフィックスを付けるのか

プレフィックスを付けないと、「64 バイトのリーフデータ」と「2 つの子ハッシュの連結」が区別できません。
これを悪用すると、元とは違うリーフ列で同じ root を作れてしまいます（第二原像攻撃）。
この攻撃は **Step 6 で実際に自分で行います**。ここでは「区別のために付ける」とだけ覚えておけば十分です。

この方式は Certificate Transparency の仕様 [RFC 6962](https://www.rfc-editor.org/rfc/rfc6962#section-2.1) と同じです。

## 実装する関数

`workspace/merkle/hash.go`

```go
// HashLeaf はリーフのデータからリーフハッシュ sha256(0x00 || data) を計算する。
func HashLeaf(data []byte) []byte

// HashNode は左右の子のハッシュから親のハッシュ sha256(0x01 || left || right) を計算する。
func HashNode(left, right []byte) []byte
```

- 戻り値は 32 バイトの `[]byte`
- `data` が `nil` でも動くこと（`sha256(0x00)` になる）

## ヒント

<details>
<summary>ヒント 1: 使うパッケージ</summary>

標準ライブラリの `crypto/sha256` を使います。
`sha256.Sum256(b)` は `[32]byte`（配列）を返します。`[]byte`（スライス）にするには `sum[:]` と書きます。

</details>

<details>
<summary>ヒント 2: 連結の方法</summary>

方法は 2 つあります。

- 新しいスライスを作って `append` で連結し、`sha256.Sum256` に渡す
- `h := sha256.New()` でハッシュ計算器を作り、`h.Write(...)` を複数回呼んで、最後に `h.Sum(nil)` で取り出す

後者は連結用のスライスを作らずに済むので、すっきり書けます。

</details>

<details>
<summary>ヒント 3: append の落とし穴</summary>

`append(left, right...)` のように **引数のスライスに直接 append しない** でください。
`left` の容量（cap）に余裕があると、呼び出し元の配列を書き換えてしまうことがあります。
連結するなら `append([]byte{0x01}, left...)` のように、新しいスライスから始めます。

</details>

## テストを実行する

```bash
make test STEP=1
```

| テスト | 確かめること |
|---|---|
| `TestStep1_HashLeaf` | `sha256(0x00 ‖ "hello")` と一致し、プレフィックスなしの `sha256("hello")` とは異なる |
| `TestStep1_HashLeafEmpty` | `nil` を渡すと `sha256(0x00)` |
| `TestStep1_HashNode` | `sha256(0x01 ‖ l ‖ r)` と一致し、左右を入れ替えると結果が変わる |
| `TestStep1_LeafAndNodeAreDistinguished` | `HashLeaf(l ‖ r)` と `HashNode(l, r)` が異なる |

## 模範実装

<details>
<summary>模範実装を見る（写経モードの人はここから）</summary>

**workspace/merkle/hash.go**

```go
// Package merkle は学習用の Merkle Tree 実装。各関数は guide/ の対応する Step で実装する。
package merkle

import "crypto/sha256"

// リーフと内部ノードで先頭 1 バイトを変える。同じバイト列がリーフとしても内部ノードとしても
// 解釈できないようにするため（RFC 6962 と同じ方式。理由は Step 6 で確かめる）。
const (
	leafPrefix byte = 0x00
	nodePrefix byte = 0x01
)

// HashLeaf はリーフのデータからリーフハッシュ sha256(0x00 || data) を計算する。
func HashLeaf(data []byte) []byte {
	h := sha256.New()
	h.Write([]byte{leafPrefix})
	h.Write(data)
	return h.Sum(nil)
}

// HashNode は左右の子のハッシュから親のハッシュ sha256(0x01 || left || right) を計算する。
func HashNode(left, right []byte) []byte {
	h := sha256.New()
	h.Write([]byte{nodePrefix})
	h.Write(left)
	h.Write(right)
	return h.Sum(nil)
}
```

</details>

## 解説

- `sha256.New()` が返す `hash.Hash` は `io.Writer` です。`Write` で少しずつデータを流し込み、
  `Sum(nil)` で結果を新しいスライスとして受け取ります。`Sum(b)` は「`b` の後ろにハッシュを追加したスライス」を返すので、`nil` を渡すと 32 バイトだけが返ります。
- `hash.Hash` の `Write` はエラーを返さない契約になっているので、戻り値は無視して構いません。
- `sha256.Sum256(append([]byte{0x00}, data...))` と 1 行で書いても正解です。
- プレフィックスを定数にしておくと、「どこで何を区別しているか」がコード上で明確になります。

## 完了条件

- [ ] `make test STEP=1` がすべて PASS する
- [ ] `HashLeaf("hello")` が `8a2a5c9b...` になる理由を説明できる

次は [Step 2: ツリーを組み立てる](02-build-tree.md) です。
