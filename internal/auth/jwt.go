package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/IsaqueB/notification-server/internal/models"
)

var (
	ErrAuthSecretNotDefined       = fmt.Errorf("auth secret not registered")
	ErrInvalidAuthenticationToken = fmt.Errorf("invalid authentication token")
	ErrDecodingHexString          = fmt.Errorf("error decoding secret")
	ErrEmptySecret                = fmt.Errorf("Empty Secret for JWT Auth")
	ErrExpiredAuthenticationToken = fmt.Errorf("Token provided has expired")
	ErrMarshallingData            = fmt.Errorf("Could not marshal data")
)

func JWTAuthentication(token string) (JWTClientPayload, error) {
	tokenSplit := strings.Split(token, ".")
	// Verify token formatting
	if len(tokenSplit) != 3 {
		return JWTClientPayload{}, ErrInvalidAuthenticationToken
	}

	header, payload, signature := tokenSplit[0], tokenSplit[1], tokenSplit[2]
	headerDecoded, err := base64.URLEncoding.DecodeString(header)
	if err != nil {
		return JWTClientPayload{}, ErrInvalidAuthenticationToken
	}

	head := JWTHeader{}
	if err := json.Unmarshal(headerDecoded, &head); err != nil {
		return JWTClientPayload{}, ErrInvalidAuthenticationToken
	}

	switch head.Algorithm {
	case "HS256":
		return VerifyJWT_HS256(header, payload, signature)
	default:
		return JWTClientPayload{}, ErrInvalidAuthenticationToken
	}
}

type JWTHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type JWTClientPayload struct {
	ClientId  string         `json:"client_id"`
	Topics    []models.Topic `json:"topics"`
	ExpiresAt time.Time      `json:"expiresAt"`
}

func Sign_HS256(message []byte, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(message)
	return h.Sum(nil)
}

func VerifyJWT_HS256(header, payload, signature string) (JWTClientPayload, error) {
	if header == "" || payload == "" || signature == "" {
		return JWTClientPayload{}, ErrInvalidAuthenticationToken
	}
	secret := os.Getenv("AUTH_SECRET")
	if secret == "" {
		return JWTClientPayload{}, ErrEmptySecret
	}
	defer func() { secret = "" }()

	signed := base64.RawURLEncoding.EncodeToString(Sign_HS256([]byte(fmt.Sprintf("%v.%v", header, payload)), []byte(secret)))
	if signature != signed {
		return JWTClientPayload{}, ErrInvalidAuthenticationToken
	}
	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return JWTClientPayload{}, ErrInvalidAuthenticationToken
	}
	payloadDecoded := JWTClientPayload{}
	if err := json.Unmarshal(decoded, &payloadDecoded); err != nil {
		return JWTClientPayload{}, ErrInvalidAuthenticationToken
	}
	if payloadDecoded.ClientId == "" {
		return JWTClientPayload{}, ErrInvalidAuthenticationToken
	}
	if time.Now().After(payloadDecoded.ExpiresAt) {
		return JWTClientPayload{}, ErrExpiredAuthenticationToken
	}

	return payloadDecoded, nil
}

func CreateJWT_HS256(clientId string, topics []models.Topic, expiresAt time.Time) (string, error) {
	header := JWTHeader{
		Algorithm: "HS256",
		Type:      "JWT",
	}
	payload := JWTClientPayload{
		ClientId:  clientId,
		Topics:    topics,
		ExpiresAt: expiresAt,
	}
	hRaw, err := json.Marshal(header)
	if err != nil {
		return "", ErrMarshallingData
	}
	pRaw, err := json.Marshal(payload)
	if err != nil {
		return "", ErrMarshallingData
	}

	hEncoded := base64.RawURLEncoding.EncodeToString(hRaw)
	pEncoded := base64.RawURLEncoding.EncodeToString(pRaw)

	secret := os.Getenv("AUTH_SECRET")
	if secret == "" {
		return "", ErrEmptySecret
	}
	defer func() { secret = "" }()

	signature := base64.RawURLEncoding.EncodeToString(
		Sign_HS256(
			[]byte(fmt.Sprintf("%s.%s", hEncoded, pEncoded)),
			[]byte(secret),
		))

	return fmt.Sprintf("%s.%s.%s", hEncoded, pEncoded, signature), nil
}
