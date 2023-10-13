package networkchecks

import (
	"fmt"
	"os"
	"strings"

	"github.com/rancher/rancher/tests/framework/clients/rancher"
	management "github.com/rancher/rancher/tests/framework/clients/rancher/generated/management/v3"
	"github.com/rancher/rancher/tests/framework/extensions/users"
	"github.com/rancher/rancher/tests/framework/pkg/config"
	"github.com/sirupsen/logrus"
)

const scirptsPathPrefix = "/../../../validation/tests/v3_api/scripts/"
const networkChecksPyTestOption = `-k "test_wl or test_connectivity or test_ingress or test_service_discovery or test_websocket"`
const jobNameParam = "JOB_NAME"
const buildNumParam = "BUILD_NUMBER"
const clusterNameParam = "RANCHER_CLUSTER_NAMES"
const rancherURLParam = "CATTLE_TEST_URL"
const adminTokenParam = "ADMIN_TOKEN"
const userTokenParam = "USER_TOKEN"
const testContainerParam = "TEST_CONTAINER"
const pytestOptionsParam = "PYTEST_OPTIONS"
const testDirParam = "TESTS_DIR"
const rootPathParam = "ROOT_PATH"

const networkChecksConfigKey = "networkChecks"

var fileContent = `
RANCHER_CLUSTER_NAMES=%s
CATTLE_TEST_URL=%s
ADMIN_TOKEN=%s
USER_TOKEN=%s
PYTEST_OPTIONS="-k \"test_wl or test_connectivity or test_ingress or test_service_discovery or test_websocket\""
`

type NetworkChecksConfig struct {
	UserToken  string `json:"userToken" yaml:"userToken"`
	AdminToken string `json:"adminToken" yaml:"adminToken"`
	RancherURL string `json:"rancherURL" yaml:"rancherURL"`
	Username   string `json:"username" yaml:"username"`
}

type NetworkChecks struct {
	config       *NetworkChecksConfig
	client       *rancher.Client
	clusterNames string
}

func (nc *NetworkChecks) InitNetChecks(client *rancher.Client) {
	nc.config = new(NetworkChecksConfig)
	config.LoadConfig(networkChecksConfigKey, nc.config)

	nc.config.AdminToken = client.RancherConfig.AdminToken
	nc.config.RancherURL = "https://" + client.RancherConfig.Host
	nc.client = client

	if nc.config.UserToken == "" {
		nc.config.UserToken = client.RancherConfig.AdminToken
		nc.config.Username = "admin"
	}
}

func (nc *NetworkChecks) RunNetworkChecks(clusterName, clusterId string) error {

	formatter := &logrus.TextFormatter{}
	formatter.DisableQuote = true
	logrus.SetFormatter(formatter)

	err := nc.granClusterOwnerAccessToUser(clusterId)
	if err != nil {
		return err
	}
	if nc.clusterNames == "" {
		nc.clusterNames = clusterName
	} else {
		nc.clusterNames = nc.clusterNames + "," + clusterName
	}

	return nil
}

func (nc *NetworkChecks) SetNetworkChecksEnv() {
	nc.clusterNames = strings.TrimSuffix(nc.clusterNames, ",")
	nc.setEnv()
}

// func (nc *NetworkChecks) RunNetworkChecks2(clusterName, clusterId string) error {

// 	formatter := &logrus.TextFormatter{}
// 	formatter.DisableQuote = true
// 	logrus.SetFormatter(formatter)

// 	nc.granClusterOwnerAccessToUser(clusterId)

// 	nc.setEnv(clusterName)

// 	_, filename, _, _ := runtime.Caller(0)
// 	dir := path.Dir(filename)
// 	configScriptPath := path.Join(dir, scirptsPathPrefix+"configure.sh")
// 	buildScriptPath := path.Join(dir, scirptsPathPrefix+"build.sh")
// 	runScriptPath := path.Join(dir, scirptsPathPrefix+"run.sh")
// 	stopScriptPath := path.Join(dir, scirptsPathPrefix+"stop.sh")

// 	logrus.Info("setting up configuration to run network checks....")
// 	out, err := exec.Command(configScriptPath).Output()
// 	if err != nil {
// 		errStr := fmt.Sprintf("error running configuration script, err:%v, scriptoutput: %v", err, string(out))
// 		return errors.New(errStr)
// 	}
// 	logrus.Infof("configuration script output....\n:%v", string(out))

// 	logrus.Info("building docker image to run network checks....")
// 	out, err = exec.Command(buildScriptPath).Output()
// 	if err != nil {
// 		errStr := fmt.Sprintf("error running biuld script, err:%v, scriptoutput: %v", err, string(out))
// 		return errors.New(errStr)
// 	}
// 	logrus.Infof("build script output....\n:%v", string(out))

