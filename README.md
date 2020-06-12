# Kubevirt-Image-Service ![kubevirt-image-service](https://github.com/tmax-cloud/kubevirt-image-service/workflows/kubevirt-image-service/badge.svg)


## What is Kubevirt-Image-Service?

Kubevirt-Image-Service (KIS) is an integration wtih [KubeVirt](https://github.com/kubevirt/kubevirt) to provide VM image and volume services. [CDI](https://github.com/kubevirt/containerized-data-importer) (Containerized-Data-Importer) project is exists for the same reason, but it is designed for using non sparse disk. Instead of using one raw disk per PVC, this project is designed to use qcow2 disk to minimize disk space take up by VM images. KIS is aim to deploy multiple VMs fast and efficiently by using little bit of storage space as possible. 

## Getting Started

For installation, deployment and further explanation, see our [user guide](docs/USERGUIDE.md).

## Contributing

We welcome contributions whether you are part of [Tmax Cloud](https://github.com/tmax-cloud) or not. See [Contributing](CONTRIBUTING.md) for more details. 

## Report a Bug

Before reporting a bug, please take a look at our [trouble shooting guide](docs/TROUBLESHOOTING.md). If you still have the same problem or if you find a bug that's not mentiond on our page, file it [here](https://github.com/tmax-cloud/kubevirt-image-service/issues). Suggest enhancements, or request new featuers as well. If you are part of Tmax Cloud, you can also open an issue at [IMS](https://ims.tmaxsoft.com/).
