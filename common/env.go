package common

import "os"

func GetEnvOrDefault(envName, defaultValue string) string {
	if envValue, set := os.LookupEnv(envName); set {
		return envValue
	} else {
		return defaultValue
	}
}
