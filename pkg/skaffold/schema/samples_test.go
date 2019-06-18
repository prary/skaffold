/*
Copyright 2019 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schema

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/validation"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

const (
	samplesRoot = "../../../docs/content/en/samples"
)

var (
	ignoredSamples = []string{"structureTest.yaml", "build.sh"}
)

func TestParseSamples(t *testing.T) {
	paths, err := findSamples(samplesRoot)
	if err != nil {
		t.Fatalf("unable to read sample files in %q", samplesRoot)
	}

	if len(paths) == 0 {
		t.Fatalf("did not find sample files in %q", samplesRoot)
	}

	for _, path := range paths {
		name := filepath.Base(path)

		testutil.Run(t, name, func(t *testutil.T) {
			for _, is := range ignoredSamples {
				if name == is {
					t.Skip()
				}
			}

			buf, err := ioutil.ReadFile(path)
			t.CheckNoError(err)

			configFile := t.TempFile("skaffold.yaml", addHeader(buf))
			cfg, err := ParseConfig(configFile, true)
			t.CheckNoError(err)

			err = validation.Process(cfg.(*latest.SkaffoldConfig))
			t.CheckNoError(err)
		})
	}
}

func findSamples(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return err
	})

	return files, err
}

func addHeader(buf []byte) []byte {
	if bytes.HasPrefix(buf, []byte("apiVersion:")) {
		return buf
	}
	return []byte(fmt.Sprintf("apiVersion: %s\nkind: Config\n%s", latest.Version, buf))
}
