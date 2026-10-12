# Step 7: ファイル改ざん検知 CLI

## この step で学ぶこと

ここまでで作った Merkle Tree を使って、ファイルの改ざんを検知する CLI を作ります。
大きなファイルを配布するとき、BitTorrent や IPFS は次のようにしています。

1. ファイルを固定サイズの **チャンク** に分ける
2. チャンクをリーフにした Merkle Tree の **root だけ** を信頼できる経路で配る
3. チャンクは信頼できない相手（他のピア）から受け取り、Proof で 1 チャンクずつ検証する

全体を受け取る前でも、届いたチャンクが正しいかをその場で確かめられます。

### 具体例

```text
入力: data = "0123456789"（10 バイト）, size = 4

Chunk(data, 4) = ["0123", "4567", "89"]   ← 最後は 4 バイト未満
```

```bash
$ go run ./workspace/cmd/merklecli root -file sample.bin
e775bffc73a1cdd672a3d4592e71552e420ef4e4d51adcb44db44a666e88f383

$ go run ./workspace/cmd/merklecli proof -file sample.bin -index 2 > proof.json
$ cat proof.json
{
  "index": 2,
  "siblings": [
    { "hash": "293f2269...", "isLeft": false },
    { "hash": "2aff67b9...", "isLeft": true },
    { "hash": "9ed34570...", "isLeft": false }
  ]
}

$ go run ./workspace/cmd/merklecli verify -file sample.bin -root e775bffc... -proof proof.json
OK: チャンク 2 は root に含まれています

# チャンク 2（2048〜3071 バイト目）の中を 1 バイト書き換えたファイルで検証
$ go run ./workspace/cmd/merklecli verify -file tampered.bin -root e775bffc... -proof proof.json
NG: チャンク 2 は root に含まれていません（改ざんの可能性）
```

（`sample.bin` は 5000 バイトのランダムなデータです。チャンクサイズの既定値は 1024 なので、5 チャンクになります）

## 実装する関数

### `workspace/merkle/chunk.go`

```go
// Chunk は data を size バイトずつに分割する。最後のチャンクは size 未満になりうる。
// data が空なら空のスライスを返す。size <= 0 なら panic する。
func Chunk(data []byte, size int) [][]byte
```

### `workspace/cmd/merklecli/main.go`

`main`、JSON との変換（`toProofJSON` / `fromProofJSON`）、`loadChunks` は実装済みです。次の 3 つを実装します。

| 関数 | サブコマンド | 動作 |
|---|---|---|
| `runRoot` | `root -file <path> [-chunk 1024]` | root を 16 進で出力 |
| `runProof` | `proof -file <path> -index <i> [-chunk 1024]` | Proof を JSON で出力 |
| `runVerify` | `verify -file <path> -root <hex> -proof <proof.json> [-chunk 1024]` | `proof.index` 番目のチャンクを検証し、OK / NG を出力。NG なら `errVerifyFailed` を返す |

## ヒント

<details>
<summary>ヒント 1: Chunk のループ</summary>

```go
for start := 0; start < len(data); start += size {
	end := min(start+size, len(data))
	// data[start:end] を追加
}
```

`min` は Go 1.21 からの組み込み関数です。

</details>

<details>
<summary>ヒント 2: サブコマンドごとのフラグ</summary>

`flag.NewFlagSet("root", flag.ExitOnError)` でサブコマンド専用のフラグセットを作り、`fs.String`, `fs.Int` で定義してから `fs.Parse(args)` します。
`main` から渡される `args` は、サブコマンド名より後ろの引数です。

</details>

<details>
<summary>ヒント 3: verify の流れ</summary>

1. `-root` を `hex.DecodeString` でバイト列に戻す
2. `-proof` のファイルを読み、`json.Unmarshal` で `proofJSON` にし、`fromProofJSON` で `merkle.Proof` にする
3. `loadChunks` でチャンクに分け、`proof.Index` が範囲内か確かめる
4. `merkle.Verify(root, chunks[proof.Index], proof)`

`encoding/json` の import を追加するのを忘れずに。

</details>

## テストを実行する

```bash
make test STEP=7
```

