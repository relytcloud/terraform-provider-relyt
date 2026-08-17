data "relyt_clouds" "all" {}

data "relyt_cloud_regions" "regions" {
  cloud = data.relyt_clouds.all.clouds[0].id
}

output "region_ids" {
  value = [for r in data.relyt_cloud_regions.regions.regions : r.id]
}
