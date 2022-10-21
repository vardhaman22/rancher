package rke2

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/rancher/rancher/pkg/api/scheme"
	kubeProvisioning "github.com/rancher/rancher/tests/framework/clients/provisioning"
	"github.com/rancher/rancher/tests/framework/clients/rancher"
	provisioningV1 "github.com/rancher/rancher/tests/framework/clients/rancher/generated/provisioning/v1"
	"github.com/rancher/rancher/tests/framework/extensions/clusters"
	"github.com/rancher/rancher/tests/framework/extensions/ingresses"
	"github.com/rancher/rancher/tests/framework/extensions/machinepools"
	"github.com/rancher/rancher/tests/framework/extensions/workloads/deployments"

	"github.com/rancher/rancher/tests/framework/pkg/config"
	"github.com/rancher/rancher/tests/framework/pkg/session"
	"github.com/rancher/rancher/tests/framework/pkg/wait"
	"github.com/rancher/rancher/tests/integration/pkg/defaults"
	provisioning "github.com/rancher/rancher/tests/v2/validation/provisioning"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	rkev1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	appv1 "k8s.io/api/apps/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

const (
	defaultNamespace = "default"
)

type RKE2EtcdSnapshotRestoreTestSuite struct {
	suite.Suite
	session            *session.Session
	client             *rancher.Client
	clusterName        string
	namespace          string
	kubernetesVersions []string
	cnis               []string
	providers          []string
	nodesAndRoles      []machinepools.NodeRoles
}

func (p *RKE2EtcdSnapshotRestoreTestSuite) TearDownSuite() {
	p.session.Cleanup()
}

var EtcdSnapshotGroupVersionResource = schema.GroupVersionResource{
	Group:    "rke.cattle.io",
	Version:  "v1",
	Resource: "etcdsnapshots",
}

func (r *RKE2EtcdSnapshotRestoreTestSuite) SetupSuite() {
	testSession := session.NewSession(r.T())
	r.session = testSession

	clustersConfig := new(provisioning.Config)
	config.LoadConfig(provisioning.ConfigurationFileKey, clustersConfig)

	r.kubernetesVersions = clustersConfig.KubernetesVersions
	r.cnis = clustersConfig.CNIs
	r.providers = clustersConfig.Providers
	r.nodesAndRoles = clustersConfig.NodesAndRoles

	client, err := rancher.NewClient("", testSession)
	require.NoError(r.T(), err)

	r.client = client

	r.clusterName = r.client.RancherConfig.ClusterName
}