| テスト | 確かめること |
|---|---|
| `TestStep7_Chunk` | 割り切れる・端数あり・size がデータより大きい・空データ。連結すると元に戻る |
| `TestStep7_ChunkInvalidSize` | size が 0 や負なら panic |
| `TestStep7_FileIntegrity` | 5000 バイトを 1024 ずつに分け、全チャンクの Proof が通る。1 バイト改ざんしたチャンクだけが NG になる |

CLI は自動テストの対象外です。上の「具体例」の手順で動作を確かめてください。

```bash
head -c 5000 /dev/urandom > sample.bin
cp sample.bin tampered.bin
printf 'X' | dd of=tampered.bin bs=1 seek=2100 conv=notrunc   # チャンク 2 の中を 1 バイト書き換える
```

## 模範実装

<details>
<summary>模範実装を見る（写経モードの人はここから）</summary>

**workspace/merkle/chunk.go**

```go
package merkle

// Chunk は data を size バイトずつに分割する。最後のチャンクは size 未満になりうる。
// data が空なら空のスライスを返す。size <= 0 なら panic する。
func Chunk(data []byte, size int) [][]byte {
	if size <= 0 {
		panic("merkle: chunk size must be positive")
	}

	chunks := make([][]byte, 0, (len(data)+size-1)/size)
	for start := 0; start < len(data); start += size {
		end := min(start+size, len(data))
		chunks = append(chunks, data[start:end])
	}
	return chunks
}
```

**workspace/cmd/merklecli/main.go**

```go
// merklecli はファイルをチャンクに分け、Merkle root の計算・Proof の生成・検証を行う。
package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"my_merkle_tree/workspace/merkle"
)

const usage = `使い方:
  merklecli root   -file <path> [-chunk 1024]
  merklecli proof  -file <path> -index <i> [-chunk 1024]
  merklecli verify -file <path> -root <hex> -proof <proof.json> [-chunk 1024]`

// errVerifyFailed は検証に失敗したことを表す。終了コード 1 で終わらせるために使う。
var errVerifyFailed = errors.New("verify failed")

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "root":
		err = runRoot(os.Args[2:])
	case "proof":
		err = runProof(os.Args[2:])
	case "verify":
		err = runVerify(os.Args[2:])
	default:
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	if errors.Is(err, errVerifyFailed) {
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// proofJSON は Proof を JSON で受け渡すための形。ハッシュは 16 進文字列にする。
type proofJSON struct {
	Index    int           `json:"index"`
	Siblings []siblingJSON `json:"siblings"`
}

type siblingJSON struct {
	Hash   string `json:"hash"`
	IsLeft bool   `json:"isLeft"`
}

func toProofJSON(p merkle.Proof) proofJSON {
	out := proofJSON{Index: p.Index, Siblings: make([]siblingJSON, len(p.Siblings))}
	for i, s := range p.Siblings {
		out.Siblings[i] = siblingJSON{Hash: hex.EncodeToString(s.Hash), IsLeft: s.IsLeft}
	}
	return out
}

func fromProofJSON(pj proofJSON) (merkle.Proof, error) {
	p := merkle.Proof{Index: pj.Index, Siblings: make([]merkle.Sibling, len(pj.Siblings))}
	for i, s := range pj.Siblings {
		h, err := hex.DecodeString(s.Hash)
		if err != nil {
			return merkle.Proof{}, fmt.Errorf("siblings[%d].hash: %w", i, err)
		}
		p.Siblings[i] = merkle.Sibling{Hash: h, IsLeft: s.IsLeft}
	}
	return p, nil
}

// loadChunks はファイルを読み込み、size バイトずつのチャンクに分ける。
func loadChunks(path string, size int) ([][]byte, error) {
	if path == "" {
		return nil, errors.New("-file を指定してください")
	}
	if size <= 0 {
		return nil, fmt.Errorf("-chunk は 1 以上を指定してください: %d", size)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return merkle.Chunk(data, size), nil
}

// runRoot はファイルの Merkle root を 16 進で出力する。
func runRoot(args []string) error {
	fs := flag.NewFlagSet("root", flag.ExitOnError)
	file := fs.String("file", "", "対象ファイル")
	chunk := fs.Int("chunk", 1024, "チャンクサイズ（バイト）")
	fs.Parse(args)

	chunks, err := loadChunks(*file, *chunk)
	if err != nil {
		return err
	}
	fmt.Println(hex.EncodeToString(merkle.NewTree(chunks).Root()))
	return nil
}

// runProof は指定したチャンクの Proof を JSON で出力する。
func runProof(args []string) error {
	fs := flag.NewFlagSet("proof", flag.ExitOnError)
	file := fs.String("file", "", "対象ファイル")
	index := fs.Int("index", 0, "Proof を作るチャンクの番号（0 始まり）")
	chunk := fs.Int("chunk", 1024, "チャンクサイズ（バイト）")
	fs.Parse(args)

	chunks, err := loadChunks(*file, *chunk)
	if err != nil {
		return err
	}
	proof, err := merkle.NewTree(chunks).Proof(*index)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(toProofJSON(proof))
}

// runVerify はファイルの proof.index 番目のチャンクが root に含まれるかを検証する。
// 検証に失敗したら errVerifyFailed を返す。
func runVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	file := fs.String("file", "", "検証するファイル（proof.index 番目のチャンクだけを検証する）")
	rootHex := fs.String("root", "", "信頼できる root（16 進）")
	proofPath := fs.String("proof", "", "proof サブコマンドが出力した JSON ファイル")
	chunk := fs.Int("chunk", 1024, "チャンクサイズ（バイト）")
	fs.Parse(args)

	root, err := hex.DecodeString(*rootHex)
	if err != nil {
		return fmt.Errorf("-root: %w", err)
	}
	raw, err := os.ReadFile(*proofPath)
	if err != nil {
		return err
	}
	var pj proofJSON
	if err := json.Unmarshal(raw, &pj); err != nil {
		return fmt.Errorf("-proof: %w", err)
	}
	proof, err := fromProofJSON(pj)
	if err != nil {
		return err
	}

	chunks, err := loadChunks(*file, *chunk)
	if err != nil {
		return err
	}
	if proof.Index < 0 || proof.Index >= len(chunks) {
		return fmt.Errorf("proof.index %d はチャンク数 %d の範囲外です", proof.Index, len(chunks))
	}

	if !merkle.Verify(root, chunks[proof.Index], proof) {
		fmt.Printf("NG: チャンク %d は root に含まれていません（改ざんの可能性）\n", proof.Index)
		return errVerifyFailed
	}
	fmt.Printf("OK: チャンク %d は root に含まれています\n", proof.Index)
	return nil
}
```

