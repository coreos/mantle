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

package qemu

import (
	"fmt"
	"io/ioutil"
	"os"
	osexec "os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/coreos/pkg/capnslog"
	"github.com/satori/go.uuid"

	"github.com/coreos/mantle/platform"
	"github.com/coreos/mantle/platform/conf"
	"github.com/coreos/mantle/platform/local"
	"github.com/coreos/mantle/system/exec"
	"github.com/coreos/mantle/system/ns"
)

const (
	primaryDiskId = "primary-disk"
)

// Options contains QEMU-specific options for the cluster.
type Options struct {
	// DiskImage is the full path to the disk image to boot in QEMU.
	DiskImage string
	Board     string

	// BIOSImage is name of the BIOS file to pass to QEMU.
	// It can be a plain name, or a full path.
	BIOSImage string

	*platform.Options
}

// Cluster is a local cluster of QEMU-based virtual machines.
//
// XXX: must be exported so that certain QEMU tests can access struct members
// through type assertions.
type Cluster struct {
	opts *Options

	mu sync.Mutex
	*local.LocalCluster
}

type MachineOptions struct {
	AdditionalDisks []Disk
}

type Disk struct {
	Size   string // disk image size in bytes, optional suffixes "K", "M", "G", "T" allowed
	Serial string // serial number to be passed to qemu via `serial=`. Disks show up under /dev/disk/by-id/virtio-<serial>
}

var (
	plog = capnslog.NewPackageLogger("github.com/coreos/mantle", "kola/platform/machine/qemu")
)

// NewCluster creates a Cluster instance, suitable for running virtual
// machines in QEMU.
func NewCluster(opts *Options, rconf *platform.RuntimeConfig) (platform.Cluster, error) {
	lc, err := local.NewLocalCluster(opts.BaseName, rconf)
	if err != nil {
		return nil, err
	}

	qc := &Cluster{
		opts:         opts,
		LocalCluster: lc,
	}

	return qc, nil
}

func (qc *Cluster) NewMachine(userdata *conf.UserData) (platform.Machine, error) {
	return qc.NewMachineWithOptions(userdata, MachineOptions{})
}

