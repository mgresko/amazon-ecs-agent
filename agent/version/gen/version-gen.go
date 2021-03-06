// Copyright 2014-2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const versiongoTemplate = `// This is an autogenerated file and should not be edited.

// Copyright 2014-2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Package version contains constants to indicate the current version of the
// agent. It is autogenerated
package version

// Please DO NOT commit any changes to this file (specifically this hash) except
// for those created by running ./scripts/update-version at the root of the
// repository

` + `//go:generate go run gen/version-gen.go

// Version is the version of the Agent
const Version = "{{.Version}}"

// GitDirty indicates the cleanliness of the git repo when this agent was built
const GitDirty = {{.Dirty}}

// GitShortHash is the short hash of this agent build
const GitShortHash = "{{.Hash}}"
`

type versionInfo struct {
	Version string
	Dirty   bool
	Hash    string
}

func gitDirty() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	err := cmd.Run()
	if err == nil {
		return false
	}
	return true
}

func gitHash() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	hash, err := cmd.Output()
	if err != nil {
		return "UNKNOWN"
	}
	return strings.TrimSpace(string(hash))
}

// version-gen is a simple program that generates the agent's version file,
// containing information about the agent's version, commit hash, and repository
// cleanliness.
func main() {

	versionStr, _ := ioutil.ReadFile(filepath.Join("..", "..", "VERSION"))

	// default values
	info := versionInfo{
		Version: strings.TrimSpace(string(versionStr)),
		Dirty:   true,
		Hash:    "UNKNOWN",
	}

	if strings.TrimSpace(os.Getenv("ECS_RELEASE")) == "cleanbuild" {
		// 'clean' release; all other releases assumed dirty
		info.Dirty = gitDirty()
	}
	if os.Getenv("ECS_UNKNOWN_VERSION") == "" {
		// When the version file is updated, the above is set
		// Setting UNKNOWN version allows the version committed in git to never
		// have a commit hash so that it does not churn with every commit. This
		// env var should not be set when building, and go generate should be
		// run before any build, such that the commithash will be set correctly.
		info.Hash = gitHash()
	}

	outFile, err := os.Create("version.go")
	if err != nil {
		log.Fatalf("Unable to create output version file: %v", err)
	}
	t := template.Must(template.New("version").Parse(versiongoTemplate))

	err = t.Execute(outFile, info)
	if err != nil {
		log.Fatalf("Error applying template: %v", err)
	}
}