func (r *RKE2EtcdSnapshotRestoreTestSuite) createSnapshot(clustername string, generation int) error {
	kubeProvisioningClient, err := r.client.GetKubeAPIProvisioningClient()
	require.NoError(r.T(), err)

	cluster, err := kubeProvisioningClient.Clusters(defaultNamespace).Get(context.TODO(), clustername, metav1.GetOptions{})
	if err != nil {
		return err
	}

	cluster.Spec.RKEConfig.ETCDSnapshotCreate = &rkev1.ETCDSnapshotCreate{
		Generation: generation,
	}

	cluster, err = kubeProvisioningClient.Clusters(defaultNamespace).Update(context.TODO(), cluster, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (r *RKE2EtcdSnapshotRestoreTestSuite) restoreSnapshot(kubeProvisioningClient *kubeProvisioning.Client, clustername string, name string, generation int, restoreconfig string) error {
	kubeProvisioningClient, err := r.client.GetKubeAPIProvisioningClient()
	require.NoError(r.T(), err)

	cluster, err := kubeProvisioningClient.Clusters(defaultNamespace).Get(context.TODO(), clustername, metav1.GetOptions{})
	if err != nil {
		return err
	}
	cluster.Spec.RKEConfig.ETCDSnapshotRestore = &rkev1.ETCDSnapshotRestore{
		Name:             name,
		Generation:       generation,
		RestoreRKEConfig: restoreconfig,
	}

	cluster, err = kubeProvisioningClient.Clusters(defaultNamespace).Update(context.TODO(), cluster, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (r *RKE2EtcdSnapshotRestoreTestSuite) GetSnapshots(client *rancher.Client,
	localClusterID string, getOpts metav1.ListOptions) (*rkev1.ETCDSnapshotList, error) {
	dynamicClient, err := client.GetDownStreamClusterClient(localClusterID)
	if err != nil {
		return nil, err
	}

	etcdResource := dynamicClient.Resource(EtcdSnapshotGroupVersionResource).Namespace(defaultNamespace)
	unstructuredResp, err := etcdResource.List(context.TODO(), getOpts)
	if err != nil {
		return nil, err
	}

	snapshots := &rkev1.ETCDSnapshotList{}
	err = scheme.Scheme.Convert(unstructuredResp, snapshots, unstructuredResp.GroupVersionKind())
	if err != nil {
		return nil, err
	}
	return snapshots, nil
}

// func (r *RKE2EtcdSnapshotRestoreTestSuite) TestEtcdSnapshotRestore() {
// 	for _, providerName := range r.providers {
// 		provider := CreateProvider(providerName)
// 		r.ProvisioningRKE2Cluster(provider)
// 	}

// 	for _, providerName := range r.providers {
// 		subSession := r.session.NewSession()

// 		provider := CreateProvider(providerName)

// 		client, err := r.client.WithSession(subSession)
// 		require.NoError(r.T(), err)

// 		cloudCredential, err := provider.CloudCredFunc(client)
// 		require.NoError(r.T(), err)

// 		for _, kubernetesVersion := range r.kubernetesVersions {
// 			for _, cni := range r.cnis {
// 				r.TestEtcdSnapshotRestoreFreshCluster(provider, kubernetesVersion, cni, r.nodesAndRoles, cloudCredential)
// 			}
// 		}

// 		subSession.Cleanup()
// 	}
// }

func TestEtcdSnapshotRestore(t *testing.T) {
	suite.Run(t, new(RKE2EtcdSnapshotRestoreTestSuite))
}

func (r *RKE2EtcdSnapshotRestoreTestSuite) EtcdSnapshotRestore(provider *Provider) {
	logrus.Infof("running etcd snapshot restore test.............")
	subSession := r.session.NewSession()
	defer subSession.Cleanup()

	client, err := r.client.WithSession(subSession)
	require.NoError(r.T(), err)

	clusterName := provisioning.AppendRandomString(provider.Name)
	localClusterName := "local"

	logrus.Infof("creating rke2Cluster.............")
	clusterResp, err := r.createRKE2NodeDriverCluster(subSession, client, provider, clusterName)
	require.NoError(r.T(), err)
	require.Equal(r.T(), clusterName, clusterResp.ObjectMeta.Name)
	logrus.Infof("rke2Cluster create request successful.............")

	logrus.Infof("creating kube provisioning client.............")
	kubeProvisioningClient, err := r.client.GetKubeAPIProvisioningClient()
	require.NoError(r.T(), err)
	logrus.Infof("kube provisioning client created.............")

	logrus.Infof("creating watch over cluster.............")
	r.watchAndWaitForCluster(kubeProvisioningClient, clusterName)
	logrus.Infof("cluster is up and running.............")

	// Get clusterID by clusterName
	logrus.Info("getting cluster id.............")
	clusterID, err := clusters.GetClusterIDByName(client, clusterName)
	require.NoError(r.T(), err)
	logrus.Info("got cluster id.............", clusterID)

	logrus.Info("getting local cluster id.............")
	localClusterID, err := clusters.GetClusterIDByName(client, localClusterName)
	require.NoError(r.T(), err)
	logrus.Info("got local cluster id.............", localClusterID)

	// creating the workload W1
	logrus.Infof("creating a workload(nginx deployment).............")
	w1Name := "w1"
	w1, err := r.createTestDeployment(client, clusterID, w1Name)
	require.NoError(r.T(), err)
	require.Equal(r.T(), w1Name, w1.ObjectMeta.Name)
	logrus.Infof("created a workload(nginx deployment).............")

	logrus.Infof("creating watch over w1(nginx-deployment).............")
	r.watchAndWaitForNginxDeploymentW1(client, clusterID, w1Name)
	logrus.Infof("w1(nginx-deployment) is ready.............")

	// creating the ingress1
	logrus.Infof("creating an ingress.............")
	ingress1Name := "ingress1"
	w1ServiceName := "w1-svc"
	ingress1, err := r.createIngress(client, clusterID, ingress1Name, w1ServiceName)
	require.NoError(r.T(), err)
	require.Equal(r.T(), ingress1Name, ingress1.ObjectMeta.Name)
	logrus.Infof("created an ingress.............")

	logrus.Infof("creating a snapshot of the cluster.............")
	err = r.createSnapshot(clusterName, 1)
	require.NoError(r.T(), err)
	logrus.Infof("created a snapshot of the cluster.............")

	logrus.Infof("creating watch over cluster after creating a snapshot.............")
	r.watchAndWaitForCluster(kubeProvisioningClient, clusterName)
	logrus.Infof("cluster is active again.............")

	snapshotList, err := r.GetSnapshots(client, localClusterID, metav1.ListOptions{})
	require.NoError(r.T(), err)

	snapshotToBeRestored := ""
	totalClusterSnapShots := 0
	for _, snapshot := range snapshotList.Items {
		if strings.Contains(snapshot.ObjectMeta.Name, clusterName) {
			if snapshotToBeRestored == "" {
				snapshotToBeRestored = snapshot.Name
			}
			totalClusterSnapShots++
		}
	}
	logrus.Info("total ", totalClusterSnapShots, " created")

	logrus.Infof("creating a workload(w2, deployment).............")
	w2Name := "w2"
	w2, err := r.createTestDeployment(client, clusterID, w2Name)
	require.NoError(r.T(), err)
	require.Equal(r.T(), w2Name, w2.ObjectMeta.Name)
	logrus.Infof("created a workload(w2, deployment).............")

	logrus.Infof("creating watch over w2.............")
	r.watchAndWaitForNginxDeploymentW1(client, clusterID, w2Name)
	logrus.Infof("w2 is ready.............")

	logrus.Infof("restoring snapshot.............")
	require.NoError(r.T(), r.restoreSnapshot(kubeProvisioningClient, clusterName, snapshotToBeRestored, 1, "all"))
	logrus.Infof("successfully submitted restoration request.............")

	logrus.Infof("creating watch over cluster after restore.............")
	r.watchAndWaitForCluster(kubeProvisioningClient, clusterName)
	logrus.Infof("cluster is active again.............")

	logrus.Infof("fetching deployment list to validate restore.............")
	deploymentsList, err := deployments.ListDeployments(client, clusterID, defaultNamespace, metav1.ListOptions{})
	require.NoError(r.T(), err)
	require.Equal(r.T(), 1, len(deploymentsList.Items))
	require.Equal(r.T(), w1Name, deploymentsList.Items[0].ObjectMeta.Name)
	logrus.Infof(" deployment list validated successfully.............")

	logrus.Infof("fetching ingresses list to validate restore.............")
	ingressesList, err := ingresses.ListIngresses(client, clusterID, defaultNamespace, metav1.ListOptions{})
	require.NoError(r.T(), err)
	require.Equal(r.T(), 1, len(ingressesList.Items))
	require.Equal(r.T(), ingress1Name, ingressesList.Items[0].ObjectMeta.Name)
	logrus.Infof("ingress list validated successfully.............")
}

func (r *RKE2EtcdSnapshotRestoreTestSuite) createRKE2NodeDriverCluster(session *session.Session,
	client *rancher.Client, provider *Provider, clusterName string) (*provisioningV1.Cluster, error) {

	nodeRoles := []machinepools.NodeRoles{
		{
			ControlPlane: true,
			Etcd:         true,
			Worker:       true,
			Quantity:     1,
		},
		{
			ControlPlane: true,
			Etcd:         true,
			Worker:       true,
			Quantity:     1,
		},
		{
			ControlPlane: false,
			Etcd:         true,
			Worker:       true,
			Quantity:     1,
		},
	}

	cloudCredential, err := provider.CloudCredFunc(client)
	require.NoError(r.T(), err)

	generatedPoolName := fmt.Sprintf("nc-%s-pool1-", clusterName)
	machinePoolConfig := provider.MachinePoolFunc(generatedPoolName, defaultNamespace)

	machineConfigResp, err := machinepools.CreateMachineConfig(provider.MachineConfig, machinePoolConfig, client)
	require.NoError(r.T(), err)

	machinePools := machinepools.MachinePoolSetup(nodeRoles, machineConfigResp)

	r.cnis = []string{"calico"}

	initialKubeVersion := "v1.24.2+rke2r1"

	cluster := clusters.NewK3SRKE2ClusterConfig(clusterName, defaultNamespace, r.cnis[0], cloudCredential.ID, initialKubeVersion, machinePools)

	return clusters.CreateK3SRKE2Cluster(client, cluster)
}

func (r *RKE2EtcdSnapshotRestoreTestSuite) watchAndWaitForCluster(kubeProvisioningClient *kubeProvisioning.Client, clusterName string) {
	result, err := kubeProvisioningClient.Clusters(defaultNamespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + clusterName,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	})
	require.NoError(r.T(), err)

	logrus.Infof("waiting for cluster to be up.............")
	checkFunc := clusters.IsProvisioningClusterReady
	err = wait.WatchWait(result, checkFunc)
	require.NoError(r.T(), err)
}

func (r *RKE2EtcdSnapshotRestoreTestSuite) createTestDeployment(
	client *rancher.Client, clusterID string, deploymentName string) (*appv1.Deployment, error) {
	podTemplateSpec := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app": "nginx",
			},
			Namespace: defaultNamespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "nginx",
					Image: "nginx",
					Ports: []v1.ContainerPort{
						{
							ContainerPort: 80,
						},
					},
				},
			},
		},
	}

	return deployments.CreateDeployment(client, clusterID, deploymentName, defaultNamespace, podTemplateSpec)
}

