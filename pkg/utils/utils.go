package utils

import (
	"archive/tar"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/apeunit/LaunchControlD/pkg/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/blake2b"
)

// BuildEnvVars generates a sane base for environment variables to run shell
// commands with. It just includes $PATH
func BuildEnvVars(settings *config.Schema) []string {
	path := append([]string{}, settings.Bin(""), os.Getenv("PATH"))
	envPath := fmt.Sprintf("PATH=%s", strings.Join(path, ":"))
	return []string{envPath}
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(folder, url string) (filename string, err error) {
	filename = path.Base(url)
	filePath := filepath.Join(folder, filename)
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer out.Close()
	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return
}

// LoadJSON load json from file into struct
func LoadJSON(filePath string, v interface{}) (err error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, v)
	return
}

// StoreJSON store a struct to a json file
func StoreJSON(filePath string, v interface{}) (err error) {
	data, err := json.Marshal(&v)
	if err != nil {
		return
	}
	// default json permission to rw- --- ---
	err = ioutil.WriteFile(filePath, data, 0600)
	return
}

// FileExists return whenever a file exists
func FileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	}
	return false
}

// DetectContentType detect the content type of a file
func DetectContentType(filePath string) (ct string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	if err != nil {
		return
	}
	ct = http.DetectContentType(buffer)
	return
}

// ExtractGzip extract a gzip file
func ExtractGzip(filePath, outFolder string) (err error) {
	gzipStream, err := os.Open(filePath)
	if err != nil {
		return
	}
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Error("ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Errorf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(filepath.Join(outFolder, header.Name), 0755); err != nil {
				log.Errorf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(filepath.Join(outFolder, header.Name))
			if err != nil {
				log.Errorf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Errorf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
		default:
			log.Errorf(
				"ExtractTarGz: unknown type: %v in %s",
				header.Typeflag,
				header.Name)
		}
	}
	return
}

// Hash calculate the hash of a string
func Hash(data ...interface{}) string {
	hash := blake2b.Sum256([]byte(fmt.Sprint(data...)))
	return hex.EncodeToString(hash[:])
}

// ShortHash calculate the hash of a string (10c)
func ShortHash(data ...string) string {
	hash := blake2b.Sum256([]byte(strings.Join(data, "")))
	return hex.EncodeToString(hash[0:10])
}

// SearchAndMove search a file in a folder and move it to another path
func SearchAndMove(root, file, targetPath string) (err error) {
	found := false
	err = filepath.Walk(root, func(subPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path.Base(subPath) == file {
			found = true
			return os.Rename(subPath, targetPath)
		}
		return err
	})
	if !found {
		err = fmt.Errorf("file %s was not found in %s", file, root)
	}
	return
}

// GetPath build the path to a file
func GetPath(pieces ...string) string {
	return filepath.Join(pieces...)
}

// GenerateRandomHash returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomHash() (h string, err error) {
	b := make([]byte, 512)
	_, err = rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return
	}
	h = Hash(b)
	return
}
