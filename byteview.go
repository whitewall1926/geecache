package geecache


type ByteView struct {
	b []byte
}

type Getter interface {
	Get(key string) ([]byte, error) 
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}
func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte {
	new_b := make([]byte, len(v.b))
	copy(new_b, v.b)
	return new_b
}

func (v ByteView) String() string {
	return string(v.b)
}