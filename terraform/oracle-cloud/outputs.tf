output "bucket_name" {
  description = "Name of the Object Storage bucket"
  value       = module.object_storage.bucket_name
}

output "bucket_namespace" {
  description = "Object Storage namespace"
  value       = module.object_storage.bucket_namespace
}

output "bucket_url" {
  description = "Object Storage bucket URL"
  value       = "https://objectstorage.${var.region}.oraclecloud.com/n/${module.object_storage.bucket_namespace}/b/${module.object_storage.bucket_name}"
}
