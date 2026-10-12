package merkle

// Chunk は data を size バイトずつに分割する。最後のチャンクは size 未満になりうる。
// data が空なら空のスライスを返す。size <= 0 なら panic する。
//
// Step 7: guide/07-file-integrity-cli.md
func Chunk(data []byte, size int) [][]byte {
	panic("TODO: Step 7 Chunk")
}
