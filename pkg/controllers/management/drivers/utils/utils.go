package utils

import "fmt"

func CredentialConfigSchemaName(driverName string) string {
	return fmt.Sprintf("%s%s", driverName, "credentialconfig")
}
