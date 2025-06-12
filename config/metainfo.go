package config

import "fmt"

func PrintMetaInfo(buildVersion, buildDate, buildCommit string) {
	printField("Build version", buildVersion)
	printField("Build date", buildDate)
	printField("Build commit", buildCommit)
}

func printField(name, value string) {
	if value != "" {
		fmt.Printf("%s: %s\n", name, value)
	} else {
		fmt.Printf("%s: N/A\n", name)
	}
}
