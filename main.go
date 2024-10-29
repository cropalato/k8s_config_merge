//
// main.go
// Copyright (C) 2021 rmelo <Ricardo Melo <rmelo@ludia.com>>
//
// Distributed under terms of the MIT license.
//

package main

import (
	//"encoding/json"
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"strings"

	"gopkg.in/yaml.v2"
)

type Rename struct {
	Old string
	New string
}

type ClusterConnection struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}

type Cluster struct {
	Cluster ClusterConnection `yaml:"cluster"`
	Name    string            `yaml:"name"`
}

type ContextInfo struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

type Context struct {
	Context ContextInfo `yaml:"context"`
	Name    string      `yaml:"name"`
}

type UserAuth struct {
	ClientCertificateData string `yaml:"client-certificate-data,omitempty"`
	ClientKeyData         string `yaml:"client-key-data,omitempty"`
	Exec                  Exec   `yaml:"exec,omitempty"`
}

type Exec struct {
	ApiVersion         string `yaml:"apiVersion"`
	Args               string `yaml:"args,omitempty"`
	Command            string `yaml:"command"`
	Env                string `yaml:"env,omitempty"`
	InstallHint        string `yaml:"installHint"`
	InteractiveMode    string `yaml:"interactiveMode"`
	ProvideClusterInfo bool   `yaml:"provideClusterInfo"`
}

type User struct {
	User UserAuth `yaml:"user"`
	Name string   `yaml:"name"`
}

type PreferencesType struct {
}

type K8SConfig struct {
	ApiVersion     string          `yaml:"apiVersion"`
	Clusters       []Cluster       `yaml:"clusters"`
	Contexts       []Context       `yaml:"contexts"`
	CurrentContext string          `yaml:"current-context"`
	Kind           string          `yaml:"kind"`
	Preferences    PreferencesType `yaml:"preferences"`
	Users          []User          `yaml:"users"`
}

// parseK8SConfig reads and parses kubernetes config from an io.Reader
func parseK8SConfig(r io.Reader) (*K8SConfig, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var config K8SConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	return &config, nil
}

// readK8SConfigFile reads kubernetes config from a file
func readK8SConfigFile(filename string) (*K8SConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", filename, err)
	}
	defer file.Close()

	return parseK8SConfig(file)
}

func writeK8SConfigFile(file string, k8sCfg *K8SConfig) {
	bytes, err := yaml.Marshal(k8sCfg)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	err = os.WriteFile(file, bytes, 0644)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func stringExists(list []string, word string) bool {
	for _, elem := range list {
		if elem == word {
			return true
		}
	}
	return false
}

func MergeCfg(dstCfg *K8SConfig, srcCfg *K8SConfig) {
	//Merge clusters map
	var knownClusterNames []string
	var knownContextNames []string
	var knownUserNames []string
	var renamedUsers map[string]string
	var renamedClusters map[string]string
	//var tmpRename Rename
	var text string
	reader := bufio.NewReader(os.Stdin)
	renamedUsers = make(map[string]string)
	renamedClusters = make(map[string]string)

	//Fixing duplicated cluster name
	for _, dstCluster := range dstCfg.Clusters {
		knownClusterNames = append(knownClusterNames, dstCluster.Name)
	}
	for _, srcCluster := range srcCfg.Clusters {
		cName := srcCluster.Name
		for stringExists(knownClusterNames, cName) {
			text = " "
			for strings.Contains(text, " ") {
				fmt.Printf("Cluster name '%v' already exists. Please provide a valid([A-Za-z0-9]+) new name: ", cName)
				text, _ = reader.ReadString('\n')
				text = strings.Replace(text, "\n", "", -1)
			}
			renamedClusters[cName] = text
			cName = text
			srcCluster.Name = text
		}
		dstCfg.Clusters = append(dstCfg.Clusters, srcCluster)
	}
	//Fixing duplicated user name
	for _, dstUser := range dstCfg.Users {
		knownUserNames = append(knownUserNames, dstUser.Name)
	}
	for _, srcUser := range srcCfg.Users {
		cName := srcUser.Name
		for stringExists(knownUserNames, cName) {
			text = " "
			for strings.Contains(text, " ") {
				fmt.Printf("User name '%v' already exists. Please provide a valid([A-Za-z0-9]+) new name: ", cName)
				text, _ = reader.ReadString('\n')
				text = strings.Replace(text, "\n", "", -1)
			}
			renamedUsers[cName] = text
			cName = text
			srcUser.Name = text
		}
		dstCfg.Users = append(dstCfg.Users, srcUser)
	}
	//Fixing duplicated context name
	for _, dstContext := range dstCfg.Contexts {
		knownContextNames = append(knownContextNames, dstContext.Name)
	}
	for _, srcContext := range srcCfg.Contexts {
		cName := srcContext.Name
		if renamedClusters[srcContext.Context.Cluster] != "" {
			cName = strings.Replace(srcContext.Name, fmt.Sprintf("@%v", srcContext.Context.Cluster), fmt.Sprintf("@%v", renamedClusters[srcContext.Context.Cluster]), 1)
			srcContext.Name = cName
			srcContext.Context.Cluster = renamedClusters[srcContext.Context.Cluster]
		}
		if renamedUsers[srcContext.Context.User] != "" {
			cName = strings.Replace(srcContext.Name, fmt.Sprintf("%v@", srcContext.Context.User), fmt.Sprintf("%v@", renamedUsers[srcContext.Context.User]), 1)
			srcContext.Name = cName
			srcContext.Context.User = renamedUsers[srcContext.Context.User]
		}
		for stringExists(knownContextNames, cName) {
			text = " "
			for strings.Contains(text, " ") {
				fmt.Printf("Context name '%v' already exists. Please provide a valid([A-Za-z0-9]+) new name: ", cName)
				text, _ = reader.ReadString('\n')
				text = strings.Replace(text, "\n", "", -1)
			}
			cName = text
			srcContext.Name = text
		}
		dstCfg.Contexts = append(dstCfg.Contexts, srcContext)
	}
}

func main() {

	var srcFiles arrayFlags
	required := []string{"s"} //, "d"}
	// load arguments
	//if len(os.Args[1:]) < 2 {
	//	log.Fatalln("Programs waiting two arguments!")
	//}
	user, err := user.Current()
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	var dstFile = flag.String("d", fmt.Sprintf("%v/%v", user.HomeDir, ".kube/config"), "File where we want to include the new config")
	flag.Var(&srcFiles, "s", "Files you want to merge with dst file.")
	flag.Parse()
	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	for _, req := range required {
		if !seen[req] {
			log.Fatalf("Missing required argument.\n")
		}
	}

	//TODO: Validate if files exist

	// Connect a client.

	var cfg, full_cfg *K8SConfig

	full_cfg, err = readK8SConfigFile(*dstFile)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	for _, file := range srcFiles {
		//log.Println(index, file)
		if file == "-" {
			cfg, err = parseK8SConfig(bufio.NewReader(os.Stdin))
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
		} else {
			cfg, err = readK8SConfigFile(file)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
		}
		MergeCfg(full_cfg, cfg)
	}
	writeK8SConfigFile(*dstFile, full_cfg)

}
