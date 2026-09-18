package clusterregistrationtoken

import (
	"fmt"
	"strings"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/rancher/pkg/capr"
)

type EnvType int

const (
	Linux EnvType = iota
	PowerShell
	Docker
)

func AgentEnvVars(cluster *v3.Cluster, envType EnvType, winsSystemAgentDir string) string {
	var agentEnvVars []string
	if cluster == nil {
		return ""
	}
	systemAgentDataDirEnvFound := false
	for _, envVar := range cluster.Spec.AgentEnvVars {
		if envVar.Value == "" {
			continue
		}
		if envVar.Name == capr.SystemAgentDataDirEnvVar {
			systemAgentDataDirEnvFound = true
		}
		switch envType {
		case Docker:
			agentEnvVars = append(agentEnvVars, fmt.Sprintf("-e \"%s=%s\"", envVar.Name, envVar.Value))
		case PowerShell:
			value := envVar.Value
			if envVar.Name == capr.SystemAgentDataDirEnvVar && winsSystemAgentDir != "" {
				value = winsSystemAgentDir
			}
			agentEnvVars = append(agentEnvVars, fmt.Sprintf("$env:%s=\"%s\";", envVar.Name, value))
		default:
			agentEnvVars = append(agentEnvVars, fmt.Sprintf("%s=\"%s\"", envVar.Name, envVar.Value))
		}
	}
	if winsSystemAgentDir != "" && envType == PowerShell && !systemAgentDataDirEnvFound {
		agentEnvVars = append(agentEnvVars, fmt.Sprintf("$env:%s=\"%s\";", capr.SystemAgentDataDirEnvVar, winsSystemAgentDir))
	}
	return strings.Join(agentEnvVars, " ")
}
