package util

import ()

func AppendTo(bA *[]byte, text string) {
	for _, b := range []byte(text) {
		(*bA) = append((*bA), b)
	}
}
