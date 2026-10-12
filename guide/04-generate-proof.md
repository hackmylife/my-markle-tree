# Step 4: Proof を生成する

## この step で学ぶこと

**Merkle Proof** は、「あるリーフが root に含まれている」ことを示すデータです。
リーフから root までの経路上で、各階層の **兄弟（sibling）ノードのハッシュ** を集めたものです。

### 具体例

`[a, b, c, d]` の `c`（index 2）の Proof を作ります。

```text
                  root
               /        \
         [H(ab)]         H(cd)
          /   \          /   \
       H(a)  H(b)    *H(c)* [H(d)]

*  * = 証明したいリーフ
[  ] = Proof に入れる兄弟
```

```text
Proof{
  Index: 2,
  Siblings: [
    { Hash: H(d)  = d070dc5b..., IsLeft: false },   ← Levels[0] での兄弟。c の右にある
    { Hash: H(ab) = b137985f..., IsLeft: true  },   ← Levels[1] での兄弟。H(cd) の左にある
  ],
}
```

検証する人は `c` と この 2 つのハッシュだけで root を計算し直せます（Step 5）。

```text
H(c)                          = HashLeaf("c")
H(cd)  = HashNode(H(c), H(d))     ← H(d) は右（IsLeft: false）なので後ろに置く
root   = HashNode(H(ab), H(cd))   ← H(ab) は左（IsLeft: true）なので前に置く
```

### 兄弟の位置の求め方

ある階層で自分の位置が `i` のとき、

| `i` | 兄弟の位置 | `IsLeft` |
|---|---|---|
| 偶数 | `i+1`（右） | `false` |
| 奇数 | `i-1`（左） | `true` |

1 つ上の階層での自分の位置は `i / 2` です。

### 持ち上げられた階層

Step 3 で、余ったノードは結合されずに持ち上げられました。そこには兄弟がいないので、**その階層では Sibling を追加しません**。

```text
leaves = [l0, l1, l2, l3, l4] の l4（index 4）

Levels[0]: H0  H1  H2  H3  *H4*     i=4, 偶数だが i+1=5 は存在しない → 追加しない
Levels[1]: H01     H23     *H4*     i=2, 偶数だが i+1=3 は存在しない → 追加しない
Levels[2]: [H0123]         *H4*     i=1, 奇数 → Levels[2][0] を IsLeft: true で追加
Levels[3]: root

Proof{ Index: 4, Siblings: [{ Hash: H0123, IsLeft: true }] }   ← Sibling は 1 個だけ
```

そのため、Proof の長さは index によって変わることがあります。

## 実装する関数

`workspace/merkle/proof.go`（`Sibling`, `Proof`, `ErrIndexOutOfRange` は定義済み）

```go
type Sibling struct {
	Hash   []byte
	IsLeft bool // 兄弟が左側にあるなら true
}

type Proof struct {
	Index    int       // 証明するリーフの位置
	Siblings []Sibling // リーフに近い階層から順に並ぶ
}

// Proof は index 番目のリーフの Merkle Proof を生成する。
// index が範囲外なら ErrIndexOutOfRange を返す。
func (t *Tree) Proof(index int) (Proof, error)
```

- 範囲外の例: `-1`、リーフ数以上、空のツリーへの `Proof(0)`
- リーフ 1 個のツリーの `Proof(0)` は、Sibling が 0 個の Proof

## ヒント

<details>
<summary>ヒント 1: どの階層を見るか</summary>

最上位（root だけの階層）には兄弟がいません。`t.Levels[:len(t.Levels)-1]` の各階層を、下から順に見ます。

</details>

<details>
<summary>ヒント 2: ループの形</summary>

```go
i := index
for _, level := range t.Levels[:len(t.Levels)-1] {
	// i が奇数なら level[i-1] が左の兄弟
	// i が偶数で i+1 < len(level) なら level[i+1] が右の兄弟
	// それ以外は兄弟なし（持ち上げ）
	i /= 2
}
```

</details>

<details>
<summary>ヒント 3: 範囲チェック</summary>

