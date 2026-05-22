package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

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

var db *maxminddb.Reader

func main() {
	var err error
	db, err = maxminddb.Open("Merged-IP.mmdb")
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/", queryIP)
	log.Println("服务启动，监听 http://127.0.0.1:8066")
	log.Fatal(http.ListenAndServe(":8066", nil))
}

func queryIP(w http.ResponseWriter, r *http.Request) {
	ipParam := r.URL.Query().Get("ip")
	if ipParam == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少 ip 参数"})
		return
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
