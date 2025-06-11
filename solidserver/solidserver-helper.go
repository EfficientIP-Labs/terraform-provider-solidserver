package solidserver

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"math/big"
	"net/netip"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Integer Absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

func IsIPAddressOrEmptyString(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return warnings, errors
	}

	if v != "" {
		return validation.IsIPAddress(i, k)
	}

	return warnings, errors
}

// Convert a Slice of interface{} into a Slice of string(s)
func interfaceSliceToStringSlice(interfaceSlice []interface{}) ([]string, error) {
	stringSlice := make([]string, len(interfaceSlice))

	for i, val := range interfaceSlice {
		strVal, ok := val.(string) // Type assertion
		if !ok {
			return nil, fmt.Errorf("element at index %d is not a string, but %v", i, reflect.TypeOf(val))
		}
		stringSlice[i] = strVal
	}

	return stringSlice, nil
}

// Return the offset of a matching string in a slice or -1 if not found
func stringOffsetInSlice(s string, list []string) int {
	for offset, entry := range list {
		if entry == s {
			return offset
		}
	}
	return -1
}

// Return an entry from a unordered slice based on an index
func removeOffsetInSlice(i int, s []string) []string {
	if i <= 0 {
		return s
	} else {
		s[i] = s[len(s)-1]
		return s[:len(s)-1]
	}
}

// Convert a Schema.TypeList interface into an array of strings
func toStringArray(in []interface{}) []string {
	out := make([]string, len(in))

	for i, v := range in {
		if v == nil {
			out[i] = ""
			continue
		}
		out[i] = v.(string)
	}

	return out
}

// Convert an array of strings into a Schema.TypeList interface
func toStringArrayInterface(in []string) []interface{} {
	out := make([]interface{}, len(in))

	for i, v := range in {
		out[i] = v
	}

	return out
}

