provider google {
    project = "census-rm-apolloakora04"
    region  = "europe-west2"
}

resource random_id postfix {
    byte_length = 4
}

resource google_storage_bucket cloud-build-bucket-no1 {
    name                        = "my-cloud-build-bucket-${random_id.postfix.hex}"
    storage_class               = "REGIONAL"
    uniform_bucket_level_access = true
}
