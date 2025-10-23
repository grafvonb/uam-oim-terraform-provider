resource "uam_module" "test_module" {
  for_each = { for group in local.groups : group.name => group }

  application_name = "Application"
  module_name      = each.value.name
  description      = each.value.description
}

resource "module_biso" "test_module_biso" {
  for_each = local.group_biso_combinations

  application_name = "Application"
  module_id        = each.value.group_name
  biso_name        = each.value.bis_name
  reason           = "AV {each.value.group_name}"
}

resource "uam_role_assignment" "test_module_role_assignment" {
  for_each = local.group_role_combinations

  application_name = "Application"
  module_id        = each.value.module_name
  group_name       = each.value.role_name
  shop_name        = each.value.shop_name
  sod_class_id     = "Keine SoD Relevanz"
  order_for        = each.value.order_for
  approval_flow    = each.value.approval_flow
  description      = each.value.description
  can_fachrolle    = true
}
