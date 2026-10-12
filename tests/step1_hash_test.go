package tests

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"my_merkle_tree/workspace/merkle"
)

func TestStep1_HashLeaf(t *testing.T) {
	data := []byte("hello")
	got := merkle.HashLeaf(data)

	if len(got) != 32 {
		t.Fatalf("HashLeaf の長さが %d バイトです。SHA-256 は 32 バイトです", len(got))
	}
	assertHash(t, "HashLeaf(\"hello\") は sha256(0x00 || \"hello\") と一致するはず", got, wantLeaf(data))

	plain := sha256.Sum256(data)
	if bytes.Equal(got, plain[:]) {
		t.Error("HashLeaf がプレフィックスなしの sha256(data) と同じです。先頭に 0x00 を付けてください")
	}
}

func TestStep1_HashLeafEmpty(t *testing.T) {
	assertHash(t, "HashLeaf(nil) は sha256(0x00) と一致するはず", merkle.HashLeaf(nil), wantLeaf(nil))
}

func TestStep1_HashNode(t *testing.T) {
	left := wantLeaf([]byte("a"))
	right := wantLeaf([]byte("b"))

	got := merkle.HashNode(left, right)
	assertHash(t, "HashNode(l, r) は sha256(0x01 || l || r) と一致するはず", got, wantNode(left, right))

	if bytes.Equal(got, merkle.HashNode(right, left)) {
		t.Error("HashNode(l, r) と HashNode(r, l) が同じです。左右の順序で結果が変わるはずです")
	}
}

func TestStep1_LeafAndNodeAreDistinguished(t *testing.T) {
	left := wantLeaf([]byte("a"))
	right := wantLeaf([]byte("b"))
	concat := append(append([]byte{}, left...), right...)

	if bytes.Equal(merkle.HashLeaf(concat), merkle.HashNode(left, right)) {
		t.Error("HashLeaf(l || r) と HashNode(l, r) が同じです。リーフと内部ノードで別のプレフィックスを使ってください")
	}
}
