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
//
// Step 6: guide/06-second-preimage.md
func Forge(leaves [][]byte) [][]byte {
	panic("TODO: Step 6 Forge")
}
