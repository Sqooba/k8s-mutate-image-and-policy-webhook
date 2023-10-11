package main

import (
	"bytes"
	simplejson "encoding/json"
	"fmt"
	"io"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

const (
	jsonContentType = `application/json`
)

var (
	groupVersion = schema.GroupVersion{
		Group:   "admission.k8s.io",
		Version: "v1",
	}
	scheme                = runtime.NewScheme()
	codecFactory          = serializer.NewCodecFactory(scheme)
	universalDeserializer = codecFactory.UniversalDeserializer()
	jsonSerializer        = json.NewSerializerWithOptions(
		json.DefaultMetaFactory, scheme, scheme,
		json.SerializerOptions{Yaml: false, Pretty: false, Strict: true},
	)
	encoder   = codecFactory.EncoderForVersion(jsonSerializer, groupVersion)
	patchType = admissionv1.PatchTypeJSONPatch
)

// patchOperation is an operation of a JSON patch, see https://tools.ietf.org/html/rfc6902 .
type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// admitFunc is a callback for admission controller logic. Given an AdmissionRequest, it returns the sequence of patch
// operations to be applied in case of success, or the error that will be shown when the operation is rejected.
type admitFunc func(*admissionv1.AdmissionRequest) ([]patchOperation, error)

// isExcludedNamespace checks if the given namespace is a excluded via the configuration.
func isExcludedNamespace(ns string, excludedNamespaces []string) bool {
	for _, a := range excludedNamespaces {
		if a == ns {
			return true
		}
	}
	return false
}

// doServeAdmitFunc parses the HTTP request for an admission controller webhook, and -- in case of a well-formed
// request -- delegates the admission control logic to the given admitFunc. The response body is then returned as raw
// bytes.
func (wh *mutationWH) doServeAdmitFunc(w http.ResponseWriter, r *http.Request, admit admitFunc) (runtime.Object, error) {
	// Step 1: Request validation. Only handle POST requests with a body and json content type.

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil, fmt.Errorf("k8s-mutate-image-and-policy-webhook: invalid method %s, only POST requests are allowed", r.Method)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("k8s-mutate-image-and-policy-webhook: could not read request body: %v", err)
	}

	if contentType := r.Header.Get("Content-Type"); contentType != jsonContentType {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("k8s-mutate-image-and-policy-webhook: unsupported content type %s, only %s is supported", contentType, jsonContentType)
	}

	// Step 2: Parse the AdmissionReview request.
	var admissionReviewReq admissionv1.AdmissionReview
	log.Tracef("About to deserialize the request, request = %s", string(body))

	// Step 3: Construct the AdmissionReview response.
	admissionReviewResponse := admissionv1.AdmissionReview{
		Response: &admissionv1.AdmissionResponse{},
	}

	if _, _, err := universalDeserializer.Decode(body, nil, &admissionReviewReq); err != nil {
		log.Printf("Got an error while deserializing the request, %v, request = %s", err, string(body))
		admissionReviewResponse.Response.Allowed = false
		admissionReviewResponse.Response.Result = &metav1.Status{
			Message: fmt.Sprintf("Got an error while deserializing the request: %s", err.Error()),
			Reason:  metav1.StatusReasonBadRequest,
		}
		//return nil, fmt.Errorf("k8s-mutate-image-and-policy-webhook: could not deserialize request: %v", err)
	} else if admissionReviewReq.Request == nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Deserializing the request produced a empty review, request = %s", string(body))
		admissionReviewResponse.Response.Allowed = false
		admissionReviewResponse.Response.Result = &metav1.Status{
			Message: "Deserializing the request produced a empty review",
			Reason:  metav1.StatusReasonInvalid,
		}
		//return admissionReviewResponse, errors.New("k8s-mutate-image-and-policy-webhook: malformed admission review: request is nil")
	} else {
		admissionReviewResponse.Response.UID = admissionReviewReq.Request.UID

		var patchOps []patchOperation

		// Apply the admit() function only for non-excluded namespaces. For objects excluded, return
		// an empty set of patch operations.
		if !isExcludedNamespace(admissionReviewReq.Request.Namespace, wh.excludedNamespaces) {
			patchOps, err = admit(admissionReviewReq.Request)
		} else {
			log.Debugf("Namespace is excluded")
		}

		if err != nil {
			// If the handler returned an error, incorporate the error message into the response and deny the object
			// creation.
			admissionReviewResponse.Response.Allowed = false
			admissionReviewResponse.Response.Result = &metav1.Status{
				Message: err.Error(),
				Reason:  metav1.StatusReasonBadRequest,
			}
		} else {
			// Otherwise, encode the patch operations to JSON and return a positive response.
			patchBytes, err := simplejson.Marshal(patchOps)
			if err != nil {
				log.Printf("Got an error while serializing the patches, %v", err)
				admissionReviewResponse.Response.Allowed = false
				admissionReviewResponse.Response.Result = &metav1.Status{
					Message: err.Error(),
					Reason:  metav1.StatusReasonInternalError,
				}
			} else {
				admissionReviewResponse.Response.Allowed = true
				admissionReviewResponse.Response.Patch = patchBytes
				admissionReviewResponse.Response.PatchType = &patchType
			}
		}
	}

	// Return the AdmissionReview with a response as JSON.
	return &admissionReviewResponse, err
}

// serveAdmitFunc is a wrapper around doServeAdmitFunc that adds error handling and logging.
func (wh *mutationWH) serveAdmitFunc(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	log.Tracef("Webhook request starts...")

	var writeErr error
	if object, err := wh.doServeAdmitFunc(w, r, admit); err != nil {
		log.Printf("Error handling webhook request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, writeErr = w.Write([]byte(err.Error()))
	} else {
		w.Header().Add("Content-Type", jsonContentType)
		buf := new(bytes.Buffer)
		writeErr = encoder.Encode(object, buf)
		if writeErr == nil {
			log.Tracef("Serialized response: %s", buf.String())
			_, writeErr = w.Write(buf.Bytes())
		}
	}

	if writeErr != nil {
		log.Printf("Could not write response: %v", writeErr)
	}
	log.Tracef("...Webhook request ends")
}

// admitFuncHandler takes an admitFunc and wraps it into a http.Handler by means of calling serveAdmitFunc.
func (wh *mutationWH) admitFuncHandler(admit admitFunc) http.Handler {

	//Some initialisation...
	scheme.AddKnownTypes(groupVersion, &admissionv1.AdmissionReview{})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wh.serveAdmitFunc(w, r, admit)
	})
}
