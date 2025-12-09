module "object_storage" {
  source = "./modules/object-storage"

  compartment_ocid = var.compartment_ocid
  project_name     = var.project_name
  environment      = var.environment
  tags             = var.tags
}
