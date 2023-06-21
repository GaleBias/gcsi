
# 静态存储

# k8s 内置动态存储

# csi spec
## 1.服务: identity,controller,node
## 2.插件能力: controllerGetCapabilities 和 nodeGetCapabilities
## 3.volume lifecycle: createVolume - controllerPublishVolume - nodeStageVolume - nodePublishVolume
## 4.增强功能: controllerExpandVolume, nodeExpandVolume
## 5.error scheme
## 6.部署架构
## 7.环境变量: CSI_ENDPOINT

# 步骤:
## 1.实现服务框架，运行后报错:CSI driver probe failed —— v0.0.1
## 2.实现identity服务的3个方法，运行后报错:Error getting CSI driver capabilities —— v0.0.2
## 3.实现controller服务的ControllerGetCapabilities方法，运行后报错:无法list-watch集群内相关资源 —— v0.0.3
## 4.创建rbac，不再报错 —— v0.0.3
## 5.在createVolume方法中打印信息，并在集群中创建sc，验证创建pvc之后，该方法被调用 —— v0.0.4
## 6.完善createVolume方法，调用云厂商创建evs，报错缺少events create/patch权限 —— v0.0.5
## 7.修改rbac，重新运行，报认证失败，错误如下：
 Post "https://evs.ap-southeast-3.myhuaweicloud.com/v2.1/2fb7832179084025b2eadab146ad3cb0/cloudvolumes": tls: failed to verify certificate: x509: certificate signed by unknown authority
## 8.尝试多次依旧不能解决此问题，修改基础镜像为centos7.9，问题解决，pvc创建成功，又报错:缺少pv create 权限 —— v0.0.8
## 9.修改rbac
## 10.创建pvc及pod，发现pvc、pv、云厂商磁盘创建成功，pod创建不成功，观察pod事件，等待几分钟，显示FailedMount及FailedAttachVolume
## 11.在deploy中添加external-attacher容器，并在ControllerPublishVolume方法中打印信息，观察到该方法被调用 —— v0.0.11
## 12.缺少list-watch csiNode的权限，修改rbac后该方法还是未被调用
## 13.将attacher容器日志-v=5，观察发现缺少patch volumeattachments/status 和 persistentvolumes的权限
## 14.日志报错Can't get nodeID from CSINode，证明需先安装node-driver-registrar
## 15.安装完成后观察日志，需要实现NodeGetInfo和NodeGetCapabilities方法 —— v0.0.13