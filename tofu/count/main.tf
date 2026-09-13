# Nothing is created: `tofu plan` only. terraform_data is built into OpenTofu, so no provider is downloaded.
resource "terraform_data" "source" {}

locals {
  # attribute `a` is unknown until apply; attribute `b` definitely differs,
  # so the two objects can never be equal
  left  = { a = terraform_data.source.id, b = "z" }
  right = { a = "x", b = "y" }
}

# The comparison is definitely false, so this should always plan with count = 0.
resource "terraform_data" "only_if_equal" {
  count = local.left == local.right ? 1 : 0
}
