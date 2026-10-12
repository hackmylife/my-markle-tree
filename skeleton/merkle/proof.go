package merkle

import "errors"

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
//
// Step 4: guide/04-generate-proof.md
func (t *Tree) Proof(index int) (Proof, error) {
	panic("TODO: Step 4 Proof")
}

// Verify は leafData が proof を通じて root に含まれるかを検証する。
//
// Step 5: guide/05-verify-proof.md
func Verify(root, leafData []byte, proof Proof) bool {
	panic("TODO: Step 5 Verify")
}
