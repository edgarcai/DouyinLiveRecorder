package crypto

import (
	"crypto/md5"
	"encoding/hex"
	"math"
	"strings"
	"time"

	"github.com/tjfoc/gmsm/sm3"
)

// RC4Encrypt implements the custom RC4 encryption from ab_sign.py
func RC4Encrypt(plaintext string, key string) string {
	s := make([]int, 256)
	for i := 0; i < 256; i++ {
		s[i] = i
	}

	j := 0
	for i := 0; i < 256; i++ {
		j = (j + s[i] + int(key[i%len(key)])) % 256
		s[i], s[j] = s[j], s[i]
	}

	i := 0
	j = 0
	result := make([]byte, 0, len(plaintext))
	for _, char := range []byte(plaintext) {
		i = (i + 1) % 256
		j = (j + s[i]) % 256
		s[i], s[j] = s[j], s[i]
		t := (s[i] + s[j]) % 256
		result = append(result, byte(int(char)^s[t]))
	}

	return string(result)
}

// ResultEncrypt implements the result_encrypt function
func ResultEncrypt(longStr string, num string) string {
	encodingTables := map[string]string{
		"s0": "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=",
		"s1": "Dkdpgh4ZKsQB80/Mfvw36XI1R25+WUAlEi7NLboqYTOPuzmFjJnryx9HVGcaStCe=",
		"s2": "Dkdpgh4ZKsQB80/Mfvw36XI1R25-WUAlEi7NLboqYTOPuzmFjJnryx9HVGcaStCe=",
		"s3": "ckdp1h4ZKsUB80/Mfvw36XIgR25+WQAlEi7NLboqYTOPuzmFjJnryx9HVGDaStCe",
		"s4": "Dkdpgh2ZmsQB80/MfvV36XI1R45-WUAlEixNLwoqYTOPuzKFjJnry79HbGcaStCe",
	}

	masks := []int{16515072, 258048, 4032, 63}
	shifts := []int{18, 12, 6, 0}

	encodingTable := encodingTables[num]
	var result strings.Builder
	roundNum := 0
	longInt := getLongInt(roundNum, longStr)

	totalChars := int(math.Ceil(float64(len(longStr)) / 3.0 * 4.0))

	for i := 0; i < totalChars; i++ {
		if i/4 != roundNum {
			roundNum++
			longInt = getLongInt(roundNum, longStr)
		}
		index := i % 4
		charIndex := (longInt & masks[index]) >> shifts[index]
		if charIndex < len(encodingTable) {
			result.WriteByte(encodingTable[charIndex])
		}
	}

	return result.String()
}

func getLongInt(roundNum int, longStr string) int {
	roundNum = roundNum * 3
	char1 := 0
	if roundNum < len(longStr) {
		char1 = int(longStr[roundNum])
	}
	char2 := 0
	if roundNum+1 < len(longStr) {
		char2 = int(longStr[roundNum+1])
	}
	char3 := 0
	if roundNum+2 < len(longStr) {
		char3 = int(longStr[roundNum+2])
	}
	return (char1 << 16) | (char2 << 8) | char3
}

func generRandom(randomNum int, option []int) []int {
	byte1 := randomNum & 255
	byte2 := (randomNum >> 8) & 255

	return []int{
		(byte1 & 170) | (option[0] & 85),
		(byte1 & 85) | (option[0] & 170),
		(byte2 & 170) | (option[1] & 85),
		(byte2 & 85) | (option[1] & 170),
	}
}

func generateRandomStr() string {
	randomValues := []float64{0.123456789, 0.987654321, 0.555555555}
	var randomBytes []int
	randomBytes = append(randomBytes, generRandom(int(randomValues[0]*10000), []int{3, 45})...)
	randomBytes = append(randomBytes, generRandom(int(randomValues[1]*10000), []int{1, 0})...)
	randomBytes = append(randomBytes, generRandom(int(randomValues[2]*10000), []int{1, 5})...)

	var sb strings.Builder
	for _, b := range randomBytes {
		sb.WriteByte(byte(b))
	}
	return sb.String()
}

// SM3Sum calculates SM3 hash and returns bytes
func SM3Sum(data []byte) []byte {
	h := sm3.New()
	h.Write(data)
	return h.Sum(nil)
}