// Consistent merge of TypeList elements, maintaining entries position within the list
// Workaround to TF Plugin SDK issue https://github.com/hashicorp/terraform-plugin-sdk/issues/477
func typeListConsistentMerge(old []string, new []string) []interface{} {
	// Step 1 Build local list of member indexed by their offset
	oldOffsets := make(map[int]string, len(old))
	diff := make([]string, 0, len(new))
	res := make([]interface{}, 0, len(new))

	for _, n := range new {
		if n != "" {
			offset := stringOffsetInSlice(n, old)

			if offset != -1 {
				oldOffsets[offset] = n
			} else {
				diff = append(diff, n)
			}
		}
	}

	// Merge sorted entries ordered by their offset with the diff array that contain the new ones
	// Step 2 Sort the index
	keys := make([]int, 0, len(old))
	for k := range oldOffsets {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// Step 3 build the result
	for _, k := range keys {
		res = append(res, oldOffsets[k])
	}
	for _, v := range diff {
		res = append(res, v)
	}

	return res
}

// BigIntToHexStr convert a Big Integer into an Hexa String
func BigIntToHexStr(bigInt *big.Int) string {
	return fmt.Sprintf("%x", bigInt)
}

// BigIntToStr convert a Big Integer to Decimal String
func BigIntToStr(bigInt *big.Int) string {
	return fmt.Sprintf("%v", bigInt)
}

// Convert hexa IPv6 address string into standard IPv6 address string
// Return an empty string in case of failure
func hexiptoip(hexip string) string {
	a, b, c, d := 0, 0, 0, 0

	count, _ := fmt.Sscanf(hexip, "%02x%02x%02x%02x", &a, &b, &c, &d)

	if count == 4 {
		return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
	}

	return ""
}

// Convert IP v4 address string into PTR record name
// Return an empty string in case of failure
func iptoptr(ip string) string {
	a, b, c, d := 0, 0, 0, 0

	count, _ := fmt.Sscanf(ip, "%03d.%03d.%03d.%03d", &a, &b, &c, &d)

	if count == 4 {
		return fmt.Sprintf("%d.%d.%d.%d.in-addr.arpa", d, c, b, a)
	}

	return ""
}

// Convert IPv6 address string into PTR record name
// Return an empty string in case of failure
func ip6toptr(ip string) string {
	buffer := strings.Split(ip, ":")
	res := ""

	for i := len(buffer) - 1; i >= 0; i-- {
		for j := len(buffer[i]) - 1; j >= 0; j-- {
			res += string(buffer[i][j]) + "."
		}
	}

	return res + "ip6.arpa"
}

// Convert hexa IPv6 address string into standard IPv6 address string
// Return an empty string in case of failure
func hexip6toip6(hexip string) string {
	res := ""

	for i, c := range hexip {
		if (i == 0) || ((i % 4) != 0) {
			res += string(c)
		} else {
			res += ":"
			res += string(c)
		}
	}

	return res
}

// Convert standard IP address string into hexa IP address string
// Return an empty string in case of failure
func iptohexip(ip string) string {
	ipDec := strings.Split(ip, ".")

	if len(ipDec) == 4 {

		a, _ := strconv.Atoi(ipDec[0])
		b, _ := strconv.Atoi(ipDec[1])
		c, _ := strconv.Atoi(ipDec[2])
		d, _ := strconv.Atoi(ipDec[3])

		if 0 <= a && a <= 255 && 0 <= b && b <= 255 &&
			0 <= c && c <= 255 && 0 <= d && d <= 255 {
			return fmt.Sprintf("%02x%02x%02x%02x", a, b, c, d)
		}

		return ""
	}

	return ""
}

// Convert standard IPv6 address string into hexa IPv6 address string
// Return an empty string in case of failure
func ip6tohexip6(ip string) string {
	ipDec := strings.Split(ip, ":")
	res := ""

	if len(ipDec) == 8 {
		for _, b := range ipDec {
			res += fmt.Sprintf("%04s", b)
		}

		return res
	}

	return ""
}

// Convert standard IPv6 address string into expanded IPv6 address string
// Return an empty string in case of failure
func longip6toshortip6(ip string) string {
	tmp, _ := netip.ParseAddr(ip)

	if tmp.Is6() {
		return tmp.String()
	}

	return ""
}

// Convert standard IPv6 address string into expanded IPv6 address string
// Return an empty string in case of failure
func shortip6tolongip6(ip string) string {
	tmp, _ := netip.ParseAddr(ip)

	if tmp.Is6() {
		return tmp.StringExpanded()
	}

	return ""
}

// Convert standard IP address string into unsigned int32
// Return 0 in case of failure
func iptolong(ip string) uint32 {
	ipDec := strings.Split(ip, ".")

	if len(ipDec) == 4 {
		a, _ := strconv.Atoi(ipDec[0])
		b, _ := strconv.Atoi(ipDec[1])
		c, _ := strconv.Atoi(ipDec[2])
		d, _ := strconv.Atoi(ipDec[3])

		var iplong uint32 = uint32(a) * 0x1000000
		iplong += uint32(b) * 0x10000
		iplong += uint32(c) * 0x100
		iplong += uint32(d) * 0x1

		return iplong
	}

	return 0
}

// Convert unsigned int32 into standard IP address string
// Return an IP formated string
func longtoip(iplong uint32) string {
	a := (iplong & 0xFF000000) >> 24
	b := (iplong & 0xFF0000) >> 16
	c := (iplong & 0xFF00) >> 8
	d := (iplong & 0xFF)

	if a < 0 {
		a = a + 0x100
	}

	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
}

// Ignore Case When comparing remote and local value
func resourcediffsuppresscase(k, old, new string, d *schema.ResourceData) bool {
	if strings.ToLower(old) == strings.ToLower(new) {
		return true
	}

	return false
}

// Ignore Different IPv6 Format
func resourcediffsuppressIPv6Format(k, old, new string, d *schema.ResourceData) bool {
	oldipv6, _ := netip.ParseAddr(old)
	newipv6, _ := netip.ParseAddr(new)

	//fmt.Printf("(%v).String() -> %v\n", ipv6, newipv6.String())

	if oldipv6.Is6() && newipv6.Is6() {
		if oldipv6.Compare(newipv6) == 0 {
			return true
		}
		return false
	}

	if new == old {
		return true
	}

	return false
}

// Compute the prefix length from the size of a CIDR prefix
// Return the prefix length
func sizetoprefixlength(size int) int {
	prefixlength := 32

	for prefixlength > 0 && size > 1 {
		size = size / 2
		prefixlength--
	}

	return prefixlength
}

// Compute the actual size of a CIDR prefix from its length
// Return -1 in case of failure
func prefixlengthtosize(length int) int {
	if length >= 0 && length <= 32 {
		return (1 << (32 - uint32(length)))
	}

	return -1
}

// Compute the netmask of a CIDR prefix from its length
// Return an empty string in case of failure
func prefixlengthtohexip(length int) string {
	if length >= 0 && length <= 32 {
		return longtoip((^((1 << (32 - uint32(length))) - 1)) & 0xffffffff)
	}

	return ""
}

// Compute the actual size of an IPv6 CIDR prefix from its length
// Return -1 in case of failure
func prefix6lengthtosize(length int64) *big.Int {
	sufix := big.NewInt(32 - (length / 4))
	size := big.NewInt(16)

	size = size.Exp(size, sufix, nil)

	//size = size.Sub(size, big.NewInt(1))

	return size
}

// Build url value object from class parameters
// Return an url.Values{} object
func urlfromclassparams(parameters interface{}) url.Values {
	classParameters := url.Values{}

	for k, v := range parameters.(map[string]interface{}) {
		classParameters.Add(k, v.(string))
	}

	return classParameters
}

// Return the port oid from an interface_id
// Or an empty string in case of failure
func portidbyinterfaceid(interfaceID string, meta interface{}) (string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("nomiface_id", interfaceID)

	// Sending the read request
	resp, body, err := s.Request("get", "rest/nom_iface_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if portID, portIDExist := buf[0]["nomport_id"].(string); portIDExist {
				return portID, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find interface (OID): %s\n", interfaceID))

	return "", err
}

// Return the oid of a device from hostdev_name
// Or an empty string in case of failure
func hostdevidbyname(hostdevName string, meta interface{}) (string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "hostdev_name='"+strings.ToLower(hostdevName)+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/hostdev_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if hostdevID, hostdevIDExist := buf[0]["hostdev_id"].(string); hostdevIDExist {
				return hostdevID, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find device: %s\n", hostdevName))

	return "", err
}

// Return an available IP addresses from site_id, block_id and expected subnet_size
// Or an empty table of string in case of failure
func ipaddressfindfree(subnetID string, poolID string, method string, meta interface{}) ([]string, error) {
	s := meta.(*SOLIDserver)

	if method == "start" || method == "end" {
		// Building parameters
		parameters := url.Values{}

		whereClause := "free_start_ip_addr != free_end_ip_addr AND subnet_id=" + subnetID

		if len(poolID) > 0 {
			whereClause += " AND pool_id = " + poolID
		}

		parameters.Add("WHERE", whereClause)

		if method == "start" {
			parameters.Add("ORDERBY", "free_start_ip_addr asc")
		} else {
			parameters.Add("ORDERBY", "free_end_ip_addr desc")
		}

		// Sending the request
		resp, body, err := s.Request("get", "rest/ip_free_address_list", &parameters)

		if err == nil {
			var buf [](map[string]interface{})
			json.Unmarshal([]byte(body), &buf)

			// Checking the answer
			if resp.StatusCode == 200 && len(buf) > 0 {
				addresses := []string{}

				for i := 0; i < len(buf); i++ {
					startIPStr, okStart := buf[i]["free_start_ip_addr"].(string)
					endIPStr, okEnd := buf[i]["free_end_ip_addr"].(string)

					if !okStart || !okEnd {
						continue
					}

					startIP, errStart := netip.ParseAddr(hexiptoip(startIPStr))
					endIP, errEnd := netip.ParseAddr(hexiptoip(endIPStr))

					if errStart != nil || errEnd != nil {
						tflog.Debug(s.Ctx, fmt.Sprintf("Unable to compute free range start/end IP addresses: %s/%s\n", startIPStr, endIPStr))
						continue
					}

					if method == "start" {
						for currentIP := startIP; currentIP.Compare(endIP) <= 0; currentIP = currentIP.Next() {
							if len(addresses) >= 32 { // Check the length before appending
								return addresses, nil
							}

							tflog.Debug(s.Ctx, fmt.Sprintf("Suggested IP address: %s\n", currentIP.String()))
							addresses = append(addresses, currentIP.String())
						}
					} else {
						for currentIP := endIP; currentIP.Compare(startIP) >= 0; currentIP = currentIP.Prev() {
							if len(addresses) >= 32 { // Check the length before appending
								return addresses, nil
							}

							tflog.Debug(s.Ctx, fmt.Sprintf("Suggested IP address: %s\n", currentIP.String()))
							addresses = append(addresses, currentIP.String())
						}
					}
				}
				return addresses, nil
			}
		}

		tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find a free IP address in subnet (oid): %s\n", subnetID))
		return []string{}, err

	} else {
		// Building parameters
		parameters := url.Values{}
		parameters.Add("subnet_id", subnetID)
		parameters.Add("max_find", "32")

		if len(poolID) > 0 {
			parameters.Add("pool_id", poolID)
		}

		// Sending the request
		resp, body, err := s.Request("get", "rpc/ip_find_free_address", &parameters)

		if err == nil {
			var buf [](map[string]interface{})
			json.Unmarshal([]byte(body), &buf)

			// Checking the answer
			if resp.StatusCode == 200 && len(buf) > 0 {
				addresses := []string{}

				for i := 0; i < len(buf); i++ {
					if addr, addrExist := buf[i]["hostaddr"].(string); addrExist {
						tflog.Debug(s.Ctx, fmt.Sprintf("Suggested IP address: %s\n", addr))
						addresses = append(addresses, addr)
					}
				}
				return addresses, nil
			}
		}

		tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find a free IP address in subnet (oid): %s\n", subnetID))
		return []string{}, err
	}
}

// Return an available IP addresses from site_id, block_id and expected subnet_size
// Or an empty table of string in case of failure
func ip6addressfindfree(subnetID string, poolID string, meta interface{}) ([]string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("subnet6_id", subnetID)
	parameters.Add("max_find", "32")

	if len(poolID) > 0 {
		parameters.Add("pool6_id", poolID)
	}

	// Sending the creation request
	resp, body, err := s.Request("get", "rpc/ip6_find_free_address6", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			addresses := []string{}

			for i := 0; i < len(buf); i++ {
				if addr, addrExist := buf[i]["hostaddr6"].(string); addrExist {
					tflog.Debug(s.Ctx, fmt.Sprintf("Suggested IP address: %s\n", addr))
					addresses = append(addresses, addr)
				}
			}
			return addresses, nil
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find a free IPv6 address in subnet (oid): %s\n", subnetID))

	return []string{}, err
}

// Return an available vlan from specified vlmdomain_name
// Or an empty table strings in case of failure
func vlanidfindfree(vlmdomainName string, meta interface{}) ([]string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("limit", "16")

	if s.Version < 700 {
		parameters.Add("WHERE", "vlmdomain_name='"+strings.ToLower(vlmdomainName)+"' AND row_enabled='2'")
	} else {
		parameters.Add("WHERE", "vlmdomain_name='"+strings.ToLower(vlmdomainName)+"' AND type='free'")
	}

	// Sending the creation request
	resp, body, err := s.Request("get", "rest/vlmvlan_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			vnIDs := []string{}

			for i := range buf {
				if s.Version < 700 {
					if vnID, vnIDExist := buf[i]["vlmvlan_vlan_id"].(string); vnIDExist {
						tflog.Debug(s.Ctx, fmt.Sprintf("Suggested vlan ID: %s\n", vnID))
						vnIDs = append(vnIDs, vnID)
					}
				} else {
					if startVlanID, startVlanIDExist := buf[i]["free_start_vlan_id"].(string); startVlanIDExist {
						if endVlanID, endVlanIDExist := buf[i]["free_end_vlan_id"].(string); endVlanIDExist {
							vnID, _ := strconv.Atoi(startVlanID)
							maxVnID, _ := strconv.Atoi(endVlanID)

							j := 0
							for vnID < maxVnID && j < 8 {
								tflog.Debug(s.Ctx, fmt.Sprintf("Suggested vlan ID: %d\n", vnID))
								vnIDs = append(vnIDs, strconv.Itoa(vnID))
								vnID++
								j++
							}
						}
					}
				}
			}
			return vnIDs, nil
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find a free vlan ID in vlan domain: %s\n", vlmdomainName))

	return []string{}, err
}

// Return the oid of a space from site_name
// Or an empty string in case of failure
func ipsiteidbyname(siteName string, meta interface{}) (string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "site_name='"+strings.ToLower(siteName)+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip_site_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if siteID, siteIDExist := buf[0]["site_id"].(string); siteIDExist {
				return siteID, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find IP space: %s\n", siteName))

	return "", err
}

// Return the oid of a vlan domain from vlmdomain_name
// Or an empty string in case of failure
func vlandomainidbyname(vlmdomainName string, meta interface{}) (string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "vlmdomain_name='"+strings.ToLower(vlmdomainName)+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/vlmdomain_name", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if vlmdomainID, vlmdomainIDExist := buf[0]["vlmdomain_id"].(string); vlmdomainIDExist {
				return vlmdomainID, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find vlan domain: %s\n", vlmdomainName))

	return "", err
}

// Return the oid of a vlan (vlmvlan_id) from vlmdomain_name and vlan_id
// Or an empty string in case of failure
func vlanidbyinfo(vlmdomainName string, vlmvlanvlanID int, meta interface{}) (string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "vlmdomain_name='"+vlmdomainName+"' AND vlmvlan_vlan_id='"+strconv.Itoa(vlmvlanvlanID)+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/vlmvlan_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if vlmvlanID, vlmvlanIDExist := buf[0]["vlmvlan_id"].(string); vlmvlanIDExist {
				return vlmvlanID, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find VLAN ID %d within VLAN Domain\n", vlmvlanvlanID, vlmdomainName))

	return "", err
}

// Return the oid of a subnet from site_id, subnet_name and is_terminal property
// Or an empty string in case of failure
func ipsubnetidbyname(siteID string, subnetName string, terminal bool, meta interface{}) (string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	whereClause := "site_id='" + siteID + "' AND " + "subnet_name='" + strings.ToLower(subnetName) + "'"

	if terminal {
		whereClause += "AND is_terminal='1'"
	} else {
		whereClause += "AND is_terminal='0'"
	}

	parameters.Add("WHERE", whereClause)

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip_block_subnet_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if subnetID, subnetIDExist := buf[0]["subnet_id"].(string); subnetIDExist {
				return subnetID, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find IP subnet: %s\n", subnetName))

	return "", err
}

// Return the oid of a pool from site_id and pool_name
// Or an empty string in case of failure
func ippoolidbyname(siteID string, poolName string, subnetName string, meta interface{}) (string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "site_id='"+siteID+"' AND "+"pool_name='"+strings.ToLower(poolName)+"' AND subnet_name='"+strings.ToLower(subnetName)+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip_pool_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if poolID, poolIDExist := buf[0]["pool_id"].(string); poolIDExist {
				return poolID, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find IP pool: %s\n", poolName))

	return "", err
}

// Return the oid of a pool from site_id and pool_name
// Or an empty string in case of failure
func ippoolinfobyname(siteID string, poolName string, subnetName string, meta interface{}) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "site_id='"+siteID+"' AND "+"pool_name='"+strings.ToLower(poolName)+"' AND subnet_name='"+strings.ToLower(subnetName)+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip_pool_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if poolID, poolIDExist := buf[0]["pool_id"].(string); poolIDExist {
				res["id"] = poolID

				if poolName, poolNameExist := buf[0]["pool_name"].(string); poolNameExist {
					res["name"] = poolName
				}

				if poolSize, poolSizeExist := buf[0]["pool_size"].(string); poolSizeExist {
					res["size"], _ = strconv.Atoi(poolSize)
				}

				if poolStartAddr, poolStartAddrExist := buf[0]["start_ip_addr"].(string); poolStartAddrExist {
					res["start_hex_addr"] = poolStartAddr
					res["start_addr"] = hexiptoip(poolStartAddr)
				}

				if poolEndAddr, poolEndAddrExist := buf[0]["end_ip_addr"].(string); poolEndAddrExist {
					res["end_hex_addr"] = poolEndAddr
					res["end_addr"] = hexiptoip(poolEndAddr)
				}

				return res, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find IP pool: %s\n", poolName))

	return nil, err
}

// Return a map of information about a subnet from site_id, subnet_name and is_terminal property
// Or nil in case of failure
func ipsubnetinfobyname(siteID string, subnetName string, terminal bool, meta interface{}) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	whereClause := "site_id='" + siteID + "' AND " + "subnet_name='" + strings.ToLower(subnetName) + "'"

	if terminal {
		whereClause += "AND is_terminal='1'"
	} else {
		whereClause += "AND is_terminal='0'"
	}

	parameters.Add("WHERE", whereClause)

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip_block_subnet_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if subnetID, subnetIDExist := buf[0]["subnet_id"].(string); subnetIDExist {
				res["id"] = subnetID

				if subnetName, subnetNameExist := buf[0]["subnet_name"].(string); subnetNameExist {
					res["name"] = subnetName
				}

				if subnetSize, subnetSizeExist := buf[0]["subnet_size"].(string); subnetSizeExist {
					res["size"], _ = strconv.Atoi(subnetSize)
					res["prefix_length"] = sizetoprefixlength(res["size"].(int))
				}

				if subnetStartAddr, subnetStartAddrExist := buf[0]["start_ip_addr"].(string); subnetStartAddrExist {
					res["start_hex_addr"] = subnetStartAddr
					res["start_addr"] = hexiptoip(subnetStartAddr)
				}

				if subnetEndAddr, subnetEndAddrExist := buf[0]["end_ip_addr"].(string); subnetEndAddrExist {
					res["end_hex_addr"] = subnetEndAddr
					res["end_addr"] = hexiptoip(subnetEndAddr)
				}

				if subnetTerminal, subnetTerminalExist := buf[0]["is_terminal"].(string); subnetTerminalExist {
					res["terminal"] = subnetTerminal
				}

				if subnetLvl, subnetLvlExist := buf[0]["subnet_level"].(string); subnetLvlExist {
					res["level"] = subnetLvl
				}

				return res, nil
			}
		}

		return nil, fmt.Errorf("SOLIDServer - Unable to find IP subnet: %s\n", subnetName)
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find IP subnet: %s\n", subnetName))

	return nil, err
}

// Return the oid of a subnet from site_id, subnet_name and is_terminal property
// Or an empty string in case of failure
func ip6subnetidbyname(siteID string, subnetName string, terminal bool, meta interface{}) (string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	whereClause := "site_id='" + siteID + "' AND " + "subnet6_name='" + strings.ToLower(subnetName) + "'"

	if terminal {
		whereClause += "AND is_terminal='1'"
	} else {
		whereClause += "AND is_terminal='0'"
	}

	parameters.Add("WHERE", whereClause)

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip6_block6_subnet6_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if subnetID, subnetIDExist := buf[0]["subnet6_id"].(string); subnetIDExist {
				return subnetID, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find IPv6 subnet: %s\n", subnetName))

	return "", err
}

// Return the oid of a pool from site_id and pool_name
// Or an empty string in case of failure
func ip6poolidbyname(siteID string, poolName string, subnetName string, meta interface{}) (string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "site_id='"+siteID+"' AND "+"pool6_name='"+strings.ToLower(poolName)+"' AND subnet6_name='"+strings.ToLower(subnetName)+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip6_pool6_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if poolID, poolIDExist := buf[0]["pool6_id"].(string); poolIDExist {
				return poolID, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find IPv6 pool: %s\n", poolName))

	return "", err
}

// Return the oid of a pool from site_id and pool_name
// Or an empty string in case of failure
func ip6poolinfobyname(siteID string, poolName string, subnetName string, meta interface{}) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "site_id='"+siteID+"' AND "+"pool6_name='"+strings.ToLower(poolName)+"' AND subnet6_name='"+strings.ToLower(subnetName)+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip6_pool6_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if poolID, poolIDExist := buf[0]["pool6_id"].(string); poolIDExist {
				res["id"] = poolID

				if poolName, poolNameExist := buf[0]["pool6_name"].(string); poolNameExist {
					res["name"] = poolName
				}

				if poolSize, poolSizeExist := buf[0]["pool6_size"].(string); poolSizeExist {
					res["size"], _ = strconv.Atoi(poolSize)
				}

				if poolStartAddr, poolStartAddrExist := buf[0]["start_ip6_addr"].(string); poolStartAddrExist {
					res["start_hex_addr"] = poolStartAddr
					res["start_addr"] = hexiptoip(poolStartAddr)
				}

				if poolEndAddr, poolEndAddrExist := buf[0]["end_ip6_addr"].(string); poolEndAddrExist {
					res["end_hex_addr"] = poolEndAddr
					res["end_addr"] = hexiptoip(poolEndAddr)
				}

				return res, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find IPv6 pool: %s\n", poolName))

	return nil, err
}

// Return a map of information about a subnet from site_id, subnet_name and is_terminal property
// Or nil in case of failure
func ip6subnetinfobyname(siteID string, subnetName string, terminal bool, meta interface{}) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	whereClause := "site_id='" + siteID + "' AND " + "subnet6_name='" + strings.ToLower(subnetName) + "'"

	if terminal {
		whereClause += "AND is_terminal='1'"
	} else {
		whereClause += "AND is_terminal='0'"
	}

	parameters.Add("WHERE", whereClause)

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip6_block6_subnet6_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if subnetID, subnetIDExist := buf[0]["subnet6_id"].(string); subnetIDExist {
				res["id"] = subnetID

				if subnetName, subnetNameExist := buf[0]["subnet6_name"].(string); subnetNameExist {
					res["name"] = subnetName
				}

				if subnetPrefixSize, subnetPrefixSizeExist := buf[0]["subnet6_prefix"].(string); subnetPrefixSizeExist {
					res["prefix_length"], _ = strconv.Atoi(subnetPrefixSize)
				}

				if subnetStartAddr, subnetStartAddrExist := buf[0]["start_ip6_addr"].(string); subnetStartAddrExist {
					res["start_hex_addr"] = subnetStartAddr
					res["start_addr"] = hexiptoip(subnetStartAddr)
				}

				if subnetEndAddr, subnetEndAddrExist := buf[0]["end_ip6_addr"].(string); subnetEndAddrExist {
					res["end_hex_addr"] = subnetEndAddr
					res["end_addr"] = hexiptoip(subnetEndAddr)
				}

				if subnetTerminal, subnetTerminalExist := buf[0]["is_terminal"].(string); subnetTerminalExist {
					res["terminal"] = subnetTerminal
				}

				if subnetLvl, subnetLvlExist := buf[0]["subnet_level"].(string); subnetLvlExist {
					res["level"] = subnetLvl
				}

				return res, nil
			}
		}

		return nil, fmt.Errorf("SOLIDServer - Unable to find IPv6 subnet: %s\n", subnetName)
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find IPv6 subnet: %s\n", subnetName))

	return nil, err
}

// Return the oid of an address from site_id, ip_address
// Or an empty string in case of failure
func ipaddressidbyip(siteID string, ipAddress string, meta interface{}) (string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "site_id='"+siteID+"' AND "+"ip_addr='"+iptohexip(ipAddress)+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip_used_address_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if ipID, ipIDExist := buf[0]["ip_id"].(string); ipIDExist {
				return ipID, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find IP address: %s\n", ipAddress))

	return "", err
}

// Return the oid of an address from site_id, ip_address
// Or an empty string in case of failure
func ip6addressidbyip6(siteID string, ipAddress string, meta interface{}) (string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "site_id='"+siteID+"' AND "+"ip6_addr='"+ip6tohexip6(ipAddress)+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip6_address6_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if ipID, ipIDExist := buf[0]["ip6_id"].(string); ipIDExist {
				return ipID, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find IPv6 address: %s\n", ipAddress))

	return "", err
}

// Return the oid of an address from ip_id, ip_name_type, alias_name
// Or an empty string in case of failure
func ipaliasidbyinfo(addressID string, aliasName string, ipNameType string, meta interface{}) (string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("ip_id", addressID)
	parameters.Add("WHERE", "ip_name_type='"+ipNameType+"' AND "+"alias_name='"+aliasName+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip_alias_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if ipNameID, ipNameIDExist := buf[0]["ipNameID"].(string); ipNameIDExist {
				return ipNameID, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find IP alias: %s - %s associated with IP address ID %s\n", aliasName, ipNameType, addressID))

	return "", err
}

// Return an available subnet address from site_id, block_id and expected subnet_size
// Or an empty string in case of failure
func ipsubnetfindbysize(siteID string, blockID string, requestedIP string, prefixSize int, meta interface{}) ([]string, error) {
	subnetAddresses := []string{}
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("site_id", siteID)
	parameters.Add("prefix", strconv.Itoa(prefixSize))
	parameters.Add("max_find", "16")

	// Specifying a suggested subnet IP address
	if len(requestedIP) > 0 {
		subnetAddresses = append(subnetAddresses, iptohexip(requestedIP))
		return subnetAddresses, nil
	}

	// Trying to create a subnet under an existing block
	parameters.Add("block_id", blockID)

	// Sending the creation request
	resp, body, err := s.Request("get", "rpc/ip_find_free_subnet", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			subnetAddresses := []string{}

			for i := 0; i < len(buf); i++ {
				if hexaddr, hexaddrExist := buf[i]["start_ip_addr"].(string); hexaddrExist {
					tflog.Debug(s.Ctx, fmt.Sprintf("Suggested IP subnet address: %s\n", hexiptoip(hexaddr)))
					subnetAddresses = append(subnetAddresses, hexaddr)
				}
			}
			return subnetAddresses, nil
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find a free IP subnet in space (oid): %s, block (oid): %s, size: %s\n", siteID, blockID, strconv.Itoa(prefixSize)))

	return []string{}, err
}

// Return an available subnet address from site_id, block_id and expected subnet_size
// Or an empty string in case of failure
func ip6subnetfindbysize(siteID string, blockID string, requestedIP string, prefixSize int, meta interface{}) ([]string, error) {
	subnetAddresses := []string{}
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("site_id", siteID)
	parameters.Add("prefix", strconv.Itoa(prefixSize))
	parameters.Add("max_find", "16")

	// Specifying a suggested subnet IP address
	if len(requestedIP) > 0 {
		subnetAddresses = append(subnetAddresses, ip6tohexip6(requestedIP))
		return subnetAddresses, nil
	}

	// Trying to create a subnet under an existing block
	parameters.Add("block6_id", blockID)

	// Sending the creation request
	resp, body, err := s.Request("get", "rpc/ip6_find_free_subnet6", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			subnetAddresses := []string{}

			for i := 0; i < len(buf); i++ {
				if hexaddr, hexaddrExist := buf[i]["start_ip6_addr"].(string); hexaddrExist {
					tflog.Debug(s.Ctx, fmt.Sprintf("Suggested IPv6 subnet address: %s\n", hexip6toip6(hexaddr)))
					subnetAddresses = append(subnetAddresses, hexaddr)
				}
			}
			return subnetAddresses, nil
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find a free IPv6 subnet in space (oid): %s, block (oid): %s, size: %s\n", siteID, blockID, strconv.Itoa(prefixSize)))

	return []string{}, err
}

// Return the oid of a Custom DB from name
// Or an empty string in case of failure
func cdbnameidbyname(name string, meta interface{}) (string, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "name='"+name+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/custom_db_name_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if cdbnameID, cdbnameIDExist := buf[0]["custom_db_name_id"].(string); cdbnameIDExist {
				return cdbnameID, nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find Custom DB: %s\n", name))

	return "", err
}

// Update a DNS SMART member's role list
// Return false in case of failure
func dnssmartmembersupdate(smartName string, smartMembersRole string, meta interface{}) bool {
	s := meta.(*SOLIDserver)

	// Building parameters for retrieving SMART vdns_dns_group_role information
	parameters := url.Values{}
	parameters.Add("dns_name", smartName)
	parameters.Add("add_flag", "edit_only")
	parameters.Add("vdns_dns_group_role", smartMembersRole)

	// Sending the update request
	resp, body, err := s.Request("put", "rest/dns_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			return true
		}

		// Log the error
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				tflog.Debug(s.Ctx, fmt.Sprintf("Unable to update members list of the DNS SMART: %s (%s)\n", smartName, errMsg))
			}
		} else {
			tflog.Debug(s.Ctx, fmt.Sprintf("Unable to update members list of the DNS SMART: %s\n", smartName))
		}
	}

	return false
}

// Get DNS Server status
// Return an empty string in case of failure the server status otherwise (Y -> OK)
func dnsserverstatus(serverID string, meta interface{}) string {
	s := meta.(*SOLIDserver)

	// Building parameters for retrieving information
	parameters := url.Values{}
	parameters.Add("dns_id", serverID)

	// Sending the get request
	resp, body, err := s.Request("get", "rest/dns_server_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if state, stateExist := buf[0]["dns_state"].(string); stateExist {
				return state
			}
			return ""
		}

		// Log the error
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				tflog.Debug(s.Ctx, fmt.Sprintf("Unable to retrieve DNS server status: %s (%s)\n", serverID, errMsg))
			}
		} else {
			tflog.Debug(s.Ctx, fmt.Sprintf("Unable to retrieve DNS server status: %s\n", serverID))
		}
	}

	return ""
}

// Get DNS Server View Support
// Return an true if the DNS Server has views
func dnsserverhasviews(serverName string, meta interface{}) bool {
	s := meta.(*SOLIDserver)

	// Building parameters for retrieving information
	parameters := url.Values{}

	whereClause := "dns_name='" + serverName + "'"
	parameters.Add("WHERE", whereClause)

	// Sending the get request
	resp, body, err := s.Request("get", "rest/dns_view_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 204 {
			return false
		}

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			return true
		}

		// Log the error
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				tflog.Debug(s.Ctx, fmt.Sprintf("Unable to retrieve DNS server views (%s)\n", errMsg))
			}
		} else {
			tflog.Debug(s.Ctx, fmt.Sprintf("Unable to retrieve DNS server views\n"))
		}
	}

	return true
}

// Get number of pending deletion operations on DNS server
// Return -1 in case of failure
func dnsserverpendingdeletions(serverID string, meta interface{}) int {
	s := meta.(*SOLIDserver)
	result := 0

	// Building parameters for retrieving information
	parameters := url.Values{}
	parameters.Add("WHERE", "delayed_delete_time='1' AND dns_id='"+serverID+"'")

	// Sending the get request
	resp, body, err := s.Request("get", "rest/dns_zone_count", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if total, totalExist := buf[0]["total"].(string); totalExist {
				inc, _ := strconv.Atoi(total)
				result += inc
			} else {
				return -1
			}
		}
		// Log the error
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				tflog.Debug(s.Ctx, fmt.Sprintf("Unable to retrieve DNS server pending operations: %s (%s)\n", serverID, errMsg))
			}
		} else {
			tflog.Debug(s.Ctx, fmt.Sprintf("Unable to retrieve DNS server pending operations: %s\n", serverID))
		}
	}

	// Building parameters for retrieving information
	parameters = url.Values{}
	parameters.Add("WHERE", "delayed_delete_time='1' AND dns_id='"+serverID+"'")

	// Sending the get request
	resp, body, err = s.Request("get", "rest/dns_view_count", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if total, totalExist := buf[0]["total"].(string); totalExist {
				inc, _ := strconv.Atoi(total)
				result += inc
			} else {
				return -1
			}
		}
		// Log the error
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				tflog.Debug(s.Ctx, fmt.Sprintf("Unable to retrieve DNS server pending operations: %s (%s)\n", serverID, errMsg))
			}
		} else {
			tflog.Debug(s.Ctx, fmt.Sprintf("Unable to retrieve DNS server pending operations: %s\n", serverID))
		}
	}

	return result
}

