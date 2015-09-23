package amazon

import (
	"bytes"
	"encoding/xml"
	"io"
)

type ByteRange struct {
	XMLName xml.Name `xml:"byte-range"`
	Start   int64    `xml:"start,attr"`
	End     int64    `xml:"end,attr"`
}

type Part struct {
	XMLName   xml.Name  `xml:"part"`
	Index     int       `xml:"index,attr"`
	ByteRange ByteRange `xml:"byte-range"`
	Key       string    `xml:"key"`
	HeadURL   string    `xml:"head-url"`
	GetURL    string    `xml:"get-url"`
	DeleteURL string    `xml:"delete-url"`
}

type Parts struct {
	XMLName xml.Name `xml:"parts"`
	Count   int      `xml:"count,attr"`
	Parts   []Part   `xml:"part"`
}

type Import struct {
	XMLName    xml.Name `xml:"import"`
	Size       int64    `xml:"size"`
	VolumeSize int64    `xml:"volume-size"`
	Parts      Parts    `xml:"parts"`
}

type Manifest struct {
	XMLName         xml.Name `xml:"manifest"`
	Version         string   `xml:"version"`
	FileFormat      string   `xml:"file-format"`
	ImporterName    string   `xml:"importer>name"`
	ImporterVersion string   `xml:"importer>version"`
	ImporterRelease string   `xml:"importer>release"`
	SelfDestructURL string   `xml:"self-destruct-url"`
	Import          Import   `xml:"import"`
}

func GenerateManifest(typ, part0url string, size int64) (io.Reader, error) {
	u := part0url

	part0 := Part{
		Index: 0,
		ByteRange: ByteRange{
			Start: 0,
			End:   size,
		},
		Key:       "",
		HeadURL:   u,
		GetURL:    u,
		DeleteURL: u,
	}

	manifest := &Manifest{
		Version:         "2010-11-15",
		FileFormat:      typ,
		ImporterName:    "plume",
		ImporterVersion: "1.0.0",
		ImporterRelease: "2015-09-22",
		SelfDestructURL: "",
		Import: Import{
			Size:       size,
			VolumeSize: 8,
			Parts: Parts{
				Count: 1,
				Parts: []Part{
					part0,
				},
			},
		},
	}

	xmlbuf := new(bytes.Buffer)

	err := xml.NewEncoder(xmlbuf).Encode(manifest)
	if err != nil {
		return nil, err
	}

	return xmlbuf, nil
}
