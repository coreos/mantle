// Copyright 2016 CoreOS, Inc.
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

package generator

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"

	"github.com/coreos/pkg/capnslog"
	"github.com/golang/protobuf/proto"

	"github.com/coreos/mantle/lang/destructor"
	"github.com/coreos/mantle/update/metadata"
	"github.com/coreos/mantle/update/signature"
)

const (
	// Default block size to use for all generated payloads.
	BlockSize = 4096

	// Default data size limit to process in a single operation.
	ChunkSize = BlockSize * 256
)

var (
	plog = capnslog.NewPackageLogger("github.com/coreos/mantle", "update/generator")

	// ErrProcedureExists indicates that a given procedure type has
	// already been added to the Generator.
	ErrProcedureExists = errors.New("generator: procedure already exists")
)

// Generator assembles an update payload from a number of sources. Each of
// its methods must only be called once, ending with Write.
type Generator struct {
	destructor.MultiDestructor
	manifest metadata.DeltaArchiveManifest
	payloads []io.Reader
}

// Procedure represent independent update within a payload.
type Procedure struct {
	metadata.InstallProcedure
	io.ReadCloser
}

// Partition adds the given /usr update Procedure to the payload.
// It must always be the first procedure added to the Generator.
func (g *Generator) Partition(proc *Procedure) error {
	if len(g.payloads) != 0 {
		return ErrProcedureExists
	}

	g.AddCloser(proc)
	g.manifest.PartitionOperations = proc.Operations
	g.manifest.OldPartitionInfo = proc.OldInfo
	g.manifest.NewPartitionInfo = proc.NewInfo
	g.payloads = append(g.payloads, proc)
	return nil
}

// Write finalizes the payload, writing it out to the given file path.
func (g *Generator) Write(path string) (err error) {
	if err = g.updateOffsets(); err != nil {
		return
	}

	// for compatibility with old update_engine versions
	g.addNoops()

	plog.Infof("Writing payload to %s", path)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return
	}
	defer func() {
		if e := f.Close(); e != nil && err == nil {
			err = e
		}
	}()

	// All payload data up until the signatures must be hashed.
	hasher := signature.NewSignatureHash()
	w := io.MultiWriter(f, hasher)

	if err = g.writeHeader(w); err != nil {
		return
	}

	if err = g.writeManifest(w); err != nil {
		return
	}

	for _, payload := range g.payloads {
		if _, err = io.Copy(w, payload); err != nil {
			return
		}
	}

	// Hashed writes complete, write signatures to payload file.
	err = g.writeSignatures(f, hasher.Sum(nil))
	return
}

func (g *Generator) updateOffsets() error {
	var offset uint32
	updateOps := func(ops []*metadata.InstallOperation) {
		for _, op := range ops {
			if op.DataLength == nil {
				op.DataOffset = nil
			} else {
				op.DataOffset = proto.Uint32(offset)
				offset += *op.DataLength
			}
		}
	}

	updateOps(g.manifest.PartitionOperations)
	for _, proc := range g.manifest.Procedures {
		updateOps(proc.Operations)
	}

	sigSize, err := signature.SignaturesSize()
	g.manifest.SignaturesOffset = proto.Uint64(uint64(offset))
	g.manifest.SignaturesSize = proto.Uint64(uint64(sigSize))
	return err
}

// The sparse hole was a feature of update_engine, it is not used
// in this new code outside of the noop compatibility goo.
const sparseHole = math.MaxUint64

// Translate a normal install operation to a dummy that discards data.
func opToNoop(op *metadata.InstallOperation) *metadata.InstallOperation {
	blocks := (uint64(*op.DataLength) + BlockSize - 1) / BlockSize
	sum := make([]byte, len(op.DataSha256Hash))
	copy(sum, op.DataSha256Hash)

	return &metadata.InstallOperation{
		Type:           metadata.InstallOperation_REPLACE.Enum(),
		DataOffset:     proto.Uint32(*op.DataOffset),
		DataLength:     proto.Uint32(*op.DataLength),
		DataSha256Hash: sum,
		DstExtents: []*metadata.Extent{&metadata.Extent{
			StartBlock: proto.Uint64(sparseHole),
			NumBlocks:  proto.Uint64(blocks),
		}},
	}
}

// Fill in the dummy noop_operations list for compatibility with old
// update_engine versions that didn't support procedures and handled
// signature data weirdly.
func (g *Generator) addNoops() {
	// Translate the new procedures list to noop operations.
	for _, proc := range g.manifest.Procedures {
		for _, op := range proc.Operations {
			if op.GetDataLength() == 0 {
				continue
			}

			g.manifest.NoopOperations = append(
				g.manifest.NoopOperations, opToNoop(op))
		}
	}

	// Create a dummy noop operation to cover trailing signature data.
	// Yes, the manifest inconsistently uses 32 and 64 bit values...
	offset := uint32(*g.manifest.SignaturesOffset)
	length := uint32(*g.manifest.SignaturesSize)
	blocks := (*g.manifest.SignaturesSize + BlockSize - 1) / BlockSize
	g.manifest.NoopOperations = append(g.manifest.NoopOperations,
		&metadata.InstallOperation{
			Type:       metadata.InstallOperation_REPLACE.Enum(),
			DataOffset: proto.Uint32(offset),
			DataLength: proto.Uint32(length),
			DstExtents: []*metadata.Extent{&metadata.Extent{
				StartBlock: proto.Uint64(sparseHole),
				NumBlocks:  proto.Uint64(blocks),
			}},
		})
}

func (g *Generator) writeHeader(w io.Writer) error {
	manifestSize := proto.Size(&g.manifest)
	header := metadata.DeltaArchiveHeader{
		Version:      metadata.Version,
		ManifestSize: uint64(manifestSize),
	}
	copy(header.Magic[:], []byte(metadata.Magic))

	return binary.Write(w, binary.BigEndian, &header)
}

func (g *Generator) writeManifest(w io.Writer) error {
	buf, err := proto.Marshal(&g.manifest)
	if err != nil {
		return err
	}

	_, err = w.Write(buf)
	return err
}

func (g *Generator) writeSignatures(w io.Writer, sum []byte) error {
	signatures, err := signature.Sign(sum)
	if err != nil {
		return err
	}

	buf, err := proto.Marshal(signatures)
	if err != nil {
		return err
	}

	_, err = w.Write(buf)
	return err
}
