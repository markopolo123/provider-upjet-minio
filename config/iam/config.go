package iam

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("minio_iam_user", func(r *config.Resource) {
		r.ShortGroup = "iam"
		r.Kind = "User"
	})

	p.AddResourceConfigurator("minio_iam_policy", func(r *config.Resource) {
		r.ShortGroup = "iam"
		r.Kind = "Policy"
	})

	p.AddResourceConfigurator("minio_iam_group", func(r *config.Resource) {
		r.ShortGroup = "iam"
		r.Kind = "Group"
	})

	p.AddResourceConfigurator("minio_iam_service_account", func(r *config.Resource) {
		r.ShortGroup = "iam"
		r.Kind = "ServiceAccount"
	})
}