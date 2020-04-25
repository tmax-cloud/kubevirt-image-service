# kubevirt-image-service
## kubevirt-image-service란?
kubevirt-image-service(이하 kis)는 kubevirt에 통합되는 이미지, 볼륨 서비스를 제공합니다. (TODO: 데모)

- `VirtualMachineImage`  : HTTP, S3, Local을 통해 외부에서 이미지를 임포트할 수 있습니다. (현재는 HTTP만 지원) 한 번 임포트된 이미지는 외부에서 다시 클론되지 않으며 내부에서 자체적으로 관리됩니다.
- `VirtualMachineVolume` : 볼륨은 이미지로부터 생성할 수 있으며 CSI(Container Storage Interface)의 스냅샷-클론 기능을 활용하여 용량을 효율적으로 관리합니다.
- `VirtualMachineExport` : 사용하던 볼륨을 qcow2 파일로 변환하여 외부에 전달할 수 있습니다.

## 시작하기
다음 [유저 가이드](docs/USERGUIDE.md)를 참조하세요.

## 컨트리뷰팅
우리는 팀/연구소와 관계 없이 자유로운 기여를 환영합니다. 자세한 내용은 [기여하기](CONTRIBUTING.md)를 참조하세요.

## 버그 리포팅
버그를 리포트하기 전에 먼저 [트러블 슈팅](docs/TROUBLESHOOTING.md)페이지를 확인해주세요. 트러블 슈팅 페이지를 통해 해결되지 않는 버그는 IMS로 리포팅할 수 있습니다. (TODO: ims 추가 및 링크)
