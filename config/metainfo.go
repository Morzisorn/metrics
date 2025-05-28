package config

import "fmt"

func PrintMetaInfo(buildVersion, buildDate, buildCommit string) {
	if buildVersion != "" {
		fmt.Printf("Build version: %s\n", buildVersion)
	} else {
		fmt.Println("Build version: N/A")
	}

	if buildDate != "" {
		fmt.Printf("Build date: %s\n", buildDate)
	} else {
		fmt.Println("Build date: N/A")
	}

	if buildCommit != "" {
		fmt.Printf("Build commit: %s\n", buildCommit)
	} else {
		fmt.Println("Build commit: N/A")
	}
}
