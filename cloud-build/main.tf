
provider "google" {
  project = "census-rm-apolloakora04"
  region  = "europe-west2"
}

resource "google_storage_bucket" "cloud-build-bucket" {

  name                        = "cloud-build-bucket-01"
  project                     = "census-rm-apolloakora04"
  storage_class               = "REGIONAL"
  location                    = "europe-west2"
  uniform_bucket_level_access = true

}
