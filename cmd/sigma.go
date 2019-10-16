/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ccdcoe/go-peek/pkg/models/rules/sigma"
	"github.com/ccdcoe/go-peek/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// sigmaCmd represents the sigma command
var sigmaCmd = &cobra.Command{
	Use:   "sigma",
	Short: "",
	Long:  ``,
	Run:   entrypoint,
}

func entrypoint(cmd *cobra.Command, args []string) {
	var err error
	var dir string
	if dir = viper.GetString("sigma.rules.dir"); dir == "" {
		log.Fatal("Missing rule directory, see --help")
	}
	if dir, err = utils.ExpandHome(dir); err != nil {
		log.Fatal(err)
	}
	rules := make([]*sigma.RawRule, 0)
	if err = filepath.Walk(dir, func(
		path string,
		info os.FileInfo,
		err error,
	) error {
		if !info.IsDir() && strings.HasSuffix(path, "yml") {
			var s sigma.RawRule
			data, err := ioutil.ReadFile(path) // just pass the file name
			if err != nil {
				log.WithFields(log.Fields{
					"file": path,
				}).Error(err)
				return nil
			}
			if err := yaml.Unmarshal([]byte(data), &s); err != nil {
				log.WithFields(log.Fields{
					"file": path,
				}).Error(err)
				return nil
			}
			if s.Detection == nil {
				log.WithFields(log.Fields{
					"title":     s.Title,
					"file":      path,
					"detection": s.Detection,
				}).Error("missing detection map, check rule")
				return nil
			}
			if err := sigma.Parse(s); err != nil {
				log.WithFields(log.Fields{
					"file":      path,
					"detection": s.Detection,
				}).Error(err)
				return nil
			}
			rules = append(rules, &s)
		}
		return err
	}); err != nil {
		log.Fatal(err)
	}
	log.Infof("Got %d rules from %s", len(rules), dir)
}

func init() {
	rootCmd.AddCommand(sigmaCmd)

	sigmaCmd.PersistentFlags().String("sigma-rules-dir", "", "Directory that contains sigma rules.")
	viper.BindPFlag("sigma.rules.dir", sigmaCmd.PersistentFlags().Lookup("sigma-rules-dir"))
}
