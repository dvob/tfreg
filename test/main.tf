
module "test_v1" {
  source = "example.com/dvob/terraform-dummy-module/base"
  version = "< 2.0.0"
}

module "test_v2" {
  source = "example.com/dvob/terraform-dummy-module/base"
}

output "output_v1" {
  value = module.test_v1.text
}

output "output_v2" {
  value = module.test_v2.text
}
