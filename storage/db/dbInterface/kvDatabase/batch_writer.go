package kvDatabase

type Batch interface {
	Put(key []byte, value []byte)
	//Write() error
}
