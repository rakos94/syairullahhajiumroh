package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"time"

	"syairullahhajiumroh/internal/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Document links signed for one jamaah stay valid this long. A leaked link
// stops working afterwards; the detail page gets a fresh one on each load.
const dokumenAksesTTL = 2 * time.Hour

func dokumenSignature(secret string, jamaahID primitive.ObjectID, exp int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("dokumen:" + jamaahID.Hex() + ":" + strconv.FormatInt(exp, 10)))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func newDokumenAkses(secret string, jamaahID primitive.ObjectID) *model.DokumenAkses {
	exp := time.Now().Add(dokumenAksesTTL).Unix()
	return &model.DokumenAkses{Exp: exp, Sig: dokumenSignature(secret, jamaahID, exp)}
}

// validDokumenAkses checks the exp/sig query parameters of a document link.
func validDokumenAkses(secret string, jamaahID primitive.ObjectID, expStr, sig string) bool {
	if expStr == "" || sig == "" {
		return false
	}
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}
	return hmac.Equal([]byte(sig), []byte(dokumenSignature(secret, jamaahID, exp)))
}
