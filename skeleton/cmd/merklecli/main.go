// merklecli はファイルをチャンクに分け、Merkle root の計算・Proof の生成・検証を行う。
//
// Step 7: guide/07-file-integrity-cli.md
// main・JSON 変換・loadChunks は実装済み。runRoot / runProof / runVerify を実装する。
package main

import (
	"encoding/hex"
	"errors"
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
	panic("TODO: Step 7 runRoot")
}

// runProof は指定したチャンクの Proof を JSON で出力する。
func runProof(args []string) error {
	panic("TODO: Step 7 runProof")
}

// runVerify はファイルの proof.index 番目のチャンクが root に含まれるかを検証する。
// 検証に失敗したら errVerifyFailed を返す。
func runVerify(args []string) error {
	panic("TODO: Step 7 runVerify")
}
