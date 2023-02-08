package main

import (
	"github.com/sqooba/go-common/logging"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestImageNotSet(t *testing.T) {

	wh := mutationWH{
		registry: "",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(patches))
}

func TestImageInvalidName(t *testing.T) {

	// This is assumed to be a container name only with a version.
	// This image name is not supported by the webhook anyway.

	wh := mutationWH{
		registry: "x.y",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "a.b:c"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "x.y/a.b:c", patches[0].Value)
}

func TestImageWithPort(t *testing.T) {
	wh := mutationWH{
		registry: "x.y",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "x.y:80/a.b:c"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "x.y/a.b:c", patches[0].Value)
}

func TestImageWithPort2(t *testing.T) {
	wh := mutationWH{
		registry: "x.y:80",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "x.y/a/b:c"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "x.y:80/a/b:c", patches[0].Value)
}

func TestImageWithDots(t *testing.T) {
	wh := mutationWH{
		registry: "x.y",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "a.b/c.d:e"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "x.y/c.d:e", patches[0].Value)
}

func TestImageWithRegistry(t *testing.T) {

	wh := mutationWH{
		registry: "x.y",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "a.b/c/d:e"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "x.y/c/d:e", patches[0].Value)
}

func TestImageWithCorrectRegistry(t *testing.T) {

	wh := mutationWH{
		registry: "a.b",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "a.b/c/d:e"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(patches))
}

func TestImageWithoutRegistry(t *testing.T) {

	wh := mutationWH{
		registry: "a.b",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "c/d:e"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "a.b/c/d:e", patches[0].Value)
}

func TestImageWithoutRegistryNorTag(t *testing.T) {

	wh := mutationWH{
		registry: "a.b",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "c/d"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "a.b/c/d", patches[0].Value)
}

func TestImageWithRegistryAndPort(t *testing.T) {

	wh := mutationWH{
		registry: "a.b",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "c.d:80/e/f:g"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "a.b/e/f:g", patches[0].Value)
}

func TestImageWithRegistryAndPort2(t *testing.T) {

	wh := mutationWH{
		registry: "a.b",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "c.d:e/f:g"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "a.b/f:g", patches[0].Value)
}

func TestRealImage(t *testing.T) {

	wh := mutationWH{
		registry: "dev-registry.metis.test.sqooba.io",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "dev-registry.metis.zoo/traefik:v1.7"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "dev-registry.metis.test.sqooba.io/traefik:v1.7", patches[0].Value)
}

func TestImageElastic(t *testing.T) {

	wh := mutationWH{
		registry: "x.y.z",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "elastic:7.4.2"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "x.y.z/elastic:7.4.2", patches[0].Value)
}

func TestImageSkaffold(t *testing.T) {

	wh := mutationWH{
		registry: "docker.sqooba.io",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "sqooba/sqooba-website:32cbc804-dirty@sha256:a4a729d8691ed70eb56cf03053333cf42e8a6c33f6ee67ea862da4459d7f70fd"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "docker.sqooba.io/sqooba/sqooba-website:32cbc804-dirty@sha256:a4a729d8691ed70eb56cf03053333cf42e8a6c33f6ee67ea862da4459d7f70fd", patches[0].Value)
}

func TestImageInternalFullRegistry(t *testing.T) {

	wh := mutationWH{
		registry: "docker.sqooba.io/public-docker-virtual",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "docker.sqooba.io/local-repo/xyz/image:snapshot"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	// actual behaviour if the registry docker.sqooba.io/local-repo is not ignored
	assert.Equal(t, "docker.sqooba.io/public-docker-virtual/local-repo/xyz/image:snapshot", patches[0].Value)
}

func TestImageInternalFullRegistryWithIgnoreNoReplace(t *testing.T) {

	wh := mutationWH{
		registry:          "docker.sqooba.io/public-docker-virtual",
		ignoredRegistries: []string{"docker.sqooba.io/local-repo", "ignoreme.io/local"},
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "docker.sqooba.io/local-repo/xyz/image:snapshot"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(patches))
}

func TestImageInternalFullRegistryWithIgnoreReplace1(t *testing.T) {

	wh := mutationWH{
		registry:          "docker.sqooba.io/public-docker-virtual",
		ignoredRegistries: []string{"docker.sqooba.io/local-repo", "ignoreme.io/local"},
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "docker.sqooba.io/local-repo-2/xyz/image:snapshot"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "docker.sqooba.io/public-docker-virtual/local-repo-2/xyz/image:snapshot", patches[0].Value)
}

func TestImageInternalFullRegistryWithIgnoreReplace2(t *testing.T) {

	wh := mutationWH{
		registry:          "docker.sqooba.io/public-docker-virtual",
		ignoredRegistries: []string{"docker.sqooba.io/local-repo", "ignoreme.io/local"},
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "any.registry.io/whatever/xyz/image:snapshot"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "docker.sqooba.io/public-docker-virtual/whatever/xyz/image:snapshot", patches[0].Value)
}

func TestImageExternal(t *testing.T) {

	wh := mutationWH{
		registry: "docker.sqooba.io/public-docker-virtual",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "quay.io/argoproj/argocd:v2.0.1"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "docker.sqooba.io/public-docker-virtual/argoproj/argocd:v2.0.1", patches[0].Value)
}

func TestImageExternalPathPrefixShortImage(t *testing.T) {
	logging.SetLogLevel(log, "trace")

	wh := mutationWH{
		registry: "docker.sqooba.io/public-docker-virtual",
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "victoriametrics/victoria-metrics:v1.40.0"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/image", patches[0].Path)
	assert.Equal(t, "docker.sqooba.io/public-docker-virtual/victoriametrics/victoria-metrics:v1.40.0", patches[0].Value)
}

func TestImagePullSecretNotPresent(t *testing.T) {
	logging.SetLogLevel(log, "debug")

	wh := mutationWH{
		imagePullSecret: "random-pull-secret",
	}

	pod := corev1.Pod{

		Spec: corev1.PodSpec{},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "add", patches[0].Op)
	assert.Equal(t, "/spec/imagePullSecrets", patches[0].Path)
	assert.Equal(t, []map[string]string{{"name": "random-pull-secret"}}, patches[0].Value)
}

func TestImagePullSecretEmpty(t *testing.T) {
	logging.SetLogLevel(log, "debug")

	wh := mutationWH{
		imagePullSecret: "random-pull-secret",
	}

	pod := corev1.Pod{

		Spec: corev1.PodSpec{
			ImagePullSecrets: []corev1.LocalObjectReference{},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/imagePullSecrets", patches[0].Path)
	assert.Equal(t, []map[string]string{{"name": "random-pull-secret"}}, patches[0].Value)
}

func TestImagePullSecretPresent(t *testing.T) {
	logging.SetLogLevel(log, "debug")

	wh := mutationWH{
		imagePullSecret: "random-pull-secret",
	}

	pod := corev1.Pod{

		Spec: corev1.PodSpec{
			ImagePullSecrets: []corev1.LocalObjectReference{
				{Name: "s1"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/imagePullSecrets", patches[0].Path)
	assert.Equal(t, []map[string]string{{"name": "random-pull-secret"}}, patches[0].Value)
}

func TestImagePullSecretWithAlreadyExistingSecret(t *testing.T) {
	logging.SetLogLevel(log, "debug")

	wh := mutationWH{
		imagePullSecret: "already-existing-pull-secret",
	}

	pod := corev1.Pod{

		Spec: corev1.PodSpec{
			ImagePullSecrets: []corev1.LocalObjectReference{
				{Name: "already-existing-pull-secret"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/imagePullSecrets", patches[0].Path)
	assert.Equal(t, []map[string]string{{"name": "already-existing-pull-secret"}}, patches[0].Value)
}

func TestImagePullSecretWithAppendAndEmptySecret(t *testing.T) {
	logging.SetLogLevel(log, "debug")

	wh := mutationWH{
		imagePullSecret:       "random-pull-secret",
		appendImagePullSecret: true,
	}

	pod := corev1.Pod{

		Spec: corev1.PodSpec{
			ImagePullSecrets: []corev1.LocalObjectReference{},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "add", patches[0].Op)
	assert.Equal(t, "/spec/imagePullSecrets/0", patches[0].Path)
	assert.Equal(t, []map[string]string{{"name": "random-pull-secret"}}, patches[0].Value)
}

func TestImagePullSecretWithAppendAndNoneExistingSecret(t *testing.T) {
	logging.SetLogLevel(log, "debug")

	wh := mutationWH{
		imagePullSecret:       "random-pull-secret",
		appendImagePullSecret: true,
	}

	pod := corev1.Pod{

		Spec: corev1.PodSpec{},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "add", patches[0].Op)
	assert.Equal(t, "/spec/imagePullSecrets", patches[0].Path)
	assert.Equal(t, []map[string]string{{"name": "random-pull-secret"}}, patches[0].Value)
}

func TestImagePullSecretWithAppendAndAlreadyExistingSecret(t *testing.T) {
	logging.SetLogLevel(log, "debug")

	wh := mutationWH{
		imagePullSecret:       "already-existing-pull-secret",
		appendImagePullSecret: true,
	}

	pod := corev1.Pod{

		Spec: corev1.PodSpec{
			ImagePullSecrets: []corev1.LocalObjectReference{
				{Name: "already-existing-pull-secret"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(patches))
}

func TestImagePullSecretAppendToExistingSecret(t *testing.T) {
	logging.SetLogLevel(log, "debug")

	wh := mutationWH{
		imagePullSecret:       "a-new-pull-secret",
		appendImagePullSecret: true,
	}

	pod := corev1.Pod{

		Spec: corev1.PodSpec{
			ImagePullSecrets: []corev1.LocalObjectReference{
				{Name: "already-existing-pull-secret"},
				{Name: "another-already-existing-pull-secret"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "add", patches[0].Op)
	assert.Equal(t, "/spec/imagePullSecrets/2", patches[0].Path)
	assert.Equal(t, []map[string]string{{"name": "a-new-pull-secret"}}, patches[0].Value)
}

func TestMissingPullPolicy(t *testing.T) {

	wh := mutationWH{
		forceImagePullPolicy: true,
	}

	pod := corev1.Pod{

		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "add", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/imagePullPolicy", patches[0].Path)
	assert.Equal(t, corev1.PullAlways, patches[0].Value)
}

func TestNotAlwaysPullPolicy(t *testing.T) {

	wh := mutationWH{
		forceImagePullPolicy: true,
	}

	pod := corev1.Pod{

		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{ImagePullPolicy: corev1.PullIfNotPresent},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/imagePullPolicy", patches[0].Path)
	assert.Equal(t, corev1.PullAlways, patches[0].Value)
}

func TestAlwaysPullPolicy(t *testing.T) {

	wh := mutationWH{
		forceImagePullPolicy: true,
	}

	pod := corev1.Pod{

		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{ImagePullPolicy: corev1.PullAlways},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(patches))
}

func TestMultipleContainers(t *testing.T) {

	wh := mutationWH{
		forceImagePullPolicy: true,
	}

	pod := corev1.Pod{

		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{ImagePullPolicy: corev1.PullIfNotPresent},
				{ImagePullPolicy: corev1.PullIfNotPresent},
				{ImagePullPolicy: corev1.PullIfNotPresent},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/containers/0/imagePullPolicy", patches[0].Path)
	assert.Equal(t, corev1.PullAlways, patches[0].Value)
	assert.Equal(t, "replace", patches[1].Op)
	assert.Equal(t, "/spec/containers/1/imagePullPolicy", patches[1].Path)
	assert.Equal(t, corev1.PullAlways, patches[1].Value)
	assert.Equal(t, "replace", patches[2].Op)
	assert.Equal(t, "/spec/containers/2/imagePullPolicy", patches[2].Path)
	assert.Equal(t, corev1.PullAlways, patches[2].Value)
}

func TestInitContainers(t *testing.T) {

	wh := mutationWH{
		forceImagePullPolicy: true,
	}

	pod := corev1.Pod{

		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{ImagePullPolicy: corev1.PullIfNotPresent},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/initContainers/0/imagePullPolicy", patches[0].Path)
	assert.Equal(t, corev1.PullAlways, patches[0].Value)
}

func TestWithAllMutations(t *testing.T) {

	wh := mutationWH{
		registry:             "x.y",
		imagePullSecret:      "s2",
		forceImagePullPolicy: true,
	}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					ImagePullPolicy: corev1.PullIfNotPresent,
					Image:           "b/c:d",
				},
			},
			ImagePullSecrets: []corev1.LocalObjectReference{
				{Name: "s1"},
			},
		},
	}

	patches, err := wh.applyMutationOnPod(pod)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(patches))
}

var storageClass1 = "storage-class-1"
var storageClass2 = "storage-class-2"

func TestWithSameStorageClass(t *testing.T) {

	wh := mutationWH{
		defaultStorageClass: storageClass1,
	}

	pvc := corev1.PersistentVolumeClaim{
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass1,
		},
	}

	patches, err := wh.applyMutationOnPvc(pvc)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(patches))
}

func TestWithNoStorageClass(t *testing.T) {

	wh := mutationWH{
		defaultStorageClass: storageClass1,
	}

	pvc := corev1.PersistentVolumeClaim{
		Spec: corev1.PersistentVolumeClaimSpec{},
	}

	patches, err := wh.applyMutationOnPvc(pvc)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "add", patches[0].Op)
	assert.Equal(t, "/spec/storageClassName", patches[0].Path)
	assert.Equal(t, storageClass1, patches[0].Value)
}

func TestWithDifferentStorageClass(t *testing.T) {

	wh := mutationWH{
		defaultStorageClass: storageClass1,
	}

	pvc := corev1.PersistentVolumeClaim{
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass2,
		},
	}

	patches, err := wh.applyMutationOnPvc(pvc)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "replace", patches[0].Op)
	assert.Equal(t, "/spec/storageClassName", patches[0].Path)
	assert.Equal(t, storageClass1, patches[0].Value)
}
