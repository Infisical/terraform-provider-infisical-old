terraform {
  required_providers {
    infisical = {
      source = "infisical/infisical"
      version = "0.1"
    }
  }
}

provider "infisical" {
  api_key  = "YOUR_API_KEY"
  host     = "https://infisical.com"
}

# List all organizations for the user's api_key.
data "infisical_organizations" "all" {}

data "infisical_projects" "all" {
  organization_id = data.infisical_organizations.all.organizations[0].id
}