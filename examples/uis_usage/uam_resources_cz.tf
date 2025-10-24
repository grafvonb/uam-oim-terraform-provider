# Create modules from groups
resource "uam_module" "this" {
  for_each         = { for g in local.groups : g.name => g }
  application_name = "Application"
  module_name      = each.value.name
  description      = each.value.description
}

# Resolve module IDs by name (so downstream resources can use ids)
data "uam_module_by_name" "mod" {
  for_each         = uam_module.this
  application_name = "Application"
  module_name      = each.value.module_name
}

# Assign BISO to each module/group pair
resource "uam_module_biso" "this" {
  for_each         = local.group_biso_combinations
  application_name = "Application"

  module_id = data.uam_module_by_name.mod[each.value.group_name].id
  biso_id   = data.uam_group_by_name.biso[each.key].id

  reason    = "AV ${each.value.group_name}"
}

# Resolve BISO group id by name
data "uam_group_by_name" "biso" {
  for_each  = local.group_biso_combinations
  group_name = each.value.biso_name
}

# Role assignment for each group/role combo
resource "uam_role_assignment" "this" {
  for_each         = local.group_role_combinations
  application_name = "Application"

  module_id     = data.uam_module_by_name.mod[each.value.module_name].id
  group_id      = data.uam_group_by_name.role[each.key].id
  shop_id       = data.uam_shop_by_name.shop[each.key].id
  sod_class_id  = data.uam_sod_class_by_name.sod[each.key].id

  order_for     = each.value.order_for
  approval_flow = each.value.approval_flow
  description   = each.value.description
  can_fachrolle = true
}

# Lookups used by role assignments
data "uam_group_by_name" "role" {
  for_each   = local.group_role_combinations
  group_name = each.value.role_name
}

data "uam_shop_by_name" "shop" {
  for_each = local.group_role_combinations
  shop_name = each.value.shop_name
}

data "uam_sod_class_by_name" "sod" {
  for_each    = local.group_role_combinations
  sod_class_name = "Keine SoD Relevanz"
}

