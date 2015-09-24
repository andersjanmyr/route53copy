package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

func getResourceRecords(domain string) ([]*route53.ResourceRecordSet, error) {
	service := route53.New(&aws.Config{Region: aws.String("eu-west-1")})

	params := &route53.ListHostedZonesByNameInput{
		DNSName:  aws.String(domain),
		MaxItems: aws.String("1"),
	}
	resp, err := service.ListHostedZonesByName(params)
	if err != nil {
		return nil, err
	}

	// resp has all of the response data, pull out instance IDs:
	fmt.Println("Number of reservation sets: ", len(resp.HostedZones))
	zone := resp.HostedZones[0]
	fmt.Println("Name: ", *zone.Name)
	fmt.Println("Id: ", *zone.Id)

	rsParams := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(*zone.Id),
	}
	rsResp, err := service.ListResourceRecordSets(rsParams)
	if err != nil {
		return nil, err
	}
	fmt.Println("Number of Records: ", len(rsResp.ResourceRecordSets))
	return rsResp.ResourceRecordSets, nil
}

func main() {
	resp, err := getResourceRecords("lifelog-dev.sonymobile.com")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", resp)
}
