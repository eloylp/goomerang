package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"time"
)

func SelfSignedCertificate() (tls.Certificate, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"goomerang"},
			CommonName:   "127.0.0.1",
		},
		IPAddresses:           []net.IP{net.ParseIP("0.0.0.0")},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		PublicKey:             publicKey(privateKey),
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(privateKey), privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	x509Cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return tls.Certificate{}, err
	}
	tlsCert := tls.Certificate{
		Certificate: [][]byte{x509Cert.Raw},
		PrivateKey:  privateKey,
		Leaf:        x509Cert,
	}
	return tlsCert, nil
}

func publicKey(privateKey interface{}) interface{} {
	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}
