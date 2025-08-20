/*
Copyright 2022 Upbound Inc.
*/

package config

import "github.com/crossplane/upjet/pkg/config"

// ExternalNameConfigs contains all external name configurations for this
// provider.
var ExternalNameConfigs = map[string]config.ExternalName{
	// S3 Resources - bucket uses "bucket" field, not "name"
	"minio_s3_bucket":              config.IdentifierFromProvider,
	"minio_s3_bucket_policy":       config.TemplatedStringAsIdentifier("bucket", "{{ .external_name }}"),
	"minio_s3_object":              config.TemplatedStringAsIdentifier("object_name", "{{ .external_name }}"),
	
	// IAM Resources - these use "name" field correctly
	"minio_iam_user":               config.NameAsIdentifier,
	"minio_iam_policy":             config.NameAsIdentifier,
	
	// Next 5 resources
	"minio_iam_group":              config.NameAsIdentifier,  // uses "name" field
	"minio_s3_bucket_versioning":   config.TemplatedStringAsIdentifier("bucket", "{{ .external_name }}"),  // uses "bucket" field
	"minio_s3_bucket_notification": config.TemplatedStringAsIdentifier("bucket", "{{ .external_name }}"),  // uses "bucket" field  
	"minio_kms_key":                config.TemplatedStringAsIdentifier("key_id", "{{ .external_name }}"),  // uses "key_id" field
	"minio_iam_service_account":    config.IdentifierFromProvider,  // uses computed "access_key" field
}

// ExternalNameConfigurations applies all external name configs listed in the
// table ExternalNameConfigs and sets the version of those resources to v1beta1
// assuming they will be tested.
func ExternalNameConfigurations() config.ResourceOption {
	return func(r *config.Resource) {
		if e, ok := ExternalNameConfigs[r.Name]; ok {
			r.ExternalName = e
		}
	}
}

// ExternalNameConfigured returns the list of all resources whose external name
// is configured manually.
func ExternalNameConfigured() []string {
	l := make([]string, len(ExternalNameConfigs))
	i := 0
	for name := range ExternalNameConfigs {
		// $ is added to match the exact string since the format is regex.
		l[i] = name + "$"
		i++
	}
	return l
}
