/*
Copyright © 2021 Theo Salvo

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
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	"github.com/aws/aws-sdk-go-v2/service/textract/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var profileName string

var validImageTypes []string = []string{"image/jpeg", "image/jpg", "image/png", "image/x-png"}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "textract",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("root called")

		if len(args) != 1 {
			log.Fatal("Need exactly one arg--how did you get here?")
		}
		filename := args[0]

		// Open and read file
		data, dataErr := ioutil.ReadFile(filename)
		if dataErr != nil {
			log.Fatal(dataErr)
		}

		// Validate that file is valid image type
		contentType := http.DetectContentType(data)
		validImage := false
		for _, v := range validImageTypes {
			if v == contentType {
				validImage = true
			}
		}

		if !validImage {
			log.Fatal("Image must be in JPEG or PNG format.")
		}

		// Send to Amazon Textract
		ctx := context.Background()
		cfg, cfgErr := config.LoadDefaultConfig(ctx)
		if cfgErr != nil {
			log.Fatalf("failed to load configuration, %v", cfgErr)
		}

		client := textract.NewFromConfig(cfg)
		result, resultErr := client.DetectDocumentText(context.Background(), &textract.DetectDocumentTextInput{
			Document: &types.Document{
				Bytes: data,
			},
		})
		if resultErr != nil {
			log.Fatalf("failed to call textrat, %v", resultErr)
		}

		for _, block := range result.Blocks {
			if block.BlockType == types.BlockTypeLine {
				fmt.Println(aws.ToString(block.Text))
			}
		}

	},
}

// Exe cute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.textract-cli.yaml)")
	rootCmd.PersistentFlags().StringVar(&profileName, "profile", "", "Use a specific profile from your credential file")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	viper.BindPFlag("toggle", rootCmd.Flags().Lookup("toggle"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".textract-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".textract-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
