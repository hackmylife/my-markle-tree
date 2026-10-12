package merkle

// Tree は Merkle Tree の全階層のハッシュを保持する。
type Tree struct {
	// Levels[0] はリーフハッシュの列、Levels[len(Levels)-1] は [root]。
	// リーフが 0 個のときは Levels は空。
	Levels [][][]byte
}

// NewTree はリーフのデータ列から Merkle Tree を構築する。
//
// Step 2: guide/02-build-tree.md （リーフ数が 2 の累乗の場合）
// Step 3: guide/03-any-size.md   （任意の個数に対応）
func NewTree(leaves [][]byte) *Tree {
	panic("TODO: Step 2 NewTree")
}

// Root は root ハッシュを返す。リーフが 0 個のときは sha256("") を返す。
//
// Step 2: guide/02-build-tree.md
// Step 3: guide/03-any-size.md   （リーフ 0 個の場合）
func (t *Tree) Root() []byte {
	panic("TODO: Step 2 Root")
}
