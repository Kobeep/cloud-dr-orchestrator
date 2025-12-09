resource "oci_objectstorage_bucket" "dr_backups" {
  compartment_id = var.compartment_ocid
  namespace      = data.oci_objectstorage_namespace.ns.namespace
  name           = "${var.project_name}-${var.environment}-backups"
  access_type    = "NoPublicAccess"

  versioning = "Enabled"

  freeform_tags = var.tags
}

data "oci_objectstorage_namespace" "ns" {
  compartment_id = var.compartment_ocid
}
