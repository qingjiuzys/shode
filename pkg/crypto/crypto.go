// Package crypto 提供加密解密功能
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// GenerateRandomBytes 生成随机字节
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	b, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}

// GenerateRandomHex 生成随机十六进制字符串
func GenerateRandomHex(length int) (string, error) {
	b, err := GenerateRandomBytes(length / 2)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashSHA256 计算SHA256哈希
func HashSHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// HashSHA512 计算SHA512哈希
func HashSHA512(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

// HashSHA256Bytes 计算SHA256哈希（字节）
func HashSHA256Bytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// HashSHA512Bytes 计算SHA512哈希（字节）
func HashSHA512Bytes(data []byte) string {
	hash := sha512.Sum512(data)
	return hex.EncodeToString(hash[:])
}

// Encrypt AES加密
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt AES解密
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EncryptString 加密字符串
func EncryptString(plaintext string, key []byte) (string, error) {
	ciphertext, err := Encrypt([]byte(plaintext), key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptString 解密字符串
func DecryptString(ciphertext string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	plaintext, err := Decrypt(data, key)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GenerateKey 生成密钥
func GenerateKey(size int) ([]byte, error) {
	return GenerateRandomBytes(size)
}

// GenerateKeyFromPassword 从密码生成密钥
func GenerateKeyFromPassword(password string, keySize int) []byte {
	hash := sha256.Sum256([]byte(password))
	if keySize > len(hash) {
		keySize = len(hash)
	}
	return hash[:keySize]
}

// Base64Encode Base64编码
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode Base64解码
func Base64Decode(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// Base64URLEncode Base64 URL编码
func Base64URLEncode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// Base64URLDecode Base64 URL解码
func Base64URLDecode(data string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(data)
}

// HexEncode 十六进制编码
func HexEncode(data []byte) string {
	return hex.EncodeToString(data)
}

// HexDecode 十六进制解码
func HexDecode(data string) ([]byte, error) {
	return hex.DecodeString(data)
}

// IsBase64 检查是否为Base64编码
func IsBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// IsHex 检查是否为十六进制编码
func IsHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

// XOREncrypt XOR加密
func XOREncrypt(plaintext []byte, key []byte) []byte {
	ciphertext := make([]byte, len(plaintext))
	keyLen := len(key)

	for i := 0; i < len(plaintext); i++ {
		ciphertext[i] = plaintext[i] ^ key[i%keyLen]
	}

	return ciphertext
}

// XORDecrypt XOR解密（与加密相同）
func XORDecrypt(ciphertext []byte, key []byte) []byte {
	return XOREncrypt(ciphertext, key)
}

// CaesarCipher 凯撒密码
func CaesarCipher(plaintext string, shift int) string {
	result := make([]rune, len(plaintext))

	for i, r := range plaintext {
		if r >= 'a' && r <= 'z' {
			result[i] = 'a' + (r-'a'+rune(shift))%26
		} else if r >= 'A' && r <= 'Z' {
			result[i] = 'A' + (r-'A'+rune(shift))%26
		} else {
			result[i] = r
		}
	}

	return string(result)
}

// Rot13 ROT13加密
func Rot13(plaintext string) string {
	return CaesarCipher(plaintext, 13)
}

// Reverse 反转字符串
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// MaskData 掩码数据
func MaskData(data string, visibleChars int) string {
	runes := []rune(data)
	length := len(runes)

	if length <= visibleChars*2 {
		return strings.Repeat("*", length)
	}

	result := make([]rune, length)

	// 保留开头
	for i := 0; i < visibleChars && i < length; i++ {
		result[i] = runes[i]
	}

	// 掩码中间
	for i := visibleChars; i < length-visibleChars; i++ {
		result[i] = '*'
	}

	// 保留结尾
	for i := length - visibleChars; i < length; i++ {
		result[i] = runes[i]
	}

	return string(result)
}

// IsValidEmail 验证邮箱格式
func IsValidEmail(email string) bool {
	pattern := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return pattern.MatchString(email)
}

// IsValidPhone 验证手机号格式（中国大陆）
func IsValidPhone(phone string) bool {
	pattern := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return pattern.MatchString(phone)
}

// IsValidIDCard 验证身份证号格式（中国大陆）
func IsValidIDCard(idCard string) bool {
	// 18位身份证
	pattern18 := regexp.MustCompile(`^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$`)
	if pattern18.MatchString(idCard) {
		return true
	}

	// 15位身份证
	pattern15 := regexp.MustCompile(`^[1-9]\d{5}\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}$`)
	return pattern15.MatchString(idCard)
}

// IsValidCreditCard 验证信用卡号
func IsValidCreditCard(cardNumber string) bool {
	// 移除空格和横线
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	// 检查长度
	if len(cardNumber) < 13 || len(cardNumber) > 19 {
		return false
	}

	// 检查是否全是数字
	pattern := regexp.MustCompile(`^\d+$`)
	if !pattern.MatchString(cardNumber) {
		return false
	}

	// Luhn算法验证
	return luhnCheck(cardNumber)
}

// luhnCheck Luhn算法检查
func luhnCheck(cardNumber string) bool {
	sum := 0
	alternate := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// HashPassword 哈希密码（简单实现，生产环境应该使用bcrypt）
func HashPassword(password string, salt string) string {
	saltedPassword := password + salt
	return HashSHA256(saltedPassword)
}

// VerifyPassword 验证密码
func VerifyPassword(password, hash, salt string) bool {
	return HashPassword(password, salt) == hash
}

// GenerateSalt 生成盐值
func GenerateSalt() (string, error) {
	salt, err := GenerateRandomBytes(16)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

// CompareHash 比较哈希
func CompareHash(hash1, hash2 string) bool {
	return hash1 == hash2
}

// PBKDF2 PBKDF2密钥派生（简化实现）
func PBKDF2(password, salt string, iterations, keySize int) ([]byte, error) {
	// 这里简化实现，实际应该使用crypto/pbkdf2
	hash := sha256.Sum256([]byte(password + salt))
	return hash[:keySize], nil
}

// HMAC HMAC签名
func HMAC(message, key string) string {
	// 简化实现，实际应该使用crypto/hmac
	return HashSHA256(message + key)
}

// CompareHMAC 比较HMAC
func CompareHMAC(message, key, mac string) bool {
	return HMAC(message, key) == mac
}

// CipherInterface 加密接口
type CipherInterface interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

// AESCipher AES加密器
type AESCipher struct {
	key []byte
}

// NewAESCipher 创建AES加密器
func NewAESCipher(key []byte) (*AESCipher, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, errors.New("key size must be 16, 24, or 32 bytes")
	}

	return &AESCipher{key: key}, nil
}

// Encrypt 加密
func (a *AESCipher) Encrypt(plaintext []byte) ([]byte, error) {
	return Encrypt(plaintext, a.key)
}

// Decrypt 解密
func (a *AESCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	return Decrypt(ciphertext, a.key)
}

// XORCipher XOR加密器
type XORCipher struct {
	key []byte
}

// NewXORCipher 创建XOR加密器
func NewXORCipher(key []byte) *XORCipher {
	return &XORCipher{key: key}
}

// Encrypt 加密
func (x *XORCipher) Encrypt(plaintext []byte) ([]byte, error) {
	return XOREncrypt(plaintext, x.key), nil
}

// Decrypt 解密
func (x *XORCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	return XORDecrypt(ciphertext, x.key), nil
}

// CipherFactory 加密器工厂
type CipherFactory struct {
	ciphers map[string]func(key []byte) (CipherInterface, error)
}

// NewCipherFactory 创建加密器工厂
func NewCipherFactory() *CipherFactory {
	return &CipherFactory{
		ciphers: make(map[string]func(key []byte) (CipherInterface, error)),
	}
}

// Register 注册加密器
func (cf *CipherFactory) Register(name string, factory func(key []byte) (CipherInterface, error)) {
	cf.ciphers[name] = factory
}

// Create 创建加密器
func (cf *CipherFactory) Create(name string, key []byte) (CipherInterface, error) {
	factory, ok := cf.ciphers[name]
	if !ok {
		return nil, fmt.Errorf("cipher not found: %s", name)
	}
	return factory(key)
}

// DefaultCipherFactory 默认加密器工厂
var DefaultCipherFactory = NewCipherFactory()

func init() {
	DefaultCipherFactory.Register("aes", func(key []byte) (CipherInterface, error) {
		return NewAESCipher(key)
	})
	DefaultCipherFactory.Register("xor", func(key []byte) (CipherInterface, error) {
		return NewXORCipher(key), nil
	})
}
