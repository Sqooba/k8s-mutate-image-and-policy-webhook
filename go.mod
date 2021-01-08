module github.com/sqooba/k8s-mutate-image-and-policy

go 1.15

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sqooba/go-common/healthchecks v0.0.0-00010101000000-000000000000
	github.com/sqooba/go-common/logging v0.0.0-00010101000000-000000000000
	github.com/sqooba/go-common/version v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/text v0.3.3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	k8s.io/api v0.17.9
	k8s.io/apimachinery v0.17.9
	k8s.io/klog v1.0.0 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)

replace (
	github.com/sqooba/go-common/healthchecks => ../sqooba-go-common/healthchecks
	github.com/sqooba/go-common/logging => ../sqooba-go-common/logging
	github.com/sqooba/go-common/version => ../sqooba-go-common/version
)
