package tests

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"my_merkle_tree/workspace/merkle"
)

func TestStep6_ForgeShape(t *testing.T) {
	a, b, c, d := []byte("a"), []byte("b"), []byte("c"), []byte("d")
	forged := merkle.Forge([][]byte{a, b, c, d})

	if len(forged) != 2 {
		t.Fatalf("[a,b,c,d] の偽造リーフ数: got %d, want 2", len(forged))
	}
	ha, hb := sha256.Sum256(a), sha256.Sum256(b)
	assertHash(t, "forged[0] は sha256(a) || sha256(b) の 64 バイト", forged[0], append(ha[:], hb[:]...))
}

func TestStep6_ForgeBreaksNaiveRoot(t *testing.T) {
	for n := 2; n <= 9; n++ {
		leaves := makeLeaves(n)
		forged := merkle.Forge(leaves)

		if len(forged) >= len(leaves) {
			t.Errorf("リーフ %d 個: 偽造リーフ列は元より短いはず（got %d 個）", n, len(forged))
		}
		if !bytes.Equal(merkle.NaiveRoot(forged), merkle.NaiveRoot(leaves)) {
			t.Errorf("リーフ %d 個: NaiveRoot(Forge(x)) が NaiveRoot(x) と一致しません。攻撃が成立していません", n)
		}
	}
}

func TestStep6_PrefixedTreeResistsForgery(t *testing.T) {
	for n := 2; n <= 9; n++ {
		leaves := makeLeaves(n)
		forged := merkle.Forge(leaves)

		if bytes.Equal(merkle.NewTree(forged).Root(), merkle.NewTree(leaves).Root()) {
			t.Errorf("リーフ %d 個: プレフィックス付きの NewTree でも偽造リーフ列の root が一致しました", n)
		}
	}
}
