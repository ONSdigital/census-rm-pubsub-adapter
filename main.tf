provider google {
    project = "census-rm-apolloakora04"
    region  = "europe-west2"
}

resource random_id postfix {
    byte_length = 4
}

resource google_storage_bucket cloud-build-bucket-no1 {
    name                        = "my-cloud-build-bucket-${random_id.postfix.hex}"
    project                     = "census-rm-apolloakora04"
    storage_class               = "REGIONAL"
    location                    = "europe-west2"
    uniform_bucket_level_access = true
}

resource google_compute_instance cloud-build-vm {
    name         = "my-cloud-build-vm-${random_id.postfix.hex}"
    machine_type = "f1-micro"
    project      = "census-rm-apolloakora04"
###    zone         = "europe-west2-a"

    boot_disk {
        initialize_params {
            image = "debian-cloud/debian-9"
        }
    }

    network_interface {
        network = "default"
        access_config {
        }
    }
#    network_interface {
#        network = google_compute_network.cloud-build-net.self_link
#        access_config {
#        }
#    }

}

#resource google_compute_network cloud-builid-net {
#    name                    = "my-cloud-build-net"
#    auto_create_subnetworks = "true"
#}
