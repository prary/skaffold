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

package cluster

import (
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestAppendCacheIfExists(t *testing.T) {
	tests := []struct {
		description  string
		cache        *latest.KanikoCache
		args         []string
		expectedArgs []string
	}{
		{
			description:  "no cache",
			cache:        nil,
			args:         []string{"some", "args"},
			expectedArgs: []string{"some", "args"},
		}, {
			description:  "cache layers",
			cache:        &latest.KanikoCache{},
			args:         []string{"some", "more", "args"},
			expectedArgs: []string{"some", "more", "args", "--cache=true"},
		}, {
			description: "cache layers to specific repo",
			cache: &latest.KanikoCache{
				Repo: "myrepo",
			},
			args:         []string{"initial", "args"},
			expectedArgs: []string{"initial", "args", "--cache=true", "--cache-repo=myrepo"},
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			actual := appendCacheIfExists(test.args, test.cache)

			t.CheckDeepEqual(test.expectedArgs, actual)
		})
	}
}

func TestAppendTargetIfExists(t *testing.T) {
	tests := []struct {
		description  string
		target       string
		args         []string
		expectedArgs []string
	}{
		{
			description:  "pass in empty target",
			args:         []string{"first", "args"},
			expectedArgs: []string{"first", "args"},
		}, {
			description:  "pass in target",
			target:       "stageOne",
			args:         []string{"first", "args"},
			expectedArgs: []string{"first", "args", "--target=stageOne"},
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			actual := appendTargetIfExists(test.args, test.target)

			t.CheckDeepEqual(test.expectedArgs, actual)
		})
	}
}

func TestAppendBuildArgsIfExists(t *testing.T) {
	tests := []struct {
		description  string
		buildArgs    map[string]*string
		args         []string
		expectedArgs []string
	}{
		{
			description:  "no build args",
			args:         []string{"first", "args"},
			expectedArgs: []string{"first", "args"},
		}, {
			description: "buid args",
			buildArgs: map[string]*string{
				"nil_key":   nil,
				"empty_key": pointer(""),
				"value_key": pointer("value"),
			},
			args:         []string{"first", "args"},
			expectedArgs: []string{"first", "args", "--build-arg", "empty_key=", "--build-arg", "nil_key", "--build-arg", "value_key=value"},
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			actual := appendBuildArgsIfExists(test.args, test.buildArgs)

			t.CheckDeepEqual(test.expectedArgs, actual)
		})
	}
}

func pointer(a string) *string {
	return &a
}
