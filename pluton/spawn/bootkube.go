// Copyright 2017 CoreOS, Inc.
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

package spawn

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/coreos/mantle/platform"
	"github.com/coreos/mantle/pluton"
	"github.com/coreos/mantle/pluton/spawn/containercache"

	"github.com/coreos/pkg/capnslog"
)

var plog = capnslog.NewPackageLogger("github.com/coreos/mantle", "pluton/spawn")

type BootkubeManager struct {
	platform.Cluster

	bastion   platform.Machine
	firstNode platform.Machine
	info      pluton.Info
	files     *inputFiles
}

func (m *BootkubeManager) AddMasters(n int) ([]platform.Machine, error) {
	masters, _, err := m.provisionNodes(n, 0)
	return masters, err
}

func (m *BootkubeManager) AddWorkers(n int) ([]platform.Machine, error) {
	_, workers, err := m.provisionNodes(0, n)
	return workers, err
}

// Ultimately a combination of global and test specific options passed via the
// harness package. We could pass those types directly but it would cause an
// import cycle. Could also move the GlobalOptions type up into the the pluton
// package.
type BootkubeConfig struct {
	BinaryPath     string
	ScriptDir      string
	InitialWorkers int
	InitialMasters int
	SelfHostEtcd   bool
}

// MakeSimpleCluster brings up a multi node bootkube cluster with static etcd
// and checks that all nodes are registered before returning.
func MakeBootkubeCluster(cloud platform.Cluster, config BootkubeConfig, bastion platform.Machine) (*pluton.Cluster, error) {
	// parse in script dir info or use defaults
	files, err := parseInputFiles(config)
	if err != nil {
		return nil, err
	}

	if config.InitialMasters < 1 {
		return nil, fmt.Errorf("Must specify at least 1 initial master for the bootstrap node")
	}

	// parse hyperkube version from service file
	info, err := getVersionFromService(files.kubeletMaster)
	if err != nil {
		return nil, fmt.Errorf("error determining kubernetes version: %v", err)
	}

	containers := []containercache.ImageName{
		{Name: fmt.Sprintf("%s:%s", info.ImageRepo, info.KubeletTag), Engine: "rkt"},
		{Name: fmt.Sprintf("%s:%s", info.ImageRepo, info.KubeletTag), Engine: "docker"},
		{Name: "nginx", Engine: "docker"},
		{Name: "busybox", Engine: "docker"},

		//TODO(pb): find a better way of ensuring we are caching the
		//right versions of these components. Its not fatal if we
		//don't, it just means those images don't get cached which
		//could cause potential test failures
		{Name: "quay.io/coreos/etcd-operator:v0.2.6", Engine: "docker"},
		{Name: "quay.io/coreos/etcd:v3.1.6", Engine: "rkt"},
		{Name: "quay.io/coreos/flannel:v0.7.1-amd64", Engine: "docker"},
		{Name: "quay.io/coreos/pod-checkpointer:2cad4cac4186611a79de1969e3ea4924f02f459e", Engine: "docker"},
		{Name: "quay.io/coreos/kenc:48b6feceeee56c657ea9263f47b6ea091e8d3035", Engine: "docker"},
		{Name: "gcr.io/google_containers/k8s-dns-kube-dns-amd64:1.14.1", Engine: "docker"},
		{Name: "gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64:1.14.1", Engine: "docker"},
		{Name: "gcr.io/google_containers/k8s-dns-sidecar-amd64:1.14.1", Engine: "docker"},
		{Name: "gcr.io/google_containers/pause-amd64:3.0", Engine: "docker"},
	}

	// start containercache on bastion machine
	err = containercache.StartBastionOnce(bastion, containers)
	if err != nil {
		return nil, fmt.Errorf("starting bastion: %v", err)
	}

	// provision master node running etcd
	masterConfig, err := renderNodeConfig(files.kubeletMaster)
	if err != nil {
		return nil, err
	}
	master, err := cloud.NewMachine(masterConfig)
	if err != nil {
		return nil, err
	}

	plog.Infof("Master VM (%s) started. It's IP is %s.", master.ID(), master.IP())

	if err := containercache.Load(bastion, []platform.Machine{master}); err != nil {
		return nil, fmt.Errorf("copying bootstrap node's containers from cache: %v", err)
	}

	// TODO(pb): as soon as we have masterIP, start additional workers/masters in parallel with bootkube start

	// start bootkube on master
	if err := bootstrapMaster(master, files, config.SelfHostEtcd); err != nil {
		return nil, fmt.Errorf("bootstrapping master node: %v", err)
	}

	// install kubectl on master
	if err := installKubectl(master, info.UpstreamVersion); err != nil {
		return nil, err
	}

	manager := &BootkubeManager{
		Cluster:   cloud,
		bastion:   bastion,
		firstNode: master,
		files:     files,
		info:      info,
	}

	// provision additional nodes
	masters, workers, err := manager.provisionNodes(config.InitialMasters-1, config.InitialWorkers)
	if err != nil {
		return nil, err
	}

	cluster := pluton.NewCluster(manager, append([]platform.Machine{master}, masters...), workers, info)

	// check that all nodes appear in kubectl
	if err := cluster.Ready(); err != nil {
		return nil, fmt.Errorf("final node check: %v", err)
	}

	return cluster, nil
}

