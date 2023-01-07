package main

import (
	"context"
	"encoding/json"
	"fmt"
	"terraform-provider-aybi-datahub/internal/datahub"
)

func main(){
	client := datahub.NewDatahubClient("https://staging.api.datahub.allyourbi.nl", "16179cd5-0483-48d2-bd7c-1d0035715728", "50YZbAU@EON1#2L!pmJ5")

	jobReq := datahub.CreateJobRequest{
		Name: "test",
		Type: "full",
		Image: "aybcr.azurecr.io/datahub/datahub-test",
		Environment: map[string]string{"TEST_API_CONF" : "TEST_ENV"},
		Secrets: map[string]string{"TEST_SECRET" : "TEST_SECRET"},
	}

	job, err := client.CreateJob(context.Background(), jobReq)
	jobId := job.JobID
	data, err := json.Marshal(&job)


	fmt.Println(job)
	fmt.Println(string(data))
	fmt.Println(err)


	fmt.Println("---------NEXT----------")
	job2, err := client.GetJob(context.Background(), jobId)
	data2, err := json.Marshal(&job2)
	fmt.Println(job2)
	fmt.Println(string(data2))
	fmt.Println(err)


	fmt.Println("---------NEXT----------")
	update := datahub.UpdateJobRequest{
		Name: "test2",
		Environment: map[string]string{"NEW_KEY": "1"},
		Deletes: &datahub.Deletes{
			Environment: []string{"TEST_API_CONF"},
		},
	}

	job3, err := client.UpdateJob(context.Background(), jobId, update)
	data3, err := json.Marshal(&job3)
	fmt.Println(job3)
	fmt.Println(string(data3))
	fmt.Println(err)


}