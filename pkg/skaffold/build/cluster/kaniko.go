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
	"context"
	"fmt"
	"io"
	"sort"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build/cache"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build/cluster/sources"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/constants"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (b *Builder) runKanikoBuild(ctx context.Context, out io.Writer, artifact *latest.Artifact, tag string) (string, error) {
	// Prepare context
	s := sources.Retrieve(b.ClusterDetails, artifact.KanikoArtifact)
	dependencies, err := b.DependenciesForArtifact(ctx, artifact)
	if err != nil {
		return "", errors.Wrapf(err, "getting dependencies for %s", artifact.ImageName)
	}
	context, err := s.Setup(ctx, out, artifact, util.RandomID(), dependencies)
	if err != nil {
		return "", errors.Wrap(err, "setting up build context")
	}
	defer s.Cleanup(ctx)

	kanikoArtifact := artifact.KanikoArtifact
	// Create pod spec
	args := []string{
		"--dockerfile", kanikoArtifact.DockerfilePath,
		"--context", context,
		"--destination", tag,
		"-v", logLevel().String()}

	// TODO: remove since AdditionalFlags will be deprecated (priyawadhwa@)
	if kanikoArtifact.AdditionalFlags != nil {
		logrus.Warn("The additionalFlags field in kaniko is deprecated, please consult the current schema at skaffold.dev to update your skaffold.yaml.")
		args = append(args, kanikoArtifact.AdditionalFlags...)
	}
	args = appendBuildArgsIfExists(args, kanikoArtifact.BuildArgs)
	args = appendTargetIfExists(args, kanikoArtifact.Target)
	args = appendCacheIfExists(args, kanikoArtifact.Cache)

	if artifact.WorkspaceHash != "" {
		hashTag := cache.HashTag(artifact)
		args = append(args, []string{"--destination", hashTag}...)
	}

	podSpec := s.Pod(args)
	// Create pod
	client, err := kubernetes.GetClientset()
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	pods := client.CoreV1().Pods(b.Namespace)

	pod, err := pods.Create(podSpec)
	if err != nil {
		return "", errors.Wrap(err, "creating kaniko pod")
	}
	defer func() {
		if err := pods.Delete(pod.Name, &metav1.DeleteOptions{
			GracePeriodSeconds: new(int64),
		}); err != nil {
			logrus.Fatalf("deleting pod: %s", err)
		}
	}()

	if err := s.ModifyPod(ctx, pod); err != nil {
		return "", errors.Wrap(err, "modifying kaniko pod")
	}

	waitForLogs := streamLogs(out, pod.Name, pods)

	if err := kubernetes.WaitForPodSucceeded(ctx, pods, pod.Name, b.timeout); err != nil {
		return "", errors.Wrap(err, "waiting for pod to complete")
	}

	waitForLogs()

	return docker.RemoteDigest(tag, b.insecureRegistries)
}

func appendCacheIfExists(args []string, cache *latest.KanikoCache) []string {
	if cache == nil {
		return args
	}
	args = append(args, "--cache=true")
	if cache.Repo != "" {
		args = append(args, fmt.Sprintf("--cache-repo=%s", cache.Repo))
	}
	if cache.HostPath != "" {
		args = append(args, fmt.Sprintf("--cache-dir=%s", constants.DefaultKanikoDockerConfigPath))
	}
	return args
}

func appendTargetIfExists(args []string, target string) []string {
	if target == "" {
		return args
	}
	return append(args, fmt.Sprintf("--target=%s", target))
}

func appendBuildArgsIfExists(args []string, buildArgs map[string]*string) []string {
	if buildArgs == nil {
		return args
	}

	var keys []string
	for k := range buildArgs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		args = append(args, "--build-arg")

		v := buildArgs[k]
		if v == nil {
			args = append(args, k)
		} else {
			args = append(args, fmt.Sprintf("%s=%s", k, *v))
		}
	}
	return args
}
