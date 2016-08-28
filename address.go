// Copyright 2016 Jeff Macdonald <macfisherman@gmail.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// stream/address is a bitcoin like address
package address

import (
    "math/big"
    "crypto/elliptic"
    "crypto/sha256"
    "crypto/rand"
    "golang.org/x/crypto/ripemd160"
    b58 "github.com/jbenet/go-base58"
    "errors"
)

// Public stores x and y of the public key
// derived from the private key
type Public struct {
    x *big.Int
    y *big.Int
}

// Key contains the curve, private and public key
// however, only the public key is exported.
type Key struct {
    curve elliptic.Curve
    private []byte
    Public
}

// NewKey creates a public/private keypair.
func NewKey() (*Key, error) {
    curve := elliptic.P256()

    private_key, public_key_x, public_key_y, err:=
        elliptic.GenerateKey(curve, rand.Reader)

    if err != nil {
        return nil, err
    }
    
    return &Key{ curve, private_key, Public{ public_key_x, public_key_y}}, nil
}

// PublicKey returns the public key.
func (k *Key) PublicKey() *Public {
    return &k.Public
}

// Marshal returns a byte slice of the public key
// into the form specified in section 4.3.6 of ANSI X9.62.
func (k *Key) Marshal() []byte {
    return elliptic.Marshal(k.curve, k.Public.x, k.Public.y)
} 

//
func (k *Key) UnMarshal(data []byte) *Public {
    x, y := elliptic.Unmarshal(k.curve, data)
    return &Public{ x, y}
}

func (k *Key) Address(p *Public) (string, error) {
    if ! k.curve.IsOnCurve(p.x, p.y) {
        return "", errors.New("invalid public key")
    }
    
    // this is the secret
    x, y := k.curve.ScalarMult(p.x, p.y, k.private)
    raw := []byte{0x04} // non-compressed
    raw = append(raw, x.Bytes()...)
    raw = append(raw, y.Bytes()...)

    digest := sha256.Sum256(raw)
    
    // why different hash implemtations?
    ripe := ripemd160.New()
    ripe.Write(digest[:])
    hashripe := ripe.Sum(nil)
    
    raw = append([]byte{62}, hashripe...)

    address := base58EncodeCheck(raw)
    
    return address, nil
}

func base58EncodeCheck(a []byte) string {
    digest := sha256.Sum256(a)
    digest = sha256.Sum256(digest[:])
    
    b := append(a, digest[:4]...)
    
    return b58.Encode(b)
}