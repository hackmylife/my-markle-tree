# Step 3: 任意の個数に対応する

## この step で学ぶこと

実際のデータは 2 の累乗個とは限りません。ペアを作るときに 1 つ余ったノードをどう扱うかは、実装によって違います。

| 方式 | 余ったノード | 採用例 |
|---|---|---|
| **持ち上げ**（この教材） | 何もせず、そのまま次の階層へ | RFC 6962（Certificate Transparency） |
| 複製 | 自分自身と結合する `HashNode(x, x)` | Bitcoin |

### 持ち上げ方式の具体例

```text
入力: leaves = ["a", "b", "c"]

Levels[2]:            root = HashNode(H(ab), H(c))
                     /    \
Levels[1]:       H(ab)    H(c)     ← H(c) は結合せずにそのまま持ち上げる
                 /   \      |
Levels[0]:    H(a)  H(b)  H(c)

root = 36642e73c2540ab121e3a6bf9545b0a24982cd830eb13d3cd19de3ce6c021ec1
```

```text
入力: leaves = [l0, l1, l2, l3, l4]（5 個）

Levels[0]: H0  H1  H2  H3  H4      (5 個)
Levels[1]: H01     H23     H4      (3 個)  ← H4 を持ち上げ
Levels[2]: H0123           H4      (2 個)  ← H4 をまた持ち上げ
Levels[3]: root = HashNode(H0123, H4)
```

### 複製方式の落とし穴

複製方式では、`[a, b, c]` の `c` が `HashNode(H(c), H(c))` になります。これは `[a, b, c, c]` の右半分とまったく同じです。

```text
複製方式:
  [a, b, c]    → root = HashNode(H(ab), HashNode(H(c), H(c)))
  [a, b, c, c] → root = HashNode(H(ab), HashNode(H(c), H(c)))   ← 同じ！
```

つまり **異なるリーフ列が同じ root になります**。
Bitcoin ではこれが原因で、正しいブロックと同じ root を持つ「トランザクションが重複した不正なブロック」を作れるという脆弱性（CVE-2012-2459）がありました。
持ち上げ方式なら `[a,b,c]` の root は `36642e73...`、`[a,b,c,c]` の root は `e9636069...` で、別の値になります。

### 空のツリー

リーフが 0 個のときは、RFC 6962 に合わせて root を `sha256("")`（空のバイト列の SHA-256）とします。

```text
NewTree(nil).Root() = e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
```

このとき `Levels` は空（`nil`）のままにします。`Levels[0]` は「リーフハッシュの列」なので、リーフが無ければ階層もありません。

## 実装する関数

Step 2 で作った `NewTree` と `Root` を修正します。

- 奇数個のとき、余った最後のノードを持ち上げる
- リーフ 0 個のとき、`Levels` を空にし、`Root()` は `sha256("")` を返す

## ヒント

<details>
<summary>ヒント 1: 余ったノードの見分け方</summary>

内側のループで `i+1 == len(level)` なら、`level[i]` にはペアの相手がいません。

</details>

<details>
<summary>ヒント 2: 次の階層の長さ</summary>

長さ `n` の階層の次は `(n+1)/2` 個です（Go の整数除算は切り捨てなので、`n` が奇数でも正しく切り上げになります）。
例: 5 → 3 → 2 → 1

</details>

<details>
<summary>ヒント 3: 空のツリー</summary>

`NewTree` の先頭で `len(leaves) == 0` なら `&Tree{}` を返し、`Root` の先頭で `len(t.Levels) == 0` なら `sha256.Sum256(nil)` を返します。

</details>

## テストを実行する

```bash
make test STEP=3
```

| テスト | 確かめること |
|---|---|
| `TestStep3_Empty` | 0 個で `sha256("")` |
| `TestStep3_ThreeLeaves` | 上の具体例のとおり |
| `TestStep3_FiveLeaves` | 階層の長さが 5 → 3 → 2 → 1 で、`Levels[1][2]` が `HashLeaf(l4)` のまま |
| `TestStep3_AllSizes` | 1〜17 個すべてで root が正しい |
| `TestStep3_NoDuplication` | `[a,b,c]` と `[a,b,c,c]` の root が異なる |

## 模範実装

<details>
<summary>模範実装を見る（写経モードの人はここから）</summary>

**workspace/merkle/tree.go**

```go
package merkle

import "crypto/sha256"

// Tree は Merkle Tree の全階層のハッシュを保持する。
type Tree struct {
	// Levels[0] はリーフハッシュの列、Levels[len(Levels)-1] は [root]。
	// リーフが 0 個のときは Levels は空。
	Levels [][][]byte
}

// NewTree はリーフのデータ列から Merkle Tree を構築する。
func NewTree(leaves [][]byte) *Tree {
	if len(leaves) == 0 {
		return &Tree{}
	}

	level := make([][]byte, len(leaves))
	for i, leaf := range leaves {
		level[i] = HashLeaf(leaf)
	}
	levels := [][][]byte{level}

	for len(level) > 1 {
		next := make([][]byte, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			if i+1 == len(level) {
				// 余ったノードは複製せずそのまま持ち上げる。複製すると [a,b,c] と [a,b,c,c] の root が一致してしまうため。
				next = append(next, level[i])
				continue
			}
			next = append(next, HashNode(level[i], level[i+1]))
		}
		levels = append(levels, next)
		level = next
	}

	return &Tree{Levels: levels}
}

// Root は root ハッシュを返す。リーフが 0 個のときは sha256("") を返す。
func (t *Tree) Root() []byte {
	if len(t.Levels) == 0 {
		empty := sha256.Sum256(nil)
		return empty[:]
	}
	return t.Levels[len(t.Levels)-1][0]
}
```

</details>

## 解説

- Step 2 からの変更は 3 か所です。空のツリーの早期 return、`i+1 == len(level)` の持ち上げ、`Root` の空チェック。
- 持ち上げられたノードは、同じ `[]byte` を次の階層でも参照しています。ハッシュは作った後に書き換えないので、共有しても問題ありません。
- **持ち上げ方式のツリーの形**: リーフ数を 2 の累乗の和に分解した形になります（5 = 4 + 1 なので、4 個の完全な木と 1 個のリーフが root で結合される）。
  RFC 6962 は「左側に `n` 未満で最大の 2 の累乗個を置く」と再帰的に定義していますが、ボトムアップで持ち上げても同じ形になります。
- **他の方法と比べて**: 複製方式は「どの階層も偶数」という単純さが利点ですが、上で見たとおり root の一意性が崩れます。
  Bitcoin はこの方式を変えられないため、検証側で重複トランザクションを拒否して対処しています。

## 完了条件

- [ ] `make test STEP=3` がすべて PASS する
- [ ] `make test STEP=2` も引き続き PASS する
- [ ] 複製方式で `[a,b,c]` と `[a,b,c,c]` の root が一致する理由を、図で説明できる

次は [Step 4: Proof を生成する](04-generate-proof.md) です。