</details>

## 解説

- **Chunk はコピーしない**: `data[start:end]` は元の `data` と同じ配列を参照するスライスです。大きなファイルでもメモリを 2 倍使いません。
  ただし、呼び出し側が `data` を書き換えるとチャンクも変わる点に注意します。
- **`errVerifyFailed`**: 検証失敗は「プログラムのエラー」ではなく「正常に判定した結果」なので、`error:` を付けずに NG メッセージだけ出し、終了コードで失敗を伝えています。
  シェルスクリプトから `if merklecli verify ...; then` のように使えます。
- **この CLI の簡略化**: 本物の配布システムでは、検証する人はファイル全体ではなく **1 チャンクだけ** を受け取ります。
  この CLI はわかりやすさのためにファイル全体から該当チャンクを切り出していますが、`Verify` に渡しているのはそのチャンクと Proof だけで、他のチャンクは使っていません。
- **チャンクサイズのトレードオフ**: 小さくすると改ざん箇所を細かく特定できますが、リーフが増えて Proof が長くなります。BitTorrent v2 は 16 KiB のブロックを使っています。

## 完了条件

- [ ] `make test-all` がすべて PASS する
- [ ] 「具体例」の手順で、改ざんしていないファイルは OK、改ざんしたファイルは NG になることを確かめる
- [ ] `-index` を変えて、改ざんしていないチャンクの検証は改ざん後のファイルでも OK になることを確かめる

## おわりに

お疲れさまでした。これで Merkle Tree の構築・Proof・安全性・応用までを一通り実装しました。
さらに学ぶなら、次のテーマがおすすめです。

- **Sparse Merkle Tree**: 2^256 個のリーフを持つ巨大な木。キーの「存在しないこと」も証明できる
- **Merkle Patricia Trie**: Ethereum の状態ルートで使われる、キーのパスで辿る 16 分木
- **RFC 6962 の Consistency Proof**: 古い root のツリーが新しい root のツリーの先頭部分であることの証明
- **Verkle Tree**: ハッシュの代わりにベクトルコミットメントを使い、Proof を大幅に小さくする木