func renderNodeConfig(kubeletService string) (string, error) {
	// render template
	tmplData := struct {
		KubeletService string
	}{
		serviceToConfig(kubeletService),
	}

	buf := new(bytes.Buffer)

	tmpl, err := template.New("nodeConfig").Parse(nodeTmpl)
	if err != nil {
		return "", err
	}
	if err := tmpl.Execute(buf, &tmplData); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// The service files we read in from the hack directory need to be indented and
// have a bash variable substituted before being placed in the cloud-config.
func serviceToConfig(s string) string {
	const indent = "        " // 8 spaces to fit in cloud-config

	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = indent + lines[i]
	}

	service := strings.Join(lines, "\n")
	service = strings.Replace(service, "${COREOS_PRIVATE_IPV4}", "$private_ipv4", -1)

	return service
}

func bootstrapMaster(m platform.Machine, files *inputFiles, selfHostEtcd bool) error {
	const startTimeout = time.Minute * 10 // stop bootkube start if it takes longer then this

	_, err := m.SSH("sudo setenforce 0")
	if err != nil {
		return err
	}

	// transfer bootkube binary to machine
	err = platform.InstallFile(bytes.NewReader(files.bootkube), m, "/home/core/bootkube")
	if err != nil {
		return fmt.Errorf("Error transferring bootkube binary to bootstrap machine: %v", err)
	}

	// transfer init-master.sh to machine
	err = platform.InstallFile(bytes.NewReader(files.initMaster), m, "/home/core/init-master.sh")
	if err != nil {
		return fmt.Errorf("Error transferring bootkube binary to bootstrap machine: %v", err)
	}

	cmd := fmt.Sprintf("sudo COREOS_PRIVATE_IPV4=%v COREOS_PUBLIC_IPV4=%v SELF_HOST_ETCD=%v /home/core/init-master.sh local",
		m.PrivateIP(), m.IP(), selfHostEtcd)

	// use ssh client to collect stderr and stdout separetly
	// TODO: make the SSH method on a platform.Machine return two slices
	// for stdout/stderr in upstream kola code.
	client, err := m.SSHClient()
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var stdout = bytes.NewBuffer(nil)
	var stderr = bytes.NewBuffer(nil)
	session.Stderr = stderr
	session.Stdout = stdout

	err = session.Start(cmd)
	if err != nil {
		return err
	}

	// add global timeout for bootkube to finish starting, this should be
	// expected to complete much faster then the built-in bootkube timeout
	// because we prefetch containers
	errc := make(chan error)
	go func() { errc <- session.Wait() }()
	select {
	case err := <-errc:
		if err != nil {
			return fmt.Errorf("SSH session returned error for cmd %s: %s\nSTDOUT:\n%s\nSTDERR:\n%s\n--\n", cmd, err, stdout, stderr)
		}
	case <-time.After(startTimeout):
		return fmt.Errorf("Timed out waiting %v for cmd %s\nSTDOUT:\n%s\nSTDERR:\n%s\n--\n", startTimeout, cmd, stdout, stderr)
	}
	plog.Infof("Success for cmd %s: %s\nSTDOUT:\n%s\nSTDERR:\n%s\n--\n", cmd, err, stdout, stderr)

	return nil
}

