package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Manasik QR tokens let a jamaah show a QR that petugas scan for attendance.
// Format: "SYH1.<jamaah id hex>.<signature>". The HMAC signature means a QR
// can't be forged for another jamaah just by knowing their ID.
const manasikTokenPrefix = "SYH1"

func manasikSignature(secret string, id primitive.ObjectID) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("manasik-qr:" + id.Hex()))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))[:22]
}

func manasikToken(secret string, id primitive.ObjectID) string {
	return manasikTokenPrefix + "." + id.Hex() + "." + manasikSignature(secret, id)
}

// parseManasikToken returns the jamaah ID if the token is valid.
func parseManasikToken(secret, token string) (primitive.ObjectID, bool) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 || parts[0] != manasikTokenPrefix {
		return primitive.NilObjectID, false
	}
	id, err := primitive.ObjectIDFromHex(parts[1])
	if err != nil {
		return primitive.NilObjectID, false
	}
	if !hmac.Equal([]byte(parts[2]), []byte(manasikSignature(secret, id))) {
		return primitive.NilObjectID, false
	}
	return id, true
}
