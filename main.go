package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/oschwald/maxminddb-golang"
)

type IPRecord struct {
	City struct {
		GeonameID uint32            `maxminddb:"geoname_id" json:"geoname_id,omitempty"`
		Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"city" json:"city,omitempty"`
	Continent struct {
		Code      string            `maxminddb:"code" json:"code,omitempty"`
		GeonameID uint32            `maxminddb:"geoname_id" json:"geoname_id,omitempty"`
		Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"continent" json:"continent,omitempty"`
	Country struct {
		GeonameID uint32            `maxminddb:"geoname_id" json:"geoname_id,omitempty"`
		ISOCode   string            `maxminddb:"iso_code" json:"iso_code,omitempty"`
		Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"country" json:"country,omitempty"`
	Location struct {
		AccuracyRadius uint16  `maxminddb:"accuracy_radius" json:"accuracy_radius,omitempty"`
		Latitude       float64 `maxminddb:"latitude" json:"latitude,omitempty"`
		Longitude      float64 `maxminddb:"longitude" json:"longitude,omitempty"`
		MetroCode      uint16  `maxminddb:"metro_code" json:"metro_code,omitempty"`
		TimeZone       string  `maxminddb:"time_zone" json:"time_zone,omitempty"`
	} `maxminddb:"location" json:"location,omitempty"`
	Postal struct {
		Code string `maxminddb:"code" json:"code,omitempty"`
	} `maxminddb:"postal" json:"postal,omitempty"`
	RegisteredCountry struct {
		GeonameID uint32            `maxminddb:"geoname_id" json:"geoname_id,omitempty"`
		ISOCode   string            `maxminddb:"iso_code" json:"iso_code,omitempty"`
		Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"registered_country" json:"registered_country,omitempty"`
	Subdivisions []struct {
		GeonameID uint32            `maxminddb:"geoname_id" json:"geoname_id,omitempty"`
		ISOCode   string            `maxminddb:"iso_code" json:"iso_code,omitempty"`
		Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"subdivisions" json:"subdivisions,omitempty"`
	ASN struct {
		Number       uint32 `maxminddb:"autonomous_system_number" json:"autonomous_system_number,omitempty"`
		Organization string `maxminddb:"autonomous_system_organization" json:"autonomous_system_organization,omitempty"`
		Domain       string `maxminddb:"as_domain" json:"as_domain,omitempty"`
	} `maxminddb:"asn" json:"asn,omitempty"`
	Proxy struct {
		IsProxy     bool `maxminddb:"is_proxy" json:"is_proxy"`
		IsVPN       bool `maxminddb:"is_vpn" json:"is_vpn"`
		IsTor       bool `maxminddb:"is_tor" json:"is_tor"`
		IsHosting   bool `maxminddb:"is_hosting" json:"is_hosting"`
		IsCDN       bool `maxminddb:"is_cdn" json:"is_cdn"`
		IsSchool    bool `maxminddb:"is_school" json:"is_school"`
		IsAnonymous bool `maxminddb:"is_anonymous" json:"is_anonymous"`
	} `maxminddb:"proxy" json:"proxy,omitempty"`
}

type IPResult struct {
	IP    string    `json:"ip"`
	Data  *IPRecord `json:"data,omitempty"`
	Error string    `json:"error,omitempty"`
}

type SimpleIPResult struct {
	Organization    string  `json:"organization"`
	City            string  `json:"city"`
	ISP             string  `json:"isp"`
	ASNOrganization string  `json:"asn_organization"`
	Latitude        float64 `json:"latitude"`
	ASN             uint32  `json:"asn"`
	ContinentCode   string  `json:"continent_code"`
	Country         string  `json:"country"`
	Timezone        string  `json:"timezone"`
	CountryCode     string  `json:"country_code"`
	Longitude       float64 `json:"longitude"`
	Region          string  `json:"region"`
	IP              string  `json:"ip"`
	RegionCode      string  `json:"region_code"`
}

var db *maxminddb.Reader

var (
	wg    = sync.WaitGroup{}
	port  = ""
	d     = "" // 下载标识
	dbUrl = "https://github.com/NetworkCats/Merged-IP-Data/releases/latest/download/Merged-IP.mmdb"
)

const (
	ipDbPath = "./Merged-IP.mmdb"
)

func init() {
	_p := flag.String("p", "8066", "本地监听的端口")
	_d := flag.String("d", "", "仅用于下载最新的ip地址库，保存在当前目录")
	flag.Parse()

	port = *_p
	d = *_d

	if d == "1" {
		downloadIpDb(dbUrl)
		os.Exit(1)
	} else if d != "" {
		downloadIpDb(d)
		os.Exit(1)
	}

	checkIpDbIsExist()
}

func main() {
	var err error
	db, err = maxminddb.Open(ipDbPath)
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/s") {
			serveSimpleIP(w, r)
		} else {
			queryIP(w, r)
		}
	})
	log.Println("服务启动，监听 http://127.0.0.1:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func serveSimpleIP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	lang := "zh-CN"

	if path != "/s" {
		prefix := "/s/"
		if strings.HasPrefix(path, prefix) {
			lang = strings.TrimPrefix(path, prefix)
			if !isValidLang(lang) {
				lang = "zh-CN"
			}
		}
	}

	querySimpleIPWithLang(w, r, lang)
}

func getRealIP(r *http.Request) string {
	// 优先级 1: X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "unknown" && net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// 优先级 2: X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// 优先级 3: Forwarded
	if forwarded := r.Header.Get("Forwarded"); forwarded != "" {
		parts := strings.Split(forwarded, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(strings.ToLower(part), "for=") {
				forVal := strings.TrimPrefix(part[4:], "\"")
				forVal = strings.TrimSuffix(forVal, "\"")
				forVal = strings.TrimSpace(forVal)
				if strings.HasPrefix(forVal, "[") {
					// IPv6
					if idx := strings.Index(forVal, "]"); idx != -1 {
						forVal = forVal[1:idx]
					}
				} else if idx := strings.Index(forVal, ":"); idx != -1 {
					// IPv4 + port
					forVal = forVal[:idx]
				}
				if net.ParseIP(forVal) != nil {
					return forVal
				}
			}
		}
	}

	// 优先级 4: RemoteAddr (去除端口)
	if r.RemoteAddr != "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			if net.ParseIP(host) != nil {
				return host
			}
		}
		if net.ParseIP(r.RemoteAddr) != nil {
			return r.RemoteAddr
		}
	}

	return "127.0.0.1"
}

func isValidLang(lang string) bool {
	validLangs := map[string]bool{
		"en":    true,
		"de":    true,
		"es":    true,
		"fr":    true,
		"ja":    true,
		"pt-BR": true,
		"ru":    true,
		"zh-CN": true,
	}
	return validLangs[lang]
}

func querySimpleIPWithLang(w http.ResponseWriter, r *http.Request, lang string) {
	ipParam := r.URL.Query().Get("ip")
	if ipParam == "" {
		ipParam = getRealIP(r)
	}

	ipList := strings.Split(ipParam, ",")
	results := make([]SimpleIPResult, len(ipList))

	var wg sync.WaitGroup
	for i, raw := range ipList {
		wg.Add(1)
		go func(idx int, raw string) {
			defer wg.Done()
			ipStr := strings.TrimSpace(raw)
			result := SimpleIPResult{IP: ipStr}

			ip := net.ParseIP(ipStr)
			if ip == nil {
				return
			}

			var record IPRecord
			if err := db.Lookup(ip, &record); err != nil {
				return
			}

			result.ASN = record.ASN.Number
			result.ASNOrganization = record.ASN.Organization
			result.Organization = record.ASN.Organization
			result.ISP = record.ASN.Domain
			result.ContinentCode = record.Continent.Code
			result.CountryCode = record.Country.ISOCode
			result.Country = record.Country.Names[lang]
			result.City = record.City.Names[lang]
			result.Latitude = record.Location.Latitude
			result.Longitude = record.Location.Longitude
			result.Timezone = record.Location.TimeZone

			if len(record.Subdivisions) > 0 {
				result.RegionCode = record.Subdivisions[0].ISOCode
				result.Region = record.Subdivisions[0].Names[lang]
			}

			results[idx] = result
		}(i, raw)
	}
	wg.Wait()

	writeJSON(w, http.StatusOK, results)
}

func queryIP(w http.ResponseWriter, r *http.Request) {
	ipParam := r.URL.Query().Get("ip")
	if ipParam == "" {
		ipParam = getRealIP(r)
	}

	ipList := strings.Split(ipParam, ",")
	results := make([]IPResult, len(ipList))

	var wg sync.WaitGroup
	for i, raw := range ipList {
		wg.Add(1)
		go func(idx int, raw string) {
			defer wg.Done()
			ipStr := strings.TrimSpace(raw)
			result := IPResult{IP: ipStr}

			ip := net.ParseIP(ipStr)
			if ip == nil {
				result.Error = "无效的 IP 地址"
				results[idx] = result
				return
			}

			var record IPRecord
			if err := db.Lookup(ip, &record); err != nil {
				result.Error = "查询失败: " + err.Error()
				results[idx] = result
				return
			}

			result.Data = &record
			results[idx] = result
		}(i, raw)
	}
	wg.Wait()

	writeJSON(w, http.StatusOK, results)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func downloadIpDb(url string) {
	log.Println("正在下载最新的 ip 地址库...：" + url)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		err := downloadFile(ctx, ipDbPath, url)
		if err != nil {
			if ctx.Err() == context.Canceled {
				log.Println("下载已取消")
				os.Remove(ipDbPath)
			} else {
				log.Println("下载失败：", err)
			}
		}
		wg.Done()
	}()
	wg.Wait()
	log.Println("下载完成")
}

func downloadFile(ctx context.Context, filepath string, url string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	total := resp.ContentLength
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	var downloaded int64
	buf := make([]byte, 32*1024)
	progressTicker := time.NewTicker(500 * time.Millisecond)
	defer progressTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("\n下载已中断，已下载: %s / %s\n", formatSize(downloaded), formatSize(total))
			return ctx.Err()
		case <-progressTicker.C:
			if total > 0 {
				percent := float64(downloaded) / float64(total) * 100
				fmt.Printf("\r下载进度: %.1f%% (%s / %s)", percent, formatSize(downloaded), formatSize(total))
			}
		default:
			n, err := resp.Body.Read(buf)
			if n > 0 {
				downloaded += int64(n)
				if _, writeErr := out.Write(buf[:n]); writeErr != nil {
					return writeErr
				}
			}
			if err != nil {
				if err == io.EOF {
					if total > 0 {
						fmt.Printf("\r下载进度: 100.0%% (%s / %s)\n", formatSize(downloaded), formatSize(total))
					}
					return nil
				}
				return err
			}
		}
	}
}

func formatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/1024/1024)
	}
	return fmt.Sprintf("%.2f GB", float64(bytes)/1024/1024/1024)
}

func checkIpDbIsExist() {
	if _, err := os.Stat(ipDbPath); os.IsNotExist(err) {
		log.Println("ip 地址库文件不存在")
		downloadIpDb(dbUrl)
	}
}
