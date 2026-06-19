terraform {
  required_providers {
    neon       = { source = "kislerdm/neon",         version = "~> 0.6"    }
    upstash    = { source = "upstash/upstash",       version = "~> 1.5"    }
    cloudflare = { source = "cloudflare/cloudflare", version = "~> 4"      }
  }

  # Remote state in Cloudflare R2 (free, S3-compatible)
  # Run `terraform init` with these env vars set:
  #   AWS_ACCESS_KEY_ID     = var.r2_access_key_id
  #   AWS_SECRET_ACCESS_KEY = var.r2_secret_access_key
  backend "s3" {
    bucket                      = "tf-state"
    key                         = "fantasy-league/terraform.tfstate"
    region                      = "auto"
    skip_credentials_validation  = true
    skip_metadata_api_check      = true
    skip_region_validation       = true
    skip_requesting_account_id   = true
    force_path_style             = true
    endpoints = {
      s3 = "https://placeholder.r2.cloudflarestorage.com"
    }
    # endpoints.s3 and credentials are overridden via -backend-config flags at init time
  }
}

provider "cloudflare" {
  api_token = var.cf_api_token
}

provider "neon" {
  api_key = var.neon_api_key
}

provider "upstash" {
  email   = var.upstash_email
  api_key = var.upstash_api_key
}

# --- Neon (Postgres) ---
resource "neon_project" "main" {
  name                       = "fantasy-league"
  region_id                  = "aws-ap-southeast-1"
  org_id                     = var.neon_org_id
  history_retention_seconds  = 21600
}

resource "neon_branch" "prod" {
  project_id = neon_project.main.id
  name       = "prod"
}

# --- Upstash Redis ---
resource "upstash_redis_database" "cache" {
  database_name  = "fantasy-league-cache"
  region         = "global"
  primary_region = "ap-southeast-1"
  tls            = true
}

# --- Cloudflare Pages ---
# The Pages project is bootstrapped manually once with:
#   cd frontend && npm run build
#   npx wrangler pages project create fantasy-league
#   npx wrangler pages deploy dist --project-name=fantasy-league
# Subsequent deploys are handled by .github/workflows/deploy-frontend.yml
# The Cloudflare Terraform provider supports Pages resources but not the
# build/deploy lifecycle, so managing it via IaC adds complexity with no benefit.

# --- Cloudflare DNS ---
# api_ip is populated after running: fly ips allocate-v4 --app fantasy-league-api
variable "fly_api_ip" {
  type        = string
  description = "IPv4 address from: fly ips allocate-v4 --app fantasy-league-api"
  default     = ""
}

resource "cloudflare_record" "api" {
  count   = var.fly_api_ip != "" ? 1 : 0
  zone_id = var.cf_zone_id
  name    = "api.fantasy"
  type    = "A"
  value   = var.fly_api_ip
  proxied = true
}
