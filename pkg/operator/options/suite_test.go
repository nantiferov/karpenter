/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package options_test

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	. "knative.dev/pkg/logging/testing"

	coreoptions "sigs.k8s.io/karpenter/pkg/operator/options"

	"github.com/aws/karpenter/pkg/operator/options"
	"github.com/aws/karpenter/pkg/test"
)

var ctx context.Context

func TestAPIs(t *testing.T) {
	ctx = TestContextWithLogger(t)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Options")
}

var _ = Describe("Options", func() {
	var fs *coreoptions.FlagSet
	var opts *options.Options

	BeforeEach(func() {
		fs = &coreoptions.FlagSet{
			FlagSet: flag.NewFlagSet("karpenter", flag.ContinueOnError),
		}
		opts = &options.Options{}
	})
	AfterEach(func() {
		os.Clearenv()
	})

	It("should correctly override default vars when CLI flags are set", func() {
		opts.AddFlags(fs)
		err := opts.Parse(fs,
			"--assume-role-arn", "env-role",
			"--assume-role-duration", "20m",
			"--cluster-ca-bundle", "env-bundle",
			"--cluster-name", "env-cluster",
			"--cluster-endpoint", "https://env-cluster",
			"--isolated-vpc",
			"--vm-memory-overhead-percent", "0.1",
			"--interruption-queue", "env-cluster",
			"--reserved-enis", "10")
		Expect(err).ToNot(HaveOccurred())
		expectOptionsEqual(opts, test.Options(test.OptionsFields{
			AssumeRoleARN:           lo.ToPtr("env-role"),
			AssumeRoleDuration:      lo.ToPtr(20 * time.Minute),
			ClusterCABundle:         lo.ToPtr("env-bundle"),
			ClusterName:             lo.ToPtr("env-cluster"),
			ClusterEndpoint:         lo.ToPtr("https://env-cluster"),
			IsolatedVPC:             lo.ToPtr(true),
			VMMemoryOverheadPercent: lo.ToPtr[float64](0.1),
			InterruptionQueue:       lo.ToPtr("env-cluster"),
			ReservedENIs:            lo.ToPtr(10),
		}))
	})
	It("should correctly fallback to env vars when CLI flags aren't set", func() {
		os.Setenv("ASSUME_ROLE_ARN", "env-role")
		os.Setenv("ASSUME_ROLE_DURATION", "20m")
		os.Setenv("CLUSTER_CA_BUNDLE", "env-bundle")
		os.Setenv("CLUSTER_NAME", "env-cluster")
		os.Setenv("CLUSTER_ENDPOINT", "https://env-cluster")
		os.Setenv("ISOLATED_VPC", "true")
		os.Setenv("VM_MEMORY_OVERHEAD_PERCENT", "0.1")
		os.Setenv("INTERRUPTION_QUEUE", "env-cluster")
		os.Setenv("RESERVED_ENIS", "10")

		// Add flags after we set the environment variables so that the parsing logic correctly refers
		// to the new environment variable values
		opts.AddFlags(fs)
		err := opts.Parse(fs)
		Expect(err).ToNot(HaveOccurred())
		expectOptionsEqual(opts, test.Options(test.OptionsFields{
			AssumeRoleARN:           lo.ToPtr("env-role"),
			AssumeRoleDuration:      lo.ToPtr(20 * time.Minute),
			ClusterCABundle:         lo.ToPtr("env-bundle"),
			ClusterName:             lo.ToPtr("env-cluster"),
			ClusterEndpoint:         lo.ToPtr("https://env-cluster"),
			IsolatedVPC:             lo.ToPtr(true),
			VMMemoryOverheadPercent: lo.ToPtr[float64](0.1),
			InterruptionQueue:       lo.ToPtr("env-cluster"),
			ReservedENIs:            lo.ToPtr(10),
		}))
	})

	Context("Validation", func() {
		BeforeEach(func() {
			opts.AddFlags(fs)
		})
		It("should fail when cluster name is not set", func() {
			err := opts.Parse(fs)
			Expect(err).To(HaveOccurred())
		})
		It("should fail when assume role duration is less than 15 minutes", func() {
			err := opts.Parse(fs, "--cluster-name", "test-cluster", "--assume-role-duration", "1s")
			Expect(err).To(HaveOccurred())
		})
		It("should fail when clusterEndpoint is invalid (not absolute)", func() {
			err := opts.Parse(fs, "--cluster-name", "test-cluster", "--cluster-endpoint", "00000000000000000000000.gr7.us-west-2.eks.amazonaws.com")
			Expect(err).To(HaveOccurred())
		})
		It("should fail when vmMemoryOverheadPercent is negative", func() {
			err := opts.Parse(fs, "--cluster-name", "test-cluster", "--vm-memory-overhead-percent", "-0.01")
			Expect(err).To(HaveOccurred())
		})
		It("should fail when reservedENIs is negative", func() {
			err := opts.Parse(fs, "--cluster-name", "test-cluster", "--reserved-enis", "-1")
			Expect(err).To(HaveOccurred())
		})
	})
})

func expectOptionsEqual(optsA *options.Options, optsB *options.Options) {
	GinkgoHelper()
	Expect(optsA.AssumeRoleARN).To(Equal(optsB.AssumeRoleARN))
	Expect(optsA.AssumeRoleDuration).To(Equal(optsB.AssumeRoleDuration))
	Expect(optsA.ClusterCABundle).To(Equal(optsB.ClusterCABundle))
	Expect(optsA.ClusterName).To(Equal(optsB.ClusterName))
	Expect(optsA.ClusterEndpoint).To(Equal(optsB.ClusterEndpoint))
	Expect(optsA.IsolatedVPC).To(Equal(optsB.IsolatedVPC))
	Expect(optsA.VMMemoryOverheadPercent).To(Equal(optsB.VMMemoryOverheadPercent))
	Expect(optsA.InterruptionQueue).To(Equal(optsB.InterruptionQueue))
	Expect(optsA.ReservedENIs).To(Equal(optsB.ReservedENIs))
}
