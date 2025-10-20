terraform {
  required_providers {
    uamoim = {
      source = "union-investment/terraform/uamoim"
    }
  }
}

provider "uamoim" {
  host     = "http://localhost:19090"
  username = "education"
  password = "test123"
}

data "uamoim_shops" "example" {}

data "uamoim_sods" "example" {}

data "uamoim_coffees" "example" {}

output "example_coffees" {
  value = data.uamoim_coffees.example
}