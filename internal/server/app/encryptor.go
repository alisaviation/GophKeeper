package app

// Encryptor определяет контракт для шифрования данных
type Encryptor interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(encryptedData []byte) ([]byte, error)
}

// NoopEncryptor - заглушка для тестов (не шифрует данные)
type NoopEncryptor struct{}

func (e *NoopEncryptor) Encrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (e *NoopEncryptor) Decrypt(encryptedData []byte) ([]byte, error) {
	return encryptedData, nil
}
