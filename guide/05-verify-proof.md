# Step 5: Proof を検証する

## この step で学ぶこと

検証する人が持っているのは次の 3 つだけです。

- 信頼できる **root**（ブロックヘッダーや公式サイトなどから入手したもの）
- 確かめたい **リーフのデータ**
- 相手から受け取った **Proof**

ツリー全体は持っていません。それでも、リーフから root までを計算し直して、手元の root と一致すれば「このデータは root に含まれている」と確信できます。

### 具体例

`[a, b, c, d]` の root `33376a3b...` に対し、`c` と Step 4 の Proof を検証します。

```text
h = HashLeaf("c")                       = 597fcb31...
Siblings[0] = { H(d),  IsLeft: false }  → h = HashNode(h, H(d))   = dbbd68c3...  (= H(cd))
Siblings[1] = { H(ab), IsLeft: true  }  → h = HashNode(H(ab), h)  = 33376a3b...  (= root)

h == root → true
```

データを `c` から `X` に改ざんすると、最初の `HashLeaf("X")` から値が変わり、最後に root と一致しなくなります。

```text
h = HashLeaf("X") → ... → 最後の h != 33376a3b... → false
```

SHA-256 の衝突は現実的に見つけられないので、「改ざんしたデータで同じ root にたどり着く Proof」を作ることはできません。

## 実装する関数

`workspace/merkle/proof.go`

```go
// Verify は leafData が proof を通じて root に含まれるかを検証する。
func Verify(root, leafData []byte, proof Proof) bool
```

- 結合の順序は `Sibling.IsLeft` だけで決めます。`proof.Index` は使いません
- Sibling が 0 個なら、`HashLeaf(leafData) == root` を比べるだけになります

## ヒント

<details>
<summary>ヒント 1: 全体の流れ</summary>

1. `h := HashLeaf(leafData)`
2. `proof.Siblings` を先頭から順に見て、`IsLeft` に応じて `h = HashNode(兄弟, h)` か `h = HashNode(h, 兄弟)` で更新する
3. 最後に `h` と `root` を比べる

</details>

<details>
<summary>ヒント 2: バイト列の比較</summary>

Go ではスライスを `==` で比べられません。`bytes.Equal(a, b)` を使います。

</details>

## テストを実行する

```bash
make test STEP=5
```

Step 5 のテストの多くは、Proof を **テスト側で手作り** しています。Step 4 の実装に誤りがあっても、Verify の誤りと区別できるようにするためです。
最後の `TestStep5_RoundTripWithProof` だけは Step 4 の `Proof` と組み合わせて確かめます。

| テスト | 確かめること |
|---|---|
| `TestStep5_ValidProof` | 正しい Proof で true |
| `TestStep5_SingleLeafNoSiblings` | Sibling 0 個なら `HashLeaf(a) == root` で true |
| `TestStep5_TamperedLeaf` | 改ざんしたデータで false |
| `TestStep5_WrongRoot` | 別の root で false |
| `TestStep5_TamperedSiblingHash` | Sibling のハッシュを 1 ビット変えると false |
| `TestStep5_FlippedDirection` | `IsLeft` を反転すると false |
| `TestStep5_MissingSibling` | Sibling が足りないと false |
| `TestStep5_RoundTripWithProof` | 1〜9 個の全 index で、正しいリーフは true、別のリーフは false |

## 模範実装

<details>
<summary>模範実装を見る（写経モードの人はここから）</summary>

Step 4 の `proof.go` の import に `"bytes"` を加え、`Verify` を実装します。ファイル全体は次のとおりです。

**workspace/merkle/proof.go**

```go
package merkle

import (
	"bytes"
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
	h := HashLeaf(leafData)
	for _, s := range proof.Siblings {
		if s.IsLeft {
			h = HashNode(s.Hash, h)
		} else {
			h = HashNode(h, s.Hash)
		}
	}
	return bytes.Equal(h, root)
}
```

</details>

## 解説

- **検証のコスト**: ハッシュ計算は「Sibling 数 + 1」回だけです。ツリー全体を作り直す必要はありません。
- **なぜ `Index` を使わないのか**: この教材の `Verify` は `IsLeft` を信じて結合します。
  そのため「このデータが root に含まれる」ことは証明できますが、「**何番目に** 含まれる」ことまでは証明しません。
  位置まで証明したい場合は、検証側が `Index` とリーフ数から左右を自分で計算し、Proof の `IsLeft` を使わないようにします（RFC 6962 の監査証明はこの方式です）。
- **`bytes.Equal` と定数時間比較**: root は秘密の値ではないので、`bytes.Equal` で十分です。
  MAC や認証トークンのように秘密に関わる値を比べる場合は、比較にかかる時間から情報が漏れないよう `crypto/subtle.ConstantTimeCompare` を使います。

## 完了条件

- [ ] `make test STEP=5` がすべて PASS する
- [ ] `make test STEP=1`〜`4` も引き続き PASS する
- [ ] 「Verify が true なら、データは何番目に含まれていると言えるか」に答えられる

次は [Step 6: 第二原像攻撃](06-second-preimage.md) です。
