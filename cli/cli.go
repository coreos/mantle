// Copyright 2014 CoreOS, Inc.
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

package cli

import (
	"os"

	"github.com/coreos/pkg/capnslog"
	"github.com/spf13/cobra"

	"github.com/coreos/mantle/system/exec"
	"github.com/coreos/mantle/version"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number and exit.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("mantle/%s version %s\n",
				cmd.Root().Name(), version.Version)
		},
	}

	logDebug   bool
	logVerbose bool
	logLevel   = capnslog.NOTICE

	plog = capnslog.NewPackageLogger("github.com/coreos/mantle", "cli")
)

// Execute sets up common features that all mantle commands should share
// and then executes the command. It does not return.
func Execute(main *cobra.Command) {
	// If we were invoked via a multicall entrypoint run it instead.
	// TODO(marineam): should we figure out a way to initialize logging?
	exec.MaybeExec()

	main.AddCommand(versionCmd)

	// TODO(marineam): pflags defines the Value interface differently,
	// update capnslog accordingly...
	//main.PersistentFlags().Var(&level, "log-level",
	//	"Set global log level. (default is NOTICE)")
	main.PersistentFlags().BoolVarP(&logVerbose, "verbose", "v", false,
		"Alias for --log-level=INFO")
	main.PersistentFlags().BoolVarP(&logDebug, "debug", "d", false,
		"Alias for --log-level=DEBUG")

	var preRun = main.PersistentPreRun
	main.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		startLogging(cmd)
		if preRun != nil {
			preRun(cmd, args)
		}
	}

	if err := main.Execute(); err != nil {
		plog.Fatal(err)
	}
	os.Exit(0)
}

func setRepoLogLevel(repo string, l capnslog.LogLevel) {
	r, err := capnslog.GetRepoLogger(repo)
	if err != nil {
		return // don't care if it isn't linked in
	}
	r.SetRepoLogLevel(l)
}

func startLogging(cmd *cobra.Command) {
	switch {
	case logDebug:
		logLevel = capnslog.DEBUG
	case logVerbose:
		logLevel = capnslog.INFO
	}

	capnslog.SetFormatter(capnslog.NewStringFormatter(os.Stderr))
	capnslog.SetGlobalLogLevel(logLevel)

	// In the context of the internally linked etcd, the NOTICE messages
	// aren't really interesting, so translate NOTICE to WARNING instead.
	if logLevel == capnslog.NOTICE {
		// etcd sure has a lot of repos in its repo
		setRepoLogLevel("github.com/coreos/etcd", capnslog.WARNING)
		setRepoLogLevel("github.com/coreos/etcd/etcdserver", capnslog.WARNING)
		setRepoLogLevel("github.com/coreos/etcd/etcdserver/etcdhttp", capnslog.WARNING)
		setRepoLogLevel("github.com/coreos/etcd/pkg", capnslog.WARNING)
	}

	plog.Infof("Started logging at level %s", logLevel)
}