空のツリーでは `t.Levels[0]` にアクセスできないので、先に `len(t.Levels) == 0` を調べます。
エラーは `fmt.Errorf("%w: %d", ErrIndexOutOfRange, index)` のように包むと、呼び出し側が `errors.Is` で判定できます。

</details>

## テストを実行する

```bash
make test STEP=4
```

| テスト | 確かめること |
|---|---|
| `TestStep4_FourLeavesIndex2` | 上の具体例のとおり |
| `TestStep4_SingleLeafHasNoSiblings` | リーフ 1 個なら Sibling は 0 個 |
| `TestStep4_PromotedLeaf` | 5 個の index 4 は Sibling 1 個 |
| `TestStep4_AllSizesAndIndices` | 1〜9 個のすべての index で、Sibling の数・Hash・IsLeft が正しい |
| `TestStep4_OutOfRange` | 範囲外の index と空のツリーでエラー |

## 模範実装

<details>
<summary>模範実装を見る（写経モードの人はここから）</summary>

`Verify` は Step 5 で実装するので、まだ `panic` のままです。

**workspace/merkle/proof.go**

```go
package merkle

import (
	"errors"
	"fmt"
)

// ErrIndexOutOfRange は Proof に範囲外の index が渡されたときのエラー。
var ErrIndexOutOfRange = errors.New("merkle: index out of range")

// Sibling は Proof の 1 階層分。検証時に結合する兄弟ノードのハッシュと、その位置を持つ。
type Sibling struct {
	Hash   []byte
	IsLeft bool // 兄弟が左側にあるなら true
}

// Proof はあるリーフが root に含まれることを示す証明。
type Proof struct {
	Index    int       // 証明するリーフの位置
	Siblings []Sibling // リーフに近い階層から順に並ぶ
}

// Proof は index 番目のリーフの Merkle Proof を生成する。
// index が範囲外なら ErrIndexOutOfRange を返す。
func (t *Tree) Proof(index int) (Proof, error) {
	if len(t.Levels) == 0 || index < 0 || index >= len(t.Levels[0]) {
		return Proof{}, fmt.Errorf("%w: %d", ErrIndexOutOfRange, index)
	}

	proof := Proof{Index: index}
	i := index
	for _, level := range t.Levels[:len(t.Levels)-1] {
		if i%2 == 1 {
			proof.Siblings = append(proof.Siblings, Sibling{Hash: level[i-1], IsLeft: true})
		} else if i+1 < len(level) {
			proof.Siblings = append(proof.Siblings, Sibling{Hash: level[i+1], IsLeft: false})
		}
		// 兄弟がいない（持ち上げられる）階層では何も追加しない。
		i /= 2
	}
	return proof, nil
}

// Verify は leafData が proof を通じて root に含まれるかを検証する。
func Verify(root, leafData []byte, proof Proof) bool {
	panic("TODO: Step 5 Verify")
}
```

</details>

## 解説

- **Proof のサイズ**: 階層数は約 `log2(n)` なので、リーフが 100 万個（約 2^20）でも Sibling は 20 個、合計 640 バイト程度です。
  全データを送る代わりにこれだけで「含まれている」ことを示せるのが Merkle Tree の最大の利点です。
- **`Levels` を保持した理由**: Step 2 で全階層を保存したので、Proof は添字をたどるだけで作れます。
  root しか持っていない場合は、Proof を作るたびに木を作り直す必要があります。
- **`IsLeft` の代わりに `Index` を使う方法**: 実は各階層の `i` の偶奇から左右は決まるので、`Index` とリーフ数があれば `IsLeft` は不要です（RFC 6962 はこの方式）。
  この教材では、検証のロジックを読みやすくするために `IsLeft` を明示しています。

## 完了条件

- [ ] `make test STEP=4` がすべて PASS する
- [ ] `make test STEP=1`〜`3` も引き続き PASS する
- [ ] リーフが 7 個のツリーで index 6 の Proof に Sibling がいくつ入るか、図を描いて答えられる

次は [Step 5: Proof を検証する](05-verify-proof.md) です。
