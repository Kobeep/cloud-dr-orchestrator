package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// KeySize is the size of AES-256 key in bytes
	KeySize = 32
	// SaltSize is the size of salt for key derivation
	SaltSize = 32
	// NonceSize is the size of GCM nonce
	NonceSize = 12
	// Iterations for PBKDF2 key derivation
	Iterations = 100000
)

// EncryptFile encrypts a file using AES-256-GCM with a password-derived key
// Returns the path to the encrypted file (original + .encrypted extension)
func EncryptFile(inputPath string, password string) (string, error) {
	// Read input file
	plaintext, err := os.ReadFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to read input file: %w", err)
	}

	// Generate random salt for key derivation
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from password using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, Iterations, KeySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Prepend salt to ciphertext (so we can derive key during decryption)
	encrypted := append(salt, ciphertext...)

	// Write encrypted file
	outputPath := inputPath + ".encrypted"
	if err := os.WriteFile(outputPath, encrypted, 0600); err != nil {
		return "", fmt.Errorf("failed to write encrypted file: %w", err)
	}

	return outputPath, nil
}

// DecryptFile decrypts a file that was encrypted with EncryptFile
// Returns the path to the decrypted file (removes .encrypted extension)
func DecryptFile(inputPath string, password string) (string, error) {
	// Read encrypted file
	encrypted, err := os.ReadFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to read encrypted file: %w", err)
	}

	// Encrypted file format: [salt(32)][nonce(12)][ciphertext]
	if len(encrypted) < SaltSize+NonceSize {
		return "", fmt.Errorf("invalid encrypted file: too short")
	}

	// Extract salt
	salt := encrypted[:SaltSize]
	ciphertext := encrypted[SaltSize:]

	// Derive key from password using the same salt
	key := pbkdf2.Key([]byte(password), salt, Iterations, KeySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce from ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("invalid encrypted file: ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: wrong password or corrupted file: %w", err)
	}

	// Determine output path (remove .encrypted extension if present)
	outputPath := inputPath
	if len(inputPath) > 10 && inputPath[len(inputPath)-10:] == ".encrypted" {
		outputPath = inputPath[:len(inputPath)-10]
	} else {
		outputPath = inputPath + ".decrypted"
	}

	// Write decrypted file
	if err := os.WriteFile(outputPath, plaintext, 0600); err != nil {
		return "", fmt.Errorf("failed to write decrypted file: %w", err)
	}

	return outputPath, nil
}

// GenerateKey generates a random 256-bit encryption key
// Returns base64-encoded key suitable for storing in config files
func GenerateKey() (string, error) {
	key := make([]byte, KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// KeyToPassword converts a base64-encoded key to a password string
// This is used when users provide a pre-generated key instead of a password
func KeyToPassword(base64Key string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return "", fmt.Errorf("invalid base64 key: %w", err)
	}
	if len(key) != KeySize {
		return "", fmt.Errorf("invalid key size: expected %d bytes, got %d", KeySize, len(key))
	}
	return base64Key, nil
}

// IsEncrypted checks if a file appears to be encrypted
// This is a simple heuristic based on file extension
func IsEncrypted(filePath string) bool {
	return len(filePath) > 10 && filePath[len(filePath)-10:] == ".encrypted"
}
