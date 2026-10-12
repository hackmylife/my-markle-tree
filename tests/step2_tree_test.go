package tests

import (
	"bytes"
	"testing"

	"my_merkle_tree/workspace/merkle"
)

func TestStep2_OneLeaf(t *testing.T) {
	a := []byte("a")
	tree := merkle.NewTree([][]byte{a})

	if len(tree.Levels) != 1 {
		t.Fatalf("リーフ 1 個の階層数: got %d, want 1", len(tree.Levels))
	}
	assertHash(t, "リーフ 1 個の root は HashLeaf(a)", tree.Root(), wantLeaf(a))
}

func TestStep2_TwoLeaves(t *testing.T) {
	a, b := []byte("a"), []byte("b")
	tree := merkle.NewTree([][]byte{a, b})

	//      root
	//     /    \
	//   H(a)   H(b)
	want := wantNode(wantLeaf(a), wantLeaf(b))
	assertHash(t, "root = HashNode(HashLeaf(a), HashLeaf(b))", tree.Root(), want)
}

func TestStep2_FourLeaves(t *testing.T) {
	a, b, c, d := []byte("a"), []byte("b"), []byte("c"), []byte("d")
	tree := merkle.NewTree([][]byte{a, b, c, d})

	//            root
	//         /        \
	//     H(ab)        H(cd)
	//     /   \        /   \
	//  H(a)  H(b)   H(c)  H(d)
	ab := wantNode(wantLeaf(a), wantLeaf(b))
	cd := wantNode(wantLeaf(c), wantLeaf(d))
	assertHash(t, "root = HashNode(HashNode(a,b), HashNode(c,d))", tree.Root(), wantNode(ab, cd))
}

func TestStep2_LevelsShape(t *testing.T) {
	leaves := makeLeaves(8)
	tree := merkle.NewTree(leaves)
	want := wantLevels(leaves)

	if len(tree.Levels) != len(want) {
		t.Fatalf("階層数: got %d, want %d（8 → 4 → 2 → 1）", len(tree.Levels), len(want))
	}
	for lv := range want {
		if len(tree.Levels[lv]) != len(want[lv]) {
			t.Fatalf("Levels[%d] の長さ: got %d, want %d", lv, len(tree.Levels[lv]), len(want[lv]))
		}
		for i := range want[lv] {
			assertHash(t, "Levels の各ハッシュ", tree.Levels[lv][i], want[lv][i])
		}
	}
	assertHash(t, "Root() は最上位の階層のハッシュ", tree.Root(), want[len(want)-1][0])
}

func TestStep2_RootChangesWhenAnyLeafChanges(t *testing.T) {
	leaves := makeLeaves(8)
	original := merkle.NewTree(leaves).Root()

	for i := range leaves {
		changed := makeLeaves(8)
		changed[i] = []byte("tampered")
		if bytes.Equal(merkle.NewTree(changed).Root(), original) {
			t.Errorf("リーフ %d を変えても root が変わりません", i)
		}
	}
}

func TestStep2_LeafOrderMatters(t *testing.T) {
	leaves := makeLeaves(4)
	swapped := [][]byte{leaves[1], leaves[0], leaves[2], leaves[3]}

	if bytes.Equal(merkle.NewTree(leaves).Root(), merkle.NewTree(swapped).Root()) {
		t.Error("リーフの順序を入れ替えても root が同じです")
	}
}
