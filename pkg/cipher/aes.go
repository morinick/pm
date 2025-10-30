package cipher

import (
	"crypto/aes"
	"crypto/cipher"
)

type AESCipher struct {
	ciph cipher.Block
	key  []byte
}

func New(key []byte) (*AESCipher, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &AESCipher{ciph: c, key: key}, nil
}

func (c *AESCipher) Encrypt(src []byte) []byte {
	srcBlocks := split(src, aes.BlockSize)
	dstBlocks := make([][]byte, len(srcBlocks))
	for i := range dstBlocks {
		dstBlocks[i] = make([]byte, aes.BlockSize)
	}

	for i := range srcBlocks {
		c.ciph.Encrypt(dstBlocks[i], srcBlocks[i])
	}

	return join(dstBlocks, 0)
}

func (c *AESCipher) Decrypt(src []byte) []byte {
	blocks := split(src, aes.BlockSize)

	dstBlocks := make([][]byte, len(blocks))
	for i := range dstBlocks {
		dstBlocks[i] = make([]byte, aes.BlockSize)
	}
	for blocki := range blocks {
		c.ciph.Decrypt(dstBlocks[blocki], blocks[blocki])
	}

	return join(dstBlocks, '-')
}

func split(src []byte, blockSize int) [][]byte {
	blocks := len(src) / blockSize
	if len(src)%blockSize > 0 {
		blocks++
	}

	result := make([][]byte, blocks)
	for i := range result {
		result[i] = make([]byte, blockSize)
	}

	for blocki := range result {
		for i := range blockSize {
			if len(src) > blocki*blockSize+i {
				result[blocki][i] = src[blocki*blockSize+i]
			} else {
				result[blocki][i] = '-'
			}
		}
	}

	return result
}

func join(src [][]byte, suffix byte) []byte {
	result := make([]byte, 0, len(src)*len(src[0]))
	for i := range src {
		result = append(result, src[i]...)
	}
	cutSuffixIndx := len(result)
	if suffix > 0 {
		for i := len(result) - 1; result[i] == suffix; i-- {
			cutSuffixIndx--
		}
	}
	return result[:cutSuffixIndx]
}
