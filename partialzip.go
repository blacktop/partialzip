package partialzip

import (
	"bufio"
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// PartialZip defines a custom partialzip object
type PartialZip struct {
	URL   string
	Size  int64
	Files []*File
}

func (p *PartialZip) init() error {
	var client http.Client
	var chuck int64 = 10 * 1024

	// get remote zip size
	req, _ := http.NewRequest("HEAD", p.URL, nil)
	resp, _ := client.Do(req)
	p.Size = resp.ContentLength

	// pull chuck from end of remote zip
	reqRange := fmt.Sprintf("bytes=%d-%d", p.Size-chuck, p.Size)
	req, _ = http.NewRequest("GET", p.URL, nil)
	req.Header.Add("Range", reqRange)
	resp, _ = client.Do(req)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read http response body")
	}
	defer resp.Body.Close()

	r := bytes.NewReader(body)

	// parse zip's directory end
	end, err := readDirectoryEnd(r, chuck)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(end)
	// z.r = r
	p.Files = make([]*File, 0, end.directoryRecords)
	// z.Comment = end.comment
	rs := io.NewSectionReader(r, 0, chuck)
	// 3642197889-1024
	offset := int64(chuck - (p.Size - int64(end.directoryOffset)))
	// fmt.Println(offset)
	if _, err = rs.Seek(offset, io.SeekStart); err != nil {
		log.Fatal(err)
	}
	buf := bufio.NewReader(rs)

	// The count of files inside a zip is truncated to fit in a uint16.
	// Gloss over this by reading headers until we encounter
	// a bad one, and then only report an ErrFormat or UnexpectedEOF if
	// the file count modulo 65536 is incorrect.
	for {
		f := &File{zipr: r, zipsize: p.Size}
		err = readDirectoryHeader(f, buf)
		if err == ErrFormat || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		p.Files = append(p.Files, f)
	}
	if uint16(len(p.Files)) != uint16(end.directoryRecords) { // only compare 16 bits here
		// Return the readDirectoryHeader error if we read
		// the wrong number of directory entries.
		log.Fatal(err)
	}

	return nil
}

// New returns a newly created partialzip object.
func New(url string) (*PartialZip, error) {

	pz := &PartialZip{URL: url}

	err := pz.init()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read http response body")
	}

	return pz, nil
}

// List lists the files in the remote zip
func (p *PartialZip) List() []string {
	filePaths := []string{}

	for _, file := range p.Files {
		filePaths = append(filePaths, file.Name)
	}

	return filePaths
}

// Get downloads a file from the re mote zip
func (p *PartialZip) Get(path string) error {

	var client http.Client
	var padding uint64 = 1024

	for _, file := range p.Files {
		if strings.EqualFold(file.Name, path) {
			// fmt.Println(file.headerOffset)
			// get ipsw's directory end bytes
			req, _ := http.NewRequest("GET", p.URL, nil)
			// off, err := file.DataOffset()
			// if err != nil {
			// 	log.Fatal(err)
			// }
			end := uint64(file.headerOffset) + file.CompressedSize64 + padding
			reqRange := fmt.Sprintf("bytes=%d-%d", file.headerOffset, end)
			// fmt.Println("reqRange: ", reqRange)
			req.Header.Add("Range", reqRange)
			resp, _ := client.Do(req)

			body, _ := ioutil.ReadAll(resp.Body)
			// fmt.Println(len(body))

			dataOffset, err := findBodyOffset(bytes.NewReader(body))
			if err != nil {
				log.Fatal(err)
			}

			enflated, err := ioutil.ReadAll(flate.NewReader(bytes.NewReader(body[dataOffset : uint64(len(body))-padding+dataOffset])))
			if err != nil {
				panic(err)
			}

			of, err := os.Create(path)
			defer of.Close()
			n, err := of.Write(enflated)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Extracting %s, wrote %d bytes\n", path, n)
		}
	}

	return nil
}
