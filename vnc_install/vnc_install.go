package vnc_install

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ssbeatty/oms/pkg/transport"
	"github.com/ssbeatty/oms/pkg/types"
	"strings"
)

const (
	StepName        = "vnc_install"
	templateService = `[Unit]
Description=x11vnc service
After=display-manager.service network.target syslog.target
StartLimitBurst=2
StartLimitIntervalSec=150s
[Service]
User=root
Type=idle
ExecStart=/usr/bin/x11vnc -forever -display :%d -auth %s %s
ExecStop=/usr/bin/killall x11vnc
Restart=on-failure
RestartSec=10
[Install]
WantedBy=multi-user.target`
)

func New() types.Step {
	return &VNCInstallStep{
		buffer: &bytes.Buffer{},
	}
}

// VNCInstallStep 安装vnc
type VNCInstallStep struct {
	id     string
	cfg    *VNCInstallStepConfig
	buffer *bytes.Buffer
}

type VNCInstallStepConfig struct {
	VNCPassWord string `json:"vnc_pass_word" jsonschema:"required=true" jsonschema_description:"VNC密码"`
	VNCDisplay  int    `json:"vnc_display"  jsonschema:"required=true,default=0" jsonschema_description:"VNC Display Port, 默认: 0"`
	Auth        string `json:"auth" jsonschema:"required=true,default=guess" jsonschema_description:"VNC Auth, Default guess"`
}

func (bs *VNCInstallStep) SetID(id string) {
	bs.id = id
}

func (bs *VNCInstallStep) ID() string {
	return bs.id
}

func (bs *VNCInstallStep) Exec(session *transport.Session, sudo bool) ([]byte, error) {
	defer bs.buffer.Reset()

	err := bs.pluginExec(session.Client)
	if err != nil {
		return bs.buffer.Bytes(), err
	}

	return bs.buffer.Bytes(), nil
}

func (bs *VNCInstallStep) Create(conf []byte) (types.Step, error) {
	cfg := &VNCInstallStepConfig{}

	err := json.Unmarshal(conf, cfg)
	if err != nil {
		return nil, err
	}
	return &VNCInstallStep{
		cfg:    cfg,
		buffer: &bytes.Buffer{},
	}, nil
}

func (bs *VNCInstallStep) Name() string {
	return StepName
}

func (bs *VNCInstallStep) Desc() string {
	return "安装vnc"
}

func (bs *VNCInstallStep) Config() interface{} {
	return bs.cfg
}

func (bs *VNCInstallStep) GetSchema() (interface{}, error) {

	return types.GetSchema(bs.cfg)
}

func (bs *VNCInstallStep) printMsg(msg string) {
	fmt.Fprintf(bs.buffer, "%s\r\n", msg)
}

func (bs *VNCInstallStep) runCommand(c *transport.Client, cmd string, sudo bool) ([]byte, error) {

	var (
		output []byte
		err    error
	)

	session, err := c.NewPty()
	if err != nil {
		return nil, err
	}

	defer session.Close()

	if sudo {
		output, err = session.Sudo(cmd, c.Conf.Password)
	} else {
		output, err = session.Output(cmd)
	}
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (bs *VNCInstallStep) runCommandNoPty(c *transport.Client, cmd string, sudo bool) ([]byte, error) {

	var (
		output []byte
		err    error
	)

	session, err := c.NewSession()
	if err != nil {
		return nil, err
	}

	defer session.Close()

	if sudo {
		output, err = session.Sudo(cmd, c.Conf.Password)
	} else {
		output, err = session.Output(cmd)
	}
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (bs *VNCInstallStep) getOsReleaseVersion(c *transport.Client) (string, error) {
	output, err := bs.runCommand(c, "cat /etc/os-release", false)
	if err != nil {
		return "", err
	}

	if strings.Contains(string(output), "Ubuntu") {
		return "ubuntu", nil
	} else if strings.Contains(string(output), "CentOS Linux 7") {
		return "centos7", nil
	} else {
		return "", errors.New("不支持的发行版")
	}
}

func (bs *VNCInstallStep) ubuntuInstallVnc(c *transport.Client) error {
	bs.printMsg("开始安装依赖...")
	_, err := bs.runCommandNoPty(c, "DEBIAN_FRONTEND=noninteractive dpkg -i .oms/ubuntu/*.deb", true)
	if err != nil {
		return err
	}

	bs.printMsg("指定lightdm默认图形界面")

	// https://askubuntu.com/questions/1114525/reconfigure-the-display-manager-non-interactively
	_, err = bs.runCommand(c, "bash -c 'echo \"/usr/sbin/lightdm\" > /etc/X11/default-display-manager'", true)
	if err != nil {
		return err
	}

	_, err = bs.runCommand(c, "DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true dpkg-reconfigure lightdm", true)
	if err != nil {
		return err
	}

	_, err = bs.runCommand(c, "bash -c 'echo \"set shared/default-x-display-manager lightdm\" | debconf-communicate'", true)
	if err != nil {
		return err
	}

	bs.printMsg("指定lightdm默认图形界面成功\r\n")

	return nil
}

// todo 其他发行版

func (bs *VNCInstallStep) registerService(c *transport.Client) error {
	servicePath := "/lib/systemd/system/x11vnc.service"

	passwd := "-passwd %s"
	if bs.cfg.VNCPassWord != "" {
		passwd = fmt.Sprintf(passwd, bs.cfg.VNCPassWord)
	} else {
		passwd = ""
	}
	if bs.cfg.Auth == "" {
		bs.cfg.Auth = "guess"
	}

	err := c.UploadFileRaw(fmt.Sprintf(templateService, bs.cfg.VNCDisplay, bs.cfg.Auth, passwd), ".oms/x11vnc.service")
	if err != nil {
		return err
	}

	output, err := bs.runCommand(c, fmt.Sprintf("cp .oms/x11vnc.service %s", servicePath), true)
	if err != nil {
		return err
	}

	output, err = bs.runCommand(c, "bash -c 'systemctl daemon-reload && systemctl enable x11vnc.service'", true)
	if err != nil {
		return err
	}

	bs.printMsg(string(output))

	return nil
}

func (bs *VNCInstallStep) clear(c *transport.Client) {
	_, _ = bs.runCommand(c, "rm -rf .oms", false)
}

func (bs *VNCInstallStep) pluginExec(c *transport.Client) error {

	if c.GetTargetMachineOs() == transport.GOOSWindows {
		return errors.New("暂不支持windows")
	}

	err := c.NewSftpClient()
	if err != nil {
		return err
	}

	release, err := bs.getOsReleaseVersion(c)
	if err != nil {
		return err
	}
	bs.printMsg("开始上传文件...")

	// todo how to get path
	err = c.UploadFile(
		fmt.Sprintf("data/plugin/data/vnc_install/files/%s.zip", release), fmt.Sprintf(".oms/%s.zip", release), "")
	if err != nil {
		return err
	}
	bs.printMsg("上传文件成功")

	output, err := bs.runCommand(c, fmt.Sprintf("unzip -o -d .oms .oms/%s.zip", release), false)
	if err != nil {
		return err
	}

	bs.printMsg(string(output))

	switch release {
	case "ubuntu":
		err = bs.ubuntuInstallVnc(c)
	default:
		return errors.New(fmt.Sprintf("暂不支持发行版: %s", release))
	}

	if err != nil {
		return err
	}

	err = bs.registerService(c)
	if err != nil {
		return err
	}

	bs.printMsg("注册服务成功, 开始清理缓存...")

	bs.clear(c)

	bs.printMsg("重启...")

	_, _ = bs.runCommand(c, "reboot", true)

	return nil
}
