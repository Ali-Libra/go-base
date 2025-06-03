package util

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

func HandleSignalFinal(quitHandler func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	<-c
	quitHandler()
}

func HandleSignal(quitHandler func(ctx context.Context), ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	go func() {
		<-c
		if quitHandler != nil {
			quitHandler(ctx)
		}
		wg.Done()
	}()

	wg.Wait()
}

// 取范围随机数[min, max)
func Random(min, max int) int {
	//	rand.Seed(time.Now().Unix())
	if min >= max {
		return min
	}
	return rand.Intn(max-min) + min
}

func RandomClosed(min, max int) int {
	//	rand.Seed(time.Now().Unix())
	if min >= max {
		return min
	}
	return rand.Intn(max+1-min) + min
}

func RandomKey(keyLen uint32) string {
	var key strings.Builder
	for i := uint32(0); i < keyLen; i++ {
		key.WriteString(strconv.Itoa(RandomClosed(1, 9)))
	}
	return key.String()
}

var letter = []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}

func RandomLetterKey(keyLen uint32) string {
	var key strings.Builder
	for i := uint32(0); i < keyLen; i++ {
		key.WriteByte(letter[Random(0, len(letter))])
	}
	return key.String()
}

func GetIpByNetCard(name string) (net.IP, error) {
	inter, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	addrs, err := inter.Addrs()
	if err != nil {
		return nil, err
	}

	// 获取IP地址，子网掩码
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ip.IP.To4() != nil {
				return ip.IP, nil
			}
		}
	}

	return nil, errors.New(name + " no ip can get")
}

func VerifyEmailFormat(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

func IsPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	defer l.Close()
	return true
}

func FindAvailablePort(startPort int, endPort int) int {
	for port := startPort; port < endPort; port++ {
		if IsPortAvailable(port) {
			return port
		}
	}
	return 0
}