// Set a DNSserver or DNSview param value
// Return false in case of failure
func dnsparamset(serverName string, viewID string, paramKey string, paramValue string, meta interface{}) bool {
	s := meta.(*SOLIDserver)

	service := "dns_server_param_add"

	// Building parameters to push information
	parameters := url.Values{}

	if viewID != "" {
		service = "dns_view_param_add"
		parameters.Add("dnsview_id", viewID)
	} else {
		parameters.Add("dns_name", serverName)
	}

	parameters.Add("param_key", paramKey)
	parameters.Add("param_value", paramValue)

	// Sending the update request
	resp, body, err := s.Request("put", "rest/"+service, &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			return true
		}

		// Log the error
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				tflog.Debug(s.Ctx, fmt.Sprintf("Unable to set DNS server or view parameter: %s on %s (%s)\n", paramKey, serverName, errMsg))
			}
		} else {
			tflog.Debug(s.Ctx, fmt.Sprintf("Unable to set DNS server or view parameter: %s on %s\n", paramKey, serverName))
		}
	}

	return false
}

// UnSet a DNSserver or DNSview param value
// Return false in case of failure
func dnsparamunset(serverName string, viewID string, paramKey string, meta interface{}) bool {
	s := meta.(*SOLIDserver)

	service := "dns_server_param_delete"

	// Building parameters to push information
	parameters := url.Values{}

	if viewID != "" {
		service = "dns_view_param_delete"
		parameters.Add("dnsview_id", viewID)
	} else {
		parameters.Add("dns_name", serverName)
	}

	parameters.Add("param_key", paramKey)

	// Sending the delete request
	resp, body, err := s.Request("delete", "rest/"+service, &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			return true
		}

		// Log the error
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				tflog.Debug(s.Ctx, fmt.Sprintf("Unable to unset DNS server or view parameter: %s on %s (%s)\n", paramKey, serverName, errMsg))
			}
		} else {
			tflog.Debug(s.Ctx, fmt.Sprintf("Unable to unset DNS server or view parameter: %s on %s\n", paramKey, serverName))
		}
	}

	return false
}

