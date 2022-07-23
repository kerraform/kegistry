package model

type Package struct {
	OS            string        `json:"os"`
	Arch          string        `json:"arch"`
	Filename      string        `json:"filename"`
	DownloadURL   string        `json:"download_url"`
	SigningKeys   []SigningKeys `json:"signing_keys"`
	SHASumsURL    string        `json:"shasums_url"`
	SHASumsSigURL string        `json:"shasums_signature_url"`
	SHASum        string        `json:"shasum"`
}

type SigningKeys struct {
	GPGPublicKeys []GPGPublicKey `json:"gpg_public_keys"`
}

type GPGPublicKey struct {
	KeyID      string `json:"key_id"`
	ASCIIArmor string `json:"ascii_armor"`
}
