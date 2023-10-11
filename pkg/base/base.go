package base

type Base interface {
	Save(key string, value string, bucket string) error
	Get(key string, bucket string) (string, error)
	SaveBytes(key string, value []byte, bucket string) error
	GetBytes(key string, bucket string) ([]byte, error)
	GetAll(bucket string) ([][]byte, error)
	Delete(key string, bucket string) error
}
