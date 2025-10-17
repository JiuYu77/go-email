package verifier

import (
	"crypto/rand"
	"errors"
	"math/big"
	"regexp"
)

// 验证邮箱格式
// @Return {bool} true 格式正确，false 格式错误
func ValidateFormat(email string) (bool, error) {
	// pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,6}$`
	// pattern := `^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*\.[a-zA-Z0-9]{2,6}$`
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*\.[a-zA-Z]{2,6}$`
	matched, err := regexp.MatchString(pattern, email)
	if err != nil {
		return false, err
	}
	return matched, nil
}

// 预定义字符集
const (
	Numbers      = "0123456789"
	UpperLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LowerLetters = "abcdefghijklmnopqrstuvwxyz"
	AlphaNumeric = Numbers + UpperLetters + LowerLetters
)

// 生成安全码
func GenerateSecureCode(length int, charset string) (code string, err error) {
	tmp := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range length {
		index, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		tmp[i] = charset[index.Int64()]
	}
	code = string(tmp)
	return code, nil
}

// 生成数字验证码
func GenerateNumberCode(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("captcha length can not less than 0")
	}

	return GenerateSecureCode(length, Numbers)
}

// 生成token
func GenerateRandomToken(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("token length can not less than 0")
	}

	return GenerateSecureCode(length, AlphaNumeric)
}
