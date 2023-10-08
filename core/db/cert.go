package db

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-acme/lego/v4/registration"
)

// TODO: Decide where to save the key
type ECPrivateKey struct {
	*ecdsa.PrivateKey
}

func (e ECPrivateKey) MarshalJSON() ([]byte, error) {
	encoded, err := x509.MarshalECPrivateKey(e.PrivateKey)
	if err != nil {
		return nil, err
	}

	// pem file
	// pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: encoded})

	return []byte(fmt.Sprintf(`"%s"`,
		base64.StdEncoding.EncodeToString(encoded),
	)), nil
}

func (e *ECPrivateKey) UnmarshalJSON(data []byte) error {
	// pem file
	// block, _ := pem.Decode(pemData)

	bytes, err := base64.StdEncoding.DecodeString(strings.Trim(string(data), `"`))
	if err != nil {
		return err
	}

	key, err := x509.ParseECPrivateKey(bytes)
	if err != nil {
		return err
	}
	e.PrivateKey = key
	return nil
}

type AcmeUser struct {
	Email        string
	Registration *registration.Resource
	PrivateKey   *ECPrivateKey
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}

func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.PrivateKey.PrivateKey
}

func (u *AcmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *AcmeUser) Save() error {
	return driver.Write("cert", u.Email, u)
}

func GetAcmeUser(email string) (*AcmeUser, error) {
	var acmeUser *AcmeUser
	err := driver.Read("cert", email, acmeUser)
	if err != nil {
		return nil, err
	}

	if acmeUser.PrivateKey == nil {
		privatekey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}
		acmeUser.PrivateKey = &ECPrivateKey{PrivateKey: privatekey}
		acmeUser.Save()
	}

	return acmeUser, nil
}
