package main

import (
    "crypto/elliptic"
    "crypto/rand"
    "fmt"
)

// this program is just to demonstrate public/private keys
// with elliptic curves. It is not needed for Stream.

func main() {
    curve := elliptic.P256()

    bob_private_key, bob_public_key_x, bob_public_key_y, err:=
        elliptic.GenerateKey(curve, rand.Reader)

    if err != nil {
        fmt.Print("error:", err)
        return
    }

    alice_private_key, alice_public_key_x, alice_public_key_y, err:=
        elliptic.GenerateKey(curve, rand.Reader)

    if err != nil {
        fmt.Print("error:", err)
        return
    }

    bob_secret_x, bob_secret_y := curve.ScalarMult(alice_public_key_x, alice_public_key_y,
                                                    bob_private_key)

    alice_secret_x, alice_secret_y := curve.ScalarMult(bob_public_key_x, bob_public_key_y,
                                                    alice_private_key)

    var match bool = true
    if bob_secret_x.Cmp(alice_secret_x) != 0 {
        fmt.Println("x's do not match")
        fmt.Printf("%v\n%v\n",bob_secret_x, alice_secret_x)
        match = false
    }

    if bob_secret_y.Cmp(alice_secret_y) != 0 {
        fmt.Println("y's do not match")
        fmt.Printf("%v\n%v\n",bob_secret_y, alice_secret_y)
        match = false
    }

    if match {
        fmt.Println("secrets match")
    }
}
