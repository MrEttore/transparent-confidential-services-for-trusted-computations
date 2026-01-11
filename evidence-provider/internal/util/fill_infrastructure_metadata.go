package util

import (
	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/compute/metadata"
	"context"
	"evidence-provider/internal/types"
)

func FillGCP(ctx context.Context, sum *types.InfrastructureSummary) error {
	sum.InstanceID, _ = metadata.InstanceIDWithContext(ctx)
	sum.Name, _ = metadata.InstanceNameWithContext(ctx)
	sum.Zone, _ = metadata.ZoneWithContext(ctx)
	sum.ProjectID, _ = metadata.ProjectIDWithContext(ctx)

	machineType, _ := metadata.GetWithContext(ctx, "instance/machine-type")
	sum.MachineType = StripResourceName(machineType)

	// Fetch *VM status* (RUNNING, TERMINATED, etc.).
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		sum.Status = "unknown"
		return err
	}
	defer client.Close()

	resp, err := client.Get(ctx, &computepb.GetInstanceRequest{
		Project:  sum.ProjectID,
		Zone:     sum.Zone,
		Instance: sum.Name,
	})
	if err != nil {
		sum.Status = "unknown"
		return err
	}
	sum.Status = resp.GetStatus()

	return nil
}

/*func fillAWS(ctx context.Context, sum *types.InfrastructureSummary) {
	sess := session.Must(session.NewSession())
	ec2Meta := ec2metadata.New(sess)
	sum.InstanceID, _ = ec2Meta.GetMetadata("instance-id")
	sum.Zone, _ = ec2Meta.GetMetadata("placement/availability-zone")
	sum.Name = sum.InstanceID // AWS doesn’t have a “name” by default; you could tag it
	sum.MachineType, _ = ec2Meta.GetMetadata("instance-type")

	// To get the instance state (pending, running, stopped), you’d need DescribeInstances via the EC2 API:
	//   svc := ec2.New(sess)
	//   out, _ := svc.DescribeInstances(&ec2.DescribeInstancesInput{InstanceIds: []*string{aws.String(sum.InstanceID)}})
	//   sum.Status = *out.Reservations[0].Instances[0].State.Name
}*/

/*func fillAzure(ctx context.Context, sum *types.InfrastructureSummary) {
	client := &http.Client{Timeout: 1 * time.Second}
	req, _ := http.NewRequest("GET",
		"http://169.254.169.254/metadata/instance?api-version=2021-02-01&format=json", nil)
	req.Header.Set("Metadata", "true")

	var az struct {
		Compute struct {
			VMID           string `json:"vmId"`
			Name           string `json:"name"`
			Location       string `json:"location"`
			SubscriptionID string `json:"subscriptionId"`
			VMSize         string `json:"vmSize"`
			ResourceGroup  string `json:"resourceGroupName"`
			// … other fields
		} `json:"compute"`
	}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&az)

	sum.InstanceID = az.Compute.VMID
	sum.Name = az.Compute.Name
	sum.Region = az.Compute.Location
	sum.Subscription = az.Compute.SubscriptionID
	sum.MachineType = az.Compute.VMSize
}*/
