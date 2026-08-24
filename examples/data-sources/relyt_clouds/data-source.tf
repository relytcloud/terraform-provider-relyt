# Discover the cloud providers served by this control plane.
# Valid `cloud` values differ per deployment — query them, don't copy them
# from another environment's examples.
data "relyt_clouds" "all" {}

output "cloud_ids" {
  value = [for c in data.relyt_clouds.all.clouds : c.id if c.is_available]
}
