package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/blacktop/partialzip"

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
	fmt.Println("===> PARSING BuildManifest.plist")
	// fmt.Println("BuildIdentities: ", data.BuildIdentities)
	// fmt.Println("ManifestVersion: ", data.ManifestVersion)
	fmt.Println("ProductVersion: ", data.ProductVersion)
	fmt.Println("ProductBuildVersion: ", data.ProductBuildVersion)
	fmt.Println("SupportedProductTypes: ")
	for _, prodType := range data.SupportedProductTypes {
		fmt.Println(" - ", prodType)
	}
}

func main() {
	pzip, err := partialzip.New(url)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pzip.List())

	n, err := pzip.Get("kernelcache.release.iphone11")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Extracting %s, wrote %d bytes\n", "kernelcache.release.iphone11", n)

	n, err = pzip.Get("BuildManifest.plist")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Extracting %s, wrote %d bytes\n", "BuildManifest.plist", n)

	parseBuildManifest()
}
