# Infrastructure

Terraform configuration that provisions the Computational Logic Attester on Google Cloud with Intel TDX support.

## Prerequisites

- [Terraform](https://developer.hashicorp.com/terraform/downloads) 1.2+
- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install) (`gcloud`)
- Access to a Google Cloud project with Intel TDX-enabled machine images
- Permissions to create Compute Engine instances and firewall rules

## Deployment

```bash
cd infrastructure
cp terraform.tfvars.example terraform.tfvars

# Edit terraform.tfvars: set project_id, region, zone, and golden_image_project_id

gcloud auth application-default login
gcloud config set project <your-project-id>

terraform init
terraform plan
terraform apply
```

Record the VM name, zone, and SSH details for Evidence Provider deployment.

## What Gets Deployed

- **Confidential VM (`llm-core-tee`):** Intel TDX-enabled instance running a dummy confidential workload (`llm-core`) under systemd.
- **Firewall rule (`allow-attestation`):** Ingress rule for attestation endpoints
- **Bootstrap script (`init-tee.sh`):** Installs Docker, Go, and configures the workload service at first boot

The CVM uses a golden reference image (`golden-reference-tee`) with pre-configured Docker support. At boot, `init-tee.sh` pulls the latest `llm-core` container and starts it as a systemd service, ensuring runtime measurements align with published baseline manifests.
