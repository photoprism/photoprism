package tensorflow

import "math/rand/v2"

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}
	return string(result)
}

func GetOne[K comparable, V any](input map[K]V) (*K, *V) {
	for k, v := range input {
		return &k, &v
	}

	return nil, nil
}

func Deref[V any](input *V, defval V) V {
	if input == nil {
		return defval
	}
	return *input
}