func (qc *Cluster) NewMachineWithOptions(userdata *conf.UserData, options MachineOptions) (platform.Machine, error) {
	id := uuid.NewV4()

	dir := filepath.Join(qc.RuntimeConf().OutputDir, id.String())
	if err := os.Mkdir(dir, 0777); err != nil {
		return nil, err
	}

	// hacky solution for cloud config ip substitution
	// NOTE: escaping is not supported
	qc.mu.Lock()
	netif := qc.Dnsmasq.GetInterface("br0")
	ip := strings.Split(netif.DHCPv4[0].String(), "/")[0]

	conf, err := qc.RenderUserData(userdata, map[string]string{
		"$public_ipv4":  ip,
		"$private_ipv4": ip,
	})
	if err != nil {
		qc.mu.Unlock()
		return nil, err
	}
	qc.mu.Unlock()

	var confPath string
	if conf.IsIgnition() {
		confPath = filepath.Join(dir, "ignition.json")
		if err := conf.WriteFile(confPath); err != nil {
			return nil, err
		}
	} else {
		confPath, err = local.MakeConfigDrive(conf, dir)
		if err != nil {
			return nil, err
		}
	}

	journal, err := platform.NewJournal(dir)
	if err != nil {
		return nil, err
	}

	qm := &machine{
		qc:          qc,
		id:          id.String(),
		netif:       netif,
		journal:     journal,
		consolePath: filepath.Join(dir, "console.txt"),
	}

	var qmCmd []string
	combo := runtime.GOARCH + "--" + qc.opts.Board
	switch combo {
	case "amd64--amd64-usr":
		qmCmd = []string{
			"qemu-system-x86_64",
			"-machine", "accel=kvm",
			"-cpu", "host",
			"-m", "1024",
		}
	case "amd64--arm64-usr":
		qmCmd = []string{
			"qemu-system-aarch64",
			"-machine", "virt",
			"-cpu", "cortex-a57",
			"-m", "2048",
		}
	case "arm64--amd64-usr":
		qmCmd = []string{
			"qemu-system-x86_64",
			"-machine", "pc-q35-2.8",
			"-cpu", "kvm64",
			"-m", "1024",
		}
	case "arm64--arm64-usr":
		qmCmd = []string{
			"qemu-system-aarch64",
			"-machine", "virt,accel=kvm,gic-version=3",
			"-cpu", "host",
			"-m", "2048",
		}
	default:
		panic("host-guest combo not supported: " + combo)
	}

	qmMac := qm.netif.HardwareAddr.String()
	qmCmd = append(qmCmd,
		"-bios", qc.opts.BIOSImage,
		"-smp", "1",
		"-uuid", qm.id,
		"-display", "none",
		"-chardev", "file,id=log,path="+qm.consolePath,
		"-serial", "chardev:log",
	)

	if conf.IsIgnition() {
		qmCmd = append(qmCmd,
			"-fw_cfg", "name=opt/com.coreos/config,file="+confPath)
	} else {
		qmCmd = append(qmCmd,
			"-fsdev", "local,id=cfg,security_model=none,readonly,path="+confPath,
			"-device", qc.virtio("9p", "fsdev=cfg,mount_tag=config-2"))
	}

	var extraFiles []*os.File
	fdnum := 3 // first additional file starts at position 3
	fdset := 1

	addDisk := func(file *os.File, serial string) {
		id := fmt.Sprintf("d%d", fdnum)
		qmCmd = append(qmCmd, "-add-fd", fmt.Sprintf("fd=%d,set=%d", fdnum, fdset),
			"-drive", fmt.Sprintf("if=none,id=%s,format=qcow2,file=/dev/fdset/%d", id, fdset),
			"-device", qc.virtio("blk", fmt.Sprintf("drive=%s,serial=%s", id, serial)))
		fdnum += 1
		fdset += 1
		extraFiles = append(extraFiles, file)
	}

	diskFile, err := setupPrimaryDisk(qc.opts.DiskImage)
	if err != nil {
		return nil, err
	}
	defer diskFile.Close()
	addDisk(diskFile, primaryDiskId)

	for _, disk := range options.AdditionalDisks {
		optionsDiskFile, err := setupDisk(disk.Size)
		if err != nil {
			return nil, err
		}
		defer optionsDiskFile.Close()
		addDisk(optionsDiskFile, disk.Serial)
	}

	qc.mu.Lock()

	tap, err := qc.NewTap("br0")
	if err != nil {
		qc.mu.Unlock()
		return nil, err
	}
	defer tap.Close()
	qmCmd = append(qmCmd, "-netdev", fmt.Sprintf("tap,id=tap,fd=%d", fdnum),
		"-device", qc.virtio("net", "netdev=tap,mac="+qmMac))
	fdnum += 1
	extraFiles = append(extraFiles, tap.File)

	plog.Debugf("NewMachine: (%s) %q", combo, qmCmd)

	qm.qemu = qm.qc.NewCommand(qmCmd[0], qmCmd[1:]...)

	qc.mu.Unlock()

	cmd := qm.qemu.(*ns.Cmd)
	cmd.Stderr = os.Stderr

	cmd.ExtraFiles = append(cmd.ExtraFiles, extraFiles...)

	if err = qm.qemu.Start(); err != nil {
		return nil, err
	}

	// wait in the background for qemu to die so we can tell if it was killed (as it should be) or ran in to
	// some other problem and died.
	go func() {
		err := qm.qemu.Wait()
		if exitErr, ok := err.(*osexec.ExitError); ok && err != nil {
			plog.Debugf("qemu exited: %v", exitErr)
			// safe type assertion - we only support Linux
			if waitStatus := exitErr.ProcessState.Sys().(syscall.WaitStatus); !waitStatus.Signaled() {
				// Under normal operation, we kill qemu. It should not exit; if it does, log it.
				plog.Errorf("qemu exited unexpectantly: %v", exitErr)
			}
		} else if err != nil {
			plog.Errorf("qemu failed unexpectedly (not killed): %v", err)
		}
	}()

	if err := platform.StartMachine(qm, qm.journal, qc.RuntimeConf()); err != nil {
		qm.Destroy()
		return nil, err
	}

	qc.AddMach(qm)

	return qm, nil
}

// The virtio device name differs between machine types but otherwise
// configuration is the same. Use this to help construct device args.
func (qc *Cluster) virtio(device, args string) string {
	var suffix string
	switch qc.opts.Board {
	case "amd64-usr":
		suffix = "pci"
	case "arm64-usr":
		suffix = "device"
	default:
		panic(qc.opts.Board)
	}
	return fmt.Sprintf("virtio-%s-%s,%s", device, suffix, args)
}

// Create a nameless temporary qcow2 image file backed by a raw image.
func setupPrimaryDisk(imageFile string) (*os.File, error) {
	// a relative path would be interpreted relative to /tmp
	backingFile, err := filepath.Abs(imageFile)
	if err != nil {
		return nil, err
	}
	// keep the COW image from breaking if the "latest" symlink changes
	backingFile, err = filepath.EvalSymlinks(backingFile)
	if err != nil {
		return nil, err
	}

	qcowOpts := fmt.Sprintf("backing_file=%s,backing_fmt=raw,lazy_refcounts=on", backingFile)
	return setupDisk("-o", qcowOpts)
}

func setupDisk(additionalOptions ...string) (*os.File, error) {
	dstFile, err := ioutil.TempFile("", "mantle-qemu")
	if err != nil {
		return nil, err
	}
	dstFileName := dstFile.Name()
	defer os.Remove(dstFileName)
	dstFile.Close()

	opts := []string{"create", "-f", "qcow2", dstFileName}
	opts = append(opts, additionalOptions...)

	qemuImg := exec.Command("qemu-img", opts...)
	qemuImg.Stderr = os.Stderr

	if err := qemuImg.Run(); err != nil {
		return nil, err
	}

	return os.OpenFile(dstFileName, os.O_RDWR, 0)
}
