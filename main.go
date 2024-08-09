package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/shirou/gopsutil/v4/process"
)

func getProcessName(pid int32) string {
	p, err := process.NewProcess(pid)
	if err != nil {
		color.Red("Error: 获取进程错误 %s", err)
		return ""
	}

	// 获取进程名称
	name, err := p.Cmdline()
	if err != nil {
		color.Red("Error: 获取进程名称错误 %s", err)
	}
	return name
}

func createTable(data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"PROCESS NAME", "PROCESS ID", "CONTAINER ID", "CONTAINER NAME"})

	for _, v := range data {
		table.Append(v)
	}
	table.SetHeaderColor(tablewriter.Colors{tablewriter.FgGreenColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgBlueColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgYellowColor},
		tablewriter.Colors{tablewriter.FgMagentaColor})

	table.SetColumnColor(tablewriter.Colors{tablewriter.FgGreenColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgBlueColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgYellowColor},
		tablewriter.Colors{tablewriter.FgMagentaColor})

	table.Render()
}

func getDockerNameByID(dockerID string) (string, error) {
	// 创建 Docker 客户端
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	defer func(cli *client.Client) {
		err := cli.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(cli)

	// 根据 ID 获取容器信息
	containerJSON, err := cli.ContainerInspect(context.Background(), dockerID)
	if err != nil {
		return "", err
	}
	name := strings.TrimPrefix(containerJSON.Name, "/")
	// 返回容器名称
	return name, nil
}

func getDockerID(pid int32) (string, error) {
	// 构建 cgroup 文件的路径
	cgroupFilePath := fmt.Sprintf("/proc/%d/cgroup", pid)
	// 读取 cgroup 文件
	content, err := os.ReadFile(cgroupFilePath)
	// 匹配文件或目录不存在的错误
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf(fmt.Sprintf("请检查PID为%d的进程是否存在", pid))
			// 在这里处理文件或目录不存在的情况
		}
		// 处理其他类型的错误
		color.Red("Error: %s", err)
	}
	//   判断content中是否包含docker
	if strings.Contains(string(content), "docker-") {
		// 正则匹配cgroup中的docker id
		re := regexp.MustCompile(`[0-9a-fA-F]{64}`)
		dockerID := re.FindString(string(content))
		return dockerID, nil
	}
	return "", fmt.Errorf(fmt.Sprintf("PID为%d的进程不是docker进程", pid))
}

func stringToInt32(s string) (int32, error) {
	// 将字符串转换为 int64
	i64, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(fmt.Sprintf("请输入正确的PID，输入为： %s", s))
	}
	// 将 int64 转换为 int32
	i32 := int32(i64)
	return i32, nil
}

func run(pid string) {
	var data [][]string
	for _, v := range strings.Split(pid, ",") {
		pid, err := stringToInt32(v)
		if err != nil {
			color.Red("Error: %s", err)
			return
		}
		dockerID, err := getDockerID(pid)
		if err != nil {
			color.Red("Error: %s", err)
			continue // return
		}
		name, err := getDockerNameByID(dockerID)
		if err != nil {
			color.Red("Error: %s", err)
			return
		}
		data = append(data, []string{getProcessName(pid), strconv.FormatInt(int64(pid), 10), dockerID[:10], name})
	}
	createTable(data)
}

func main() {
	var pid string
	fmt.Println("说明：根据你输入的进程 ID (PID)，找到对应的 Docker 容器")
	fmt.Print("请输入PID，多个PID用英文逗号分隔：")
	_, err := fmt.Scanf("%s", &pid)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	run(pid)
}
