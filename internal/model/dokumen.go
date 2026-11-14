package model

// DokumenAkses is a short-lived signature that lets the admin UI load a
// jamaah's documents with plain <img>/<a> links (which can't send the
// Authorization header). Returned only to logged-in admins.
type DokumenAkses struct {
	Exp int64  `json:"exp"`
	Sig string `json:"sig"`
}
