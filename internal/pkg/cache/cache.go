package cache

type Cache interface {
	Get(key string) (interface{}, bool)
	Pull(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl int) bool
	Delete(key string) bool
}
