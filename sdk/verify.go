// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sdk

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/coreos/mantle/Godeps/_workspace/src/golang.org/x/crypto/openpgp"
)

// pub   2048R/5A50CE28 2015-12-17
// uid       [ultimate] CoreOS Developer Key <coreos-developer-key@quantum.com>
// sub   2048R/47B4220F 2015-12-17
const buildbotPubKey = `
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v2

mQENBFZzEN4BCACnq4cZNJDbmvrOM27CxfWZUqi4zj8FsqGLdRA0OlNHf/zMgdKI
Cdx8H6QNNtjvJYSoHD2RncCOVxmltWqER2YyXfJOTboUog8b/wFhkk4DzPKfDDjL
MdmGlnFEZNAOvUMA4d15vuTCVOhRVczAjxbA0cPe4ePeY81UM6yeuyIq6PXj8GQt
1BEdYl4eV6+41PGdePc07eHHqFe4cnQBM42N+JpDBcIwR+70ATrybniIQZPNW87x
NhOFdhfEfXguYOhDjc4l108FfG9voQumdD9WJk/mxt/a6YHqaIvUEdzHzFvWSyFi
CGqpNfuH0D6CtD9xO0dJfpxKi9J4Kv1xhejvABEBAAG0N0NvcmVPUyBEZXZlbG9w
ZXIgS2V5IDxjb3Jlb3MtZGV2ZWxvcGVyLWtleUBxdWFudHVtLmNvbT6JATgEEwEC
ACIFAlZzEN4CGwMGCwkIBwMCBhUIAgkKCwQWAgMBAh4BAheAAAoJEF0oqhNaUM4o
CsoH/3ftVj76x2vWi3WG0KBNpS2NW09J4ce2/HFH0CCz1skkVb7p3QK5ocm+kNS5
2Vck0NZr26f8tMfGRsIxJJPsb3DadykuJ4wOl1WxvsEbaKV1BtUsAsY752lqAkeK
3lUWyY7acQWQBcal63ddvevULMOPwgL7VkroebcL/MIAJ3B6IWhWe/5TwjVVKS1F
mwfJzfsfEEsH+QiUxaWVzFZpPiga0BMHDcS2iRtJnto1MDA0/BT4cCKLQ0Adgn5M
KWf9M+lo4lGfEad915luZXKF0OIXIvVOdvucLhRK+kBJVjugwSzDdy55Gqk/v9WQ
Dx+WWGrUQYVW3bPT6xQVPVfJkU+5AQ0EVnMQ3gEIALXcDvr3c7lJQ/k/IQn1cEmU
IB8cZO8ggAVgyVhGugJc/+Hd5XDJ8RyWRwbvcvXNu2ns7cnIwMIuvSLnoDfOFwma
Wt3XawgH++v2upMp6xmVC0BbYicVdr9yvLadgbUC8ao8TiiZfMX6v1dfjOTHlJz1
nlVN1jo2u+C+NyW0GyeKMARo/SbRg0ay7ly3qvvme+hqo6/varzFcTNrPMCQ030h
T1LnhL3urrFrNTieOoWo3qqdt92xKIiqcDmtERTsPXXvCFfluxYrYkp8wVTFqa9j
rBE5owV4u2cdp9NP7lyZPuZxFt1TsSnOt2xjRzFdJukPe+uFh921vsjhDf16a3EA
EQEAAYkBHwQYAQIACQUCVnMQ3gIbDAAKCRBdKKoTWlDOKE7hB/9sHl256i/zslDw
P8Q/ssq2ckmisUPSDfXgVitWYwFRjc5bQCf3by4dJM7KpLbvQLcskACUXjjIZ0f9
RiaeTBEcKsKfCFQ2VqrG6qfNnizhBQMlbCMAK9JivyZ0IY6hsAlCFt27a5GvPs+6
92q3usuiKqEJVvIha2jGbOosMyfrSRQR35i/K9Y7+18LPMCvMcLUWwi3//ehnpXy
4XQMLwP1Hvv5bICfbGcWzIiInpdSXPxm2BSOnK7ExSht4hHcjxob0ApXTG/Y9A8b
u52r5p8z5hThZX65muPIHkdXWXVzU2ER8VrTyy09rSaxumN04SGneYbnkMiF19FD
HznYL5Wa
=HQOM
-----END PGP PUBLIC KEY BLOCK-----
`

func Verify(signed, signature io.Reader) error {
	keyring, err := openpgp.ReadArmoredKeyRing(strings.NewReader(buildbotPubKey))
	if err != nil {
		panic(err)
	}

	_, err = openpgp.CheckDetachedSignature(keyring, signed, signature)
	return err
}

func VerifyFile(file string) error {
	signed, err := os.Open(file)
	if err != nil {
		return err
	}
	defer signed.Close()

	signature, err := os.Open(file + ".sig")
	if err != nil {
		return err
	}
	defer signature.Close()

	if err := Verify(signed, signature); err != nil {
		return fmt.Errorf("%v: %s", err, file)
	}
	return nil
}
