

resource "relyt_dps" "abc" {
  dwsu_id     = "dwsu-id-from-an-duws-resource"
  name        = "edps-exp"
  description = "An EDPS Example" #optional
  engine      = "extreme"
  size        = "S"
}