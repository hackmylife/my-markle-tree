# Step 6: 第二原像攻撃

## この step で学ぶこと

Step 1 で、リーフには `0x00`、内部ノードには `0x01` のプレフィックスを付けました。
この step では、**プレフィックスを付けない素朴な Merkle Tree を自分で攻撃して**、プレフィックスが何を防いでいるのかを確かめます。

### 攻撃対象: NaiveRoot

`workspace/merkle/attack.go` には、プレフィックスなしで root を計算する `NaiveRoot` が **実装済み** で入っています。

```text
リーフ:     sha256(data)            ← プレフィックスなし
内部ノード: sha256(left ‖ right)    ← プレフィックスなし
奇数個:     持ち上げ（NewTree と同じ）
```

### 攻撃の目標

ある root について、別のデータ列で同じ root を作れることを **第二原像（second preimage）** を見つけると言います。
ここでは「元のリーフ列 `x` とは違うのに `NaiveRoot` が同じになるリーフ列」を作ります。

```text
入力:  x = ["a", "b", "c", "d"]
NaiveRoot(x) = 14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7

出力:  forged = [ sha256("a") ‖ sha256("b"),  sha256("c") ‖ sha256("d") ]   ← 64 バイトのリーフが 2 個

       forged[0] = ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb   ← sha256("a")
                   3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d   ← sha256("b")

NaiveRoot(forged) = 14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7   ← 一致！
```

### なぜ一致するのか

```text
元のツリー x:
                 root
              /        \
     sha256(Ha‖Hb)    sha256(Hc‖Hd)        ← 内部ノード
        /   \            /   \
      Ha     Hb        Hc     Hd           ← Ha = sha256("a") など

偽造したツリー forged:
                 root
              /        \
     sha256(Ha‖Hb)    sha256(Hc‖Hd)        ← リーフハッシュ = sha256(リーフのデータ)
         |                |
      "Ha‖Hb"          "Hc‖Hd"             ← 64 バイトのリーフのデータ
```

内部ノード `sha256(Ha ‖ Hb)` と、データが `Ha ‖ Hb` のリーフのハッシュ `sha256(Ha ‖ Hb)` は **同じ計算** です。
NaiveRoot はリーフと内部ノードを区別しないので、1 つ下の階層を切り落としても見分けがつきません。

### 何が困るのか

例えば「root に含まれるトランザクション」を Proof で確かめる仕組みで、攻撃者が `Ha ‖ Hb` という 64 バイトを **1 件のトランザクション** として提示すると、
Proof の検証が通ってしまいます。実在しないトランザクションが「含まれている」ことになります。
Bitcoin では、64 バイトのトランザクションがこの種の攻撃に使えることが知られています（SPV クライアントへの攻撃）。

### プレフィックスで防げる理由

プレフィックス付きの `NewTree` では、

```text
内部ノード:  sha256(0x01 ‖ Ha ‖ Hb)
偽造リーフ:  sha256(0x00 ‖ Ha ‖ Hb)   ← 先頭 1 バイトが違うので別の値
```

となり、偽造したリーフ列の root は元と一致しません。
`[a,b,c,d]` の場合、元の root は `33376a3b...`、偽造したリーフ列の root は `22c47d71...` です。

## 実装する関数

`workspace/merkle/attack.go`

```go
// Forge は NaiveRoot(Forge(leaves)) == NaiveRoot(leaves) となる、元より短いリーフ列を作る。
// len(leaves) >= 2 を前提とする。
func Forge(leaves [][]byte) [][]byte
```

- 隣り合うペア `(leaves[0], leaves[1])`, `(leaves[2], leaves[3])`, ... ごとに、`sha256(左) ‖ sha256(右)` の 64 バイトを 1 つのリーフにする
- 奇数個で余った最後のリーフは、**元のデータのまま** 末尾に残す

```text
入力: ["a", "b", "c"]
出力: [ sha256("a") ‖ sha256("b"),  "c" ]
```

## ヒント

<details>
<summary>ヒント 1: 余ったリーフをそのまま残してよい理由</summary>

NaiveRoot では、余ったリーフ `c` のハッシュ `sha256("c")` は Levels[1] へそのまま持ち上げられます。
偽造リーフ列の Levels[0] にも `sha256("c")` を置けば、偽造側の Levels[0] は元の Levels[1] とまったく同じになります。

</details>

<details>
<summary>ヒント 2: 64 バイトの作り方</summary>

`sha256.Sum256` は `[32]byte` を返します。`append(left[:], right[:]...)` で 64 バイトのスライスにできます。
`left` はローカル変数の配列なので、`left[:]` に append しても他のデータを壊しません（容量が 32 なので新しい配列が確保されます）。

