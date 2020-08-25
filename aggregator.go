package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

func getLatestOutput(instanceId string, sess *session.Session, latest bool) (string, error) {
	consoleInput := &ec2.GetConsoleOutputInput{
		InstanceId: aws.String(instanceId),
		Latest:     aws.Bool(latest),
	}
	ec2Svc := ec2.New(sess)
	if result, err := ec2Svc.GetConsoleOutput(consoleInput); err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// Non Nitro AWS instance types don't support retrieving the latest console logs
			if awsErr.Code() == "UnsupportedOperation" && latest {
				return getLatestOutput(instanceId, sess, false)
			} else {
				return "", awsErr
			}
		} else {
			return "", err
		}
	} else {
		if output, err := base64.StdEncoding.DecodeString(*result.Output); err != nil {
			log.Printf("Failed decoding console output from base64\n%s", err)
			return "", err
		} else {
			return string(output), nil
		}
	}
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func findOverlappingEndingIndex(prev, latest string) int {
	//This function deals with 2 strings where the "latest" first characters may overlap with the "prev" last characters
	//The function will return "latest" index which starts a new string that isn't overlapping with "prev"
	//examples:
	//prev="abcd", latest="cdef" returns 2
	//prev="abcd", latest="efgh" returns 0

	var i int
	prevLen := len(prev)
	for i = min(prevLen, len(latest)); i > 0; i-- {
		if latest[:i] == prev[prevLen-i:] {
			break
		}
	}
	return i
}

func aggregateConsoleOutput(wg *sync.WaitGroup, folder, instanceId, instanceAlias string, sess *session.Session) {
	defer wg.Done()

	log.Printf("Instance %s: start aggregating console log", instanceId)

	logPath := fmt.Sprintf("%s/%s.log", folder, instanceAlias)

	if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		log.Printf("Instance %s: failed opening file: %s", instanceId, logPath)
		log.Fatalln(err)
	} else {
		defer f.Close()
		prev := ""
		total := 0
		for {
			if output, err := getLatestOutput(instanceId, sess, true); err != nil {
				log.Println(err.Error())
			} else {
				ind := findOverlappingEndingIndex(prev, output)
				newData := output[ind:]
				if newData != "" {
					log.Printf("Instance %s: got new console output, updating log file '%s'", instanceId, logPath)
					if _, err := f.WriteString(newData + "\n"); err != nil {
						log.Printf("Instance %s: failed writing new console output to file\n%s", instanceId, err)
					} else {
						total += len(newData)
						log.Printf("Instance %s: total characters written to log file so far: %d", instanceId, total)
					}
				} else {
					log.Printf("Instance %s: no new console output", instanceId)
				}
				prev = output
			}
			time.Sleep(5 * time.Minute)
		}
	}
}

type arrayFlag []string

func (i *arrayFlag) String() string {
	return "my string representation"
}

func (i *arrayFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func parseParams() (string, string, arrayFlag) {
	var ids arrayFlag

	region := flag.String("region", "eu-central-1", "AWS instance region")
	folder := flag.String("folder", ".", "Aggregated AWS console log folder")
	flag.Var(&ids, "id", "List of AWS instances ids")
	flag.Parse()
	log.Printf("Region:%s OutputFolder:'%s' InstancesIds:%s", *region, *folder, ids)
	return *region, *folder, ids
}

func getInstanceIdAndAlias(id string) (string, string) {
	var instanceAlias string
	split := strings.Split(id, ":")
	instanceId := split[0]
	if len(split) > 1 {
		instanceAlias = split[1]
	} else {
		instanceAlias = instanceId
	}
	return instanceId, instanceAlias
}

func main() {
	log.Println("Running aws console log aggregator service...")

	region, folder, ids := parseParams()

	if sess, err := session.NewSession(&aws.Config{Region: aws.String(region)}); err != nil {
		log.Println("Failed creating session")
		log.Fatalln(err)
	} else {
		var wg sync.WaitGroup
		for _, id := range ids {
			wg.Add(1)
			instanceId, instanceAlias := getInstanceIdAndAlias(id)
			go aggregateConsoleOutput(&wg, folder, instanceId, instanceAlias, sess)
		}
		wg.Wait()
	}
}