func (r *RKE2EtcdSnapshotRestoreTestSuite) watchAndWaitForNginxDeploymentW1(client *rancher.Client, clusterID string, deploymentName string) {
	deploymentResource, err := deployments.GetDeploymentResource(client, clusterID, defaultNamespace)
	require.NoError(r.T(), err)

	deploymentResult, err := deploymentResource.Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + deploymentName,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	})
	require.NoError(r.T(), err)

	logrus.Infof("waiting for deployment to be created.............")
	deploymentCheckFunc := deployments.IsDeploymentReady
	err = wait.WatchWait(deploymentResult, deploymentCheckFunc)
	assert.NoError(r.T(), err)
}

func (r *RKE2EtcdSnapshotRestoreTestSuite) createIngress(
	client *rancher.Client, clusterID string, ingressName string, serviceName string) (*networkingv1.Ingress, error) {
	exactPath := networkingv1.PathTypeExact
	ingressSpec := &networkingv1.IngressSpec{
		Rules: []networkingv1.IngressRule{
			{
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/index.html",
								PathType: &exactPath,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: serviceName,
										Port: networkingv1.ServiceBackendPort{
											Number: 80,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return ingresses.CreateIngress(client, clusterID, ingressName, defaultNamespace, ingressSpec)
}

func (r *RKE2EtcdSnapshotRestoreTestSuite) TestEtcdSnapshotRestore() {
	for _, providerName := range r.providers {
		provider := CreateProvider(providerName)
		r.EtcdSnapshotRestore(&provider)
	}
}