</details>

## テストを実行する

```bash
make test STEP=6
```

| テスト | 確かめること |
|---|---|
| `TestStep6_ForgeShape` | `[a,b,c,d]` から 2 個のリーフを作り、1 つ目が `sha256(a) ‖ sha256(b)` |
| `TestStep6_ForgeBreaksNaiveRoot` | 2〜9 個で、偽造リーフ列は元より短く、`NaiveRoot` が一致する（攻撃成功） |
| `TestStep6_PrefixedTreeResistsForgery` | 同じ偽造リーフ列でも、あなたの `NewTree` では root が一致しない（防御成功） |

## 模範実装

<details>
<summary>模範実装を見る（写経モードの人はここから）</summary>

`NaiveRoot` は雛形のままです。`Forge` だけを書き換えます。ファイル全体は次のとおりです。

**workspace/merkle/attack.go**

```go
package merkle

import "crypto/sha256"

// NaiveRoot はプレフィックスを付けない SHA-256 で root を計算する。
// 奇数個の扱いは NewTree と同じ（余ったノードを持ち上げる）。Step 6 の攻撃対象として実装済みで提供する。
func NaiveRoot(leaves [][]byte) []byte {
	if len(leaves) == 0 {
		empty := sha256.Sum256(nil)
		return empty[:]
	}

	level := make([][]byte, len(leaves))
	for i, leaf := range leaves {
		sum := sha256.Sum256(leaf)
		level[i] = sum[:]
	}

	for len(level) > 1 {
		next := make([][]byte, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			if i+1 == len(level) {
				next = append(next, level[i])
				continue
			}
			pair := make([]byte, 0, len(level[i])+len(level[i+1]))
			pair = append(pair, level[i]...)
			pair = append(pair, level[i+1]...)
			sum := sha256.Sum256(pair)
			next = append(next, sum[:])
		}
		level = next
	}
	return level[0]
}

// Forge は NaiveRoot(Forge(leaves)) == NaiveRoot(leaves) となる、元より短いリーフ列を作る。
// len(leaves) >= 2 を前提とする。
func Forge(leaves [][]byte) [][]byte {
	forged := make([][]byte, 0, (len(leaves)+1)/2)
	for i := 0; i+1 < len(leaves); i += 2 {
		left := sha256.Sum256(leaves[i])
		right := sha256.Sum256(leaves[i+1])
		forged = append(forged, append(left[:], right[:]...))
	}
	if len(leaves)%2 == 1 {
		forged = append(forged, leaves[len(leaves)-1])
	}
	return forged
}
```

</details>

## 解説

- **攻撃は 1 段だけではない**: `Forge(Forge(x))` のように繰り返すと、さらに上の階層を切り落とせます。
  最終的には root の 1 つ下の 2 つのハッシュを連結した、1 個の 64 バイトのリーフになります。
- **対策は 2 通り**:
  1. この教材のようにリーフと内部ノードでプレフィックスを変える（RFC 6962、多くの新しい実装）
  2. リーフのデータ長を内部ノードの入力長（64 バイト）と必ず変える、またはリーフ数を root と一緒にコミットする
- **Step 3 の複製問題との違い**: 複製方式の問題は「同じ形のツリーで、リーフが重複した別のリーフ列」が同じ root になるものでした。
  第二原像攻撃は「ツリーの高さ自体を変える」もので、原因はリーフと内部ノードを区別しないことです。どちらも「root が同じなら同じリーフ列」という前提を壊します。

## 完了条件

- [ ] `make test STEP=6` がすべて PASS する
- [ ] プレフィックスが無いと何が起きるか、図で説明できる

### 発展課題（テストなし）

「プレフィックスを付けること」ではなく「**リーフと内部ノードで違う** プレフィックスを付けること」が対策になっていることを確かめます。

1. `workspace/merkle/hash.go` の `nodePrefix` を一時的に `0x00`（`leafPrefix` と同じ）に変える
2. 小さな `main` かテストを書いて、`x = [a, b, c, d]` と
   `forged = [HashLeaf(a) ‖ HashLeaf(b), HashLeaf(c) ‖ HashLeaf(d)]` の `NewTree(...).Root()` を比べる
3. 一致することを確かめたら、`nodePrefix` を `0x01` に戻す

`Forge` は NaiveRoot 用の偽造（`sha256` で連結）なので、そのままでは使えません。プレフィックス付きのツリーを攻撃するには `HashLeaf` で連結する必要があります。

次は [Step 7: ファイル改ざん検知 CLI](07-file-integrity-cli.md) です。
