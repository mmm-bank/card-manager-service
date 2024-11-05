package generate

import (
	"fmt"
	"math/rand"
	"strconv"
)

func calculateLuhnCheckDigit(number string) string {
	sum := 0
	for i := len(number) - 1; i >= 0; i -= 2 {
		digit := 2 * (number[i] - '0')
		if digit > 9 {
			digit -= 9
		}
		sum += int(digit)
	}
	for i := len(number) - 2; i >= 0; i -= 2 {
		sum += int(number[i] - '0')
	}
	return strconv.Itoa((10 - (sum % 10)) % 10)
}

func CardNumber() string {
	randNumber := rand.Intn(1000000000)
	number := fmt.Sprintf("777777%09d", randNumber)
	return number + calculateLuhnCheckDigit(number)
}

func AccountNumber() string {
	digits := make([]byte, 20)
	for i := 0; i < 20; i++ {
		digits[i] = byte(rand.Intn(10) + '0')
	}
	return string(digits)
}
