package hash

import "golang.org/x/crypto/bcrypt"

func GenerateFromPassword(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func CompareWithPassword(raw, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw))
}
