package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/utils"
)

func main() {
	info, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalf("failed to stat stdin: %v", err)
	}
	if (info.Mode()&os.ModeCharDevice) != 0 || info.Size() < 0 {
		log.Fatal("usage: terraform output -json | tf2kops")
	}

	var parsed struct {
		VPCID struct {
			Value string `json:"value"`
		} `json:"vpc_id"`

		PrivateSubnets struct {
			Value []string `json:"value"`
		} `json:"private_subnets"`

		UtilitySubnets struct {
			Value []string `json:"value"`
		} `json:"utility_subnets"`
	}

	if err := json.NewDecoder(os.Stdin).Decode(&parsed); err != nil {
		log.Fatalf("failed to parse tf output: %v\n", err)
	}

	var subnets []kops.ClusterSubnetSpec
	for _, ps := range parsed.PrivateSubnets.Value {
		cc := strings.Split(ps, " ")
		id := strings.Split(cc[0], "=")[1]
		az := strings.Split(cc[1], "=")[1]
		nat := strings.Split(cc[2], "=")[1]

		s := kops.ClusterSubnetSpec{
			ProviderID: id,
			Name:       az,
			Type:       kops.SubnetTypePrivate,
			Zone:       az,
			Egress:     nat,
		}
		subnets = append(subnets, s)
	}
	for _, ps := range parsed.UtilitySubnets.Value {
		cc := strings.Split(ps, " ")
		id := strings.Split(cc[0], "=")[1]
		az := strings.Split(cc[1], "=")[1]
		s := kops.ClusterSubnetSpec{
			ProviderID: id,
			Name:       "utility-" + az,
			Type:       kops.SubnetTypeUtility,
			Zone:       az,
		}
		subnets = append(subnets, s)
	}

	type marsh struct {
		NetworkID string                   `json:"networkID"`
		Subnets   []kops.ClusterSubnetSpec `json:"subnets"`
	}

	bb, err := utils.YamlMarshal(marsh{
		NetworkID: parsed.VPCID.Value,
		Subnets:   subnets,
	})
	if err != nil {
		log.Fatalf("failed to encode subnets as yaml: %v", err)
	}

	// indent everything to kops default indent
	fmt.Println("  # tf2kops start")
	s := bufio.NewScanner(bytes.NewReader(bb))
	for s.Scan() {
		fmt.Println("  " + s.Text())
	}
	fmt.Println("  # tf2kops end")
	if s.Err() != nil {
		log.Fatal(err)
	}
}
