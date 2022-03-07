package validation

func Luhn(number string) bool {
	var sum int64
	for i := 1; i <= len(number); i++ {
		digit := number[len(number)-i] - '0'
		if i%2 == 0 {
			if digit = digit * 2; digit > 9 {
				digit -= 9
			}
		}
		sum += int64(digit)
	}
	return sum%10 == 0
}
