package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"oms_plugin_vnc/transport"
	"os"
	"strings"
)

func init() {
	execCmd.PersistentFlags().StringVarP(&clients, "client", "", "", "Client Config")
	execCmd.PersistentFlags().StringVarP(&params, "params", "", "", "Schema Params")
}

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "plugin exec",
	Args:  cobra.MinimumNArgs(0),
	Long:  `plugin exec`,
	Run: func(cmd *cobra.Command, args []string) {
		pluginExec(args)
	},
}

func printMsg(msg string) {
	fmt.Fprintf(os.Stdout, "%s\r\n", msg)
}

const TemplateService = `[Unit]
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

func runCommand(c *transport.Client, cmd string, sudo bool) ([]byte, error) {

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

func runCommandNoPty(c *transport.Client, cmd string, sudo bool) ([]byte, error) {

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

func getOsReleaseVersion(c *transport.Client) (string, error) {
	output, err := runCommand(c, "cat /etc/os-release", false)
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

func ubuntuInstallVnc(c *transport.Client) error {
	printMsg("开始安装依赖...")
	_, err := runCommandNoPty(c, "DEBIAN_FRONTEND=noninteractive dpkg -i .oms/ubuntu/*.deb", true)
	if err != nil {
		return err
	}

	printMsg("指定lightdm默认图形界面")

	// https://askubuntu.com/questions/1114525/reconfigure-the-display-manager-non-interactively
	_, err = runCommand(c, "bash -c 'echo \"/usr/sbin/lightdm\" > /etc/X11/default-display-manager'", true)
	if err != nil {
		return err
	}

	_, err = runCommand(c, "DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true dpkg-reconfigure lightdm", true)
	if err != nil {
		return err
	}

	_, err = runCommand(c, "bash -c 'echo \"set shared/default-x-display-manager lightdm\" | debconf-communicate'", true)
	if err != nil {
		return err
	}

	fmt.Printf("指定lightdm默认图形界面成功\r\n")

	return nil
}

// todo 其他发行版

func registerService(c *transport.Client, params Params) error {
	servicePath := "/lib/systemd/system/x11vnc.service"

	passwd := "-passwd %s"
	if params.VNCPassWord != "" {
		passwd = fmt.Sprintf(passwd, params.VNCPassWord)
	} else {
		passwd = ""
	}
	if params.Auth == "" {
		params.Auth = "guess"
	}

	err := c.UploadFileRaw(fmt.Sprintf(TemplateService, params.VNCDisplay, params.Auth, passwd), ".oms/x11vnc.service")
	if err != nil {
		return err
	}

	output, err := runCommand(c, fmt.Sprintf("cp .oms/x11vnc.service %s", servicePath), true)
	if err != nil {
		return err
	}

	output, err = runCommand(c, "bash -c 'systemctl daemon-reload && systemctl enable x11vnc.service'", true)
	if err != nil {
		return err
	}

	printMsg(string(output))

	return nil
}

func clear(c *transport.Client) {
	_, _ = runCommand(c, "rm -rf .oms", false)
}

func pluginExec(args []string) {
	var (
		param  Params
		client ClientConfig
	)

	err := json.Unmarshal([]byte(clients), &client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析ssh参数出错, err: %v\r\n", err)
		os.Exit(-1)
	}
	err = json.Unmarshal([]byte(params), &param)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析plugin参数出错, err: %v\r\n", err)
		os.Exit(-1)
	}

	c, err := transport.New(client.Host, client.User, client.Password, client.Passphrase, client.KeyBytes, client.Port)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "创建ssh客户端出错, err: %v\r\n", err)
		os.Exit(-1)
	}

	if c.GetTargetMachineOs() == transport.GOOSWindows {
		_, _ = fmt.Fprintf(os.Stderr, "暂不支持windows\r\n")
		os.Exit(-1)
	}

	err = c.NewSftpClient()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "创建sftp客户端出错, err: %v\r\n", err)
		os.Exit(-1)
	}

	release, err := getOsReleaseVersion(c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "解析发行版出错, err: %v\r\n", err)
		os.Exit(-1)
	}
	printMsg("开始上传文件...")

	err = c.UploadFile(
		fmt.Sprintf("files/%s.zip", release), fmt.Sprintf(".oms/%s.zip", release), "")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "上传文件出错, err: %v\r\n", err)
		os.Exit(-1)
	}
	printMsg("上传文件成功")

	output, err := runCommand(c, fmt.Sprintf("unzip -o -d .oms .oms/%s.zip", release), false)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "解压失败, err: %v, ouput: %s\r\n", err, output)
		os.Exit(-1)
	}

	printMsg(string(output))

	switch release {
	case "ubuntu":
		err = ubuntuInstallVnc(c)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "暂不支持发行版: %s\r\n", release)
		os.Exit(-1)
	}
	err = ubuntuInstallVnc(c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "安装失败, err: %v\r\n", err)
		os.Exit(-1)
	}

	err = registerService(c, param)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "注册服务失败, err: %v\r\n", err)
		os.Exit(-1)
	}

	printMsg("注册服务成功, 开始清理缓存...")

	clear(c)

	printMsg("重启...")

	_, _ = runCommand(c, "reboot", true)
}
