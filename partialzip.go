package partialzip

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// PartialZip defines a custom partialzip object
type PartialZip struct {
	URL   string
	Files []PartialFile
}

// FileRange is a partialzip's file's start and end byte range
type FileRange struct {
	Start uint64
	End   uint64
}

// PartialFile is a file inside the remote partialzip
type PartialFile struct {
	URL   string
	Name  string
	Size  uint64
	Range FileRange
}

func getZipSize(url string) int64 {
	var client http.Client
	req, _ := http.NewRequest("HEAD", url, nil)
	resp, _ := client.Do(req)
	return resp.ContentLength
}

// New returns a newly created partialzip object.
func New() *PartialZip {
	pz := &PartialZip{}
	return pz
}

// List lists the files in the remote zip
func (p *PartialZip) List() []string {
	return nil
}

// Get downloads a file from the re mote zip
func (p *PartialFile) Get(path string) error {

	req, _ := http.NewRequest("GET", p.URL, nil)
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", p.Range.Start, p.Range.End))
	fmt.Println(req)
	var client http.Client
	resp, _ := client.Do(req)
	fmt.Println(resp)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(len(body))

	return nil
}
