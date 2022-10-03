// Copyright © 2022 Jinwoo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backend

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/k0kubun/pp/v3"
	"golang.org/x/crypto/sha3"
	"math/big"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"time"
	"unicode"
	"unsafe"
)

func FPrintln(a ...interface{}) {
	_, _ = pp.Println(a...)
	return
}

func ToJsonString(v interface{}) string {
	//s := base64.StdEncoding.EncodeToString(bytes)
	bs, err := json.Marshal(v)
	if err != nil {
		fmt.Errorf("err %v", err)
	}
	return string(bs)
}

func TimeStampNow() string {
	return "0x" + strconv.FormatInt(time.Now().UnixNano()/1000, 16)
}

func TimeStampNowInt() int64 {
	return time.Now().UnixNano() / 1000000
}

func contains(arr []*big.Int, value *big.Int) bool {
	for _, a := range arr {
		if a.Cmp(value) == 0 {
			return true
		}
	}
	return false
}

func ForceDecodeString(s string) []byte {
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		s = s[2:]
	}
	res, _ := hex.DecodeString(s)
	return res
}

func DecodeStringToBytes(s string) ([]byte, error) {
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		s = s[2:]
	}
	res, err := hex.DecodeString(s)

	return res, err
}

func IsValidHexString(s string) bool {

	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		s = s[2:]
	}
	res, _ := hex.DecodeString(s)
	if len(res) > 0 {
		return true
	}
	return false
}

func IsValidIconAddress(s string) bool {
	if len(s) != 42 {
		return false
	}
	if s[:2] != "hx" {
		return false
	}
	s = s[2:]
	if IsValidHexString(s) {
		return true
	}
	return false
}

// ParseBig256 parses s as a 256 bit integer in decimal or hexadecimal syntax.
// Leading zeros are accepted. The empty string parses as zero.
func ParseBig256(s string) (*big.Int, bool) {
	if s == "" {
		return new(big.Int), true
	}
	var bigint *big.Int
	var ok bool
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		bigint, ok = new(big.Int).SetString(s[2:], 16)
	} else {
		bigint, ok = new(big.Int).SetString(s, 10)
	}
	if ok && bigint.BitLen() > 256 {
		bigint, ok = nil, false
	}
	return bigint, ok
}

// MustParseBig256 parses s as a 256 bit big integer and panics if the string is invalid.
func MustParseBig256(s string) *big.Int {
	v, ok := ParseBig256(s)
	if !ok {
		panic("invalid 256 bit integer: " + s)
	}
	return v
}

func ValidNumber(input string) *big.Int {
	if input == "" {
		return big.NewInt(0)
	}

	matched, err := regexp.MatchString("([0-9])", input)
	if !matched || err != nil {
		return nil
	}
	amount := MustParseBig256(input)
	return amount.Abs(amount)
}

func PrintType(x interface{}) string {
	return reflect.TypeOf(x).String()
}

func IsInt(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func RandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func BytesToString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{bh.Data, bh.Len}
	return *(*string)(unsafe.Pointer(&sh))
}

// SHA3Sum256 returns the SHA3-256 digest of the data
func SHA3Sum256(m []byte) []byte {
	d := sha3.Sum256(m)
	return d[:]
}
