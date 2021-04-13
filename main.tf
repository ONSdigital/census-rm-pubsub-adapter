provider google {
    region  = "europe-west2"
}

resource google_storage_bucket cloud-build-bucket {
    name                        = "my-cloud-build-bucket"
    storage_class               = "REGIONAL"
    location                    = "europe-west2"
    uniform_bucket_level_access = true
}

resource google_compute_instance cloud-build-vm {
    name         = "my-cloud-build-vm"
    machine_type = "f1-micro"

    boot_disk {
        initialize_params {
            image = "debian-cloud/debian-9"
        }
    }

    network_interface {
        network = google_compute_network.cloud-build-net.self_link
        access_config {
        }
    }
}

resource google_compute_network cloud-builid-net {
    name                    = "my-cloud-build-net"
    auto_create_subnetworks = "true"
}
