package kms

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("minio_kms_key", func(r *config.Resource) {
		r.ShortGroup = "kms"
		r.Kind = "Key"
	})
}