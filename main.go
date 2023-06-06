package main

import (
	"flag"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"github.com/sqooba/go-common/logging"
	"github.com/sqooba/go-common/version"
)

type envConfig struct {
	TLSCertFile            string   `envconfig:"TLS_CERT_FILE" default:"/run/secrets/tls/webhook-server-tls.crt"`
	TLSKeyFile             string   `envconfig:"TLS_KEY_FILE" default:"/run/secrets/tls/webhook-server-tls.key"`
	Port                   string   `envconfig:"PORT" default:"8443"`
	LogLevel               string   `envconfig:"LOG_LEVEL" default:"info"`
	Registry               string   `envconfig:"REGISTRY"`
	ImagePullSecret        string   `envconfig:"IMAGE_PULL_SECRET"`
	AppendImagePullSecret  bool     `envconfig:"IMAGE_PULL_SECRET_APPEND" default:"false"`
	ForceImagePullPolicy   bool     `envconfig:"FORCE_IMAGE_PULL_POLICY"`
	ImagePullPolicyToForce string   `envconfig:"IMAGE_PULL_POLICY_TO_FORCE" default:"Always"`
	DefaultStorageClass    string   `envconfig:"DEFAULT_STORAGE_CLASS"`
	ExcludeNamespaces      []string `envconfig:"EXCLUDE_NAMESPACES"`
	IgnoredRegistries      []string `envconfig:"IGNORED_REGISTRIES"`
}

var (
	log = logging.NewLogger()
)

type mutationWH struct {
	registry               string
	imagePullSecret        string
	appendImagePullSecret  bool
	forceImagePullPolicy   bool
	imagePullPolicyToForce string
	defaultStorageClass    string
	excludedNamespaces     []string
	ignoredRegistries      []string
}

func main() {
	log.Println("k8s-mutate-image-and-policy-webhook is starting...")
	log.Printf("Version    : %s", version.Version)
	log.Printf("Commit     : %s", version.GitCommit)
	log.Printf("Build date : %s", version.BuildDate)
	log.Printf("OSarch     : %s", version.OsArch)

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s\n", err)
		return
	}

	flag.Parse()
	err := logging.SetLogLevel(log, env.LogLevel)
	if err != nil {
		log.Fatalf("Logging level %s do not seem to be right. Err = %v", env.LogLevel, err)
	}

	// Validate pull policy
	pullPolicyValid := isPullPolicyValid(env.ImagePullPolicyToForce)
	if !pullPolicyValid {
		log.Fatalf("Pull policy %s is not valid. Fix IMAGE_PULL_POLICY_TO_FORCE and retry", env.ImagePullPolicyToForce)
	}

	wh := mutationWH{
		registry:               env.Registry,
		imagePullSecret:        env.ImagePullSecret,
		appendImagePullSecret:  env.AppendImagePullSecret,
		forceImagePullPolicy:   env.ForceImagePullPolicy,
		imagePullPolicyToForce: env.ImagePullPolicyToForce,
		defaultStorageClass:    env.DefaultStorageClass,
		excludedNamespaces:     env.ExcludeNamespaces,
		ignoredRegistries:      env.IgnoredRegistries,
	}

	mux := http.NewServeMux()

	wh.routes(mux, env)

	server := &http.Server{
		// We listen on port 8443 such that we do not need root privileges or extra capabilities for this server.
		// The Service object will take care of mapping this port to the HTTPS port 443.
		Addr:    ":" + env.Port,
		Handler: mux,
	}

	log.Fatal(server.ListenAndServeTLS(env.TLSCertFile, env.TLSKeyFile))
}
