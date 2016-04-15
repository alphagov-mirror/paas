variable "system_dns_zone_id" {
  description = "Amazon Route53 DNS zone identifier for the system components. Different per account."
}

variable "system_dns_zone_name" {
  description = "Amazon Route53 DNS zone name for the provisioned environment."
}

variable "git_rsa_id_pub" {
  description = "Public SSH key for the git user"
}

variable "git_default_branch_workaround" {
  description = "Value of current default branch for codecommit git repo. Temporary workaround."
  default     = ""
}