// 	errStr := ""

// 	logrus.Info("running network checks tests inside docker contianer....")
// 	out, err = exec.Command(runScriptPath).Output()
// 	if err != nil {
// 		errStr = errStr + fmt.Sprintf("error during network checks tests execution:%v", err.Error())
// 		logrus.Infof("error running run.sh script\n err:%v \nscriptoutput: %v", err, string(out))
// 	} else {
// 		logrus.Infof("run.sh script output....\n:%v", string(out))
// 	}

// 	logrus.Info("cleaning up container and network checks docker image....")
// 	out, err = exec.Command(stopScriptPath).Output()
// 	if err != nil {
// 		errStr = errStr + fmt.Sprintf("error during docker image and contianer cleanup:%v, scriptout:%v", err.Error(), string(out))
// 	}
// 	if errStr != "" {
// 		return errors.New(errStr)
// 	}

// 	logrus.Infof("network checks successful for cluster %s (%s)!!!", clusterName, clusterId)

// 	return nil
// }

func (nc *NetworkChecks) granClusterOwnerAccessToUser(clusterID string) error {

	userId, err := users.GetUserIDByName(nc.client, nc.config.Username)
	if err != nil {
		return err
	}

	_, err = nc.client.Management.ClusterRoleTemplateBinding.Create(&management.ClusterRoleTemplateBinding{
		Name:            "cluster-role-template-binding-1",
		ClusterID:       clusterID,
		RoleTemplateID:  "cluster-owner",
		UserPrincipalID: fmt.Sprintf("%s://%s", "local", userId),
	})

	return err
}

func (nc *NetworkChecks) setEnv() {
	content := fmt.Sprintf(fileContent,
		nc.clusterNames, nc.config.RancherURL, nc.config.AdminToken, nc.config.UserToken)

	file, errs := os.Create("myfile.txt")
	if errs != nil {
		fmt.Println("Failed to create file:", errs)
		return
	}
	defer file.Close()

	// Write the string "Hello, World!" to the file
	_, errs = file.WriteString(content)
	if errs != nil {
		fmt.Println("Failed to write to file:", errs) //print the failed message
		return
	}
}

// func (nc *NetworkChecks) setEnv() {

// 	jobName := os.Getenv(jobNameParam)
// 	if jobName == "" {
// 		jobName = "provisioning-network-checks"
// 		os.Setenv(jobNameParam, jobName)
// 	}

// 	buildNum := os.Getenv(buildNumParam)
// 	if buildNum == "" {
// 		buildNum = fmt.Sprintf("%d", time.Now().UnixMilli())
// 		os.Setenv(buildNumParam, buildNum)
// 	}

// 	os.Setenv(clusterNameParam, nc.clusterNames)
// 	os.Setenv(rancherURLParam, nc.config.RancherURL)
// 	os.Setenv(adminTokenParam, nc.config.AdminToken)
// 	os.Setenv(userTokenParam, nc.config.AdminToken)

// 	os.Setenv(pytestOptionsParam, networkChecksPyTestOption)
// 	os.Setenv(testDirParam, "tests/v3_api/")
// 	os.Setenv(rootPathParam, "/src/rancher-validation/")

// 	os.Setenv("RANCHER_VALIDATE_RESOURCES_PREFIX", "test-nc")
// 	os.Setenv("RANCHER_CREATE_RESOURCES_PREFIX", "test-nc")
// 	os.Setenv("RANCHER_ENABLE_HOST_NODE_PORT_TESTS", "True")
// 	os.Setenv("RANCHER_CHECK_FOR_LB", "False")
// 	os.Setenv("RANCHER_SKIP_INGRESS", "False")
// 	os.Setenv("RANCHER_PROJECT_ISOLATION", "disabled")
// 	os.Setenv("RANCHER_TEST_RBAC", "False")
// 	os.Setenv("RANCHER_CLEANUP_CLUSTER", "False")
// 	os.Setenv("RANCHER_HARDENED_CLUSTER", "False")
// 	os.Setenv("RANCHER_SKIP_PING_CHECK_TEST", "False")
// 	os.Setenv("RANCHER_UPGRADE_CHECK", "preupgrade")

// 	if nc.config.AdminToken == nc.config.UserToken {
// 		os.Setenv("USER", "admin")
// 		os.Setenv("USERNAME", "admin")
// 	} else {
// 		os.Setenv("USER", nc.config.Username)
// 		os.Setenv("USERNAME", nc.config.Username)
// 	}

// 	fmt.Sprintf(fileContent,
// 		nc.clusterNames, nc.config.RancherURL, nc.config.AdminToken, nc.config.UserToken)

// }
