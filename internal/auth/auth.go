package auth

func HashPassword(password string) (string, error) {
	// Implement password hashing logic here, e.g., using bcrypt
	hashed_password_bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed_password_bytes), nil
}

func CheckPasswordHash(password, hash string) error {
	// Implement password hash comparison logic here, e.g., using bcrypt
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}