package main

import (
	"fmt"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	podResource         = metav1.GroupVersionResource{Version: "v1", Resource: "pods"}
	volumeClaimResource = metav1.GroupVersionResource{Version: "v1", Resource: "persistentvolumeclaims"}
)

// applyMutations implements the logic of our admission controller webhook.
func (wh *mutationWH) applyMutations(req *admissionv1.AdmissionRequest) ([]patchOperation, error) {
	// This handler should only get called on Pod or Pvc objects as per the MutatingWebhookConfiguration in the YAML file.
	// However, if (for whatever reason) this gets invoked on an object of a different kind, issue a log message but
	// let the object request pass through otherwise.
	if req.Resource == podResource {

		// Parse the Pod object.
		raw := req.Object.Raw
		pod := corev1.Pod{}
		if _, _, err := universalDeserializer.Decode(raw, nil, &pod); err != nil {
			return nil, fmt.Errorf("could not deserialize pod object: %v", err)
		}

		return wh.applyMutationOnPod(pod)

	} else if req.Resource == volumeClaimResource {

		// Parse the Pvc object.
		raw := req.Object.Raw
		pvc := corev1.PersistentVolumeClaim{}
		if _, _, err := universalDeserializer.Decode(raw, nil, &pvc); err != nil {
			return nil, fmt.Errorf("could not deserialize pvc object: %v", err)
		}

		return wh.applyMutationOnPvc(pvc)
	}

	log.Printf("Got an unexpected resource %s, don't know what to do with...", req.Resource)
	return nil, nil
}

// applyMutationOnPod gets the deserialized pod spec and returns the patch operations
// to apply, if any, or an error if something went wrong.
func (wh *mutationWH) applyMutationOnPod(pod corev1.Pod) ([]patchOperation, error) {

	var patches []patchOperation

	if wh.registry != "" {
		if pod.Spec.InitContainers != nil {
			for i, c := range pod.Spec.InitContainers {
				log.Tracef("/spec/initContainers/%d/image = %s", i, c.Image)

				if !containsAnyRegistry(c.Image, append(wh.ignoredRegistries, wh.registry)) {
					patches = append(patches, patchOperation{
						Op:    "replace",
						Path:  fmt.Sprintf("/spec/initContainers/%d/image", i),
						Value: replaceRegistryIfSet(c.Image, wh.registry),
					})
				}
			}
		}

		if pod.Spec.Containers != nil {
			for i, c := range pod.Spec.Containers {
				log.Tracef("/spec/containers/%d/image = %s", i, c.Image)

				if !containsAnyRegistry(c.Image, append(wh.ignoredRegistries, wh.registry)) {
					patches = append(patches, patchOperation{
						Op:    "replace",
						Path:  fmt.Sprintf("/spec/containers/%d/image", i),
						Value: replaceRegistryIfSet(c.Image, wh.registry),
					})
				}
			}
		}
	}

	if wh.forceImagePullPolicy {
		if pod.Spec.InitContainers != nil {
			for i, c := range pod.Spec.InitContainers {
				log.Tracef("/spec/initContainers/%d/imagePullPolicy = %s", i, c.ImagePullPolicy)
				op := "replace"
				// still take the case when ImagePullPolicy is empty, but this case should not happen.
				// Policy defaults to Always if tag is latest, IfNotPresent otherwise.
				if c.ImagePullPolicy == "" {
					op = "add"
				}
				if wh.imagePullPolicyToForce != string(c.ImagePullPolicy) {
					patches = append(patches, patchOperation{
						Op:    op,
						Path:  fmt.Sprintf("/spec/initContainers/%d/imagePullPolicy", i),
						Value: wh.imagePullPolicyToForce,
					})
				}
			}
		}

		if pod.Spec.Containers != nil {
			for i, c := range pod.Spec.Containers {
				log.Tracef("/spec/containers/%d/imagePullPolicy = %s", i, c.ImagePullPolicy)
				if c.ImagePullPolicy != corev1.PullAlways {
					op := "replace"
					if c.ImagePullPolicy == "" {
						op = "add"
					}
					patches = append(patches, patchOperation{
						Op:    op,
						Path:  fmt.Sprintf("/spec/containers/%d/imagePullPolicy", i),
						Value: corev1.PullAlways,
					})
				}
			}
		}
	}

	if wh.imagePullSecret != "" {
		if pod.Spec.ImagePullSecrets == nil {
			patches = append(patches, patchOperation{
				Op:    "add",
				Path:  "/spec/imagePullSecrets",
				Value: []map[string]string{{"name": wh.imagePullSecret}},
			})
		} else if wh.appendImagePullSecret {
			imagePullSecretsAlreadyExist := false
			for _, i := range pod.Spec.ImagePullSecrets {
				if i.Name == wh.imagePullSecret {
					imagePullSecretsAlreadyExist = true
					break
				}
			}
			if !imagePullSecretsAlreadyExist {
				patches = append(patches, patchOperation{
					Op:    "add",
					Path:  fmt.Sprintf("/spec/imagePullSecrets/%d", len(pod.Spec.ImagePullSecrets)),
					Value: []map[string]string{{"name": wh.imagePullSecret}},
				})
			}
		} else {
			if !(len(pod.Spec.ImagePullSecrets) == 1 && pod.Spec.ImagePullSecrets[0].Name == wh.imagePullSecret) {
				patches = append(patches, patchOperation{
					Op:    "replace",
					Path:  "/spec/imagePullSecrets",
					Value: []map[string]string{{"name": wh.imagePullSecret}},
				})
			}
		}
	}

	log.Debugf("Patch applied: %v", patches)

	return patches, nil
}

