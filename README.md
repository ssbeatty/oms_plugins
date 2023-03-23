# oms_plugins

## 介绍
[oms](https://github.com/ssbeatty/oms)的插件示例

### 目录组织
每个文件夹是一个插件的目录
```text
vnc_install              # 可以自动在目标机器安装x11vnc
```

## 原理
基于https://github.com/traefik/yaegi  实现的go代码动态解释

> 怎么安装？
1. 将插件目录zip或者tar.gz打包在界面上上传
2. 直接将插件目录放入`data/plugin/src/下`，根据当前的data目录位置可能有所不同，详细请看oms文档
3. 重启oms服务（每次修改了插件的内容也要重启oms服务）

> 为什么这样设计？
1. 一些不通用的插件没必要进行二次开发oms，能复用即可
2. 这样设计可以减少开发量

## 如何开发
代码请参考vnc_install下的内容

manifest.yaml是插件的元文件，需要放在插件目录下
```yaml
name: vnc_install        # 插件的名称（oms全局唯一）
import: vnc_install      # 插件的导入package
```

> 关于import

如果你的目录结构为`/data/plugin/src/github.com/someone/package`

则填写`github.com/someone/package`

其原理等同于go path的导入规则

### 代码里面包含的symbols
1. 通用的golang模块，但是不包含`os/exec`等危险操作，[ISSUE](https://github.com/traefik/yaegi/issues/1160)
2. 插件必要的包
```go
//go:generate yaegi extract github.com/ssbeatty/oms/pkg/transport
//go:generate yaegi extract github.com/ssbeatty/oms/pkg/types
//go:generate yaegi extract github.com/ssbeatty/jsonschema
```
这意味着导入安全的通用模块和上面的三个包是不需要依赖的，但是如果使用了其他依赖，请加入vendor中


## 怎样离线安装apt的包？

### ubuntu
参考[wine](https://wiki.winehq.org/Ubuntu_zhcn)

```shell
sudo apt-get clean
sudo apt-get --download-only install lightdm x11vnc
cp -R /var/cache/apt/archives/ /media/deb-pkgs/

cd /media/deb-pkgs/
sudo dpkg -i *.deb
```

### centos

```shell
yum -y install yum-utils

mkdir /media/yum-pkgs/ && cd /media/yum-pkgs/
repotrack x11vnc

rpm -Uvh --force --nodeps *.rpm
```
