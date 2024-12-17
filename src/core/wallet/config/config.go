// MIT License
//
// Copyright (c) 2024 sphinx-core
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

// walletConfig handles the storage and retrieval of keys in the keystore directory.
type walletConfig struct {
	db *leveldb.DB // LevelDB database instance for storing keys.
}

// NewWalletConfig initializes a new walletConfig with a LevelDB instance for key storage.
// It creates the keystore directory if it doesn't exist.
func NewWalletConfig() (*walletConfig, error) {
	// Define the path to the LevelDB database and keystore directory
	keystoreDir := "src/accounts/keystore"

	// Create the keystore directory if it doesn't already exist
	err := os.MkdirAll(keystoreDir, os.ModePerm)
	if err != nil {
		log.Fatal("Failed to create keystore directory:", err) // Log and exit if directory creation fails
		return nil, fmt.Errorf("failed to create keystore directory: %v", err)
	}

	// Open the LevelDB database for storing keys
	db, err := leveldb.OpenFile(keystoreDir+"/sphinxkeys", nil)
	if err != nil {
		log.Fatal("Failed to open LevelDB:", err) // Log and exit if database opening fails
		return nil, fmt.Errorf("failed to open LevelDB: %v", err)
	}

	// Return the walletConfig with the LevelDB instance
	return &walletConfig{db: db}, nil
}

// SaveKeyPair saves a serialized SPHINCS secret (sk) and public (pk) key pair in LevelDB in a .dat file format.
func (config *walletConfig) SaveKeyPair(sk []byte, pk []byte) error {
	if sk == nil || pk == nil {
		return errors.New("secret or public key is nil")
	}

	// Combine the secret and public keys into one byte slice
	combinedKeys := append(sk, pk...)

	// Define the key to store the combined keys (you can use a unique identifier here)
	key := []byte("sphinxKeys")

	// Save the combined keys in a .dat file format inside LevelDB (by storing the byte slice)
	err := config.db.Put(key, combinedKeys, nil)
	if err != nil {
		return fmt.Errorf("failed to save keys in LevelDB: %v", err)
	}

	// Optionally, save the .dat file to the disk as well, if needed
	err = os.WriteFile("sphinxKeys.dat", combinedKeys, 0644)
	if err != nil {
		return fmt.Errorf("failed to save keys to file sphinxKeys.dat: %v", err)
	}

	return nil
}

// LoadKeyPair retrieves the serialized SPHINCS secret (sk) and public (pk) key pair from LevelDB, interpreting it as a .dat file.
func (config *walletConfig) LoadKeyPair() ([]byte, []byte, error) {
	// Define the key used to retrieve the combined keys
	key := []byte("sphinxKeys")

	// Retrieve the combined keys (as a .dat file) from LevelDB
	combinedKeys, err := config.db.Get(key, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load keys from LevelDB: %v", err)
	}

	// Check that the combined keys length matches the expected total length (96 bytes for SK + 48 bytes for PK)
	if len(combinedKeys) != 144 { // 96 bytes for SK + 48 bytes for PK
		return nil, nil, errors.New("invalid combined keys length")
	}

	// Split the keys back into separate secret key (sk) and public key (pk)
	sk := combinedKeys[:96] // First 96 bytes are for the secret key
	pk := combinedKeys[96:] // Last 48 bytes are for the public key

	return sk, pk, nil
}

// Close closes the LevelDB database when done.
func (config *walletConfig) Close() {
	if err := config.db.Close(); err != nil {
		log.Fatal("Failed to close LevelDB:", err)
	}
}
