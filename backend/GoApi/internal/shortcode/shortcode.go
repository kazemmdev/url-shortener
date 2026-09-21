package shortcode

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
)

// Same scheme as DotnetApi's ShortcodeService: pad the counter, shuffle its
// digits, then base62-encode — so a decoded short code can't be walked
// sequentially to enumerate other links.
const (
	alphabet  = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	Length    = 7
	paddedLen = 9
	salt      = 123
)

func Generate(counter int64) (string, error) {
	padded := fmt.Sprintf("%0*d", paddedLen, counter+salt)

	shuffled, err := shuffle(padded)
	if err != nil {
		return "", err
	}

	number, err := strconv.ParseInt(shuffled, 10, 64)
	if err != nil {
		return "", err
	}

	return toBase62(number), nil
}

func shuffle(s string) (string, error) {
	chars := []byte(s)
	for i := len(chars) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		j := jBig.Int64()
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars), nil
}

func toBase62(value int64) string {
	buf := make([]byte, Length)
	pos := Length

	for value > 0 {
		pos--
		buf[pos] = alphabet[value%62]
		value /= 62
	}
	for pos > 0 {
		pos--
		buf[pos] = alphabet[0]
	}

	return string(buf)
}
