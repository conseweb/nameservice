/*
Copyright Mojing Inc. 2016 All Rights Reserved.
Written by mint.zhao.chiu@gmail.com. github.com: https://www.github.com/mintzhao

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package passphrase

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	pb "github.com/conseweb/common/protos"
	"golang.org/x/crypto/pbkdf2"
	"math/big"
	"strings"
)

var (
	wordsStore map[pb.PassphraseLanguage][]string
	// Some bitwise operands for working with big.Ints
	Last11BitsMask          = big.NewInt(2047)
	RightShift11BitsDivider = big.NewInt(2048)
	BigOne                  = big.NewInt(1)
	BigTwo                  = big.NewInt(2)
)

func init() {
	wordsStore = make(map[pb.PassphraseLanguage][]string)

	// english
	wordsStore[pb.PassphraseLanguage_English] = strings.Split(words_en, "\n")
	// simplified chinese
	wordsStore[pb.PassphraseLanguage_SimplifiedChinese] = strings.Split(words_zh_SC, "\n")
	// traditional chinese
	wordsStore[pb.PassphraseLanguage_TraditionalChinese] = strings.Split(words_zh_TC, "\n")
	// japanese
	wordsStore[pb.PassphraseLanguage_JAPANESE] = strings.Split(words_jp, "\n")
	// spanish
	wordsStore[pb.PassphraseLanguage_SPANISH] = strings.Split(words_sp, "\n")
	// french
	wordsStore[pb.PassphraseLanguage_FRENCH] = strings.Split(words_fr, "\n")
	// italian
	wordsStore[pb.PassphraseLanguage_ITALIAN] = strings.Split(words_it, "\n")
}

func Passphrase(bitSize int, lang pb.PassphraseLanguage) (string, error) {
	entropy, err := newEntropy(bitSize)
	if err != nil {
		return "", err
	}

	return newMnemonic(entropy, lang)
}

// newEntropy will create random entropy bytes
// so long as the requested size bitSize is an appropriate size.
func newEntropy(bitSize int) ([]byte, error) {
	err := validateEntropyBitSize(bitSize)
	if err != nil {
		return nil, err
	}

	entropy := make([]byte, bitSize/8)
	_, err = rand.Read(entropy)
	return entropy, err
}

// newMnemonic will return a string consisting of the mnemonic words for
// the given entropy.
// If the provide entropy is invalid, an error will be returned.
func newMnemonic(entropy []byte, lang pb.PassphraseLanguage) (string, error) {
	// Compute some lengths for convenience
	entropyBitLength := len(entropy) * 8
	checksumBitLength := entropyBitLength / 32
	sentenceLength := (entropyBitLength + checksumBitLength) / 11

	err := validateEntropyBitSize(entropyBitLength)
	if err != nil {
		return "", err
	}

	// Add checksum to entropy
	entropy = addChecksum(entropy)

	// Break entropy up into sentenceLength chunks of 11 bits
	// For each word AND mask the rightmost 11 bits and find the word at that index
	// Then bitshift entropy 11 bits right and repeat
	// Add to the last empty slot so we can work with LSBs instead of MSB

	// Entropy as an int so we can bitmask without worrying about bytes slices
	entropyInt := new(big.Int).SetBytes(entropy)

	// Slice to hold words in
	words := make([]string, sentenceLength)

	// Throw away big int for AND masking
	word := big.NewInt(0)

	for i := sentenceLength - 1; i >= 0; i-- {
		// Get 11 right most bits and bitshift 11 to the right for next time
		word.And(entropyInt, Last11BitsMask)
		entropyInt.Div(entropyInt, RightShift11BitsDivider)

		// Get the bytes representing the 11 bits as a 2 byte slice
		wordBytes := padByteSlice(word.Bytes(), 2)

		//fmt.Printf("wordBytes: %v, index: %v\n", wordBytes, binary.BigEndian.Uint16(wordBytes))
		// Convert bytes to an index and add that word to the list
		words[i] = wordsStore[lang][binary.BigEndian.Uint16(wordBytes)]
	}

	return strings.Join(words, " "), nil
}

// NewSeed creates a hashed seed output given a provided string and password.
// No checking is performed to validate that the string provided is a valid mnemonic.
func NewSeed(mnemonic string, password string) []byte {
	return pbkdf2.Key([]byte(mnemonic), []byte("mnemonic"+password), 4096, 64, sha512.New)
}

func padByteSlice(slice []byte, length int) []byte {
	newSlice := make([]byte, length-len(slice))
	return append(newSlice, slice...)
}

// Appends to data the first (len(data) / 32)bits of the result of sha256(data)
// Currently only supports data up to 32 bytes
func addChecksum(data []byte) []byte {
	// Get first byte of sha256
	hasher := sha256.New()
	hasher.Write(data)
	hash := hasher.Sum(nil)
	firstChecksumByte := hash[0]

	// len() is in bytes so we divide by 4
	checksumBitLength := uint(len(data) / 4)

	// For each bit of check sum we want we shift the data one the left
	// and then set the (new) right most bit equal to checksum bit at that index
	// staring from the left
	dataBigInt := new(big.Int).SetBytes(data)
	for i := uint(0); i < checksumBitLength; i++ {
		// Bitshift 1 left
		dataBigInt.Mul(dataBigInt, BigTwo)

		// Set rightmost bit if leftmost checksum bit is set
		if uint8(firstChecksumByte&(1<<(7-i))) > 0 {
			dataBigInt.Or(dataBigInt, BigOne)
		}
	}

	return dataBigInt.Bytes()
}

func validateEntropyBitSize(bitSize int) error {
	if (bitSize%32) != 0 || bitSize < 128 || bitSize > 256 {
		return errors.New("Entropy length must be [128, 256] and a multiple of 32")
	}

	return nil
}

func validateEntropyWithChecksumBitSize(bitSize int) error {
	if (bitSize != 128+4) && (bitSize != 160+5) && (bitSize != 192+6) && (bitSize != 224+7) && (bitSize != 256+8) {
		return fmt.Errorf("Wrong entropy + checksum size - expected %v, got %v", int((bitSize-bitSize%32)+(bitSize-bitSize%32)/32), bitSize)
	}
	return nil
}

// IsMnemonicValid attempts to verify that the provided mnemonic is valid.
// Validity is determined by both the number of words being appropriate,
// and that all the words in the mnemonic are present in the word list.
func IsMnemonicValid(mnemonic string, lang pb.PassphraseLanguage) bool {
	// Create a list of all the words in the mnemonic sentence
	words := strings.Fields(mnemonic)

	//Get num of words
	numOfWords := len(words)

	// The number of words should be 12, 15, 18, 21 or 24
	if numOfWords%3 != 0 || numOfWords < 12 || numOfWords > 24 {
		return false
	}

	// Check if all words belong in the wordlist
	for i := 0; i < numOfWords; i++ {
		if !contains(wordsStore[lang], words[i]) {
			return false
		}
	}

	return true
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
