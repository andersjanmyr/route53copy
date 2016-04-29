package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

func connect(profile string) *route53.Route53 {
	return route53.New(session.New(), &aws.Config{
		Region: aws.String("eu-west-1"),
		Credentials: credentials.NewCredentials(&credentials.SharedCredentialsProvider{
			Profile: profile,
		}),
	})
}

func getHostedZone(service *route53.Route53, domain string) (*route53.HostedZone, error) {
	params := &route53.ListHostedZonesByNameInput{
		DNSName:  aws.String(domain),
		MaxItems: aws.String("1"),
	}
	resp, err := service.ListHostedZonesByName(params)
	if err != nil {
		return nil, err
	}

	zone := resp.HostedZones[0]
	return zone, nil
}

func getResourceRecords(profile string, domain string) ([]*route53.ResourceRecordSet, error) {
	service := connect(profile)
	zone, err := getHostedZone(service, domain)
	if err != nil {
		return nil, err
	}

	params := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(*zone.Id),
	}
	resp, err := service.ListResourceRecordSets(params)
	if err != nil {
		return nil, err
	}
	return resp.ResourceRecordSets, nil
}

func createChanges(srcDomain string, destDomain string, recordSets []*route53.ResourceRecordSet) []*route53.Change {
	//srcDomain = normalizeDomain(srcDomain)
	var changes []*route53.Change
	//s := []string{srcDomain, "$"}
	re := regexp.MustCompile(strings.Join([]string{srcDomain, ".$"}, ""))
	for _, recordSet := range recordSets {
		if (*recordSet.Type == "NS" || *recordSet.Type == "SOA" || *recordSet.Type == "TXT") && *recordSet.Name == normalizeDomain(srcDomain) {
			log.Printf("Skipping %s %s", *recordSet.Name, *recordSet.Type)
			continue
		}

		//log.Printf(*recordSet.Name)
		*recordSet.Name = normalizeDomain(re.ReplaceAllLiteralString(*recordSet.Name, destDomain))
		//log.Printf(*recordSet.Name)
		change := &route53.Change{
			Action:            aws.String("UPSERT"),
			ResourceRecordSet: recordSet,
		}
		changes = append(changes, change)
	}
	return changes

}

func normalizeDomain(domain string) string {
	if strings.HasSuffix(domain, ".") {
		return domain
	}
	return domain + "."

}

func updateRecords(sourceProfile, destProfile, domain string, changes []*route53.Change) (*route53.ChangeInfo, error) {
	service := connect(destProfile)
	zone, err := getHostedZone(service, domain)
	if err != nil {
		return nil, err
	}
	params := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: zone.Id,
		ChangeBatch: &route53.ChangeBatch{
			Changes: changes,
			Comment: aws.String("Importing ALL records from " + sourceProfile),
		},
	}
	resp, _ := service.ChangeResourceRecordSets(params)
	return resp.ChangeInfo, nil
}

func main() {
	log.SetFlags(0)

	var version bool
	var dry bool
	var help bool

	program := path.Base(os.Args[0])
	flag.BoolVar(&dry, "dry", false, "Don't make any changes")
	flag.BoolVar(&help, "help", false, "Show help text")
	flag.BoolVar(&version, "version", false, "Show version")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <source_profile> <dest_profile> <domain>\n", program)
		flag.PrintDefaults()
	}
	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}
	if version {
		fmt.Println(Version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 3 {
		fmt.Fprintf(os.Stderr, "Wrong number of arguments, %d < 3\n", len(args))
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <source_profile> <dest_profile> <srcDomain> [destDomain]\n", program)
		flag.PrintDefaults()
		os.Exit(1)
	}

	sourceProfile := args[0]
	destProfile := args[1]
	srcDomain := args[2]
	destDomain := srcDomain
	if len(args) == 4 {
		destDomain = args[3]
	}

	recordSets, err := getResourceRecords(sourceProfile, srcDomain)
	if err != nil {
		panic(err)
	}
	//log.Printf(" %s ", recordSets)
	changes := createChanges(srcDomain, destDomain, recordSets)
	log.Println("Number of records to copy", len(changes))

	if dry {
		log.Printf("Not copying records to %s since -dry is given\n", destProfile)
		service := connect(destProfile)
		zone, err := getHostedZone(service, destDomain)
		if err != nil {
			panic(err)
		}
		log.Printf("Destination profile contains %d records, including NS and SOA\n",
			*zone.ResourceRecordSetCount)
	} else {
		changeInfo, err := updateRecords(sourceProfile, destProfile, destDomain, changes)
		if err != nil {
			panic(err)
		}
		log.Printf("%d records in '%s' are copied from %s to %s\n",
			len(changes), destDomain, sourceProfile, destProfile)
		log.Printf("%#v\n", changeInfo)
	}

}
