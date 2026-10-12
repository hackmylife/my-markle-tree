// Package merkle は学習用の Merkle Tree 実装。各関数は guide/ の対応する Step で実装する。
package merkle

// HashLeaf はリーフのデータからリーフハッシュ sha256(0x00 || data) を計算する。
//
// Step 1: guide/01-hash.md
func HashLeaf(data []byte) []byte {
	panic("TODO: Step 1 HashLeaf")
}

// HashNode は左右の子のハッシュから親のハッシュ sha256(0x01 || left || right) を計算する。
//
// Step 1: guide/01-hash.md
func HashNode(left, right []byte) []byte {
	panic("TODO: Step 1 HashNode")
}
