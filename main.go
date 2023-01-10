// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"terraform-provider-datahub/internal/datahub"
// )

// func main(){
// 	client := datahub.NewDatahubClient("https://staging.api.datahub.allyourbi.nl", "16179cd5-0483-48d2-bd7c-1d0035715728", "50YZbAU@EON1#2L!pmJ5")

// 	jobReq := datahub.CreateJobRequest{
// 		Name: "test",
// 		Type: "full",
// 		Image: "aybcr.azurecr.io/datahub/datahub-test",
// 		Environment: map[string]string{"TEST_API_CONF" : "TEST_ENV"},
// 		Secrets: map[string]string{"TEST_SECRET" : "TEST_SECRET"},
// 	}

// 	job, err := client.CreateJob(context.Background(), jobReq)
// 	jobId := job.JobID
// 	data, err := json.Marshal(&job)

// 	fmt.Println(job)
// 	fmt.Println(string(data))
// 	fmt.Println(err)

// 	fmt.Println("---------NEXT----------")
// 	job2, err := client.GetJob(context.Background(), jobId)
// 	data2, err := json.Marshal(&job2)
// 	fmt.Println(job2)
// 	fmt.Println(string(data2))
// 	fmt.Println(err)

// 	fmt.Println("---------NEXT----------")
// 	update := datahub.UpdateJobRequest{
// 		Name: "test2",
// 		Environment: map[string]string{"NEW_KEY": "1"},
// 		Deletes: &datahub.Deletes{
// 			Environment: []string{"TEST_API_CONF"},
// 		},
// 	}

// 	job3, err := client.UpdateJob(context.Background(), jobId, update)
// 	data3, err := json.Marshal(&job3)
// 	fmt.Println(job3)
// 	fmt.Println(string(data3))
// 	fmt.Println(err)

// }

package main

import (
	"context"
	"flag"
	"log"
	"terraform-provider-datahub/internal/terraform"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name datahub

func main() {

	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), terraform.New, providerserver.ServeOpts{
		// NOTE: This is not a typical Terraform Registry provider address,
		// such as registry.terraform.io/hashicorp/hashicups. This specific
		// provider address is used in these tutorials in conjunction with a
		// specific Terraform CLI configuration for manual development testing
		// of this provider.
		Address: "registry.terraform.io/aybi/datahub",
		Debug:   debug,
		// ProtocolVersion: 6,
	})

	if err != nil {
		log.Fatal(err.Error())
	}
}
