package krbd

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Options is per client instance and per mapping (block device) rbd device map options.
// krbd tag is the string of the option passed via sysfs.
//
//   - Reference: https://docs.ceph.com/docs/master/man/8/rbd/#kernel-rbd-krbd-options
type Options struct {
	// Client Options
	Fsid                     string `krbd:"fsid"`
	IP                       string `krbd:"ip"`
	Share                    bool   `krbd:"share"`
	NoShare                  bool   `krbd:"noshare"`
	CRC                      bool   `krbd:"crc"`
	NoCRC                    bool   `krbd:"nocr"`
	CephxRequireSignatures   bool   `krbd:"cephx_require_signatures"`
	NoCephxRequireSignatures bool   `krbd:"nocephx_require_signatures"`
	TCPNoDelay               bool   `krbd:"tcp_nodelay"`
	NoTCPNoDelay             bool   `krbd:"notcp_nodelay"`
	CephxSignMessages        bool   `krbd:"cephx_sign_messages"`
	NoCephxSignMessages      bool   `krbd:"nocephx_sign_messages"`
	MountTimeout             int    `krbd:"mount_timeout"`
	OSDKeepAlive             int    `krbd:"osd_keepalive"`
	OSDIdleTTL               int    `krbd:"osd_idlettl"`

	// RBD Block Options
	Force       bool   `krbd:"force"` // Force specifies that the unmap operation should be forced; this option is only recognized and used during Unmap operations.
	ReadWrite   bool   `krbd:"rw"`
	ReadOnly    bool   `krbd:"ro"`
	QueueDepth  int    `krbd:"queue_depth"`
	LockOnRead  bool   `krbd:"lock_on_read"`
	Exclusive   bool   `krbd:"exclusive"`
	LockTimeout uint64 `krbd:"lock_timeout"`
	NoTrim      bool   `krbd:"notrim"`
	AbortOnFull bool   `krbd:"abort_on_full"`
	AllocSize   int    `krbd:"alloc_size"`
	Name        string `krbd:"name"`
	Secret      string `krbd:"secret"`
	Namespace   string `krbd:"_pool_ns"`
}

func (o Options) MarshalText() (text []byte, err error) {
	output := []string{}
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)

	// Iterate over all available struct fields
	for i := 0; i < t.NumField(); i++ {
		// Skip values that are zero values of the struct. Otherwise Options would have
		// to track the upstream default values to always provide all options.
		if v.Field(i).Interface() == reflect.Zero(v.Field(i).Type()).Interface() {
			continue
		}
		tag := t.Field(i).Tag.Get("krbd")
		// Bool types don't include their value just the tag.
		if v.Field(i).Kind() == reflect.Bool {
			output = append(output, tag)
		} else {
			output = append(output, fmt.Sprintf("%s=%v", tag, v.Field(i)))
		}
	}
	text = []byte(strings.Join(output, ","))
	return
}

func (o Options) String() string {
	b, err := o.MarshalText()
	if err != nil {
		return ""
	}
	return string(b)
}

// Decode parses a comma separated option string (opt1,opt2=val,...) into Options.
// Each option key is resolved from the struct field `krbd` tag.
func (o *Options) UnmarshalText(text []byte) error {
	if o == nil {
		return fmt.Errorf("options is nil")
	}

	options := string(text)
	options = strings.TrimSpace(options)
	if options == "" {
		return nil
	}

	t := reflect.TypeOf(*o)
	v := reflect.ValueOf(o).Elem()

	fieldsByTag := map[string]int{}
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("krbd")
		if tag == "" {
			continue
		}
		fieldsByTag[tag] = i
	}

	for _, rawOpt := range strings.Split(options, ",") {
		opt := strings.TrimSpace(rawOpt)
		if opt == "" {
			continue
		}

		kv := strings.SplitN(opt, "=", 2)
		key := strings.TrimSpace(kv[0])
		val := ""
		if len(kv) == 2 {
			val = strings.TrimSpace(kv[1])
		}

		idx, ok := fieldsByTag[key]
		if !ok {
			return fmt.Errorf("unsupported option %q", key)
		}

		field := v.Field(idx)
		switch field.Kind() {
		case reflect.Bool:
			// krbd options are typically passed as key presence for booleans.
			// key=true/false is also accepted for convenience.
			if len(kv) == 1 || val == "" {
				field.SetBool(true)
				continue
			}
			b, err := strconv.ParseBool(val)
			if err != nil {
				return fmt.Errorf("invalid bool value for %q: %q", key, val)
			}
			field.SetBool(b)
		case reflect.String:
			if len(kv) != 2 {
				return fmt.Errorf("option %q requires a value", key)
			}
			field.SetString(val)
		case reflect.Int:
			if len(kv) != 2 {
				return fmt.Errorf("option %q requires a value", key)
			}
			iv, err := strconv.ParseInt(val, 10, 0)
			if err != nil {
				return fmt.Errorf("invalid int value for %q: %q", key, val)
			}
			field.SetInt(iv)
		case reflect.Uint64:
			if len(kv) != 2 {
				return fmt.Errorf("option %q requires a value", key)
			}
			uv, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid uint64 value for %q: %q", key, val)
			}
			field.SetUint(uv)
		default:
			return fmt.Errorf("unsupported field type for option %q", key)
		}
	}

	return nil
}