func (m *BootkubeManager) provisionNodes(masters, workers int) ([]platform.Machine, []platform.Machine, error) {
	if masters == 0 && workers == 0 {
		return []platform.Machine{}, []platform.Machine{}, nil
	} else if masters < 0 || workers < 0 {
		return nil, nil, fmt.Errorf("can't provision negative number of nodes")
	}

	configM, err := renderNodeConfig(m.files.kubeletMaster)
	if err != nil {
		return nil, nil, err
	}

	configW, err := renderNodeConfig(m.files.kubeletWorker)
	if err != nil {
		return nil, nil, err
	}

	// NewMachines already does parallelization but doesn't guarentee the
	// order of the nodes returned which matters when we have heterogenious
	// cloudconfigs here
	var wg sync.WaitGroup
	var masterNodes, workerNodes []platform.Machine
	var merror, werror error

	wg.Add(2)
	go func() {
		defer wg.Done()
		if masters > 0 {
			masterNodes, merror = platform.NewMachines(m, configM, masters)
		} else {
			masterNodes = []platform.Machine{}
		}
	}()
	go func() {
		defer wg.Done()
		if workers > 0 {
			workerNodes, werror = platform.NewMachines(m, configW, workers)
		} else {
			workerNodes = []platform.Machine{}
		}
	}()
	wg.Wait()
	if merror != nil || werror != nil {
		return nil, nil, fmt.Errorf("error calling NewMachines: %v %v", merror, werror)
	}

	// populate machines with containers from cache
	err = containercache.Load(m.bastion, append(masterNodes, workerNodes...))
	if err != nil {
		return nil, nil, fmt.Errorf("populating cache: %v", err)
	}

	// start kubelet
	for _, node := range append(masterNodes, workerNodes...) {
		// transfer kubeconfig from existing node
		err := platform.TransferFile(m.firstNode, "/etc/kubernetes/kubeconfig", node, "/etc/kubernetes/kubeconfig")
		if err != nil {
			return nil, nil, err
		}

		// transfer client ca cert but soft fail for older verions of bootkube
		err = platform.TransferFile(m.firstNode, "/etc/kubernetes/ca.crt", node, "/etc/kubernetes/ca.crt")
		if err != nil {
			plog.Infof("Warning: unable to transfer client cert to worker: %v", err)
		}

		if err := installKubectl(node, m.info.UpstreamVersion); err != nil {
			return nil, nil, err
		}

		// disable selinux
		_, err = node.SSH("sudo setenforce 0")
		if err != nil {
			return nil, nil, err
		}

		// start kubelet
		_, err = node.SSH("sudo systemctl -q enable --now kubelet.service")
		if err != nil {
			return nil, nil, err
		}
	}

	return masterNodes, workerNodes, nil

}

type inputFiles struct {
	kubeletMaster string
	kubeletWorker string
	initMaster    []byte
	bootkube      []byte
}

