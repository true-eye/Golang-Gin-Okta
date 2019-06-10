package config

import (
	"flag"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// ConfigFile is the default config file
var ConfigFile = "./config.yml"

// GlobalConfig is the global config
type GlobalConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Okta	 OktaConfig	 	`yaml:"okta"`
	EndpointACL EndpointConfig	`yaml:"endpointacl"`
}

type EndpointConfig struct {
	ACL 	map[string][]string 	`yaml:"acl"`
}

type OktaConfig struct {
	ClientId		   string  `yaml:"client_id"`
	ClientSecret	   string  `yaml:"client_secret"`
	Issuer			   string  `yaml:"issuer"`
	State		  	   string  `yaml:"state"`
	Nonce				string	`yaml:"nonce"`
	APIURL				string	`yaml:"apiurl"`
	APIToken			string	`yaml:"apitoken"`
}

// ServerConfig is the server config
type ServerConfig struct {
	Addr               string
	Mode               string
	Version            string
	StaticDir          string `yaml:"static_dir"`
	ViewDir            string `yaml:"view_dir"`
	LogDir             string `yaml:"log_dir"`
	UploadDir          string `yaml:"upload_dir"`
	MaxMultipartMemory int64  `yaml:"max_multipart_memory"`
}

// DatabaseConfig is the database config
type DatabaseConfig struct {
	Dialect      string
	DSN          string `yaml:"datasource"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	MaxOpenConns int    `yaml:"max_open_conns"`
}

// global configs
var (
	Global   GlobalConfig
	Server   ServerConfig
	Database DatabaseConfig
	Okta	 OktaConfig
	EndpointACL EndpointConfig
)

// Load config from file
func Load(file string) (GlobalConfig, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("%v", err)
		return Global, err
	}

	err = yaml.Unmarshal(data, &Global)
	if err != nil {
		log.Printf("%v", err)
		return Global, err
	}

	Server = Global.Server
	Database = Global.Database
	Okta = Global.Okta
	EndpointACL = Global.EndpointACL

	// set log dir flag for glog
	flag.CommandLine.Set("log_dir", Server.LogDir)

	return Global, nil
}

// loads configs
func init() {
	if flag.Lookup("test.v") != nil { // run under go test
		ConfigFile = "../config.yml"
	}
	Load(ConfigFile)
}
