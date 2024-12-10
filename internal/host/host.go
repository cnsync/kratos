package host

import (
	"fmt"
	"net"
	"strconv"
)

// ExtractHostPort 从地址中提取主机名和端口号
func ExtractHostPort(addr string) (host string, port uint64, err error) {
	var ports string
	// 使用 net.SplitHostPort 函数将地址分割为主机名和端口号
	host, ports, err = net.SplitHostPort(addr)
	// 如果发生错误，返回错误
	if err != nil {
		return
	}
	// 将端口号字符串转换为 uint64 类型
	port, err = strconv.ParseUint(ports, 10, 16) //nolint:mnd
	return
}

// isValidIP 检查给定的地址是否是有效的全局单播 IP 地址
func isValidIP(addr string) bool {
	// 使用 net.ParseIP 函数解析地址
	ip := net.ParseIP(addr)
	// 检查 IP 地址是否为全局单播地址且不是接口本地多播地址
	return ip.IsGlobalUnicast() && !ip.IsInterfaceLocalMulticast()
}

// Port 从监听器中获取实际的端口号
func Port(lis net.Listener) (int, bool) {
	// 检查监听器的地址是否为 *net.TCPAddr 类型
	if addr, ok := lis.Addr().(*net.TCPAddr); ok {
		// 返回端口号和 true 表示成功
		return addr.Port, true
	}
	// 返回 0 和 false 表示失败
	return 0, false
}

// Extract 从给定的主机端口字符串和监听器中提取私有地址和端口号
func Extract(hostPort string, lis net.Listener) (string, error) {
	// 使用 net.SplitHostPort 函数将主机端口字符串分割为主机名和端口号
	addr, port, err := net.SplitHostPort(hostPort)
	// 如果发生错误且监听器为 nil，返回错误
	if err != nil && lis == nil {
		return "", err
	}
	// 如果监听器不为 nil
	if lis != nil {
		// 从监听器中获取端口号
		p, ok := Port(lis)
		// 如果获取失败，返回错误
		if !ok {
			return "", fmt.Errorf("failed to extract port: %v", lis.Addr())
		}
		// 将端口号转换为字符串
		port = strconv.Itoa(p)
	}
	// 如果主机名不为空且不是特殊地址
	if len(addr) > 0 && (addr != "0.0.0.0" && addr != "[::]" && addr != "::") {
		// 返回拼接后的主机名和端口号
		return net.JoinHostPort(addr, port), nil
	}
	// 获取本地网络接口列表
	ifaces, err := net.Interfaces()
	// 如果发生错误，返回错误
	if err != nil {
		return "", err
	}
	// 初始化最小索引值为最大整数
	minIndex := int(^uint(0) >> 1)
	// 初始化 IP 地址列表
	ips := make([]net.IP, 0)
	// 遍历网络接口列表
	for _, iface := range ifaces {
		// 检查网络接口是否已启用
		if (iface.Flags & net.FlagUp) == 0 {
			continue
		}
		// 如果网络接口索引大于最小索引且 IP 地址列表不为空，跳过当前接口
		if iface.Index >= minIndex && len(ips) != 0 {
			continue
		}
		// 获取网络接口的地址列表
		addrs, err := iface.Addrs()
		// 如果发生错误，跳过当前接口
		if err != nil {
			continue
		}
		// 遍历地址列表
		for i, rawAddr := range addrs {
			var ip net.IP
			// 根据地址类型获取 IP 地址
			switch addr := rawAddr.(type) {
			case *net.IPAddr:
				ip = addr.IP
			case *net.IPNet:
				ip = addr.IP
			default:
				continue
			}
			// 检查 IP 地址是否有效
			if isValidIP(ip.String()) {
				// 更新最小索引值
				minIndex = iface.Index
				// 如果是第一个有效 IP 地址，初始化 IP 地址列表
				if i == 0 {
					ips = make([]net.IP, 0, 1)
				}
				// 将有效 IP 地址添加到列表中
				ips = append(ips, ip)
				// 如果是 IPv4 地址，停止遍历
				if ip.To4() != nil {
					break
				}
			}
		}
	}
	// 如果找到了有效 IP 地址
	if len(ips) != 0 {
		// 返回最后一个有效 IP 地址和端口号
		return net.JoinHostPort(ips[len(ips)-1].String(), port), nil
	}
	// 返回空字符串和 nil 表示没有找到有效地址
	return "", nil
}
