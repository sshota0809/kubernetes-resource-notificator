package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/argoproj/notifications-engine/pkg/api"
	"github.com/argoproj/notifications-engine/pkg/controller"
	"github.com/argoproj/notifications-engine/pkg/services"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type controllerOptions struct {
	clientConfig clientcmd.ClientConfig
	configMap string
	secret string
	apiGroup string
	apiVersion string
	apiResource string
}

type command struct {
	cmd cobra.Command
}

var co controllerOptions

func (c *command) setK8SFlagsToCmd() {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	kflags := clientcmd.RecommendedConfigOverrideFlags("")
	c.cmd.PersistentFlags().StringVar(&loadingRules.ExplicitPath, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster")
	clientcmd.BindOverrideFlags(&overrides, c.cmd.PersistentFlags(), kflags)
	co.clientConfig = clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
}

func (c *command) setControllerConfigToCmd() {
	c.cmd.Flags().StringVarP(&co.configMap, "configmap", "c", "kubernetes-notificator-cm", "ConfigMap name for controller")
	c.cmd.Flags().StringVarP(&co.secret, "secret", "s", "kubernetes-notificator-secret", "Secret name for controller")
	c.cmd.Flags().StringVarP(&co.apiGroup, "group", "g", "", "apiGroup monitored by controller")
	c.cmd.Flags().StringVarP(&co.apiVersion, "version", "v", "v1", "apiVersion monitored by controller")
	c.cmd.Flags().StringVarP(&co.apiResource, "resource", "r", "pods", "apiResource monitored by controller")
}

func (c *command) Execute() {
	if err := c.cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func NewCommand() *command {
	cmd := command {
		cobra.Command{
			Use: "controller",
			Run: func(c *cobra.Command, args []string) {
				// Get Kubernetes REST Config and current Namespace so we can talk to Kubernetes
				restConfig, err := co.clientConfig.ClientConfig()
				if err != nil {
					log.Fatalf("Failed to get Kubernetes config")
				}
				namespace, _, err := co.clientConfig.Namespace()
				if err != nil {
					log.Fatalf("Failed to get namespace from Kubernetes config")
				}

				// Create ConfigMap and Secret informer to access notifications configuration
				informersFactory := informers.NewSharedInformerFactoryWithOptions(
					kubernetes.NewForConfigOrDie(restConfig),
					time.Minute,
					informers.WithNamespace(namespace))
				secrets := informersFactory.Core().V1().Secrets().Informer()
				configMaps := informersFactory.Core().V1().ConfigMaps().Informer()

				// Create "Notifications" API factory that handles notifications processing
				notificationsFactory := api.NewFactory(api.Settings{
					ConfigMapName: co.configMap,
					SecretName:    co.secret,
					InitGetVars: func(cfg *api.Config, configMap *v1.ConfigMap, secret *v1.Secret) (api.GetVars, error) {
						return func(obj map[string]interface{}, dest services.Destination) map[string]interface{} {
							return map[string]interface{}{"pod": obj}
						}, nil
					},
				}, namespace, secrets, configMaps)

				// Create notifications controller that handles Kubernetes resources processing
				notificationClient := dynamic.NewForConfigOrDie(restConfig).Resource(schema.GroupVersionResource{
					Group:    co.apiGroup,
					Version:  co.apiVersion,
					Resource: co.apiResource,
				})
				notificationsInformer := cache.NewSharedIndexInformer(&cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						return notificationClient.List(context.Background(), options)
					},
					WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
						return notificationClient.Watch(context.Background(), metav1.ListOptions{})
					},
				}, &unstructured.Unstructured{}, time.Minute, cache.Indexers{})
				ctrl := controller.NewController(notificationClient, notificationsInformer, notificationsFactory)

				// Start informers and controller
				go informersFactory.Start(context.Background().Done())
				go notificationsInformer.Run(context.Background().Done())
				if !cache.WaitForCacheSync(context.Background().Done(), secrets.HasSynced, configMaps.HasSynced, notificationsInformer.HasSynced) {
					log.Fatalf("Failed to synchronize informers")
				}

				ctrl.Run(10, context.Background().Done())
			},
		},
	}
	cmd.setK8SFlagsToCmd()
	cmd.setControllerConfigToCmd()
	return &cmd
}
