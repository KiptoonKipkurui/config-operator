package v1alpha1

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Webhook test", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		ConfigOpName      = "test-configop"
		ConfigOpNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("When Validating Created Object", func() {
		It("Should Create ManagedConfigMap and ManagedSecret", func() {
			By("By creating a new ConfigOp")
			ctx := context.Background()
			configOp := &ConfigOp{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1apha1",
					Kind:       "ConfigOp",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ConfigOpName,
					Namespace: ConfigOpNamespace,
				},
				Spec: ConfigOpSpec{
					Namespaces: []string{"default", "test"},
					ConfigMaps: []ManagedConfigMap{
						{
							Name: "test-configmap",
							Data: map[string]string{
								"name": "test",
							},
						},
					},
					Secrets: []ManagedSecret{
						{
							Name: "example-secret",
							Data: map[string][]byte{
								"username": []byte("YWRtaW4="),
							},
							StringData: map[string]string{
								"password": "MWYyZDFlMmU2N2Rm",
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, configOp)).Should(Succeed())

			/*
				After creating this ConfigOp, let's check that the ConfigOp's Spec fields match what we passed in.
				Note that, because the k8s apiserver may not have finished creating a ConfigOp after our `Create()` call from earlier, we will use Gomega’s Eventually() testing function instead of Expect() to give the apiserver an opportunity to finish creating our ConfigOp.

				`Eventually()` will repeatedly run the function provided as an argument every interval seconds until
				(a) the function’s output matches what’s expected in the subsequent `Should()` call, or
				(b) the number of attempts * interval period exceed the provided timeout value.

				In the examples below, timeout and interval are Go Duration values of our choosing.
			*/

			configOpLookupKey := types.NamespacedName{Name: ConfigOpName, Namespace: ConfigOpNamespace}
			createdConfigOp := &ConfigOp{}

			// We'll need to retry getting this newly created ConfigOp, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, configOpLookupKey, createdConfigOp)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			data := make(map[string]string)
			data["app.kubernetes.io/name"] = "configOp"
			data["app.kubernetes.io/instance"] = "configOp-instance"
			data["app.kubernetes.io/version"] = "v1alpha1"
			data["app.kubernetes.io/component"] = "configuration"
			data["app.kubernetes.io/part-of"] = "configop-operator"
			data["app.kubernetes.io/managed-by"] = "configop-operator"
			data["addonmanager.kubernetes.io/mode"] = "Reconcile"

			Expect(data).To(Equal(configOp.Labels))
		})
	})
})
