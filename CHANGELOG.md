# Release Note

## 1.0.0

### Features:

- Implement VirtualMachineImage CR to import image from HTTP source 
- Implement VirtualMachineVolume CR to create volume for VM from image
- Implement VirtualMachineExport CR to export volume to local destination

### Enhancements:

- Update operator-sdk version from v0.17.0 to v0.17.1
- Apply condition array to each CRs' status 

### Bug Fixs:

- Modify to use accessMode RWO when creating scratch pvc for VirtualMachineImage
