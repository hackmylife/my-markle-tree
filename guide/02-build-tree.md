# Step 2: ツリーを組み立てる

## この step で学ぶこと

リーフのハッシュ列から、2 つずつペアにして親を作る操作を、1 つになるまで繰り返します。
この step では **リーフの個数が 2 の累乗（1, 2, 4, 8, ...）の場合だけ** を扱います。

```text
Levels[2]:               root
                      /        \
Levels[1]:        H(ab)        H(cd)
                  /   \        /   \
Levels[0]:     H(a)  H(b)   H(c)  H(d)
```

各階層（level）のハッシュ列を `Levels` にすべて保存しておきます。
root を計算するだけなら最上位の 1 つだけで足りますが、Step 4 の Proof 生成で途中の階層が必要になります。

### 具体例

```text
入力: leaves = ["a", "b", "c", "d"]

Levels[0] = [HashLeaf("a"), HashLeaf("b"), HashLeaf("c"), HashLeaf("d")]
          = [022a6979..., 57eb3561..., 597fcb31..., d070dc5b...]

Levels[1] = [HashNode(022a6979..., 57eb3561...), HashNode(597fcb31..., d070dc5b...)]
          = [b137985f..., dbbd68c3...]

Levels[2] = [HashNode(b137985f..., dbbd68c3...)]
          = [33376a3bd63e9993708a84ddfe6c28ae58b83505dd1fed711bd924ec5a6239f0]   ← root
```

リーフが 1 個のときは、`Levels = [[HashLeaf(a)]]` で root は `HashLeaf(a)` です。

## 実装する関数

`workspace/merkle/tree.go`（`Tree` 型は定義済み）

```go
type Tree struct {
	// Levels[0] はリーフハッシュの列、Levels[len(Levels)-1] は [root]。
	Levels [][][]byte
}

// NewTree はリーフのデータ列から Merkle Tree を構築する。
func NewTree(leaves [][]byte) *Tree

// Root は root ハッシュを返す。
func (t *Tree) Root() []byte
```

`[][][]byte` は「ハッシュ（`[]byte`）の列（`[][]byte`）の列」です。`Levels[階層][位置]` で 1 つのハッシュを取り出せます。

## ヒント

<details>
<summary>ヒント 1: 全体の流れ</summary>

1. 各リーフに `HashLeaf` をかけた列を作る。これが `Levels[0]`
2. 今の階層の長さが 1 になるまで、次を繰り返す
   - 隣り合う 2 つ（`i` と `i+1`）を `HashNode` で結合した列を作る
   - それを `Levels` に追加し、次の「今の階層」にする

</details>

<details>
<summary>ヒント 2: ループの形</summary>

```go
for len(level) > 1 {
	next := ...
	for i := 0; i < len(level); i += 2 {
		// level[i] と level[i+1] を結合
	}
	levels = append(levels, next)
	level = next
}
```

リーフが 1 個のときは、外側のループに 1 度も入らずに終わります。

</details>

<details>
<summary>ヒント 3: Root</summary>

最上位の階層は `t.Levels[len(t.Levels)-1]` で、その中の唯一の要素が root です。

</details>

## テストを実行する

```bash
make test STEP=2
```

| テスト | 確かめること |
|---|---|
| `TestStep2_OneLeaf` | リーフ 1 個なら階層は 1 つで、root は `HashLeaf(a)` |
| `TestStep2_TwoLeaves` | `HashNode(HashLeaf(a), HashLeaf(b))` |
| `TestStep2_FourLeaves` | 上の具体例のとおり |
| `TestStep2_LevelsShape` | 8 個のとき `Levels` の長さが 8 → 4 → 2 → 1 で、各ハッシュが正しい |
| `TestStep2_RootChangesWhenAnyLeafChanges` | どのリーフを変えても root が変わる |
| `TestStep2_LeafOrderMatters` | リーフの順序を入れ替えると root が変わる |

## 模範実装

<details>
<summary>模範実装を見る（写経モードの人はここから）</summary>

**workspace/merkle/tree.go**

```go
package merkle

// Tree は Merkle Tree の全階層のハッシュを保持する。
type Tree struct {
	// Levels[0] はリーフハッシュの列、Levels[len(Levels)-1] は [root]。
	// リーフが 0 個のときは Levels は空。
	Levels [][][]byte
}

// NewTree はリーフのデータ列から Merkle Tree を構築する。
// この版はリーフの個数が 2 の累乗の場合だけを扱う（Step 3 で一般化する）。
func NewTree(leaves [][]byte) *Tree {
	level := make([][]byte, len(leaves))
	for i, leaf := range leaves {
		level[i] = HashLeaf(leaf)
	}
	levels := [][][]byte{level}

	for len(level) > 1 {
		next := make([][]byte, 0, len(level)/2)
		for i := 0; i < len(level); i += 2 {
			next = append(next, HashNode(level[i], level[i+1]))
		}
		levels = append(levels, next)
		level = next
	}

	return &Tree{Levels: levels}
}

// Root は root ハッシュを返す。
func (t *Tree) Root() []byte {
	return t.Levels[len(t.Levels)-1][0]
}
```

</details>

## 解説

- **ボトムアップ構築**: 下から 1 階層ずつ作るので、再帰を使わずループで書けます。
  `Left`/`Right` ポインタを持つノード構造体で木を表す方法もありますが、階層ごとのスライスで持つと Proof 生成（Step 4）で「`i` 番目の兄弟は `i^1` 番目」と添字計算だけで辿れて簡単です。
- **計算量**: 内部ノードは「リーフ数 − 1」個なので、ハッシュ計算は合計 `2n − 1` 回です。
- **`make([][]byte, 0, len(level)/2)`**: 次の階層の長さは分かっているので、容量を先に確保しています。省略しても動きます。
- この実装はリーフが 3 個だと `level[i+1]` で範囲外アクセスになり、0 個だと `Root` が範囲外アクセスになります。これは Step 3 で直します。

## 完了条件

- [ ] `make test STEP=2` がすべて PASS する
- [ ] `make test STEP=1` も引き続き PASS する
- [ ] リーフが 8 個のとき、ハッシュ計算が何回行われるか説明できる

次は [Step 3: 任意の個数に対応する](03-any-size.md) です。
