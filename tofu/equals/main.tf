# Nothing is created: `tofu plan` only. terraform_data is built into OpenTofu, so no provider is downloaded.
resource "terraform_data" "source" {}

locals {
  # attribute `a` is unknown until apply; attribute `b` definitely differs
  left  = { a = terraform_data.source.id, b = "z" }
  right = { a = "x", b = "y" }
}

output "objects_equal" {
  value = local.left == local.right
}

output "maps_equal" {
  value = tomap(local.left) == tomap(local.right)
}
