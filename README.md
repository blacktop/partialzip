# partialzip

[![GoDoc](https://godoc.org/github.com/blacktop/partialzip?status.svg)](https://godoc.org/github.com/blacktop/partialzip) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

> Partial Implementation of PartialZip in Go

---

## Why?

If you need to download a VERY large zip, but you only need some small file from it, this is a nice optimization to allow you to only extract what you need and nothing else. Saving time, bandwidth and space.

## Install

```bash
go get github.com/blacktop/partialzip
```

## Example

```golang
import (
    "fmt"
    "log"

    "github.com/blacktop/partialzip"
)

func main() {
    pzip, err := partialzip.New("https://apple.com/ipsw/download/link")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(pzip.List())

    n, err := pzip.Download("kernelcache.release.iphone11")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("extracting kernelcache.release.iphone11, wrote %d bytes\n", n)
}
```

```bash
extracting "kernelcache.release.iphone11", wrote 17842148 bytes
```

## CLI

### Install

Download **punzip** from [releases](https://github.com/blacktop/partialzip/releases)

### Usage

List zipped files

```bash
$ punzip list http://updates-http.cdn-apple.com/download/ipsw

NAME                                                         |SIZE
Firmware/                                                    |0 B
...SNIP...
kernelcache.release.iphone11                                 |18 MB
Firmware/all_flash/LLB.d321.RELEASE.im4p.plist               |331 B
048-16246-211.dmg                                            |107 MB
048-15811-213.dmg                                            |106 MB
Firmware/ICE18-1.00.08.Release.bbfw                          |39 MB
```

Download a file from the remote zip

```bash
$ punzip get --path kernelcache.release.iphone11 http://updates-http.cdn-apple.com/download/ipsw

Successfully downloaded "kernelcache.release.iphone11"
```

## Credits

- [planetbeing/partial-zip](https://github.com/planetbeing/partial-zip) _(written in C)_
- [marcograss/partialzip](https://github.com/marcograss/partialzip) _(written in Rust)_

## License

MIT Copyright (c) 2018 blacktop
