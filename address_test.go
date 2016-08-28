package address

import (
    "testing"
    "bytes"
)

func TestAddress(t *testing.T) {
    
    alice, err := NewKey()
    if err != nil {
        t.Fatal("error creating key", err)
    }
    
    bob, err := NewKey()
    if err != nil {
       t.Fatal("error creating key", err) 
    }
    
    if bytes.Compare(bob.Marshal(), alice.Marshal()) == 0 {
        t.Error("public keys match")
    }

    address1, err := alice.Address(&bob.Public)
    if err != nil {
        t.Fatal("error getting address")
    }
    
    address2, err := bob.Address(&alice.Public)
    if err != nil {
        t.Fatal("error getting address")
    }
    
    if address1 != address2 {
        t.Error("addresses do not match")
    }
    
    t.Logf("address: %v\n", address1)
}