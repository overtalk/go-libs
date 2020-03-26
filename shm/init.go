package shm

type SerializeFunc func(in interface{}) ([]byte, error)
type DeserializeFunc func(bytes []byte) (interface{}, error)

// Resolver is to make transform between interface{} & []byte
type Resolver interface {
	Serialize() ([]byte, error)
	Deserialize(bytes []byte) error
}

// Cache is a common struct for mMap share memory
// serializeFunc & deserializeFunc defines the func to make transform between interface{} & []byte
// so you can store anything you want
// thread safety with a pair of producer & consumer
type Cache struct {
	shm             *SHM
	serializeFunc   SerializeFunc
	deserializeFunc DeserializeFunc
}

// NewCache defines the constructor of Cache
func NewCache(mMapPath string, shmCacheSize uint64) (*Cache, error) {
	sharedMem, err := NewShm(mMapPath, shmCacheSize)
	if err != nil {
		return nil, err
	}

	return &Cache{
		shm:             sharedMem,
		serializeFunc:   Serialize,
		deserializeFunc: Deserialize,
	}, nil

}

// Set func transform interface{} to []byte, and save to mMap safely
func (cache *Cache) Set(data interface{}) error {
	bytes, err := cache.serializeFunc(data)
	if err != nil {
		return err
	}

	return cache.shm.Save(bytes)
}

// Set func get []byte from mMap, and transform to self defined structure by `deserializeFunc`
func (cache *Cache) Get() (interface{}, error) {
	data := cache.shm.Get()
	ret, err := cache.deserializeFunc(data)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// Rewrite replace shm mem with new data
// it's atomicity
func (cache *Cache) Rewrite(data interface{}) error {
	bytes, err := cache.serializeFunc(data)
	if err != nil {
		return err
	}

	return cache.shm.Rewrite(bytes)
}