// Get a DNSserver or DNSview param's value
// Return an empty string and an error in case of failure
func dnsparamget(serverName string, viewID string, paramKey string, meta interface{}) (string, error) {
	s := meta.(*SOLIDserver)

	service := "dns_server_param_list"
	if viewID != "" {
		service = "dns_view_param_list"
	}

	// Building parameters for retrieving information
	parameters := url.Values{}

	if viewID == "" {
		parameters.Add("WHERE", "dns_name='"+serverName+"' AND param_key='"+paramKey+"'")
	} else {
		parameters.Add("WHERE", "dns_name='"+serverName+"' AND dnsview_id='"+viewID+"' AND param_key='"+paramKey+"'")
	}

	// Sending the read request
	resp, body, err := s.Request("get", "rest/"+service, &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if paramValue, paramValueExist := buf[0]["param_value"].(string); paramValueExist {
				return paramValue, nil
			} else {
				return "", nil
			}
		}
	}

	tflog.Debug(s.Ctx, fmt.Sprintf("Unable to find DNS Param Key: %s\n", paramKey))

	return "", err
}

// Add a DNS server to a SMART with the required role, return the
// Return false in case of failure
func dnsaddtosmart(smartName string, serverName string, serverRole string, meta interface{}) bool {
	s := meta.(*SOLIDserver)

	parameters := url.Values{}
	parameters.Add("vdns_name", smartName)
	parameters.Add("dns_name", serverName)
	parameters.Add("dns_role", serverRole)

	// Sending the read request
	resp, body, err := s.Request("post", "rest/dns_smart_member_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			return true
		}

		// Atomic SMART registration service unavailable attempting to use existing services
		if resp.StatusCode == 400 || resp.StatusCode == 404 {
			// Random Delay (in case of concurrent resources creation - until 8.0 and service dns_smart_member_add)
			//time.Sleep(time.Duration((rand.Intn(600) / 10) * time.Second))

			// Otherwise proceed using the previous method
			// Building parameters for retrieving SMART vdns_dns_group_role information
			parameters := url.Values{}

			parameters.Add("WHERE", "vdns_parent_name='"+smartName+"' AND dns_type!='vdns'")

			// Sending the read request
			resp, body, err := s.Request("get", "rest/dns_server_list", &parameters)

			if err == nil {
				var buf [](map[string]interface{})
				json.Unmarshal([]byte(body), &buf)

				// Checking the answer
				if resp.StatusCode == 200 || resp.StatusCode == 204 {

					// Building vdns_dns_group_role parameter from the SMART member list
					membersRole := ""

					if len(buf) > 0 {
						for _, smartMember := range buf {
							membersRole += smartMember["dns_name"].(string) + "&" + smartMember["dns_role"].(string) + ";"
						}
					}

					membersRole += serverName + "&" + serverRole

					if dnssmartmembersupdate(smartName, membersRole, meta) {
						return true
					}

					return false
				}

				// Log the error
				if len(buf) > 0 {
					if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
						tflog.Debug(s.Ctx, fmt.Sprintf("Unable to retrieve members list of the DNS SMART: %s (%s)\n", smartName, errMsg))
					}
				} else {
					tflog.Debug(s.Ctx, fmt.Sprintf("Unable to retrieve members list of the DNS SMART: %s\n", smartName))
				}
			}

			return false
		}

		// Log the error
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				tflog.Debug(s.Ctx, fmt.Sprintf("Unable to update the member list of the DNS SMART: %s (%s)\n", smartName, errMsg))
			}
		} else {
			tflog.Debug(s.Ctx, fmt.Sprintf("Unable to update the member list of the DNS SMART: %s\n", smartName))
		}
	}

	return false
}

