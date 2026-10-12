package tests

import (
	"bytes"
	"testing"

	"my_merkle_tree/workspace/merkle"
)

func TestStep3_Empty(t *testing.T) {
	assertHash(t, "リーフ 0 個の root は sha256(\"\")", merkle.NewTree(nil).Root(), sha(nil))
	assertHash(t, "空スライスでも同じ", merkle.NewTree([][]byte{}).Root(), sha(nil))
}

func TestStep3_ThreeLeaves(t *testing.T) {
	a, b, c := []byte("a"), []byte("b"), []byte("c")
	tree := merkle.NewTree([][]byte{a, b, c})

	//         root
	//        /    \
	//    H(ab)    H(c)   ← H(c) は複製せずそのまま持ち上げる
	//    /   \      |
	//  H(a) H(b)  H(c)
	want := wantNode(wantNode(wantLeaf(a), wantLeaf(b)), wantLeaf(c))
	assertHash(t, "root = HashNode(HashNode(a,b), HashLeaf(c))", tree.Root(), want)
}

func TestStep3_FiveLeaves(t *testing.T) {
	leaves := makeLeaves(5)
	tree := merkle.NewTree(leaves)

	wantShape := []int{5, 3, 2, 1}
	if len(tree.Levels) != len(wantShape) {
		t.Fatalf("階層数: got %d, want %d（5 → 3 → 2 → 1）", len(tree.Levels), len(wantShape))
	}
	for lv, n := range wantShape {
		if len(tree.Levels[lv]) != n {
			t.Fatalf("Levels[%d] の長さ: got %d, want %d", lv, len(tree.Levels[lv]), n)
		}
	}
	assertHash(t, "持ち上げられた Levels[1][2] は HashLeaf(leaf-4) のまま", tree.Levels[1][2], wantLeaf(leaves[4]))
	assertHash(t, "root", tree.Root(), wantRoot(leaves))
}

func TestStep3_AllSizes(t *testing.T) {
	for n := 1; n <= 17; n++ {
		leaves := makeLeaves(n)
		got := merkle.NewTree(leaves).Root()
		if !bytes.Equal(got, wantRoot(leaves)) {
			t.Errorf("リーフ %d 個の root が一致しません\n  got:  %s\n  want: %s", n, short(got), short(wantRoot(leaves)))
		}
	}
}

func TestStep3_NoDuplication(t *testing.T) {
	a, b, c := []byte("a"), []byte("b"), []byte("c")
	three := merkle.NewTree([][]byte{a, b, c}).Root()
	four := merkle.NewTree([][]byte{a, b, c, c}).Root()

	if bytes.Equal(three, four) {
		t.Error("[a,b,c] と [a,b,c,c] の root が同じです。余ったノードを複製していませんか？")
	}
}
