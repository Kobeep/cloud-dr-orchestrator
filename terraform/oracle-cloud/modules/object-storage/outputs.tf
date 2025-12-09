output "bucket_name" {
  description = "Name of the Object Storage bucket"
  value       = oci_objectstorage_bucket.dr_backups.name
}

output "bucket_namespace" {
  description = "Object Storage namespace"
  value       = data.oci_objectstorage_namespace.ns.namespace
}

output "bucket_id" {
  description = "OCID of the bucket"
  value       = oci_objectstorage_bucket.dr_backups.id
}