func generateRc4BbStr(urlSearchParams, userAgent, windowEnvStr string) string {
	suffix := "cus"
	arguments := []int{0, 1, 14}

	startTime := int64(time.Now().UnixNano() / 1e6)

	// 1. url_search_params twice SM3
	hash1 := SM3Sum([]byte(urlSearchParams + suffix))
	urlSearchParamsList := SM3Sum(hash1)

	// 2. suffix twice SM3
	hash2 := SM3Sum([]byte(suffix))
	cus := SM3Sum(hash2)

	// 3. ua processing
	uaKey := string([]byte{0, 1, 14})
	uaEnc := ResultEncrypt(RC4Encrypt(userAgent, uaKey), "s3")
	ua := SM3Sum([]byte(uaEnc))

	endTime := startTime + 100

	// Build config object b (simulated with map/array)
	b := make(map[int]interface{})
	b[8] = 3
	b[10] = endTime
	// b[15] is complex object, simplified for byte generation
	b[16] = startTime
	b[18] = 44
	// b[19] = [1, 0, 1, 5]

	splitToBytes := func(num int64) []int {
		return []int{
			int((num >> 24) & 255),
			int((num >> 16) & 255),
			int((num >> 8) & 255),
			int(num & 255),
		}
	}

	// Process start time
	startTimeBytes := splitToBytes(startTime)
	b[20], b[21], b[22], b[23] = startTimeBytes[0], startTimeBytes[1], startTimeBytes[2], startTimeBytes[3]
	b[24] = int(startTime/256/256/256/256) & 255
	b[25] = int(startTime/256/256/256/256/256) & 255

	// Process arguments
	arg0Bytes := splitToBytes(int64(arguments[0]))
	b[26], b[27], b[28], b[29] = arg0Bytes[0], arg0Bytes[1], arg0Bytes[2], arg0Bytes[3]
	b[30] = int(arguments[1]/256) & 255
	b[31] = (arguments[1] % 256) & 255

	arg1Bytes := splitToBytes(int64(arguments[1]))
	b[32], b[33] = arg1Bytes[0], arg1Bytes[1]

	arg2Bytes := splitToBytes(int64(arguments[2]))
	b[34], b[35], b[36], b[37] = arg2Bytes[0], arg2Bytes[1], arg2Bytes[2], arg2Bytes[3]

	// Process encryption results
	b[38] = int(urlSearchParamsList[21])
	b[39] = int(urlSearchParamsList[22])
	b[40] = int(cus[21])
	b[41] = int(cus[22])
	b[42] = int(ua[23])
	b[43] = int(ua[24])

	// Process end time
	endTimeBytes := splitToBytes(endTime)
	b[44], b[45], b[46], b[47] = endTimeBytes[0], endTimeBytes[1], endTimeBytes[2], endTimeBytes[3]
	b[48] = b[8]
	b[49] = int(endTime/256/256/256/256) & 255
	b[50] = int(endTime/256/256/256/256/256) & 255

	// Config items
	pageId := 110624
	b[51] = pageId
	pageIdBytes := splitToBytes(int64(pageId))
	b[52], b[53], b[54], b[55] = pageIdBytes[0], pageIdBytes[1], pageIdBytes[2], pageIdBytes[3]

	aid := 6383
	b[56] = aid
	b[57] = aid & 255
	b[58] = (aid >> 8) & 255
	b[59] = (aid >> 16) & 255
	b[60] = (aid >> 24) & 255

	// Window env
	windowEnvList := []int{}
	for _, char := range windowEnvStr {
		windowEnvList = append(windowEnvList, int(char))
	}
	b[64] = len(windowEnvList)
	b[65] = b[64].(int) & 255
	b[66] = (b[64].(int) >> 8) & 255

	b[69], b[70], b[71] = 0, 0, 0

	// Checksum
	checksum := b[18].(int) ^ b[20].(int) ^ b[26].(int) ^ b[30].(int) ^ b[38].(int) ^ b[40].(int) ^ b[42].(int) ^ b[21].(int) ^ b[27].(int) ^ b[31].(int) ^
		b[35].(int) ^ b[39].(int) ^ b[41].(int) ^ b[43].(int) ^ b[22].(int) ^ b[28].(int) ^ b[32].(int) ^ b[36].(int) ^ b[23].(int) ^ b[29].(int) ^
		b[33].(int) ^ b[37].(int) ^ b[44].(int) ^ b[45].(int) ^ b[46].(int) ^ b[47].(int) ^ b[48].(int) ^ b[49].(int) ^ b[50].(int) ^ b[24].(int) ^
		b[25].(int) ^ b[52].(int) ^ b[53].(int) ^ b[54].(int) ^ b[55].(int) ^ b[57].(int) ^ b[58].(int) ^ b[59].(int) ^ b[60].(int) ^ b[65].(int) ^
		b[66].(int) ^ b[70].(int) ^ b[71].(int)

	b[72] = checksum

	// Build final byte array
	bb := []int{
		b[18].(int), b[20].(int), b[52].(int), b[26].(int), b[30].(int), b[34].(int), b[58].(int), b[38].(int), b[40].(int), b[53].(int), b[42].(int), b[21].(int),
		b[27].(int), b[54].(int), b[55].(int), b[31].(int), b[35].(int), b[57].(int), b[39].(int), b[41].(int), b[43].(int), b[22].(int), b[28].(int), b[32].(int),
		b[60].(int), b[36].(int), b[23].(int), b[29].(int), b[33].(int), b[37].(int), b[44].(int), b[45].(int), b[59].(int), b[46].(int), b[47].(int), b[48].(int),
		b[49].(int), b[50].(int), b[24].(int), b[25].(int), b[65].(int), b[66].(int), b[70].(int), b[71].(int),
	}
	bb = append(bb, windowEnvList...)
	bb = append(bb, b[72].(int))

	var sb strings.Builder
	for _, byteVal := range bb {
		sb.WriteByte(byte(byteVal))
	}

	return RC4Encrypt(sb.String(), string([]byte{121}))
}

// GenerateSignature generates the a_bogus signature
func GenerateSignature(urlSearchParams, userAgent string) string {
	windowEnvStr := "1920|1080|1920|1040|0|30|0|0|1872|92|1920|1040|1857|92|1|24|Win32"

	randomStr := generateRandomStr()
	rc4BbStr := generateRc4BbStr(urlSearchParams, userAgent, windowEnvStr)

	return ResultEncrypt(randomStr+rc4BbStr, "s4") + "="
}

// Helper for MD5
func MD5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
