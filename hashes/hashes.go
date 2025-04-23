package hashes

import "golang.org/x/crypto/bcrypt"

func HashCleartextWithBCrypt(cleartext string) (string, error) {
	hashedCleartext, err := bcrypt.GenerateFromPassword([]byte(cleartext), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedCleartext), nil
}

func ValidateBCryptHashedCleartext(hashedCleartext, cleartextCleartext string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedCleartext), []byte(cleartextCleartext))
}
