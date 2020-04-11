# HyperCloud 이미지 제공

## Background
### 볼륨
볼륨은 VM을 띄울 수 있는 디스크입니다. VM 내에서 파일 쓰기가 발생하면 볼륨에 데이터가 직접 쓰여집니다. 볼륨은 PVC로써 관리됩니다.

### 이미지
이미지는 볼륨을 생성할 때 사용되는 디스크 템플릿입니다. 이미지로는 VM을 생성할 수 없으며 VM을 생성하기 위해서는 반드시 이미지로부터 볼륨을 생성해야 합니다. 이미지로부터 볼륨을 생성하면 내부적으로 이미지를 read-only로 관리해서 볼륨의 변경 부분만이 추가로 기록됩니다.

### 스냅샷
스냅샷은 볼륨의 특정 시점을 저장하는 것을 뜻합니다. 필요할 경우 볼륨에 롤백 명령을 통해 해당 시점으로 돌아갈 수 있습니다. 스냅샷은 [Kubernetes Volume Snapshot](https://kubernetes.io/docs/concepts/storage/volume-snapshots/) 메카니즘을 사용합니다.

### 클론
클론은 볼륨 혹은 이미지로부터 새로운 볼륨을 생성하는 것을 뜻합니다. 클론은 [CSI Volume Cloning](https://kubernetes.io/docs/concepts/storage/volume-pvc-datasource/) 메카니즘을 사용합니다.

## 사용자 스토리
1. 사용자는 원하는 OS 이미지를 선택합니다. 이 과정을 이미지 스테이징 과정이라고 합니다. 파일 형태는 qcow2와 raw를 지원하며 소스는 HTTP와 객체 스토리지를 지원합니다.
- HTTP
  - 사용자는 OS 이미지의 http 주소를 지정할 수 있습니다.
- 객체 스토리지
  - HyperCloud에서는 이미지 업/다운로드를 위한 객체 스토리지를 제공합니다. 이 객체 스토리지는 aws S3와 호환되기 때문에 awscli 혹은 HTTP 메소드를 통해 이미지를 업로드할 수 있습니다.
2. `VirtualMachineImage` CR을 생성합니다. 아래 CRD는 예제이며 실제 CRD 구현이 업데이트 되면 아래 예제도 업데이트 됩니다. VirtualMachineImage는 HTTP 혹은 objectStore로부터 이미지를 다운받아서 자체적으로 관리하는 PVC를 생성합니다. 이미지를 임포트 한다는 측면에서는 CDI의 DataVolume과 비슷하지만, VM을 직접 띄울 수 있는 DataVolume과 달리 VirtualMachineImage로는 VM을 생성할 수 없습니다.

```yaml
apiVersion: hypercloud.tmaxanc.com/v1alpha1
kind: VirtualMachineImage
metadata:
  name: ubuntu18.04
spec:
  storageClassName: cephrbd
  source:
    # HTTP로부터 이미지를 가져옵니다. objectStore와 동시에 사용할 수 없습니다.
    http: https://cloud-images.ubuntu.com/bionic/current/bionic-server-cloudimg-amd64.img
    # 객체 스토리지로부터 이미지를 가져옵니다. http와 동시에 사용할 수 없습니다.
    objectStore:
      bucketName: VmImage
      objectName: bionic-server-cloudimg-amd64.img
```

3. 사용자는 다음과 같이 이미지의 목록을 확인할 수 있습니다.

```yaml
$ kubectl get virtualmachineimage
NAMESPACE  NAME
default    ubuntu18.04
```

4. VM 생성을 위해서는 이미지를 이용해서 볼륨을 생성해야 합니다. 기존의 CDI에 DataVolume 객체와 비슷한 역할을 하기 떄문에 DataVolume을 수정할 수도 있지만 가능하면 CDI를 완전히 제거하는 것을 목표로 합니다. `virtualMachineImageName`의 PVC를 가져와서 해당 PVC를 k8s pvc cloning을 통해 복제합니다.

```yaml
apiVersion: hypercloud.tmaxanc.com/v1alpha1
kind: VirtualMachinePvc
metadata:
  name: vmRootDisk
spec:
  virtualMachineImageName: ubuntu18.04
  pvc:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
```

5. 사용자는 생성된 PVC를 사용해서 VM을 생성합니다. VM의 디스크 스냅샷을 위해서는 해당 PVC의 스냅샷 기능을 사용합니다. VM을 복제할 경우 PVC 클론 기능을 사용합니다.

6. 사용자가 디스크 Export를 원할 경우 `VirtualMachineVolumeExport` CR을  생성해야 합니다. 오퍼레이터는 해당 pvc의 디스크 이미지를 qcow2로 변환하여 objectStore에 지정된 객체 스토리지에 업로드합니다. 사용자는 필요할 경우 해당 객체 스토리지에 awscli 혹은 HTTP GET 메소드를 통해 이미지를 다운로드 할 수 있습니다.

```yaml
apiVersion: hypercloud.tmaxanc.com/v1alpha1
kind: VirtualMachineVolumeExport
metadata:
  name: export
spec:
  pvcName: mydisk
  objectStore:
    bucketName: VmImage
    objectName: mydisk
```

## 구현 세부 사항
이 프로젝트는 기존의 CDI를 대체하지만 Kubevirt와는 완전히 통합되어야 합니다. 따라서 Kubevirt의 소스코드 수정은 없으며 Kubevirt와는 오로지 PVC를 통해 연결됩니다. 개발이 어느 정도 진행되고 유용함이 확인되면 Kubevirt의 공식 레파지토리에 편입되어 CDI를 완전히 대체하는 것을 목표로 합니다.
