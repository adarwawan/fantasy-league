variable "cf_zone_id" {
  type        = string
  description = "Cloudflare Zone ID for adarw.xyz (shown on the zone overview page)"
}

variable "cf_api_token" {
  type        = string
  sensitive   = true
  description = "Cloudflare API token with DNS Edit + Pages Edit permissions"
}

variable "neon_api_key" {
  type        = string
  sensitive   = true
  description = "Neon API key (console.neon.tech → Account → API keys)"
}

variable "neon_org_id" {
  type        = string
  description = "Neon Organization ID (console.neon.tech → Settings → Organization)"
}

variable "upstash_email" {
  type        = string
  description = "Email address on your Upstash account"
}

variable "upstash_api_key" {
  type        = string
  sensitive   = true
  description = "Upstash API key (console.upstash.com → Account → API keys)"
}

variable "cf_account_id" {
  type        = string
  description = "Cloudflare Account ID (shown on the right sidebar of any zone overview)"
}

variable "r2_access_key_id" {
  type        = string
  sensitive   = true
  description = "R2 API token Access Key ID (used for Terraform remote state)"
}

variable "r2_secret_access_key" {
  type        = string
  sensitive   = true
  description = "R2 API token Secret Access Key"
}
