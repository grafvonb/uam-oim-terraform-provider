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

resource "uamoim_order" "example" {
  items = [{
    coffee = {
      id = 3
    }
    quantity = 2
  }, {
    coffee = {
      id = 1
    }
    quantity = 2
  }
  ]
}

output "example_order" {
  value = uamoim_order.example
}

output "example_coffees" {
  value = data.uamoim_coffees.example
}