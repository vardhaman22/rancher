package networkchecks

import (
	"testing"

	"github.com/rancher/rancher/tests/framework/clients/rancher"
	"github.com/rancher/rancher/tests/framework/pkg/session"
	"github.com/stretchr/testify/require"
)

func TestRunNetworkChecks(t *testing.T) {
	testSession := session.NewSession()
	netChecks := NetworkChecks{}
	client, err := rancher.NewClient("", testSession)
	require.NoError(t, err)
	netChecks.InitNetChecks(client)
	err = netChecks.RunNetworkChecks("vardhaman-rke1", "c-pjq9t")
	require.NoError(t, err)
	// RunNetworkChecks("https://3.136.116.194", "token-clgdg:jtp47xpjjkbfmnf8wbtbbxj5dbh6tzgg7gtthwcfcv7l7227sr4jj2",
	// "token-clgdg:jtp47xpjjkbfmnf8wbtbbxj5dbh6tzgg7gtthwcfcv7l7227sr4jj2", "vardhaman-rke1", "admin")
}
