package enc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)
func NewEncryptionKey() []byte {

	keyBuffer := make([]byte, 32)
	io.ReadFull(rand.Reader, keyBuffer)
	return keyBuffer
}

func StreamDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	// Reading the iv from the file
	iv := make([]byte, block.BlockSize())
	if _, err := src.Read(iv); err != nil {
		return 0, err
	}

	var (
		buffer = make([] byte, 32 * 1024)
		stream = cipher.NewCTR(block, iv)
		totalBytes = block.BlockSize()
	)

	for {
		n, err := src.Read(buffer)
		if n > 0 {
			stream.XORKeyStream(buffer, buffer[:n])
			writtenBytes, err := dst.Write(buffer[:n])
			if err != nil {
				return 0, err
			}
			totalBytes = totalBytes + writtenBytes
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}

	return totalBytes, nil
}

func StreamEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	//making initialisation vector
	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}

	// writing the IV + fileContent
	if _, err := dst.Write(iv); err != nil {
		return 0, err
	}

	var (
		buffer = make([] byte, 32 * 1024)
		stream = cipher.NewCTR(block, iv)
		totalBytes = block.BlockSize()
	)

	for {
		n, err := src.Read(buffer)
		if n > 0 {
			stream.XORKeyStream(buffer, buffer[:n])
			writtenBytes, err := dst.Write(buffer[:n])
			if err != nil {
				return 0, err
			}
			totalBytes = totalBytes + writtenBytes
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return totalBytes, nil
}
