# The spec list differs per cloud/region and may skip sizes, so look it up
# instead of hardcoding a size that only exists elsewhere.
data "relyt_dps_specs" "hybrid" {
  engine = "hybrid" # hybrid | extreme
  cloud  = "your_cloud_id"
  region = "your_region_id"
  # edition defaults to "standard"
}

resource "relyt_dwsu" "dwsu" {
  cloud  = "your_cloud_id"
  region = "your_region_id"
  domain = "your-subdomain-prefix"
  default_dps = {
    name   = "hdps"
    engine = "hybrid"
    size   = data.relyt_dps_specs.hybrid.specs[0].name
  }
}
