locals {
  groups_yaml = yamldecode(file("groups.yaml"))
  groups      = local.groups_yaml.gitlab_groups

  roles = {
    owner = {
      name          = "Administrator"
      order_for     = "Alle internen und externen Mitarbeiter"
      approval_flow = "Vorgesetzter, Applikationsverantwortlicher und BISO"
    }
    maintainer = {
      name          = "Betreuer"
      order_for     = "Alle internen und externen Mitarbeiter"
      approval_flow = "Vorgesetzter und BISO"
    }
    developer = {
      name          = "Entwickler"
      order_for     = "Alle internen und externen Mitarbeiter"
      approval_flow = "Vorgesetzter"
    }
    reporter = {
      name          = "Leser"
      order_for     = "Alle internen und externen Mitarbeiter"
      approval_flow = "Automatisch"
    }
  }

  group_biso_combinations = merge([
    for group in local.groups : {
      for biso in lookup(group, "BISOS", []) :
      "${group.path}-${biso}" => {
        group_path        = group.path
        group_name        = group.name
        group_description = group.description
        biso_name         = biso
      }
    }
  ]...)

  group_role_combinations = merge([
    for group in local.groups : {
      for role_key, role in local.roles :
      "${group.path}-${role_key}" => {
        module_name = group.name
        role_name   = "App.Application.PROD.${group.path}.${role.name}"
        shop_name   = "${group.path} - ${role.name}"
        order_for   = role.order_for
        approval_flow = role.approval_flow
        description = "Role for ${group.path} ${role.name}"
      }
    }
  ]...)
}
