package tests

import (
	"bytes"
	"testing"

	"my_merkle_tree/workspace/merkle"
)

func assertChunkLens(t *testing.T, name string, chunks [][]byte, want []int) {
	t.Helper()
	got := make([]int, len(chunks))
	for i, c := range chunks {
		got[i] = len(c)
	}
	if len(got) != len(want) {
		t.Errorf("%s: チャンクの長さ got %v, want %v", name, got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s: チャンクの長さ got %v, want %v", name, got, want)
			return
		}
	}
}

func TestStep7_Chunk(t *testing.T) {
	data := []byte("0123456789")

	assertChunkLens(t, "10 バイトを 5 ずつ", merkle.Chunk(data, 5), []int{5, 5})
	assertChunkLens(t, "10 バイトを 4 ずつ", merkle.Chunk(data, 4), []int{4, 4, 2})
	assertChunkLens(t, "10 バイトを 100 ずつ", merkle.Chunk(data, 100), []int{10})
	assertChunkLens(t, "空データ", merkle.Chunk(nil, 4), []int{})

	if got := bytes.Join(merkle.Chunk(data, 4), nil); !bytes.Equal(got, data) {
		t.Errorf("チャンクを連結すると元のデータに戻るはず: got %q", got)
	}
}

func TestStep7_ChunkInvalidSize(t *testing.T) {
	for _, size := range []int{0, -1} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Chunk(data, %d) は panic するはず", size)
				}
			}()
			merkle.Chunk([]byte("abc"), size)
		}()
	}
}

func TestStep7_FileIntegrity(t *testing.T) {
	data := make([]byte, 5000)
	for i := range data {
		data[i] = byte(i * 7)
	}
	chunks := merkle.Chunk(data, 1024)
	tree := merkle.NewTree(chunks)
	root := tree.Root()

	for i := range chunks {
		proof, err := tree.Proof(i)
		if err != nil {
			t.Fatalf("Proof(%d): %v", i, err)
		}
		if !merkle.Verify(root, chunks[i], proof) {
			t.Errorf("チャンク %d の検証が false", i)
		}
	}

	tampered := append([]byte{}, data...)
	tampered[2*1024+10] ^= 0x01
	tamperedChunks := merkle.Chunk(tampered, 1024)

	if bytes.Equal(merkle.NewTree(tamperedChunks).Root(), root) {
		t.Error("1 バイト改ざんしても root が変わりません")
	}
	proof, _ := tree.Proof(2)
	if merkle.Verify(root, tamperedChunks[2], proof) {
		t.Error("改ざんしたチャンク 2 の検証が true になりました")
	}
	proof0, _ := tree.Proof(0)
	if !merkle.Verify(root, tamperedChunks[0], proof0) {
		t.Error("改ざんしていないチャンク 0 は検証が true のはず")
	}
}
