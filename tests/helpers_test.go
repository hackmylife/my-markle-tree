// tests は workspace/merkle の実装を検査する。学習者はこのディレクトリを編集しない。
package tests

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

// sha は prefix と parts を連結した SHA-256 を返す。期待値はすべてこれを使ってテスト側で組み立てる。
func sha(prefix []byte, parts ...[]byte) []byte {
	h := sha256.New()
	h.Write(prefix)
	for _, p := range parts {
		h.Write(p)
	}
	return h.Sum(nil)
}

func wantLeaf(data []byte) []byte        { return sha([]byte{0x00}, data) }
func wantNode(left, right []byte) []byte { return sha([]byte{0x01}, left, right) }

// makeLeaves は "leaf-0", "leaf-1", ... のリーフデータを n 個作る。
func makeLeaves(n int) [][]byte {
	leaves := make([][]byte, n)
	for i := range leaves {
		leaves[i] = fmt.Appendf(nil, "leaf-%d", i)
	}
	return leaves
}

// wantLevels は仕様どおりの各階層を返す（余ったノードは持ち上げる）。
func wantLevels(leaves [][]byte) [][][]byte {
	if len(leaves) == 0 {
		return nil
	}
	level := make([][]byte, len(leaves))
	for i, l := range leaves {
		level[i] = wantLeaf(l)
	}
	levels := [][][]byte{level}
	for len(level) > 1 {
		var next [][]byte
		for i := 0; i < len(level); i += 2 {
			if i+1 == len(level) {
				next = append(next, level[i])
			} else {
				next = append(next, wantNode(level[i], level[i+1]))
			}
		}
		levels = append(levels, next)
		level = next
	}
	return levels
}

func wantRoot(leaves [][]byte) []byte {
	levels := wantLevels(leaves)
	if levels == nil {
		return sha(nil)
	}
	return levels[len(levels)-1][0]
}

func short(b []byte) string {
	s := hex.EncodeToString(b)
	if len(s) > 16 {
		return s[:16] + "..."
	}
	return s
}

func assertHash(t *testing.T, what string, got, want []byte) {
	t.Helper()
	if !bytes.Equal(got, want) {
		t.Errorf("%s\n  got:  %s\n  want: %s", what, short(got), short(want))
	}
}
