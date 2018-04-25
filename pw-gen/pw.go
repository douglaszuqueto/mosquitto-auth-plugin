package main

import (
	"flag"
	"fmt"

	"../common"
)

// saltSize defines the salt size
const saltSize = 16

func main() {

	var algorithm = flag.String("a", "sha512", "algorithm (sha256 or default: sha512)")
	var HashIterations = flag.Int("i", 100000, "hash iterations (default: 100000)")
	var password = flag.String("p", "", "password")

	flag.Parse()

	pwHash, err := common.Hash(*password, saltSize, *HashIterations, *algorithm)
	if err != nil {
		fmt.Errorf("error: %s\n", err)
	} else {
		fmt.Println(pwHash)
	}

}
