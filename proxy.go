package main

import (
	"crypto/aes"
	"crypto/cipher"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"io/ioutil"
	"strconv"
	"bytes"
	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/sha3"
	"golang.org/x/sync/singleflight"

	"github.com/golang/groupcache/lru"
)

func main() {

	// Define the cache size and eviction policy
	const cacheSize = 1 << 20 // 1 MB
	lruCache := lru.New(cacheSize)

	// Create a new groupcache with the defined cache
	/*cache := groupcache.NewGroup("myCache", cacheSize, groupcache.GetterFunc(
		func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
			// Return the cached value if it exists
			if v, ok := lruCache.Get(key); ok {
				dest.SetBytes(v.([]byte))
				return nil
			}

			// Otherwise, return an error
			return fmt.Errorf("key %q not found", key)
		}))*/

	targetStr := flag.String("target", "", "Target server URL (required)")
	username := flag.String("username", "", "Username for basic authentication (optional)")
	password := flag.String("password", "", "Password for basic authentication (optional)")
	port := flag.Int("port", 8080, "Port number to listen on (optional, default: 8080)")
	keyStr := flag.String("key", "", "AES key for decryption (optional)")
	verbose := flag.Bool("verbose", false, "Enable verbose output (optional)")

	flag.Parse()

	if *targetStr == "" {
		fmt.Println("Error: Target URL is required")
		return
	}

	target, err := url.Parse(*targetStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create a proxy handler that forwards requests to the target server
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Set the proxy handler's ModifyResponse function to add authentication
	// to the request if necessary and decrypt the response if necessary
	var block cipher.Block
	if *keyStr != "" {
		// Validate the key length
		if len(*keyStr) != 32 {
			fmt.Println("Error: AES key must be 32 bytes long")
			return
		}
		// Create a cipher block for decryption
		var err error
		block, err = aes.NewCipher([]byte(*keyStr))
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	if *username != "" && *password != "" || block != nil {
		proxy.ModifyResponse = func(resp *http.Response) error {
			// Add the "Authorization" header to the request with basic
			// authentication credentials if necessary
			if *username != "" && *password != "" {
				resp.Header.Set("Authorization", "Basic "+basicAuth(*username, *password))
			}

			// Decrypt the response if necessary
			if block != nil {
				// Create a new reader for the encrypted response body
				encryptedReader := resp.Body
				defer encryptedReader.Close()

				// Use a singleflight group to avoid decrypting the same response concurrently
				g := singleflight.Group{}
				key := hashResponse(resp)
				decrypted, err, _ := g.Do(key, func() (interface{}, error) {
					// Read the encrypted response into a byte slice
					encrypted, err := ioutil.ReadAll(encryptedReader)
					if err != nil {
						return nil, err
					}

					// Decrypt the response
					iv := encrypted[:aes.BlockSize]
					encrypted = encrypted[aes.BlockSize:]
					stream := cipher.NewCFBDecrypter(block, iv)
					stream.XORKeyStream(encrypted, encrypted)

					// Return the decrypted response as a new reader
					return bytes.NewReader(encrypted), nil
				})
				if err != nil {
					return err
				}

				// Set the response body to the decrypted reader
				resp.Body = decrypted.(io.ReadCloser)
			}

			// Use the cache package to store the decrypted response in a cache
			key := hashResponse(resp)
			// Read the response body into a byte slice
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			// Set the value of the key in the cache
			lruCache.Add(key, body)

			if *verbose {
				// Print a message to the standard output
				fmt.Println("Cached response for key:", key)
			}

			return nil
		}
	}

	// Create an HTTP server and set the handler to the proxy
	server := http.Server{
		Addr:    ":" + strconv.Itoa(*port),
		Handler: proxy,
	}

	// Start the server
	fmt.Println("Starting proxy server on port", *port)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
		return
	}
}

// basicAuth returns the base64-encoded basic authentication string for the given
// username and password.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// hashResponse hashes the given HTTP response and returns a string key that can
// be used to cache the response.
func hashResponse(resp *http.Response) string {
	// Hash the status code, headers, and body
	hasher := sha3.New256()
	hasher.Write([]byte(strconv.Itoa(resp.StatusCode)))
	for k, vals := range resp.Header {
		for _, val := range vals {
			hasher.Write([]byte(k))
			hasher.Write([]byte(val))
		}
	}
	body, _ := ioutil.ReadAll(resp.Body)
	hasher.Write(body)

	// Return the hex-encoded hash as the key
	return hex.EncodeToString(hasher.Sum(nil))
}