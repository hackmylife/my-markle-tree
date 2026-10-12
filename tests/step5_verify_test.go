package tests

import (
	"testing"

	"my_merkle_tree/workspace/merkle"
)

// fourLeavesIndex2 は [a,b,c,d] の c に対する Proof を手で組み立てる。
// Step 5 のテストを Step 4 の実装に依存させないため。
func fourLeavesIndex2() (root []byte, leaf []byte, proof merkle.Proof) {
	a, b, c, d := []byte("a"), []byte("b"), []byte("c"), []byte("d")
	ab := wantNode(wantLeaf(a), wantLeaf(b))
	cd := wantNode(wantLeaf(c), wantLeaf(d))
	proof = merkle.Proof{
		Index: 2,
		Siblings: []merkle.Sibling{
			{Hash: wantLeaf(d), IsLeft: false},
			{Hash: ab, IsLeft: true},
		},
	}
	return wantNode(ab, cd), c, proof
}

// cloneProof は Siblings を複製する。改ざんテストで元の Proof を書き換えないため。
func cloneProof(p merkle.Proof) merkle.Proof {
	out := merkle.Proof{Index: p.Index, Siblings: make([]merkle.Sibling, len(p.Siblings))}
	for i, s := range p.Siblings {
		out.Siblings[i] = merkle.Sibling{Hash: append([]byte{}, s.Hash...), IsLeft: s.IsLeft}
	}
	return out
}

func TestStep5_ValidProof(t *testing.T) {
	root, leaf, proof := fourLeavesIndex2()
	if !merkle.Verify(root, leaf, proof) {
		t.Error("正しい Proof の検証が false になりました")
	}
}

func TestStep5_SingleLeafNoSiblings(t *testing.T) {
	a := []byte("a")
	if !merkle.Verify(wantLeaf(a), a, merkle.Proof{Index: 0}) {
		t.Error("リーフ 1 個（Sibling なし）の検証が false になりました。root == HashLeaf(a) のはずです")
	}
}

func TestStep5_TamperedLeaf(t *testing.T) {
	root, _, proof := fourLeavesIndex2()
	if merkle.Verify(root, []byte("X"), proof) {
		t.Error("改ざんしたリーフデータで検証が true になりました")
	}
}

func TestStep5_WrongRoot(t *testing.T) {
	_, leaf, proof := fourLeavesIndex2()
	if merkle.Verify(wantLeaf([]byte("other")), leaf, proof) {
		t.Error("別の root で検証が true になりました")
	}
}

func TestStep5_TamperedSiblingHash(t *testing.T) {
	root, leaf, proof := fourLeavesIndex2()
	for k := range proof.Siblings {
		bad := cloneProof(proof)
		bad.Siblings[k].Hash[0] ^= 0xff
		if merkle.Verify(root, leaf, bad) {
			t.Errorf("Siblings[%d].Hash を改ざんしても検証が true になりました", k)
		}
	}
}

func TestStep5_FlippedDirection(t *testing.T) {
	root, leaf, proof := fourLeavesIndex2()
	for k := range proof.Siblings {
		bad := cloneProof(proof)
		bad.Siblings[k].IsLeft = !bad.Siblings[k].IsLeft
		if merkle.Verify(root, leaf, bad) {
			t.Errorf("Siblings[%d].IsLeft を反転しても検証が true になりました。左右の結合順を IsLeft で切り替えていますか？", k)
		}
	}
}

func TestStep5_MissingSibling(t *testing.T) {
	root, leaf, proof := fourLeavesIndex2()
	bad := cloneProof(proof)
	bad.Siblings = bad.Siblings[:1]
	if merkle.Verify(root, leaf, bad) {
		t.Error("Sibling が足りない Proof で検証が true になりました")
	}
}

// Step 4 の Proof と組み合わせた往復テスト。
func TestStep5_RoundTripWithProof(t *testing.T) {
	for n := 1; n <= 9; n++ {
		leaves := makeLeaves(n)
		tree := merkle.NewTree(leaves)
		root := tree.Root()
		for i := 0; i < n; i++ {
			proof, err := tree.Proof(i)
			if err != nil {
				t.Fatalf("リーフ %d 個, Proof(%d): %v", n, i, err)
			}
			if !merkle.Verify(root, leaves[i], proof) {
				t.Errorf("リーフ %d 個, index %d: 正しいリーフの検証が false", n, i)
			}
			for j := 0; j < n; j++ {
				if j != i && merkle.Verify(root, leaves[j], proof) {
					t.Errorf("リーフ %d 個: index %d の Proof で leaf-%d の検証が true になりました", n, i, j)
				}
			}
		}
	}
}
