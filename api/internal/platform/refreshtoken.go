package platform

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// NewRefreshTokenValue genera un token opaco de 32 bytes. El valor crudo se
// entrega al cliente una sola vez; el servidor solo guarda su hash.
func NewRefreshTokenValue() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// NewJoinCode genera un codigo corto (8 caracteres) para compartir e
// ingresar a un gimnasio. Base32 evita 0/1/8/9 ambiguos con letras.
func NewJoinCode() (string, error) {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)), nil
}