func parseInputFiles(config BootkubeConfig) (*inputFiles, error) {
	var files = new(inputFiles)

	b, err := ioutil.ReadFile(filepath.Join(config.ScriptDir, "kubelet.master"))
	if err != nil {
		return nil, fmt.Errorf("failed to read expected kubelet.master file: %v", err)
	}
	files.kubeletMaster = string(b)

	b, err = ioutil.ReadFile(filepath.Join(config.ScriptDir, "kubelet.worker"))
	if err != nil {
		return nil, fmt.Errorf("failed to read expected kubelet.worker file: %v", err)
	}
	files.kubeletWorker = string(b)

	b, err = ioutil.ReadFile(filepath.Join(config.ScriptDir, "init-master.sh"))
	if err != nil {
		return nil, fmt.Errorf("failed to read expected init-master.sh file: %v", err)
	}
	files.initMaster = b

	b, err = ioutil.ReadFile(config.BinaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read expected bootkube binary: %v", err)
	}
	files.bootkube = b

	return files, nil
}

func getVersionFromService(kubeletService string) (pluton.Info, error) {
	var kubeletTag, kubeletImageRepo string
	lines := strings.Split(kubeletService, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Environment=KUBELET_IMAGE_TAG=") {
			kubeletTag = strings.TrimPrefix(strings.TrimSpace(line), "Environment=KUBELET_IMAGE_TAG=")
		} else if strings.Contains(line, "Environment=KUBELET_IMAGE_URL=") {
			kubeletImageRepo = strings.TrimPrefix(strings.TrimSpace(line), "Environment=KUBELET_IMAGE_URL=")
		}
	}
	if kubeletTag == "" || kubeletImageRepo == "" {
		return pluton.Info{}, fmt.Errorf("could not find kubelet version from service file")
	}

	// Special logic for testing upstream branches.
	if kubeletTag == "master" || strings.HasPrefix(kubeletTag, "release") {
		var err error
		kubeletTag, err = fetchUpstreamBranchVersion(kubeletTag)
		if err != nil {
			return pluton.Info{}, err
		}
	}
	upstream, err := stripSemverSuffix(kubeletTag)
	if err != nil {
		return pluton.Info{}, fmt.Errorf("tag %v: %v", kubeletTag, err)
	}
	semVer := strings.Replace(kubeletTag, "_", "+", 1)

	// hack to handle upstream pre-release versions TODO(pb): simpify this
	// parsing and only have conformance test rely on accurate upstream
	// versions, not having kubectl will fail all tests. kubectl can be
	// copied from hyperkube
	if strings.Contains(semVer, "-") {
		upstream = strings.Split(semVer, "+")[0]
	}

	s := pluton.Info{
		KubeletTag:      kubeletTag,
		Version:         semVer,
		UpstreamVersion: upstream,
		ImageRepo:       kubeletImageRepo,
	}
	plog.Infof("version detection: %#v", s)

	return s, nil
}

func stripSemverSuffix(v string) (string, error) {
	semverPrefix := regexp.MustCompile(`^v[\d]+\.[\d]+\.[\d]+`)
	v = semverPrefix.FindString(v)
	if v == "" {
		return "", fmt.Errorf("error stripping semver suffix")
	}

	return v, nil
}

func installKubectl(m platform.Machine, upstreamVersion string) error {
	kubeURL := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%v/bin/linux/amd64/kubectl", upstreamVersion)
	if _, err := m.SSH("wget -q " + kubeURL); err != nil {
		return fmt.Errorf("curling kubectl: %v", err)
	}
	if _, err := m.SSH("chmod +x ./kubectl"); err != nil {
		return err
	}

	return nil
}

// fetchUpstreamBranchVersion fetches the latest version for the given upstream kubernetes branch.
func fetchUpstreamBranchVersion(branch string) (string, error) {
	var releaseSuffix string
	switch branch {
	case "master":
		releaseSuffix = ""
	default:
		releaseSuffix = strings.TrimPrefix(branch, "release") // e.g. "-1.6"
	}
	url := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/latest%s.txt", releaseSuffix)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}
