package main

import (
	"archive/zip"
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

	"howett.net/plist"
)

const url = "http://updates-http.cdn-apple.com/2018FallFCS/fullrestores/041-11466/CCF56996-C7FE-11E8-84F3-F1077A11A89E/iPhone11,2_12.0.1_16A405_Restore.ipsw"

type buildManifest struct {
	BuildIdentities       interface{} `plist:"BuildIdentities,omitempty"`
	ManifestVersion       uint64      `plist:"ManifestVersion,omitempty"`
	ProductBuildVersion   string      `plist:"ProductBuildVersion,omitempty"`
	ProductVersion        string      `plist:"ProductVersion,omitempty"`
	SupportedProductTypes []string    `plist:"SupportedProductTypes,omitempty"`
}

func parseBuildManifest() {
	dat, err := ioutil.ReadFile("BuildManifest.plist")
	if err != nil {
		log.Fatal(err)
	}
	var data buildManifest
	decoder := plist.NewDecoder(bytes.NewReader(dat))
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(data)
}

func parseRemoteZip() {

	var client http.Client
	var ipswSize int64
	// TODO: make dEndChunk a lot smaller
	var dEndChunk int64 = 10240
	var padding uint64 = 1024

	// get ipsw total size
	req, _ := http.NewRequest("HEAD", url, nil)
	resp, _ := client.Do(req)
	ipswSize = resp.ContentLength

	// get ipsw's directory end bytes
	req, _ = http.NewRequest("GET", url, nil)
	reqRange := fmt.Sprintf("bytes=%d-%d", ipswSize-dEndChunk, ipswSize)
	fmt.Println("reqRange: ", reqRange)
	req.Header.Add("Range", reqRange)
	resp, _ = client.Do(req)

	// fmt.Println(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	readerAt := bytes.NewReader(body)

	// parse ipsw's directory end
	end, err := readDirectoryEnd(readerAt, dEndChunk)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(end)
	// z.r = r
	files := make([]*File, 0, end.directoryRecords)
	// z.Comment = end.comment
	rs := io.NewSectionReader(readerAt, 0, dEndChunk)
	// 3642197889-1024
	offset := int64(dEndChunk - (ipswSize - int64(end.directoryOffset)))
	fmt.Println(offset)
	if _, err = rs.Seek(offset, io.SeekStart); err != nil {
		log.Fatal(err)
	}
	buf := bufio.NewReader(rs)

	// The count of files inside a zip is truncated to fit in a uint16.
	// Gloss over this by reading headers until we encounter
	// a bad one, and then only report an ErrFormat or UnexpectedEOF if
	// the file count modulo 65536 is incorrect.
	for {
		f := &File{zipr: readerAt, zipsize: ipswSize}
		err = readDirectoryHeader(f, buf)
		if err == ErrFormat || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		files = append(files, f)
	}
	if uint16(len(files)) != uint16(end.directoryRecords) { // only compare 16 bits here
		// Return the readDirectoryHeader error if we read
		// the wrong number of directory entries.
		log.Fatal(err)
	}

	for _, file := range files {
		if strings.EqualFold(file.Name, "kernelcache.release.iphone11") {
			fmt.Println(file.headerOffset)
			// get ipsw's directory end bytes
			req, _ = http.NewRequest("GET", url, nil)
			// off, err := file.DataOffset()
			// if err != nil {
			// 	log.Fatal(err)
			// }
			end := uint64(file.headerOffset) + file.CompressedSize64 + padding
			reqRange := fmt.Sprintf("bytes=%d-%d", file.headerOffset, end)
			fmt.Println("reqRange: ", reqRange)
			req.Header.Add("Range", reqRange)
			resp, _ = client.Do(req)

			body, _ = ioutil.ReadAll(resp.Body)
			fmt.Println(len(body))

			dataOffset, err := findBodyOffset(bytes.NewReader(body))
			if err != nil {
				log.Fatal(err)
			}

			enflated, err := ioutil.ReadAll(flate.NewReader(bytes.NewReader(body[dataOffset : uint64(len(body))-padding+dataOffset])))
			if err != nil {
				panic(err)
			}
			// fmt.Println("Enflated:\n", enflated)
			of, err := os.Create("kernelcache.release.iphone11")
			defer of.Close()
			n, err := of.Write(enflated)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("wrote %d bytes\n", n)

		}
	}
	// return nil
}

// kernelcache.release.iphone11
// example/BuildManifest.plist
func getBuildManifest(start, end uint64) error {
	req, _ := http.NewRequest("GET", url, nil)
	bRange := fmt.Sprintf("bytes=%d-%d", start, start+end)
	fmt.Println("bRange: ", bRange)
	req.Header.Add("Range", bRange)
	fmt.Println(req)
	var client http.Client
	resp, _ := client.Do(req)
	fmt.Println(resp)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(len(body))

	enflated, err := ioutil.ReadAll(flate.NewReader(bytes.NewReader(body)))
	if err != nil {
		panic(err)
	}
	fmt.Println("Enflated:\n", enflated)
	of, err := os.Create("BuildManifest.plist")
	defer of.Close()
	n, err := of.Write(enflated)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %d bytes\n", n)
	return nil
}

// 773 827 2024
func main() {
	// Open a zip archive for reading.
	r, err := zip.OpenReader("../iPhone11,2_12.0.1_16A405_Restore.ipsw")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		if !f.FileInfo().IsDir() {
			fmt.Printf("Contents of %s:\n", f.Name)

			off, err := f.DataOffset()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Offset: %d\n", off)
			// rc, err := f.Open()
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// _, err = io.CopyN(os.Stdout, rc, 68)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// rc.Close()
			// fmt.Println()
			parseRemoteZip()
			return
			getBuildManifest(uint64(off), f.CompressedSize64)
			return
			// ipsw, err := os.Open("../iPhone11,2_12.0.1_16A405_Restore.ipsw")
			// defer ipsw.Close()
			// o, err := ipsw.Seek(off, 0)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// buf := make([]byte, f.CompressedSize64)
			// n2, err := ipsw.Read(buf)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// fmt.Printf("%d bytes @ %d: 0x%x\n", n2, o, buf)

			// enflated, err := ioutil.ReadAll(flate.NewReader(bytes.NewReader(buf)))
			// if err != nil {
			// 	panic(err)
			// }
			// fmt.Println("Enflated:\n", enflated)
			// of, err := os.Create(f.Name)
			// defer of.Close()
			// n2, err = of.Write(enflated)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// fmt.Printf("wrote %d bytes\n", n2)

			// buf := []byte{}
			// Create a new zip archive.
			// nr, err := zip.NewReader(bytes.NewReader(buf), int64(f.CompressedSize64))
			// if err != nil {
			// 	fmt.Println(err)
			// }
			// // Register a custom Deflate compressor.
			// nr.RegisterDecompressor(zip.Deflate, func(in io.Reader) io.ReadCloser {
			// 	return flate.NewReader(in)
			// })
			// for _, nf := range nr.File {
			// 	rc, err := nf.Open()
			// 	if err != nil {
			// 		log.Fatal(err)
			// 	}
			// 	_, err = io.CopyN(os.Stdout, rc, 68)
			// 	if err != nil {
			// 		log.Fatal(err)
			// 	}
			// 	rc.Close()
			// 	fmt.Println()
			// }
		}
	}
}
