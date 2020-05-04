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

package buildpacks

import (
	"context"
	"io"
	"io/ioutil"
	"testing"

	"github.com/buildpacks/pack"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

type fakePack struct {
	Opts pack.BuildOptions
}

func (f *fakePack) runPack(_ context.Context, _ io.Writer, _ docker.LocalDaemon, opts pack.BuildOptions) error {
	f.Opts = opts
	return nil
}

func TestBuild(t *testing.T) {
	tests := []struct {
		description     string
		artifact        *latest.Artifact
		tag             string
		api             *testutil.FakeAPIClient
		pushImages      bool
		devMode         bool
		shouldErr       bool
		expectedOptions *pack.BuildOptions
	}{
		{
			description: "success",
			artifact:    buildpacksArtifact("my/builder", "my/run"),
			tag:         "img:tag",
			api:         &testutil.FakeAPIClient{},
			expectedOptions: &pack.BuildOptions{
				AppPath:  ".",
				Builder:  "my/builder",
				RunImage: "my/run",
				Env:      map[string]string{},
				Image:    "img:latest",
			},
		},
		{
			description: "success with buildpacks",
			artifact:    withBuildpacks([]string{"my/buildpack", "my/otherBuildpack"}, buildpacksArtifact("my/otherBuilder", "my/otherRun")),
			tag:         "img:tag",
			api:         &testutil.FakeAPIClient{},
			expectedOptions: &pack.BuildOptions{
				AppPath:    ".",
				Builder:    "my/otherBuilder",
				RunImage:   "my/otherRun",
				Buildpacks: []string{"my/buildpack", "my/otherBuildpack"},
				Env:        map[string]string{},
				Image:      "img:latest",
			},
		},
		{
			description: "dev mode",
			artifact:    withSync(&latest.Sync{Auto: &latest.Auto{}}, buildpacksArtifact("another/builder", "another/run")),
			tag:         "img:tag",
			api:         &testutil.FakeAPIClient{},
			devMode:     true,
			expectedOptions: &pack.BuildOptions{
				AppPath:  ".",
				Builder:  "another/builder",
				RunImage: "another/run",
				Env: map[string]string{
					"GOOGLE_DEVMODE": "1",
				},
				Image: "img:latest",
			},
		},
		{
			description: "dev mode but no sync",
			artifact:    buildpacksArtifact("my/other-builder", "my/run"),
			tag:         "img:tag",
			api:         &testutil.FakeAPIClient{},
			devMode:     true,
			expectedOptions: &pack.BuildOptions{
				AppPath:  ".",
				Builder:  "my/other-builder",
				RunImage: "my/run",
				Env:      map[string]string{},
				Image:    "img:latest",
			},
		},
		{
			description: "invalid ref",
			artifact:    buildpacksArtifact("my/builder", "my/run"),
			tag:         "in valid ref",
			api:         &testutil.FakeAPIClient{},
			shouldErr:   true,
		},
		{
			description: "push error",
			artifact:    buildpacksArtifact("my/builder", "my/run"),
			tag:         "img:tag",
			pushImages:  true,
			api: &testutil.FakeAPIClient{
				ErrImagePush: true,
			},
			shouldErr: true,
		},
		{
			description: "invalid env",
			artifact:    withEnv([]string{"INVALID"}, buildpacksArtifact("my/builder", "my/run")),
			tag:         "img:tag",
			api:         &testutil.FakeAPIClient{},
			shouldErr:   true,
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			t.NewTempDir().Touch("file").Chdir()
			pack := &fakePack{}
			t.Override(&runPackBuildFunc, pack.runPack)

			test.api.
				Add(test.artifact.BuildpackArtifact.Builder, "builderImageID").
				Add(test.artifact.BuildpackArtifact.RunImage, "runImageID").
				Add("img:latest", "builtImageID")
			localDocker := docker.NewLocalDaemon(test.api, nil, false, nil)

			builder := NewArtifactBuilder(localDocker, test.pushImages, test.devMode)
			_, err := builder.Build(context.Background(), ioutil.Discard, test.artifact, test.tag)

			t.CheckError(test.shouldErr, err)
			if test.expectedOptions != nil {
				t.CheckDeepEqual(*test.expectedOptions, pack.Opts)
			}
		})
	}
}

func buildpacksArtifact(builder, runImage string) *latest.Artifact {
	return &latest.Artifact{
		Workspace: ".",
		ArtifactType: latest.ArtifactType{
			BuildpackArtifact: &latest.BuildpackArtifact{
				Builder:  builder,
				RunImage: runImage,
				Dependencies: &latest.BuildpackDependencies{
					Paths: []string{"."},
				},
			},
		},
	}
}

func withEnv(env []string, artifact *latest.Artifact) *latest.Artifact {
	artifact.BuildpackArtifact.Env = env
	return artifact
}

func withSync(sync *latest.Sync, artifact *latest.Artifact) *latest.Artifact {
	artifact.Sync = sync
	return artifact
}

func withBuildpacks(buildpacks []string, artifact *latest.Artifact) *latest.Artifact {
	artifact.BuildpackArtifact.Buildpacks = buildpacks
	return artifact
}
