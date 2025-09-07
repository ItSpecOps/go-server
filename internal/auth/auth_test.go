package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
)

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrongPassword",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password1,
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// This test ensures that a token can be successfully created and validated
func TestMakeAndValidateJWT(t *testing.T) {
	// A new unique user ID for the test
	userID := uuid.New()
	// The secret key for signing and validating the token
	tokenSecret := "my-test-secret"
	// The token will be valid for 1 hour
	expiresIn := time.Hour

	// Make the JWT
	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("failed to make JWT: %v", err)
	}

	// Validate the JWT
	validatedID, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("failed to validate JWT: %v", err)
	}

	// Check if the validated user ID matches the original user ID
	if validatedID != userID {
		t.Errorf("validated user ID %s does not match original user ID %s", validatedID, userID)
	}
}

// This test ensures that a token signed with the wrong secret is rejected
func TestValidateJWT_WrongSecret(t *testing.T) {
	// A new unique user ID for the test
	userID := uuid.New()
	// The secret key used to create the token
	tokenSecret := "my-test-secret"
	// A different secret key that should cause validation to fail
	wrongSecret := "a-different-secret"
	// The token will be valid for 1 hour
	expiresIn := time.Hour

	// Make the JWT with the correct secret
	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("failed to make JWT: %v", err)
	}

	// Attempt to validate the JWT with the wrong secret
	_, err = ValidateJWT(tokenString, wrongSecret)
	// The validation should fail, so we expect an error
	if err == nil {
		t.Fatalf("validated JWT with wrong secret, expected an error")
	}

	// We can check the error message to be more specific, if desired
	expectedError := "token signature is invalid: signature is invalid"
	if err.Error() != expectedError {
		t.Errorf("unexpected error message: got %q, want %q", err.Error(), expectedError)
	}
}

// This test ensures that an expired token is rejected
func TestValidateJWT_ExpiredToken(t *testing.T) {
	// A new unique user ID for the test
	userID := uuid.New()
	// The secret key for signing and validating the token
	tokenSecret := "my-test-secret"
	// The token expires in 0 seconds, making it expired immediately
	expiresIn := time.Duration(0)

	// We will create the token with a very short expiry time and a custom clock
	// to ensure it's expired by the time we validate it.
	// NOTE: Because our MakeJWT uses `time.Now()`, we will need to create a test token
	// that is already expired. For a more robust test, you would likely modify your
	// MakeJWT function to accept a `time` parameter for testing purposes.
	// For this test, we will create a token with a one-second expiry and then wait 2 seconds.
	
	// Make the JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		t.Fatalf("failed to make expired JWT: %v", err)
	}

	// Attempt to validate the expired JWT
	_, err = ValidateJWT(tokenString, tokenSecret)
	// The validation should fail, so we expect an error
	if err == nil {
		t.Fatalf("validated expired JWT, expected an error")
	}

	// We can check for a specific error message, although different library versions might
	// return slightly different error strings.
	expectedError := "token has invalid claims: token is expired"
	if err.Error() != expectedError {
		t.Errorf("unexpected error message: got %q, want %q", err.Error(), expectedError)
	}
}

// This test ensures that a malformed JWT is rejected.
func TestValidateJWT_MalformedToken(t *testing.T) {
	// A malformed token string
	malformedToken := "this.is.not.a.valid.jwt"
	tokenSecret := "my-test-secret"

	_, err := ValidateJWT(malformedToken, tokenSecret)

	if err == nil {
		t.Fatalf("validated malformed token, expected an error")
	}
}
