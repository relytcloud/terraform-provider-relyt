output "matched_dwsu" {
  value = one([for i, v in data.relyt_dwsus.dwsus.dwsu_list : v if v.alias == var.alias])
}