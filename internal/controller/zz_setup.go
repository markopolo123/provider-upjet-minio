// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/pkg/controller"

	group "github.com/markopolo123/provider-upjet-minio/internal/controller/iam/group"
	policy "github.com/markopolo123/provider-upjet-minio/internal/controller/iam/policy"
	serviceaccount "github.com/markopolo123/provider-upjet-minio/internal/controller/iam/serviceaccount"
	user "github.com/markopolo123/provider-upjet-minio/internal/controller/iam/user"
	key "github.com/markopolo123/provider-upjet-minio/internal/controller/kms/key"
	providerconfig "github.com/markopolo123/provider-upjet-minio/internal/controller/providerconfig"
	bucket "github.com/markopolo123/provider-upjet-minio/internal/controller/s3/bucket"
	bucketnotification "github.com/markopolo123/provider-upjet-minio/internal/controller/s3/bucketnotification"
	bucketpolicy "github.com/markopolo123/provider-upjet-minio/internal/controller/s3/bucketpolicy"
	bucketversioning "github.com/markopolo123/provider-upjet-minio/internal/controller/s3/bucketversioning"
	object "github.com/markopolo123/provider-upjet-minio/internal/controller/s3/object"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		group.Setup,
		policy.Setup,
		serviceaccount.Setup,
		user.Setup,
		key.Setup,
		providerconfig.Setup,
		bucket.Setup,
		bucketnotification.Setup,
		bucketpolicy.Setup,
		bucketversioning.Setup,
		object.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
