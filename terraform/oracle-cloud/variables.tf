variable "tenancy_ocid" {
  description = "OCID of your OCI tenancy"
  type        = string
  sensitive   = true
}

variable "user_ocid" {
  description = "OCID of the user calling the API"
  type        = string
  sensitive   = true
}

variable "fingerprint" {
  description = "Fingerprint for the API key"
  type        = string
  sensitive   = true
}

variable "private_key_path" {
  description = "Path to the private key file for API authentication"
  type        = string
  default     = "~/.oci/oci_api_key.pem"
}

variable "region" {
  description = "OCI region (e.g., eu-frankfurt-1, eu-amsterdam-1)"
  type        = string
  default     = "eu-frankfurt-1"
}

variable "compartment_ocid" {
  description = "OCID of the compartment where resources will be created"
  type        = string
}

variable "environment" {
  description = "Environment name (e.g., prod, dr, staging)"
  type        = string
  default     = "dr"
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "cloud-dr-orchestrator"
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default = {
    Project     = "cloud-dr-orchestrator"
    ManagedBy   = "terraform"
    Environment = "dr"
  }
}
