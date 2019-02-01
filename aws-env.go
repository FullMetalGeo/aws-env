package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"log"
	"os"
	"strings"
)

const (
	formatExports = "exports"
	formatDotenv  = "dotenv"
)

func main() {
	if os.Getenv("AWS_ENV_PATH") == "" {
		log.Println("aws-env running locally, without AWS_ENV_PATH")
		return
	}

	recursivePtr := flag.Bool("recursive", false, "recursively process parameters on path")
	format := flag.String("format", formatExports, "output format")
	flag.Parse()

	if *format == formatExports || *format == formatDotenv {
	} else {
		log.Fatal("Unsupported format option. Must be 'exports' or 'dotenv'")
	}

	sess := CreateSession()
	client := CreateClient(sess)

        env_paths := strings.Split(os.Getenv("AWS_ENV_PATH"), ":")

        for i := range env_paths {
		ExportVariables(client, env_paths[i], *recursivePtr, *format, "")
        }
}

func CreateSession() *session.Session {
	return session.Must(session.NewSession())
}

func CreateClient(sess *session.Session) *ssm.SSM {
	return ssm.New(sess)
}

func ExportVariables(client *ssm.SSM, path string, recursive bool, format string, nextToken string) {
	input := &ssm.GetParametersByPathInput{
		Path:           &path,
		WithDecryption: aws.Bool(true),
		Recursive:      aws.Bool(recursive),
	}

	if nextToken != "" {
		input.SetNextToken(nextToken)
	}

	output, err := client.GetParametersByPath(input)

	if err != nil {
		log.Panic(err)
	}

        if len(output.Parameters) == 0 {
                input := &ssm.GetParameterInput{
		        Name:           &path,
		        WithDecryption: aws.Bool(true),
	        }
                paramOutput, _ := client.GetParameter(input)

                if paramOutput.Parameter != nil {
                        paramName := strings.Split(path, "/")
                        name := fmt.Sprintf("%s/%s", *paramOutput.Parameter.Name, paramName[len(paramName)-1])
                        OutputParameter(path, name, *paramOutput.Parameter.Value, format)
                }
        }

	for _, element := range output.Parameters {
		OutputParameter(path, *element.Name, *element.Value, format)
	}

	if output.NextToken != nil {
		ExportVariables(client, path, recursive, format, *output.NextToken)
	}
}

func OutputParameter(path string, name string, value string, format string) {
	env := strings.ToUpper(strings.Replace(strings.Replace(strings.Trim(name[len(path):], "/"), "/", "_", -1), "-", "_", -1))
	value = strings.Replace(value, "\n", "\\n", -1)

	switch format {
	case formatExports:
		fmt.Printf("export %s=$'%s'\n", env, value)
	case formatDotenv:
		fmt.Printf("%s=\"%s\"\n", env, value)
	}
}