// Remove a DNS server from a SMART
// Return false in case of failure
func dnsdeletefromsmart(smartName string, serverName string, meta interface{}) bool {
	s := meta.(*SOLIDserver)

	parameters := url.Values{}
	parameters.Add("vdns_name", smartName)
	parameters.Add("dns_name", serverName)

	// Sending the read request
	resp, body, err := s.Request("delete", "rest/dns_smart_member_delete", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 || resp.StatusCode == 204 {
			return true
		}

		// Atomic SMART registration service unavailable attempting to use existing services
		if resp.StatusCode == 400 || resp.StatusCode == 404 {
			// Random Delay (in case of concurrent resources creation - until 8.0 and service dns_smart_member_add)
			//time.Sleep(time.Duration((rand.Intn(600) / 10) * time.Second))

			// Building parameters for retrieving SMART vdns_dns_group_role information
			parameters := url.Values{}

			parameters.Add("WHERE", "vdns_parent_name='"+smartName+"' AND dns_type!='vdns'")

			// Sending the read request
			resp, body, err := s.Request("get", "rest/dns_server_list", &parameters)

			if err == nil {
				var buf [](map[string]interface{})
				json.Unmarshal([]byte(body), &buf)

				// Checking the answer
				if resp.StatusCode == 200 || resp.StatusCode == 204 {

					// Building vdns_dns_group_role parameter from the SMART member list
					membersRole := ""

					if len(buf) > 0 {
						for _, smartMember := range buf {
							if smartMember["dns_name"].(string) != serverName {
								membersRole += smartMember["dns_name"].(string) + "&" + smartMember["dns_role"].(string) + ";"
							}
						}
					}

					if dnssmartmembersupdate(smartName, membersRole, meta) {
						return true
					}

					return false
				}

				// Log the error
				if len(buf) > 0 {
					if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
						tflog.Debug(s.Ctx, fmt.Sprintf("Unable to retrieve members list of the DNS SMART: %s (%s)\n", smartName, errMsg))
					}
				} else {
					tflog.Debug(s.Ctx, fmt.Sprintf("Unable to retrieve members list of the DNS SMART: %s\n", smartName))
				}
			}

			return false
		}

		// Log the error
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				tflog.Debug(s.Ctx, fmt.Sprintf("Unable to update the member list of the DNS SMART: %s (%s)\n", smartName, errMsg))
			}
		} else {
			tflog.Debug(s.Ctx, fmt.Sprintf("Unable to update the member list of the DNS SMART: %s\n", smartName))
		}
	}

	return false
}
