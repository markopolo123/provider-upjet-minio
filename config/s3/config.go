package s3

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("minio_s3_bucket", func(r *config.Resource) {
		r.ShortGroup = "s3"
		r.Kind = "Bucket"
	})

	p.AddResourceConfigurator("minio_s3_bucket_policy", func(r *config.Resource) {
		r.ShortGroup = "s3"
		r.Kind = "BucketPolicy"
		r.References["bucket"] = config.Reference{
			TerraformName: "minio_s3_bucket",
		}
	})

	p.AddResourceConfigurator("minio_s3_object", func(r *config.Resource) {
		r.ShortGroup = "s3"
		r.Kind = "Object"
		r.References["bucket_name"] = config.Reference{
			TerraformName: "minio_s3_bucket",
		}
	})

	p.AddResourceConfigurator("minio_s3_bucket_versioning", func(r *config.Resource) {
		r.ShortGroup = "s3"
		r.Kind = "BucketVersioning"
		r.References["bucket"] = config.Reference{
			TerraformName: "minio_s3_bucket",
		}
	})

	p.AddResourceConfigurator("minio_s3_bucket_notification", func(r *config.Resource) {
		r.ShortGroup = "s3"
		r.Kind = "BucketNotification"
		r.References["bucket"] = config.Reference{
			TerraformName: "minio_s3_bucket",
		}
	})
}