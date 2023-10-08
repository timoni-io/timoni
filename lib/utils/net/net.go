package net

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
)

const DNSTimeout = 100 * time.Millisecond

func GetFile(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get file: %d", res.StatusCode)
	}

	file, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func GetFileWriter(url string, w io.Writer) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get file: %d", res.StatusCode)
	}

	_, err = io.Copy(w, res.Body)
	if err != nil {
		return err
	}
	return nil
}

// Get preferred outbound ip of this machine
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return ""
	}
	return localAddr.IP.String()
}

func RequestIP(r *http.Request) string {
	if prior := r.Header.Get("X-Forwarded-For"); prior != "" {
		proxies := strings.Split(prior, ",")
		if len(proxies) > 0 {
			return strings.Trim(proxies[0], " ")
		}
	}

	// X-Real-Ip is less supported, but worth checking in the
	// absence of X-Forwarded-For
	if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		return realIP
	}

	return strings.Split(r.RemoteAddr, ":")[0]
}

func DNSLookup(ip string) string {
	if ip == "" {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), DNSTimeout)
	defer cancel()

	var r net.Resolver

	addrs, err := r.LookupAddr(ctx, ip)
	if err != nil {
		return ""
	}
	return addrs[0]
}

func DNSResolve(url string) []string {
	if url == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), DNSTimeout)
	defer cancel()

	var r net.Resolver

	addrs, err := r.LookupIP(ctx, "ip4", url)
	if err != nil {
		return nil
	}

	ips := make([]string, len(addrs))
	for i, addr := range addrs {
		ips[i] = addr.String()
	}
	sort.Strings(ips)
	return ips
}

// LocalIP returns the non loopback local IP of the host
func LocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