// replaceRegistryIfSet assumes the image format is a.b[:port]/c/d:e
// if a.b is present, it is replaced by the registry given as argument.
func replaceRegistryIfSet(image string, registry string) string {

	imageParts := strings.Split(image, "/")

	if len(imageParts) == 1 {
		// case imagename or imagename:version, where version can contains .
		imageParts = append([]string{registry}, imageParts...)
	} else {
		// case something/imagename:version, assessing the something part.
		if strings.Contains(imageParts[0], ".") {
			imageParts[0] = registry
		} else {
			imageParts = append([]string{registry}, imageParts...)
		}
	}

	return strings.Join(imageParts, "/")
}

// applyMutationOnPvc gets the deserialized pvc spec and returns the patch operations
// to apply, if any, or an error if something went wrong.
func (wh *mutationWH) applyMutationOnPvc(pvc corev1.PersistentVolumeClaim) ([]patchOperation, error) {

	var patches []patchOperation

	if wh.defaultStorageClass != "" {
		if pvc.Spec.StorageClassName != nil {
			if *pvc.Spec.StorageClassName != wh.defaultStorageClass {
				patches = append(patches, patchOperation{
					Op:    "replace",
					Path:  "/spec/storageClassName",
					Value: wh.defaultStorageClass,
				})
			}
		} else {
			patches = append(patches, patchOperation{
				Op:    "add",
				Path:  "/spec/storageClassName",
				Value: wh.defaultStorageClass,
			})
		}
	}

	log.Debugf("Patch applied: %v", patches)

	return patches, nil
}

// containsRegistry returns true if the image "contains",
// i.e. start with the registry prefix.
// A tailing / is added during the comparison to ensure
// the registry is not only a prefix of the image.
func containsRegistry(image string, registry string) bool {
	return strings.HasPrefix(image, registry+"/")
}

func containsAnyRegistry(image string, registries []string) bool {
	for _, s := range registries {
		if containsRegistry(image, s) {
			return true
		}
	}
	return false
}

func isPullPolicyValid(policy string) bool {
	switch policy {
	case string(corev1.PullAlways):
		return true
	case string(corev1.PullIfNotPresent):
		return true
	case string(corev1.PullNever):
		return true
	default:
		return false
	}
}
