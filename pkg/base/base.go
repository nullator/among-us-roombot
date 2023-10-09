package base

type Base interface {
	Save(key string, value string, bucket string) error
	Get(key string, bucket string) (string, error)
}
