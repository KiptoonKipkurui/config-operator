package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	configv1alpha1 "github.com/kiptoonkipkurui/config-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

/*
The first step to writing a simple integration test is to actually create an instance of ConfigOp you can run tests against.
Note that to create a ConfigOp, you’ll need to create a stub ConfigOp struct that contains your CronJob’s specifications.

Note that when we create a stub ConfigOp, the ConfigOp also needs stubs of its required downstream objects.
Without the stubbed Job template spec and the Pod template spec below, the Kubernetes API will not be able to
create the ConfigOp.
*/
var _ = Describe("ConfigOp controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		ConfigOpName      = "test-configop"
		ConfigOpNamespace = "default"
		JobName           = "test-job"

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

	Context("When creating ConfigOp Status", func() {
		It("Should Create ManagedConfigMap and ManagedSecret", func() {
			By("By creating a new ConfigOp")
			ctx := context.Background()
			configOp := &configv1alpha1.ConfigOp{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1apha1",
					Kind:       "ConfigOp",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ConfigOpName,
					Namespace: ConfigOpNamespace,
				},
				Spec: configv1alpha1.ConfigOpSpec{
					Namespaces: []string{"default", "test"},
					ConfigMaps: []configv1alpha1.ManagedConfigMap{
						{
							Name: "test-configmap",
							Data: map[string]string{
								"name": "test",
							},
						},
					},
					Secrets: []configv1alpha1.ManagedSecret{
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
			createdConfigOp := &configv1alpha1.ConfigOp{}

			// We'll need to retry getting this newly created ConfigOp, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, configOpLookupKey, createdConfigOp)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our the namespaces in our spec have been created.
			Expect(createdConfigOp.Spec.Namespaces).Should(Equal([]string{"default", "test"}))
			Expect(len(createdConfigOp.Spec.ConfigMaps)).Should(Equal(1))
			Expect(len(createdConfigOp.Spec.Secrets)).Should(Equal(1))

			// ensure that the config map has been created
			configmapKey := types.NamespacedName{Name: "test-configmap", Namespace: "test"}
			createdConfigMap := &v1.ConfigMap{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, configmapKey, createdConfigMap)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdConfigMap).ToNot(BeNil())

			// ensure that the secret example-secret has been created
			secretKey := types.NamespacedName{Name: "example-secret", Namespace: "test"}
			createdSecret := &v1.Secret{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretKey, createdSecret)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdSecret).ToNot(BeNil())

			// ensure that the test namespace has been created
			ns := &v1.Namespace{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: "test"}, ns)
				return err == nil
			})

			Expect(ns).ToNot(BeNil())

			//update spec
			// Update
			updated := &configv1alpha1.ConfigOp{}
			Expect(k8sClient.Get(context.Background(), configOpLookupKey, updated)).Should(Succeed())

			updated.Spec.Namespaces = append(updated.Spec.Namespaces, "config-ns")
			Expect(k8sClient.Update(context.Background(), updated)).Should(Succeed())

			// ensure new namespace has been created
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: "config-ns"}, ns)
				return err == nil
			})
			// ensure that the config map has been created
			configmapKey = types.NamespacedName{Name: "test-configmap", Namespace: "config-ns"}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, configmapKey, createdConfigMap)
				return err == nil
			}, timeout, interval).Should(BeTrue())

		})
	})
})

/*
	After writing all this code, you can run `go test ./...` in your `controllers/` directory again to run your new test!
*/
