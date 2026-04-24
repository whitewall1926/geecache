package geecache

type LocalCache interface {
	Add(key string, value ByteView)
	Get(key string) (ByteView, bool)
}
