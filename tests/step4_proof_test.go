package tests

import (
	"bytes"
	"testing"

	"my_merkle_tree/workspace/merkle"
)

// wantProof は仕様どおりの Proof を返す。兄弟がいない階層では Sibling を追加しない。
func wantProof(leaves [][]byte, index int) merkle.Proof {
	levels := wantLevels(leaves)
	p := merkle.Proof{Index: index}
	i := index
	for _, level := range levels[:len(levels)-1] {
		switch {
		case i%2 == 1:
			p.Siblings = append(p.Siblings, merkle.Sibling{Hash: level[i-1], IsLeft: true})
		case i+1 < len(level):
			p.Siblings = append(p.Siblings, merkle.Sibling{Hash: level[i+1], IsLeft: false})
		}
		i /= 2
	}
	return p
}

func assertProof(t *testing.T, n, index int, got, want merkle.Proof) {
	t.Helper()
	if got.Index != want.Index {
		t.Errorf("リーフ %d 個, index %d: Proof.Index got %d, want %d", n, index, got.Index, want.Index)
	}
	if len(got.Siblings) != len(want.Siblings) {
		t.Errorf("リーフ %d 個, index %d: Siblings の数 got %d, want %d", n, index, len(got.Siblings), len(want.Siblings))
		return
	}
	for k := range want.Siblings {
		g, w := got.Siblings[k], want.Siblings[k]
		if !bytes.Equal(g.Hash, w.Hash) || g.IsLeft != w.IsLeft {
			t.Errorf("リーフ %d 個, index %d, Siblings[%d]\n  got:  {%s IsLeft=%v}\n  want: {%s IsLeft=%v}",
				n, index, k, short(g.Hash), g.IsLeft, short(w.Hash), w.IsLeft)
		}
	}
}

func TestStep4_FourLeavesIndex2(t *testing.T) {
	a, b, c, d := []byte("a"), []byte("b"), []byte("c"), []byte("d")
	proof, err := merkle.NewTree([][]byte{a, b, c, d}).Proof(2)
	if err != nil {
		t.Fatalf("Proof(2) がエラーを返しました: %v", err)
	}

	//            root
	//         /        \
	//    [H(ab)]       H(cd)
	//     /   \        /   \
	//  H(a)  H(b)  *H(c)* [H(d)]     [ ] が Proof に入る兄弟
	want := merkle.Proof{
		Index: 2,
		Siblings: []merkle.Sibling{
			{Hash: wantLeaf(d), IsLeft: false},
			{Hash: wantNode(wantLeaf(a), wantLeaf(b)), IsLeft: true},
		},
	}
	assertProof(t, 4, 2, proof, want)
}

func TestStep4_SingleLeafHasNoSiblings(t *testing.T) {
	proof, err := merkle.NewTree(makeLeaves(1)).Proof(0)
	if err != nil {
		t.Fatalf("Proof(0) がエラーを返しました: %v", err)
	}
	if len(proof.Siblings) != 0 {
		t.Errorf("リーフ 1 個の Proof に Sibling が %d 個あります。0 個のはずです", len(proof.Siblings))
	}
}

func TestStep4_PromotedLeaf(t *testing.T) {
	leaves := makeLeaves(5)
	proof, err := merkle.NewTree(leaves).Proof(4)
	if err != nil {
		t.Fatalf("Proof(4) がエラーを返しました: %v", err)
	}

	// leaf-4 は Levels[1], Levels[2] へ持ち上げられ、最上位でだけ兄弟と結合される。
	want := merkle.Proof{
		Index:    4,
		Siblings: []merkle.Sibling{{Hash: wantLevels(leaves)[2][0], IsLeft: true}},
	}
	assertProof(t, 5, 4, proof, want)
}

func TestStep4_AllSizesAndIndices(t *testing.T) {
	for n := 1; n <= 9; n++ {
		leaves := makeLeaves(n)
		tree := merkle.NewTree(leaves)
		for i := 0; i < n; i++ {
			got, err := tree.Proof(i)
			if err != nil {
				t.Errorf("リーフ %d 個, Proof(%d) がエラーを返しました: %v", n, i, err)
				continue
			}
			assertProof(t, n, i, got, wantProof(leaves, i))
		}
	}
}

func TestStep4_OutOfRange(t *testing.T) {
	tree := merkle.NewTree(makeLeaves(4))
	for _, i := range []int{-1, 4, 100} {
		if _, err := tree.Proof(i); err == nil {
			t.Errorf("Proof(%d) は範囲外なのでエラーを返すはず", i)
		}
	}
	if _, err := merkle.NewTree(nil).Proof(0); err == nil {
		t.Error("空のツリーの Proof(0) はエラーを返すはず")
	}
}
